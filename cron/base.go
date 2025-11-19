package cron

import (
	"net/http"
	"time"

	"github.com/geerew/off-course/app"
	"github.com/geerew/off-course/dao"
	"github.com/robfig/cron/v3"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Global release checker instance (exported so API can access it)
var ReleaseChecker *releaseChecker

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

	// Release checker
	ReleaseChecker = &releaseChecker{
		logger:     app.Logger.WithCron(),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	// Run release check immediately on startup
	go func() { ReleaseChecker.run() }()

	// Check for new releases every 5 minutes
	c.AddFunc("@every 5m", func() { ReleaseChecker.run() })

	c.Start()
}
