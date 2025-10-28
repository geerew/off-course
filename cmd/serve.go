package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/geerew/off-course/api"
	"github.com/geerew/off-course/cron"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/media"
	"github.com/geerew/off-course/utils/media/hls"
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

		// Logger
		appLogger := logger.New(&logger.Config{
			Level:         logger.LevelInfo,
			ConsoleOutput: true,
		})

		if appLogger == nil {
			panic("Failed to initialize logger")
		}

		mainLogger := appLogger.WithMain()

		// AppFS (filesystem)
		appFs := appfs.New(afero.NewOsFs())

		// FFmpeg
		ffmpeg, err := media.NewFFmpeg()
		if err != nil {
			mainLogger.Error().Err(err).Msg("Failed to initialize FFmpeg")
			os.Exit(1)
		}

		// Database manager
		dbManager, err := database.NewSQLiteManager(&database.DatabaseManagerConfig{
			DataDir: dataDir,
			AppFs:   appFs,
		})

		if err != nil {
			mainLogger.Error().Err(err).Msg("Failed to create database manager")
			os.Exit(1)
		}

		// Course scanner
		courseScan := coursescan.New(&coursescan.CourseScanConfig{
			Db:     dbManager.DataDb,
			AppFs:  appFs,
			Logger: appLogger.WithCourseScan(),
			FFmpeg: ffmpeg,
		})

		// Start the course scan worker
		go courseScan.Worker(ctx, coursescan.Processor, nil)

		// Start cron
		cron.StartCron(&cron.CronConfig{
			Db:     dbManager.DataDb,
			AppFs:  appFs,
			Logger: appLogger.WithCron(),
		})

		// HLS
		transcoder, err := hls.NewTranscoder(&hls.TranscoderConfig{
			CachePath: dataDir,
			AppFs:     appFs,
			Logger:    appLogger.WithHLS(),
			Dao:       dao.New(dbManager.DataDb),
		})

		if err != nil {
			mainLogger.Error().Err(err).Msg("Failed to create HLS transcoder")
			os.Exit(1)
		}

		// Router
		router := api.NewRouter(&api.RouterConfig{
			DbManager:     dbManager,
			Logger:        appLogger.WithAPI(),
			AppFs:         appFs,
			CourseScan:    courseScan,
			FFmpeg:        ffmpeg,
			HttpAddr:      httpAddr,
			IsProduction:  !isDev,
			SignupEnabled: enableSignup,
			DataDir:       dataDir,
			Transcoder:    transcoder,
		})

		// Check bootstrap status and generate token if needed
		router.InitBootstrap()
		if !router.IsBootstrapped() {
			bootstrapToken, err := auth.GenerateBootstrapToken(dataDir, appFs.Fs)
			if err != nil {
				mainLogger.Error().Err(err).Msg("Failed to generate bootstrap token")
				os.Exit(1)
			}

			bootstrapURL := fmt.Sprintf("http://%s/auth/bootstrap/%s", httpAddr, bootstrapToken.Token)
			mainLogger.Info().
				Str("bootstrap_url", bootstrapURL).
				Str("expires_in", "5 minutes").
				Msg("Bootstrap required")
		} else {
			auth.DeleteBootstrapToken(dataDir, appFs.Fs)
			mainLogger.Info().Msg("Application bootstrapped")
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
				mainLogger.Error().Err(err).Msg("Failed to start router")
				os.Exit(1)
			}
		}()

		wg.Wait()

		mainLogger.Info().Msg("Shutting down...")

		// Delete all scans
		_, err = dbManager.DataDb.Exec("DELETE FROM " + models.SCAN_TABLE)
		if err != nil {
			mainLogger.Error().Err(err).Msg("Failed to delete scans")
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
