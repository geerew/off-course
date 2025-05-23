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

	return Create(ctx, dao, scan)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetScan retrieves a scan
//
// When options is nil or options.Where is nil, the models ID will be used
func (dao *DAO) GetScan(ctx context.Context, scan *models.Scan, options *database.Options) error {
	if scan == nil {
		return utils.ErrNilPtr
	}

	if options == nil {
		options = &database.Options{}
	}

	if options.Where == nil {
		if scan.Id() == "" {
			return utils.ErrInvalidId
		}

		options.Where = squirrel.Eq{models.SCAN_TABLE_ID: scan.Id()}
	}

	return Get(ctx, dao, scan, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListScans retrieves a list of scans
func (dao *DAO) ListScans(ctx context.Context, scans *[]*models.Scan, options *database.Options) error {
	if scans == nil {
		return utils.ErrNilPtr
	}

	return List(ctx, dao, scans, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateScan updates a scan
func (dao *DAO) UpdateScan(ctx context.Context, scan *models.Scan) error {
	if scan == nil {
		return utils.ErrNilPtr
	}

	_, err := Update(ctx, dao, scan)
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
