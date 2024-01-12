package models

import (
	"database/sql"
	"fmt"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CountAssets(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		count, err := CountAssets(db, nil)
		require.Nil(t, err)
		assert.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		NewTestData(t, db, 5, false, 1, 0)

		count, err := CountAssets(db, nil)
		require.Nil(t, err)
		assert.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 3, false, 2, 0)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := CountAssets(db, &database.DatabaseParams{Where: sq.Eq{TableAssets() + ".id": workingData[0].Assets[1].ID}})
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = CountAssets(db, &database.DatabaseParams{Where: sq.NotEq{TableAssets() + ".id": workingData[0].Assets[1].ID}})
		require.Nil(t, err)
		assert.Equal(t, 5, count)

		// ----------------------------
		// EQUALS COURSE_ID
		// ----------------------------
		count, err = CountAssets(db, &database.DatabaseParams{Where: sq.Eq{TableAssets() + ".course_id": workingData[1].ID}})
		require.Nil(t, err)
		assert.Equal(t, 2, count)

		// ----------------------------
		// ERROR
		// ----------------------------
		count, err = CountAssets(db, &database.DatabaseParams{Where: sq.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, -1, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableAssets())
		require.Nil(t, err)

		_, err = CountAssets(db, nil)
		require.ErrorContains(t, err, "no such table: "+TableAssets())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAssets(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		assets, err := GetAssets(db, nil)
		require.Nil(t, err)
		require.Zero(t, assets)
	})

	t.Run("found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 5, false, 1, 3)

		result, err := GetAssets(db, nil)
		require.Nil(t, err)
		require.Len(t, result, 5)

		// ----------------------------
		// Progress
		// ----------------------------
		for _, a := range result {
			require.Zero(t, a.VideoPos)
			require.False(t, a.Completed)
			require.True(t, a.CompletedAt.IsZero())
		}

		// Update 1 video position and 1 completed
		_, err = UpdateAssetProgressVideoPos(db, workingData[0].Assets[0].ID, 50)
		require.Nil(t, err)
		_, err = UpdateAssetProgressCompleted(db, workingData[1].Assets[0].ID, true)
		require.Nil(t, err)

		// Find all started videos
		dbParams := &database.DatabaseParams{
			Where: sq.And{sq.Eq{TableAssets() + ".type": string(types.AssetVideo)}, sq.Gt{TableAssetsProgress() + ".video_pos": 0}},
		}
		result, err = GetAssets(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, workingData[0].Assets[0].ID, result[0].ID)

		// Find completed assets
		result, err = GetAssets(db, &database.DatabaseParams{Where: sq.Eq{TableAssetsProgress() + ".completed": true}})
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, workingData[1].Assets[0].ID, result[0].ID)

		// ----------------------------
		// Attachments
		// ----------------------------
		for _, a := range result {
			require.Len(t, a.Attachments, 3)
		}
	})

	t.Run("orderby", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 3, false, 1, 0)

		// ----------------------------
		// Descending
		// ----------------------------
		dbParams := &database.DatabaseParams{OrderBy: []string{"created_at desc"}}
		result, err := GetAssets(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, workingData[2].Assets[0].ID, result[0].ID)

		// ----------------------------
		// Ascending
		// ----------------------------
		result, err = GetAssets(db, &database.DatabaseParams{OrderBy: []string{"created_at asc"}})
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, workingData[0].Assets[0].ID, result[0].ID)

		// ----------------------------
		// Error
		// ----------------------------
		dbParams = &database.DatabaseParams{OrderBy: []string{"unit_test asc"}}
		result, err = GetAssets(db, dbParams)
		require.ErrorContains(t, err, "no such column")
		assert.Nil(t, result)
	})

	t.Run("where", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 3, false, 2, 0)

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		result, err := GetAssets(db, &database.DatabaseParams{Where: sq.Eq{TableAssets() + ".id": workingData[0].Assets[1].ID}})
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, workingData[0].Assets[1].ID, result[0].ID)

		// ----------------------------
		// EQUALS ID OR ID
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where:   sq.Or{sq.Eq{TableAssets() + ".id": workingData[0].Assets[1].ID}, sq.Eq{TableAssets() + ".id": workingData[1].Assets[1].ID}},
			OrderBy: []string{"created_at asc"},
		}
		result, err = GetAssets(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, workingData[0].Assets[1].ID, result[0].ID)
		assert.Equal(t, workingData[1].Assets[1].ID, result[1].ID)

		// ----------------------------
		// EQUALS COURSE_ID
		// ----------------------------
		dbParams = &database.DatabaseParams{
			Where:   sq.Eq{TableAssets() + ".course_id": workingData[1].ID},
			OrderBy: []string{"created_at asc"},
		}
		result, err = GetAssets(db, dbParams)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, workingData[1].Assets[0].ID, result[0].ID)
		assert.Equal(t, workingData[1].Assets[1].ID, result[1].ID)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = GetAssets(db, &database.DatabaseParams{Where: sq.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 17, 0)

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := GetAssets(db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, workingData[0].Assets[0].ID, result[0].ID)
		assert.Equal(t, workingData[0].Assets[9].ID, result[9].ID)

		// ----------------------------
		// Page 2 with 7 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = GetAssets(db, &database.DatabaseParams{Pagination: p})
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, workingData[0].Assets[10].ID, result[0].ID)
		assert.Equal(t, workingData[0].Assets[16].ID, result[6].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableAssets())
		require.Nil(t, err)

		_, err = GetAssets(db, nil)
		require.ErrorContains(t, err, "no such table: "+TableAssets())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAsset(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		a, err := GetAsset(db, "1234")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, a)
	})

	t.Run("found", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 2, false, 1, 2)

		a, err := GetAsset(db, workingData[0].Assets[0].ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].Assets[0].ID, a.ID)

		// ----------------------------
		// Progress
		// ----------------------------
		require.Zero(t, a.VideoPos)
		require.False(t, a.Completed)
		require.True(t, a.CompletedAt.IsZero())

		// Set video pos
		_, err = UpdateAssetProgressVideoPos(db, a.ID, 50)
		require.Nil(t, err)

		a, err = GetAsset(db, a.ID)
		require.Nil(t, err)
		assert.Equal(t, 50, a.VideoPos)
		assert.False(t, a.Completed)
		assert.True(t, a.CompletedAt.IsZero())

		// Set completed
		_, err = UpdateAssetProgressCompleted(db, a.ID, true)
		require.Nil(t, err)

		a, err = GetAsset(db, a.ID)
		require.Nil(t, err)
		assert.Equal(t, 50, a.VideoPos)
		assert.True(t, a.Completed)
		assert.False(t, a.CompletedAt.IsZero())

		// ----------------------------
		// Attachments
		// ----------------------------
		require.Len(t, a.Attachments, 2)
		assert.Equal(t, workingData[0].Assets[0].Attachments[0].ID, a.Attachments[0].ID)
		assert.Equal(t, workingData[0].Assets[0].Attachments[1].ID, a.Attachments[1].ID)
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		c, err := GetAsset(db, "")
		require.EqualError(t, err, "id cannot be empty")
		assert.Nil(t, c)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableAssets())
		require.Nil(t, err)

		_, err = GetAsset(db, "1234")
		require.ErrorContains(t, err, "no such table: "+TableAssets())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateAsset(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)
		a := newTestAssets(t, nil, workingData[0].Course, 1, 0)[0]

		err := CreateAsset(db, a)
		require.Nil(t, err)

		newA, err := GetAsset(db, a.ID)
		require.Nil(t, err)
		assert.Equal(t, a.ID, newA.ID)
		assert.Equal(t, a.CourseID, newA.CourseID)
		assert.Equal(t, a.Title, newA.Title)
		assert.Equal(t, a.Prefix, newA.Prefix)
		assert.Equal(t, a.Chapter, newA.Chapter)
		assert.Equal(t, a.Type, newA.Type)
		assert.Equal(t, a.Path, newA.Path)
		assert.False(t, a.CreatedAt.IsZero())
		assert.False(t, a.UpdatedAt.IsZero())
	})

	t.Run("duplicate paths", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)
		a := newTestAssets(t, nil, workingData[0].Course, 1, 0)[0]

		err := CreateAsset(db, a)
		require.Nil(t, err)

		err = CreateAsset(db, a)
		assert.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.path", TableAssets()))
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)

		// No course ID
		asset := &Asset{}
		assert.ErrorContains(t, CreateAsset(db, asset), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableAssets()))
		asset.CourseID = ""
		assert.ErrorContains(t, CreateAsset(db, asset), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableAssets()))
		asset.CourseID = "1234"

		// No title
		assert.ErrorContains(t, CreateAsset(db, asset), fmt.Sprintf("NOT NULL constraint failed: %s.title", TableAssets()))
		asset.Title = ""
		assert.ErrorContains(t, CreateAsset(db, asset), fmt.Sprintf("NOT NULL constraint failed: %s.title", TableAssets()))
		asset.Title = "Course 1"

		// No/invalid prefix
		assert.ErrorContains(t, CreateAsset(db, asset), "NOT NULL constraint failed: assets.prefix")
		asset.Prefix = sql.NullInt16{Int16: -1, Valid: true}
		assert.ErrorContains(t, CreateAsset(db, asset), "prefix must be greater than 0")
		asset.Prefix = sql.NullInt16{Int16: 1, Valid: true}

		// No type
		assert.ErrorContains(t, CreateAsset(db, asset), "NOT NULL constraint failed: assets.type")
		asset.Type = types.Asset{}
		assert.ErrorContains(t, CreateAsset(db, asset), "NOT NULL constraint failed: assets.type")
		asset.Type = *types.NewAsset("mp4")

		// No path
		assert.ErrorContains(t, CreateAsset(db, asset), "NOT NULL constraint failed: assets.path")
		asset.Path = ""
		assert.ErrorContains(t, CreateAsset(db, asset), "NOT NULL constraint failed: assets.path")
		asset.Path = "/course 1/01 asset"

		// Invalid Course ID
		assert.ErrorContains(t, CreateAsset(db, asset), "FOREIGN KEY constraint failed")

		// Success
		asset.CourseID = workingData[0].ID
		assert.Nil(t, CreateAsset(db, asset))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteAsset(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		workingData := NewTestData(t, db, 1, false, 2, 0)

		err := DeleteAsset(db, workingData[0].Assets[1].ID)
		require.Nil(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		err := DeleteAsset(db, "")
		assert.ErrorContains(t, err, "id cannot be empty")
	})

	t.Run("invalid id", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		err := DeleteAsset(db, "1234")
		assert.Nil(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		_, db, teardown := setup(t)
		defer teardown(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + TableAssets())
		require.Nil(t, err)

		err = DeleteAsset(db, "1234")
		require.ErrorContains(t, err, "no such table: "+TableAssets())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteAssetCascade(t *testing.T) {
	_, db, teardown := setup(t)
	defer teardown(t)

	workingData := NewTestData(t, db, 1, false, 1, 0)

	err := DeleteCourse(db, workingData[0].ID)
	require.Nil(t, err)

	s, err := GetAsset(db, workingData[0].Assets[0].ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, s)
}
