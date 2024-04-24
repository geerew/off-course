package daos

import (
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

func tagSetup(t *testing.T) (*appFs.AppFs, *TagDao, database.Database) {
	appFs, db := setup(t)
	tagDao := NewTagDao(db)
	return appFs, tagDao, db
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestTag_Count(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, dao, _ := tagSetup(t)

		count, err := dao.Count(nil)
		require.Nil(t, err)
		assert.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		_, dao, _ := tagSetup(t)

		// Add test_tags into the database
		for _, tag := range test_tags {
			require.Nil(t, dao.Create(&models.Tag{Tag: tag}, nil))
		}

		count, err := dao.Count(nil)
		require.Nil(t, err)
		assert.Equal(t, count, len(test_tags))
	})

	t.Run("where", func(t *testing.T) {
		_, dao, _ := tagSetup(t)

		// Add test_tags into the database
		for _, tag := range test_tags {
			require.Nil(t, dao.Create(&models.Tag{Tag: tag}, nil))
		}

		// ----------------------------
		// EQUALS
		// ----------------------------
		count, err := dao.Count(&database.DatabaseParams{Where: squirrel.Eq{dao.Table + ".tag": test_tags[0]}})
		require.Nil(t, err)
		assert.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.NotEq{dao.Table + ".tag": test_tags[0]}})
		require.Nil(t, err)
		assert.Equal(t, 19, count)

		// ----------------------------
		//  STARTS WITH (Java%)
		// ----------------------------
		count, err = dao.Count(&database.DatabaseParams{Where: squirrel.Like{dao.Table + ".tag": "Java%"}})
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
		_, dao, db := tagSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		_, err = dao.Count(nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestTag_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, _ := tagSetup(t)

		tag := &models.Tag{
			Tag: "JavaScript",
		}

		err := dao.Create(tag, nil)
		require.Nil(t, err)
	})

	t.Run("duplicate tags", func(t *testing.T) {
		_, dao, _ := tagSetup(t)

		tag := &models.Tag{
			Tag: "JavaScript",
		}

		// Create the tag
		require.Nil(t, dao.Create(tag, nil))

		// Create the asset (again)
		require.ErrorContains(t, dao.Create(tag, nil), fmt.Sprintf("UNIQUE constraint failed: %s.tag", dao.Table))
	})

	t.Run("constraints", func(t *testing.T) {
		_, dao, _ := tagSetup(t)

		// Empty tag ID
		tag := &models.Tag{}
		assert.ErrorContains(t, dao.Create(tag, nil), fmt.Sprintf("NOT NULL constraint failed: %s.tag", dao.Table))

		// Success
		tag.Tag = "JavaScript"
		assert.Nil(t, dao.Create(tag, nil))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestTag_List(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		_, dao, _ := tagSetup(t)

		tags, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Zero(t, tags)
	})

	t.Run("found", func(t *testing.T) {
		_, dao, _ := tagSetup(t)

		NewTestBuilder(t).Db(dao.db).Courses(1).Tags([]string{"PHP", "Go"}).Build()
		NewTestBuilder(t).Db(dao.db).Courses(1).Tags([]string{"Go", "C"}).Build()
		NewTestBuilder(t).Db(dao.db).Courses(1).Tags([]string{"C", "TypeScript"}).Build()

		result, err := dao.List(nil, nil)
		require.Nil(t, err)
		require.Len(t, result, 4)

		// ----------------------------
		// Course tags
		// ----------------------------
		assert.Len(t, result[0].CourseTags, 1) // PHP
		assert.Len(t, result[1].CourseTags, 2) // GO
		assert.Len(t, result[2].CourseTags, 2) // C
		assert.Len(t, result[3].CourseTags, 1) // TypeScript

	})

	t.Run("orderby", func(t *testing.T) {
		_, dao, _ := tagSetup(t)

		testData := NewTestBuilder(t).
			Db(dao.db).
			Courses([]string{"course 1", "course 2", "course 3"}).
			Tags([]string{"PHP", "Go", "Java", "TypeScript", "C"}).Build()

		// ----------------------------
		// TAG DESC
		// ----------------------------
		result, err := dao.List(&database.DatabaseParams{OrderBy: []string{"tag desc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 5)
		assert.Equal(t, "TypeScript", result[0].Tag)

		// ----------------------------
		// TAG ASC
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"tag asc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 5)
		assert.Equal(t, "C", result[0].Tag)

		// ----------------------------
		// CREATED_AT ASC + COURSES.TITLE DESC
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"tag asc", NewCourseDao(dao.db).Table + ".title desc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 5)
		assert.Equal(t, "C", result[0].Tag)
		assert.Equal(t, testData[2].ID, result[0].CourseTags[0].CourseId)

		// ----------------------------
		// Error
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{OrderBy: []string{"unit_test asc"}}, nil)
		require.ErrorContains(t, err, "no such column")
		assert.Nil(t, result)
	})

	t.Run("where", func(t *testing.T) {
		_, dao, _ := tagSetup(t)

		for _, tag := range test_tags {
			require.Nil(t, dao.Create(&models.Tag{Tag: tag}, nil))
		}

		// ----------------------------
		// EQUALS (PHP)
		// ----------------------------
		result, err := dao.List(&database.DatabaseParams{Where: squirrel.Eq{dao.Table + ".tag": "PHP"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 1)

		// ----------------------------
		// LIKE (Java%)
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Like{dao.Table + ".tag": "Java%"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 2)

		// ----------------------------
		// ERROR
		// ----------------------------
		result, err = dao.List(&database.DatabaseParams{Where: squirrel.Eq{"": ""}}, nil)
		require.ErrorContains(t, err, "syntax error")
		assert.Nil(t, result)
	})

	t.Run("pagination", func(t *testing.T) {
		_, dao, _ := tagSetup(t)

		for _, tag := range test_tags {
			require.Nil(t, dao.Create(&models.Tag{Tag: tag}, nil))
		}

		// ----------------------------
		// Page 1 with 10 items
		// ----------------------------
		p := pagination.New(1, 10)

		result, err := dao.List(&database.DatabaseParams{Pagination: p, OrderBy: []string{"tag asc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 20, p.TotalItems())
		assert.Equal(t, "C", result[0].Tag)

		// ----------------------------
		// Page 2 with 10 items
		// ----------------------------
		p = pagination.New(2, 10)

		result, err = dao.List(&database.DatabaseParams{Pagination: p, OrderBy: []string{"tag asc"}}, nil)
		require.Nil(t, err)
		require.Len(t, result, 10)
		require.Equal(t, 20, p.TotalItems())
		assert.Equal(t, "Perl", result[0].Tag)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := tagSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		_, err = dao.List(nil, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestTag_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, _ := tagSetup(t)

		// Add test_tags into the database
		for _, tag := range test_tags {
			require.Nil(t, dao.Create(&models.Tag{Tag: tag}, nil))
		}

		err := dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"tag": test_tags[0]}}, nil)
		require.Nil(t, err)
	})

	t.Run("no db params", func(t *testing.T) {
		_, dao, _ := scanSetup(t)

		err := dao.Delete(nil, nil)
		assert.ErrorIs(t, err, ErrMissingWhere)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := tagSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.Table)
		require.Nil(t, err)

		err = dao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"tag": "1234"}}, nil)
		require.ErrorContains(t, err, "no such table: "+dao.Table)
	})
}
