package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/fatih/color"
	"github.com/geerew/off-course/api"
	"github.com/geerew/off-course/cron"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/media"
	"github.com/geerew/off-course/utils/media/hls"
	"github.com/geerew/off-course/utils/security"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the application",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		httpAddr := viper.GetString("http")
		isDev := viper.GetBool("dev")
		dataDir := viper.GetString("data-dir")
		enableSignup := viper.GetBool("enable-signup")

		appFs := appfs.New(afero.NewOsFs(), nil)

		dbManager, err := database.NewSQLiteManager(&database.DatabaseManagerConfig{
			DataDir: dataDir,
			AppFs:   appFs,
		})

		if err != nil {
			fmt.Printf("ERR - Failed to create database manager: %s", err)
			os.Exit(1)
		}

		logger, loggerDone, err := logger.InitLogger(&logger.BatchOptions{
			BatchSize:   200,
			BeforeAddFn: loggerBeforeAddFn(),
			WriteFn:     loggerWriteFn(dbManager.LogsDb),
		})

		if err != nil {
			utils.Errf("Failed to initialize logger: %s", err)
			os.Exit(1)
		}
		defer close(loggerDone)

		// Set DB loggers
		dbManager.DataDb.SetLogger(logger)
		appFs.SetLogger(logger)

		// Initialize FFmpeg
		ffmpeg, err := media.NewFFmpeg()
		if err != nil {
			utils.Errf("Failed to initialize FFmpeg: %s", err)
			os.Exit(1)
		}

		courseScan := coursescan.New(&coursescan.CourseScanConfig{
			Db:     dbManager.DataDb,
			AppFs:  appFs,
			Logger: logger,
			FFmpeg: ffmpeg,
		})

		// Start the course scan worker
		go courseScan.Worker(ctx, coursescan.Processor, nil)

		cron.InitCron(&cron.CronConfig{
			Db:     dbManager.DataDb,
			AppFs:  appFs,
			Logger: logger,
		})

		hls.InitSettings(dataDir)

		router := api.NewRouter(&api.RouterConfig{
			DbManager:     dbManager,
			Logger:        logger,
			AppFs:         appFs,
			CourseScan:    courseScan,
			FFmpeg:        ffmpeg,
			HttpAddr:      httpAddr,
			IsProduction:  !isDev,
			SignupEnabled: enableSignup,
			DataDir:       dataDir,
		})

		// Check bootstrap status and generate token if needed
		router.InitBootstrap()
		if !router.IsBootstrapped() {
			// Generate bootstrap token
			bootstrapToken, err := auth.GenerateBootstrapToken(dataDir, appFs.Fs)
			if err != nil {
				utils.Errf("Failed to generate bootstrap token: %s", err)
				os.Exit(1)
			}

			// Print bootstrap URL to console
			bootstrapURL := fmt.Sprintf("http://%s/auth/bootstrap/%s", httpAddr, bootstrapToken.Token)
			utils.Infof(
				"%s %s\n",
				"⚠️  Bootstrap required:",
				color.CyanString(bootstrapURL),
			)
			utils.Infof("Token expires in 5 minutes\n")
		} else {
			// Clean up any leftover bootstrap token files
			auth.DeleteBootstrapToken(dataDir, appFs.Fs)
			utils.Infof("Application bootstrapped\n")
		}

		var wg sync.WaitGroup
		wg.Add(1)

		// Listen for shutdown signals
		go func() {
			defer wg.Done()
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
			<-quit
		}()

		// Serve the UI/API
		go func() {
			defer wg.Done()
			if err := router.Serve(); err != nil {
				utils.Errf("Failed to start router:: %s", err)
				os.Exit(1)
			}
		}()

		wg.Wait()

		utils.Infof("Shutting down...")

		// Delete all scans
		_, err = dbManager.DataDb.Exec("DELETE FROM " + models.SCAN_TABLE)
		if err != nil {
			utils.Errf("Failed to delete scans: %s", err)
		}
	},
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().BoolP("dev", "d", false, "Run in development mode")
	serveCmd.Flags().String("http", "127.0.0.1:9081", "TCP address to listen for the HTTP server")
	serveCmd.Flags().String("data-dir", "./oc_data", "Directory to store data files")
	serveCmd.Flags().Bool("enable-signup", false, "Allow users to create new accounts")

	// Bind flags
	viper.SetEnvPrefix("OC")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Bind each flag
	_ = viper.BindPFlag("dev", serveCmd.Flags().Lookup("dev"))
	_ = viper.BindPFlag("http", serveCmd.Flags().Lookup("http"))
	_ = viper.BindPFlag("data-dir", serveCmd.Flags().Lookup("data-dir"))
	_ = viper.BindPFlag("enable-signup", serveCmd.Flags().Lookup("enable-signup"))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// loggerBeforeAddFunc is a logger.BeforeAddFn
func loggerBeforeAddFn() logger.BeforeAddFn {
	return func(ctx context.Context, log *logger.Log) bool {
		// Skip calls to the logs API
		if strings.HasPrefix(log.Message, "GET /api/logs") {
			return false
		}

		// This should never happen as the logsDb should be nil, but in the event it is not, skip
		// logging log writes as it will cause an infinite loop
		if strings.HasPrefix(log.Message, "INSERT INTO "+models.LOG_TABLE) ||
			strings.HasPrefix(log.Message, "SELECT "+models.LOG_TABLE) ||
			strings.HasPrefix(log.Message, "UPDATE "+models.LOG_TABLE) ||
			strings.HasPrefix(log.Message, "DELETE FROM "+models.LOG_TABLE) {
			return false
		}

		return true
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// loggerWriteFn returns a logger.WriteFn that writes logs to the database
func loggerWriteFn(db database.Database) logger.WriteFn {
	return func(ctx context.Context, logs []*logger.Log) error {
		logDao := dao.New(db)

		// Write accumulated logs
		db.RunInTransaction(ctx, func(txCtx context.Context) error {
			model := &models.Log{}

			for _, l := range logs {
				model.ID = security.PseudorandomString(10)
				model.Level = int(l.Level)
				model.Message = l.Message
				model.Data = l.Data
				model.CreatedAt = l.Time
				model.UpdatedAt = model.CreatedAt

				// Write the log
				err := logDao.CreateLog(txCtx, model)
				if err != nil {
					log.Println("Failed to write log", model, err)
				}
			}

			return nil
		})

		return nil
	}
}
