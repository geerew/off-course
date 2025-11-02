package cron

import (
	"github.com/geerew/off-course/app"
	"github.com/geerew/off-course/dao"
	"github.com/robfig/cron/v3"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// StartCron initializes the cron jobs
func StartCron(application *app.App) {
	c := cron.New()

	// Course availability
	ca := &courseAvailability{
		db:        application.DbManager.DataDb,
		dao:       dao.New(application.DbManager.DataDb),
		appFs:     application.AppFs,
		logger:    application.Logger.WithCron(),
		batchSize: 200,
	}

	// When cron is started, run the course availability job immediately
	go func() { ca.run() }()

	c.AddFunc("@every 5m", func() { ca.run() })

	c.Start()
}
