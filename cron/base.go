package cron

import (
	"github.com/geerew/off-course/app"
	"github.com/geerew/off-course/dao"
	"github.com/robfig/cron/v3"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// StartCron initializes the cron jobs
func StartCron(app *app.App) {
	c := cron.New()

	// Course availability
	ca := &courseAvailability{
		db:        app.DbManager.DataDb,
		dao:       dao.New(app.DbManager.DataDb),
		appFs:     app.AppFs,
		logger:    app.Logger.WithCron(),
		batchSize: 200,
	}

	// When cron is started, run the course availability job immediately
	go func() { ca.run() }()

	c.AddFunc("@every 5m", func() { ca.run() })

	c.Start()
}
