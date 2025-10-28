package cron

import (
	"log/slog"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/types"
	"github.com/robfig/cron/v3"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// loggerType is the type of logger
var loggerType = slog.Any("type", types.LogTypeCron)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type CronConfig struct {
	Db     database.Database
	AppFs  *appfs.AppFs
	Logger *logger.Logger
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// StartCron initializes the cron jobs
func StartCron(config *CronConfig) {
	c := cron.New()

	// Course availability
	ca := &courseAvailability{
		db:        config.Db,
		dao:       dao.New(config.Db),
		appFs:     config.AppFs,
		logger:    config.Logger,
		batchSize: 200,
	}

	// When cron is started, run the course availability job immediately
	go func() { ca.run() }()

	c.AddFunc("@every 5m", func() { ca.run() })

	c.Start()
}
