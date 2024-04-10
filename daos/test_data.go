package daos

import (
	"database/sql"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Slice of 20 tags for testing (programming languages)
var test_tags = []string{
	"JavaScript", "Python", "Java", "Ruby", "PHP",
	"TypeScript", "C#", "C++", "C", "Swift",
	"Kotlin", "Rust", "Go", "Perl", "Scala",
	"R", "Objective-C", "Shell", "PowerShell", "Haskell",
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type TestBuilder struct {
	t                   *testing.T
	db                  database.Database
	numberOfCourses     int
	scan                bool
	assetsPerCourse     int
	attachmentsPerAsset int

	//
	tagsPerCourse          int
	specifiedTagsPerCourse []string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type TestCourse struct {
	*models.Course
	Scan   *models.Scan
	Assets []*models.Asset
	Tags   []*models.CourseTag
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func NewTestBuilder(t *testing.T) *TestBuilder {
	return &TestBuilder{
		t: t,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Db sets the database
func (builder *TestBuilder) Db(db database.Database) *TestBuilder {
	builder.db = db
	return builder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NumberOfCourses sets the number of courses
func (builder *TestBuilder) Courses(numberOfCourses int) *TestBuilder {
	builder.numberOfCourses = numberOfCourses
	return builder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan sets a scan per course
func (builder *TestBuilder) Scan() *TestBuilder {
	builder.scan = true
	return builder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Assets sets the number of assets per course
func (builder *TestBuilder) Assets(assetsPerCourse int) *TestBuilder {
	builder.assetsPerCourse = assetsPerCourse
	return builder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Attachments sets the number of attachments per asset
func (builder *TestBuilder) Attachments(attachmentsPerAsset int) *TestBuilder {
	builder.attachmentsPerAsset = attachmentsPerAsset
	return builder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tags sets either a random number of tags per course or a specific set of tags
func (builder *TestBuilder) Tags(tags any) *TestBuilder {
	switch t := tags.(type) {
	case int:
		builder.tagsPerCourse = t
	case []string:
		builder.specifiedTagsPerCourse = t
	}
	return builder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (builder *TestBuilder) Build() []*TestCourse {
	var testCourses []*TestCourse

	for i := 0; i < builder.numberOfCourses; i++ {
		tc := &TestCourse{}

		tc.Course = builder.newTestCourse()

		if builder.scan {
			tc.Scan = builder.newTestScan(tc.Course.ID)
		}

		if builder.assetsPerCourse > 0 {
			tc.Assets = builder.newTestAssets(tc.Course)
		}

		if (builder.tagsPerCourse > 0 || len(builder.specifiedTagsPerCourse) > 0) && builder.db != nil {
			tc.Tags = builder.newTestTags(tc.Course)
		}

		testCourses = append(testCourses, tc)
	}

	return testCourses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (builder *TestBuilder) newTestCourse() *models.Course {
	c := &models.Course{}

	c.RefreshId()
	c.RefreshCreatedAt()
	c.RefreshUpdatedAt()

	c.Title = fmt.Sprintf("Course %s", security.PseudorandomString(5))
	c.Path = fmt.Sprintf("/%s/%s", security.PseudorandomString(5), c.Title)

	if builder.db != nil {
		dao := NewCourseDao(builder.db)

		err := dao.Create(c)
		require.NoError(builder.t, err, "Failed to create course")

		time.Sleep(time.Millisecond * 1)
	}

	return c
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (builder *TestBuilder) newTestScan(courseId string) *models.Scan {
	s := &models.Scan{}

	s.RefreshId()
	s.RefreshCreatedAt()
	s.RefreshUpdatedAt()

	s.CourseID = courseId
	s.Status = types.NewScanStatus(types.ScanStatusWaiting)

	if builder.db != nil {
		dao := NewScanDao(builder.db)

		err := dao.Create(s)
		require.Nil(builder.t, err)

		time.Sleep(time.Millisecond * 1)
	}

	return s
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (builder *TestBuilder) newTestAssets(course *models.Course) []*models.Asset {
	assets := []*models.Asset{}

	for i := 0; i < builder.assetsPerCourse; i++ {
		a := &models.Asset{}

		a.RefreshId()
		a.RefreshCreatedAt()
		a.RefreshUpdatedAt()

		a.CourseID = course.ID
		a.Title = security.PseudorandomString(6)
		a.Prefix = sql.NullInt16{Int16: int16(rand.Intn(100-1) + 1), Valid: true}
		a.Chapter = fmt.Sprintf("%d chapter %s", i+1, security.PseudorandomString(2))
		a.Type = *types.NewAsset("mp4")
		a.Path = fmt.Sprintf("%s/%s/%d %s.mp4", course.Path, a.Chapter, a.Prefix.Int16, a.Title)

		if builder.db != nil {
			dao := NewAssetDao(builder.db)

			err := dao.Create(a)
			require.Nil(builder.t, err)

			time.Sleep(time.Millisecond * 1)
		}

		if builder.attachmentsPerAsset > 0 {
			a.Attachments = builder.newTestAttachments(a)
		}

		assets = append(assets, a)
	}

	return assets
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (builder *TestBuilder) newTestAttachments(asset *models.Asset) []*models.Attachment {
	attachments := []*models.Attachment{}

	for i := 0; i < builder.attachmentsPerAsset; i++ {
		a := &models.Attachment{}

		a.RefreshId()
		a.RefreshCreatedAt()
		a.RefreshUpdatedAt()

		a.CourseID = asset.CourseID
		a.AssetID = asset.ID
		a.Title = security.PseudorandomString(6)
		a.Path = fmt.Sprintf("%s/%d %s", filepath.Dir(asset.Path), asset.Prefix.Int16, a.Title)

		if builder.db != nil {
			dao := NewAttachmentDao(builder.db)

			err := dao.Create(a)
			require.Nil(builder.t, err)

			time.Sleep(time.Millisecond * 1)
		}

		attachments = append(attachments, a)

	}

	return attachments
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (builder *TestBuilder) newTestTags(course *models.Course) []*models.CourseTag {
	if builder.db == nil {
		return nil
	}

	tags := []*models.CourseTag{}
	chosenTags := map[string]bool{}

	count := builder.tagsPerCourse
	if len(builder.specifiedTagsPerCourse) > 0 {
		count = len(builder.specifiedTagsPerCourse)
	}

	for i := 0; i < count; i++ {
		tag := &models.Tag{}

		if len(builder.specifiedTagsPerCourse) > 0 {
			tag.Tag = builder.specifiedTagsPerCourse[i]
		} else {
			for {
				randomTag := test_tags[rand.Intn(len(test_tags))]
				if !chosenTags[randomTag] {
					tag = &models.Tag{
						Tag: randomTag,
					}
					chosenTags[randomTag] = true
					break
				}
			}
		}

		ct := &models.CourseTag{
			CourseId: course.ID,
		}

		dao := NewCourseTagDao(builder.db)
		require.Nil(builder.t, dao.Create(ct, tag.Tag, nil))

		tags = append(tags, ct)
	}

	return tags
}