package dao

import (
	"context"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WriteLog writes a new log
func (dao *DAO) WriteLog(ctx context.Context, log *models.Log) error {
	if log == nil {
		return utils.ErrNilPtr
	}

	return Create(ctx, dao, log)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListLogs retrieves a list of logs
func (dao *DAO) ListLogs(ctx context.Context, logs *[]*models.Log, options *database.Options) error {
	if logs == nil {
		return utils.ErrNilPtr
	}

	return List(ctx, dao, logs, options)
}
