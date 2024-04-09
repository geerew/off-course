package daos

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func attachmentSetup(t *testing.T) (*appFs.AppFs, *AttachmentDao, database.Database) {
	appFs, db := setup(t)
	attachmentDao := NewAttachmentDao(db)
	return appFs, attachmentDao, db
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachment_Count(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, dao, _ := attachmentSetup(t)

		count, err := dao.Count(nil)
		require.Nil(t, err)
		assert.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		NewTestBuilder(t).Db(db).Courses(5).Assets(1).Attachments(1).Build()

		count, err := dao.Count(nil)
		require.Nil(t, err)
		assert.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(3).Assets(1).Attachments(2).Build()

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := dao.Count(&database.DatabaseParams{Where: squirrel.Eq{TableAttachments() + ".id": testData[1].Assets[0].Attachments[1].ID}})
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.NotEq{TableAttachments() + ".id": testData[1].Assets[0].Attachments[1].ID}})
		require.Nil(t, err)
		assert.Equal(t, 5, count)

		// ----------------------------
		// EQUALS ASSET_ID
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.Eq{TableAttachments() + ".asset_id": testData[1].Assets[0].ID}})
		require.Nil(t, err)
		assert.Equal(t, 2, count)

		// ----------------------------
		// ERROR
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		assert.Equal(t, 0, count)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.Count(nil)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachment_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		testData := NewTestBuilder(t).Courses(1).Assets(1).Attachments(1).Build()

		// Create the course
		courseDao := NewCourseDao(db)
		require.Nil(t, courseDao.Create(testData[0].Course))

		// Create the asset
		assetDao := NewAssetDao(db)
		err := assetDao.Create(testData[0].Assets[0])
		require.Nil(t, err)

		// Create the attachment
		err = dao.Create(testData[0].Assets[0].Attachments[0])
		require.Nil(t, err)

		newA, err := dao.Get(testData[0].Assets[0].Attachments[0].ID, nil)
		require.Nil(t, err)
		assert.Equal(t, testData[0].Assets[0].Attachments[0].ID, newA.ID)
		assert.Equal(t, testData[0].Assets[0].Attachments[0].CourseID, newA.CourseID)
		assert.Equal(t, testData[0].Assets[0].Attachments[0].AssetID, newA.AssetID)
		assert.Equal(t, testData[0].Assets[0].Attachments[0].Title, newA.Title)
		assert.Equal(t, testData[0].Assets[0].Attachments[0].Path, newA.Path)
		assert.False(t, newA.CreatedAt.IsZero())
		assert.False(t, newA.UpdatedAt.IsZero())
	})

	t.Run("duplicate paths", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Assets(1).Attachments(1).Build()

		// Create the attachment (again)
		err := dao.Create(testData[0].Assets[0].Attachments[0])
		require.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.path", dao.table))
	})

	t.Run("constraints", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Assets(1).Build()

		// No course ID
		attachment := &models.Attachment{}
		assert.ErrorContains(t, dao.Create(attachment), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableAttachments()))
		attachment.CourseID = ""
		assert.ErrorContains(t, dao.Create(attachment), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", TableAttachments()))
		attachment.CourseID = "1234"

		// No asset ID
		assert.ErrorContains(t, dao.Create(attachment), fmt.Sprintf("NOT NULL constraint failed: %s.asset_id", TableAttachments()))
		attachment.AssetID = ""
		assert.ErrorContains(t, dao.Create(attachment), fmt.Sprintf("NOT NULL constraint failed: %s.asset_id", TableAttachments()))
		attachment.AssetID = "1234"

		// No title
		assert.ErrorContains(t, dao.Create(attachment), fmt.Sprintf("NOT NULL constraint failed: %s.title", TableAttachments()))
		attachment.Title = ""
		assert.ErrorContains(t, dao.Create(attachment), fmt.Sprintf("NOT NULL constraint failed: %s.title", TableAttachments()))
		attachment.Title = "Course 1"

		// No path
		assert.ErrorContains(t, dao.Create(attachment), "NOT NULL constraint failed: attachments.path")
		attachment.Path = ""
		assert.ErrorContains(t, dao.Create(attachment), "NOT NULL constraint failed: attachments.path")
		attachment.Path = "/course 1/01 attachment"

		// Invalid course ID
		assert.ErrorContains(t, dao.Create(attachment), "FOREIGN KEY constraint failed")
		attachment.CourseID = testData[0].ID

		// Invalid asset ID
		assert.ErrorContains(t, dao.Create(attachment), "FOREIGN KEY constraint failed")
		attachment.AssetID = testData[0].Assets[0].ID

		// Success
		assert.Nil(t, dao.Create(attachment))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachment_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(2).Assets(1).Attachments(1).Build()

		a, err := dao.Get(testData[1].Assets[0].Attachments[0].ID, nil)
		require.Nil(t, err)
		assert.Equal(t, testData[1].Assets[0].Attachments[0].ID, a.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, dao, _ := attachmentSetup(t)

		c, err := dao.Get("1234", nil)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, c)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := attachmentSetup(t)

		c, err := dao.Get("", nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, c)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.Get("1234", nil)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachment_List(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, dao, _ := attachmentSetup(t)

		assets, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Zero(t, assets)
	})

	t.Run("found", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		NewTestBuilder(t).Db(db).Courses(5).Assets(1).Attachments(1).Build()

		result, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Len(t, result, 5)

	})

	t.Run("orderby", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(3).Assets(1).Attachments(1).Build()

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		result, err := dao.List(&database.DatabaseParams{OrderBy: []string{"created_at desc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, testData[2].Assets[0].Attachments[0].ID, result[0].ID)
		assert.Equal(t, testData[1].Assets[0].Attachments[0].ID, result[1].ID)
		assert.Equal(t, testData[0].Assets[0].Attachments[0].ID, result[2].ID)

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"created_at asc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, testData[0].Assets[0].Attachments[0].ID, result[0].ID)
		assert.Equal(t, testData[1].Assets[0].Attachments[0].ID, result[1].ID)
		assert.Equal(t, testData[2].Assets[0].Attachments[0].ID, result[2].ID)

		// ----------------------------
		// Error
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"unit_test asc"}}, nil)
		require.ErrorContains(t, err, "no such column")
		assert.Nil(t, result)
	})

	t.Run("where", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(3).Assets(2).Attachments(2).Build()

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		result, err := dao.List(&database.DatabaseParams{Where: squirrel.Eq{TableAttachments() + ".id": testData[1].Assets[1].Attachments[0].ID}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, testData[1].Assets[1].Attachments[0].ID, result[0].ID)

		// ----------------------------
		// EQUALS ID OR ID
		// ----------------------------
		dbParams := &database.DatabaseParams{
			Where: squirrel.Or{
				squirrel.Eq{TableAttachments() + ".id": testData[1].Assets[1].Attachments[0].ID},
				squirrel.Eq{TableAttachments() + ".id": testData[2].Assets[0].Attachments[1].ID},
			},
			OrderBy: []string{"created_at asc"},
		}

		result, err = dao.List(dbParams, nil)
		require.Nil(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, testData[1].Assets[1].Attachments[0].ID, result[0].ID)
		assert.Equal(t, testData[2].Assets[0].Attachments[1].ID, result[1].ID)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{"": ""}}, nil)
		require.ErrorContains(t, err, "syntax error")
		assert.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Assets(1).Attachments(17).Build()

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := dao.List(&database.DatabaseParams{Pagination: p}, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, testData[0].Assets[0].Attachments[0].ID, result[0].ID)
		assert.Equal(t, testData[0].Assets[0].Attachments[9].ID, result[9].ID)

		// ----------------------------
		// Page 2 with 7 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = dao.List(&database.DatabaseParams{Pagination: p}, nil)
		require.Nil(t, err)
		require.Len(t, result, 7)
		require.Equal(t, 17, p.TotalItems())
		assert.Equal(t, testData[0].Assets[0].Attachments[10].ID, result[0].ID)
		assert.Equal(t, testData[0].Assets[0].Attachments[16].ID, result[6].ID)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.List(nil, nil)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachment_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(1).Assets(1).Attachments(1).Build()
		err := dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": testData[0].Assets[0].Attachments[0].ID}}, nil)
		require.Nil(t, err)
	})

	t.Run("no db params", func(t *testing.T) {
		_, dao, _ := attachmentSetup(t)

		err := dao.Delete(nil, nil)
		assert.ErrorIs(t, err, ErrMissingWhere)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		err = dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": "1234"}}, nil)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAttachment_DeleteCascade(t *testing.T) {
	t.Run("delete course", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(2).Assets(1).Attachments(1).Build()

		// Delete the course
		courseDao := NewCourseDao(db)
		err := courseDao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": testData[0].ID}}, nil)
		require.Nil(t, err)

		// Check the asset was deleted
		s, err := dao.Get(testData[0].Assets[0].Attachments[0].ID, nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, s)
	})

	t.Run("delete asset", func(t *testing.T) {
		_, dao, db := attachmentSetup(t)

		testData := NewTestBuilder(t).Db(db).Courses(2).Assets(1).Attachments(1).Build()

		// Delete the asset
		assetDao := NewAssetDao(db)
		err := assetDao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": testData[0].Assets[0].ID}}, nil)
		require.Nil(t, err)

		_, err = dao.Get(testData[0].Assets[0].Attachments[0].ID, nil)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}
