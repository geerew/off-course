package models

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type Scan struct {
	BaseModel
	CourseID string           `bun:",unique,notnull"`
	Status   types.ScanStatus `bun:",notnull"`

	// Belongs to
	Course *Course `bun:"rel:belongs-to,join:course_id=id"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountScans returns the number of scans
func CountScans(db database.Database, params *database.DatabaseParams, ctx context.Context) (int, error) {
	q := db.DB().NewSelect().Model((*Scan)(nil))

	if params != nil && params.Where != nil {
		q = selectWhere(q, params)
	}

	return q.Count(ctx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetScans returns a slice of scans
func GetScans(db database.Database, params *database.DatabaseParams, ctx context.Context) ([]*Scan, error) {
	var scans []*Scan

	q := db.DB().NewSelect().Model(&scans)

	if params != nil {
		// Pagination
		if params.Pagination != nil {
			if count, err := CountScans(db, params, ctx); err != nil {
				return nil, err
			} else {
				params.Pagination.SetCount(count)
			}

			q = q.Offset(params.Pagination.Offset()).Limit(params.Pagination.Limit())
		}

		if params.Relation != nil {
			q = selectRelation(q, params)
		}

		// Order by
		if len(params.OrderBy) > 0 {
			q = q.Order(params.OrderBy...)
		}

		// Where
		if params.Where != nil {
			if params.Where != nil {
				q = selectWhere(q, params)
			}
		}
	}

	err := q.Scan(ctx)

	return scans, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetScanById returns a scan with the given ID
func GetScanById(db database.Database, id string, params *database.DatabaseParams, ctx context.Context) (*Scan, error) {
	scan := &Scan{}
	scan.SetId(id)

	q := db.DB().NewSelect().Model(scan)

	if params != nil && params.Relation != nil {
		q = selectRelation(q, params)
	}

	if err := q.Where("scan.id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}

	return scan, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetScanByCourseId returns a scan with the given course ID
func GetScanByCourseId(db database.Database, id string, params *database.DatabaseParams, ctx context.Context) (*Scan, error) {
	scan := &Scan{}

	q := db.DB().NewSelect().Model(scan)

	if params != nil && params.Relation != nil {
		q = selectRelation(q, params)
	}

	if err := q.Where("course_id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}

	return scan, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateScan inserts a new scan with a status of waiting
func CreateScan(db database.Database, scan *Scan, ctx context.Context) error {
	scan.RefreshId()
	scan.RefreshCreatedAt()
	scan.RefreshUpdatedAt()
	scan.Status = types.NewScanStatus(types.ScanStatusWaiting)

	_, err := db.DB().NewInsert().Model(scan).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateScanStatus updates the scan status
func UpdateScanStatus(db database.Database, scan *Scan, newStatus types.ScanStatusType, ctx context.Context) error {
	// Do nothing when the status is the same
	ss := types.NewScanStatus(newStatus)
	if scan.Status == ss {
		return nil
	}

	// Require an ID
	if scan.ID == "" {
		return errors.New("scan ID cannot be empty")
	}

	// Set a new timestamp
	ts := types.NowDateTime()

	// Update the status
	if res, err := db.DB().NewUpdate().Model(scan).
		Set("status = ?", ss).
		Set("updated_at = ?", ts).
		WherePK().Exec(ctx); err != nil {
		return err
	} else {
		count, _ := res.RowsAffected()
		if count == 0 {
			return nil
		}
	}

	// Update the original scan struct
	scan.Status = ss
	scan.UpdatedAt = ts

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteScan deletes a scan with the given ID
func DeleteScan(db database.Database, id string, ctx context.Context) (int, error) {
	scan := &Scan{}
	scan.SetId(id)

	if res, err := db.DB().NewDelete().Model(scan).WherePK().Exec(ctx); err != nil {
		return 0, err
	} else {
		count, _ := res.RowsAffected()
		return int(count), err
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NextScan returns the next scan to be processed whose status is `waiting“
func NextScan(db database.Database, ctx context.Context) (*Scan, error) {
	var scan Scan

	err := db.DB().NewSelect().
		Model(&scan).
		Relation("Course").
		Where("scan.status = ?", types.ScanStatusWaiting).
		Order("scan.created_at ASC").
		Limit(1).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &scan, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTestScans creates a scan for each course in the slice. If a db is provided, the scans will
// be inserted into the db
//
// THIS IS FOR TESTING PURPOSES
func NewTestScans(t *testing.T, db database.Database, courses []*Course) []*Scan {
	scans := []*Scan{}
	for i := 0; i < len(courses); i++ {
		s := &Scan{
			CourseID: courses[i].ID,
			Status:   types.NewScanStatus(types.ScanStatusWaiting),
		}

		s.RefreshId()
		s.RefreshCreatedAt()
		s.RefreshUpdatedAt()

		if db != nil {
			_, err := db.DB().NewInsert().Model(s).Exec(context.Background())
			require.Nil(t, err)
		}

		scans = append(scans, s)
		time.Sleep(1 * time.Millisecond)
	}

	return scans
}
