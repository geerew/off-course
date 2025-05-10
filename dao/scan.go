package dao

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateScan creates a scan
func (dao *DAO) CreateScan(ctx context.Context, scan *models.Scan) error {
	if scan == nil {
		return utils.ErrNilPtr
	}

	// A scan should always be in the waiting state when created
	if !scan.Status.IsWaiting() {
		scan.Status.SetWaiting()
	}

	return dao.Create(ctx, scan)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetScan retrieves a scan
//
// When options is nil or options.Where is nil, the function will use the ID to filter scans
func (dao *DAO) GetScan(ctx context.Context, scan *models.Scan, options *database.Options) error {
	if scan == nil {
		return utils.ErrNilPtr
	}

	if options == nil || options.Where == nil {
		if scan.Id() == "" {
			return utils.ErrInvalidId
		}

		options = &database.Options{
			Where: squirrel.Eq{scan.Table() + "." + models.BASE_ID: scan.Id()},
		}
	}

	if options.Where == nil {
	}

	return dao.Get(ctx, scan, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateScan updates a scan
func (dao *DAO) UpdateScan(ctx context.Context, scan *models.Scan) error {
	if scan == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, scan)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NextWaitingScan gets the next scan whose status is `waiting“ based upon the created_at column
func (dao *DAO) NextWaitingScan(ctx context.Context, scan *models.Scan) error {
	if scan == nil {
		return utils.ErrNilPtr
	}

	options := &database.Options{
		Where:   squirrel.Eq{models.SCAN_TABLE_STATUS: types.ScanStatusWaiting},
		OrderBy: []string{models.SCAN_TABLE_CREATED_AT + " ASC"},
	}

	return dao.GetScan(ctx, scan, options)
}
