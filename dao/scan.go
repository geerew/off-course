package dao

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateScan inserts a new scan record
func (dao *DAO) CreateScan(ctx context.Context, scan *models.Scan) error {
	if scan == nil {
		return utils.ErrNilPtr
	}

	if scan.ID == "" {
		scan.RefreshId()
	}

	scan.RefreshCreatedAt()
	scan.RefreshUpdatedAt()

	if !scan.Status.IsWaiting() {
		scan.Status.SetWaiting()
	}

	builderOptions := newBuilderOptions(models.SCAN_TABLE).
		WithData(
			map[string]interface{}{
				models.BASE_ID:         scan.ID,
				models.SCAN_COURSE_ID:  scan.CourseID,
				models.SCAN_STATUS:     scan.Status,
				models.BASE_CREATED_AT: scan.CreatedAt,
				models.BASE_UPDATED_AT: scan.UpdatedAt,
			},
		)

	return createGeneric(ctx, dao, *builderOptions)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountScans counts the number of scan records
func (dao *DAO) CountScans(ctx context.Context, dbOpts *database.Options) (int, error) {
	builderOpts := newBuilderOptions(models.SCAN_TABLE).SetDbOpts(dbOpts)
	return countGeneric(ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetScan gets a record from the scans table based upon the where clause in the options. If
// there is no where clause, it will return the first record in the table
func (dao *DAO) GetScan(ctx context.Context, dbOpts *database.Options) (*models.Scan, error) {
	builderOpts := newBuilderOptions(models.SCAN_TABLE).
		WithColumns(
			models.SCAN_TABLE+".*",
			models.COURSE_TABLE_PATH+" AS course_path",
		).
		WithJoin(models.COURSE_TABLE, fmt.Sprintf("%s = %s", models.COURSE_TABLE_ID, models.SCAN_TABLE_COURSE_ID)).
		SetDbOpts(dbOpts).
		WithLimit(1)

	return getGeneric[models.Scan](ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListScans gets all records from the scans table based upon the where clause and pagination
// in the options
func (dao *DAO) ListScans(ctx context.Context, dbOpts *database.Options) ([]*models.Scan, error) {
	builderOpts := newBuilderOptions(models.SCAN_TABLE).
		WithColumns(
			models.SCAN_TABLE+".*",
			models.COURSE_TABLE_PATH+" AS course_path",
		).
		WithJoin(models.COURSE_TABLE, fmt.Sprintf("%s = %s", models.COURSE_TABLE_ID, models.SCAN_TABLE_COURSE_ID)).
		SetDbOpts(dbOpts)

	return listGeneric[models.Scan](ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateScan updates a scan record
func (dao *DAO) UpdateScan(ctx context.Context, scan *models.Scan) error {
	if scan == nil {
		return utils.ErrNilPtr
	}

	if scan.ID == "" {
		return utils.ErrId
	}

	scan.RefreshUpdatedAt()

	dbOpts := &database.Options{
		Where: squirrel.Eq{models.BASE_ID: scan.ID},
	}

	builderOptions := newBuilderOptions(models.SCAN_TABLE).
		WithData(
			map[string]interface{}{
				models.SCAN_STATUS:     scan.Status,
				models.BASE_UPDATED_AT: scan.UpdatedAt,
			},
		).
		SetDbOpts(dbOpts)

	_, err := updateGeneric(ctx, dao, *builderOptions)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteScans deletes records from the scans table
//
// Errors when a where clause is not provided
func (dao *DAO) DeleteScans(ctx context.Context, dbOpts *database.Options) error {
	if dbOpts == nil || dbOpts.Where == nil {
		return utils.ErrWhere
	}

	builderOpts := newBuilderOptions(models.SCAN_TABLE).SetDbOpts(dbOpts)
	sqlStr, args, _ := deleteBuilder(*builderOpts)

	q := database.QuerierFromContext(ctx, dao.db)
	_, err := q.ExecContext(ctx, sqlStr, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NextWaitingScan gets the next scan whose status is `waiting“ based upon the created_at column
func (dao *DAO) NextWaitingScan(ctx context.Context) (*models.Scan, error) {
	dbOpts := database.NewOptions().
		WithWhere(squirrel.Eq{models.SCAN_TABLE_STATUS: types.ScanStatusWaiting}).
		WithOrderBy(models.SCAN_TABLE_CREATED_AT + " ASC")

	return dao.GetScan(ctx, dbOpts)
}
