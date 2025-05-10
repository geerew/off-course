package dao

import (
	"context"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WriteLog writes a new log
func (dao *DAO) WriteLog(ctx context.Context, log *models.Log) error {
	if log == nil {
		return utils.ErrNilPtr
	}

	return dao.Create(ctx, log)
}
