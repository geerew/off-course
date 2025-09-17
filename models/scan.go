package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"fmt"

	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	SCAN_TABLE = "scans"

	SCAN_COURSE_ID   = "course_id"
	SCAN_STATUS      = "status"
	SCAN_COURSE_PATH = "path"

	SCAN_TABLE_ID         = SCAN_TABLE + "." + BASE_ID
	SCAN_TABLE_CREATED_AT = SCAN_TABLE + "." + BASE_CREATED_AT
	SCAN_TABLE_UPDATED_AT = SCAN_TABLE + "." + BASE_UPDATED_AT
	SCAN_TABLE_COURSE_ID  = SCAN_TABLE + "." + SCAN_COURSE_ID
	SCAN_TABLE_STATUS     = SCAN_TABLE + "." + SCAN_STATUS
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan defines the model for a scan
type Scan struct {
	Base
	CourseID string           `db:"course_id"` // Immutable
	Status   types.ScanStatus `db:"status"`    // Mutable, defaults to "waiting" if not set

	// Joins
	CoursePath string `db:"course_path"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ScanColumns returns the list of columns to use when populating `Scan`
func ScanColumns() []string {
	return []string{
		fmt.Sprintf("%s AS id", SCAN_TABLE_ID),
		fmt.Sprintf("%s AS created_at", SCAN_TABLE_CREATED_AT),
		fmt.Sprintf("%s AS updated_at", SCAN_TABLE_UPDATED_AT),
		fmt.Sprintf("%s AS course_id", SCAN_TABLE_COURSE_ID),
		fmt.Sprintf("%s AS status", SCAN_TABLE_STATUS),
		// Join columns
		fmt.Sprintf("%s AS course_path", SCAN_COURSE_PATH),
	}
}
