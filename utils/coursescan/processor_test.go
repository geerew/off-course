package coursescan

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanner_Processor(t *testing.T) {
	t.Run("scan nil", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		err := Processor(ctx, scanner, nil)
		require.ErrorIs(t, err, ErrNilScan)
	})

	t.Run("error getting course", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		_, err := scanner.db.Exec("DROP TABLE IF EXISTS " + course.Table())
		require.NoError(t, err)

		err = Processor(ctx, scanner, scan)
		require.ErrorContains(t, err, fmt.Sprintf("no such table: %s", course.Table()))
	})

	t.Run("course unavailable", func(t *testing.T) {
		scanner, ctx, logs := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1", Available: true}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		err := Processor(ctx, scanner, scan)
		require.NoError(t, err)

		require.NotEmpty(t, *logs)
		require.Greater(t, len(*logs), 1)
		require.Equal(t, "Skipping unavailable course", (*logs)[len(*logs)-2].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-2].Level)
	})

	t.Run("mark course available", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1", Available: false}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)

		err := Processor(ctx, scanner, scan)
		require.NoError(t, err)

		courseResult := &models.Course{}
		options := &database.Options{
			Where:            squirrel.Eq{models.COURSE_TABLE_ID: course.ID},
			ExcludeRelations: []string{models.COURSE_RELATION_PROGRESS},
		}
		err = scanner.dao.GetCourse(ctx, courseResult, options)
		require.NoError(t, err)
		require.True(t, courseResult.Available)
	})

	t.Run("card", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		options := &database.Options{
			Where:            squirrel.Eq{models.COURSE_TABLE_ID: course.ID},
			ExcludeRelations: []string{models.COURSE_RELATION_PROGRESS},
		}

		// Add card at the root
		{
			scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
			scanner.appFs.Fs.Create(filepath.Join(course.Path, "card.jpg"))

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			courseResult := &models.Course{}
			err = scanner.dao.GetCourse(ctx, courseResult, options)
			require.NoError(t, err)
			require.Equal(t, filepath.Join(course.Path, "card.jpg"), courseResult.CardPath)
		}

		// Ignore card in chapter
		{
			scanner.appFs.Fs.Remove(filepath.Join(course.Path, "card.jpg"))
			scanner.appFs.Fs.Create(filepath.Join(course.Path, "01 Chapter 1", "card.jpg"))

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			courseResult := &models.Course{}
			err = scanner.dao.GetCourse(ctx, courseResult, options)
			require.NoError(t, err)
			require.Empty(t, courseResult.CardPath)
		}

		// Ignore additional cards at the root
		{
			scanner.appFs.Fs.Create(filepath.Join(course.Path, "card.jpg"))
			scanner.appFs.Fs.Create(filepath.Join(course.Path, "card.png"))

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			courseResult := &models.Course{}
			err = scanner.dao.GetCourse(ctx, courseResult, options)
			require.NoError(t, err)
			require.Equal(t, filepath.Join(course.Path, "card.jpg"), courseResult.CardPath)
		}
	})

	t.Run("ignore files", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/file 1", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/file.file", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/file.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/ - file.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/- - file.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/-1 - file.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/a - file.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/1.1 - file.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/2.3-file.avi", course.Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/1file.avi", course.Path))

		err := Processor(ctx, scanner, scan)
		require.NoError(t, err)

		count, err := dao.Count(ctx, scanner.dao, &models.Asset{}, nil)
		require.NoError(t, err)
		require.Zero(t, count)
	})

	t.Run("assets", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		assetGroups := []*models.AssetGroup{}
		assetGroupOptions := &database.Options{
			OrderBy: []string{models.ASSET_GROUP_TABLE_MODULE + " asc", models.ASSET_GROUP_TABLE_PREFIX + " asc"},
			Where:   squirrel.Eq{models.ASSET_GROUP_TABLE_COURSE_ID: course.ID},
		}

		// Add file 1, file 2 and file 3 (create op)
		{
			scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 file 1.mkv", course.Path), []byte("hash 1"), os.ModePerm)
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/02 file 2.html", course.Path), []byte("hash 2"), os.ModePerm)
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/03 file 3.pdf", course.Path), []byte("hash 3"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 3)

			require.Len(t, assetGroups[0].Assets, 1)
			require.Equal(t, "file 1", assetGroups[0].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[0].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[0].Prefix.Int16))
			require.Empty(t, assetGroups[0].Assets[0].Module)
			require.True(t, assetGroups[0].Assets[0].Type.IsVideo())
			require.Equal(t, "0657190350cbea662b6c15d703d9c7482308e511504d3308306d0f1ede153a34", assetGroups[0].Assets[0].Hash)

			require.Len(t, assetGroups[1].Assets, 1)
			require.Equal(t, "file 2", assetGroups[1].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[1].Assets[0].CourseID)
			require.Equal(t, 2, int(assetGroups[1].Assets[0].Prefix.Int16))
			require.Empty(t, assetGroups[1].Assets[0].Module)
			require.True(t, assetGroups[1].Assets[0].Type.IsHTML())
			require.Equal(t, "ac4f5d7f5ca1f7b2a9e8107ca793b5ead43a1d04afdafabc9488e93b5d738b41", assetGroups[1].Assets[0].Hash)

			require.Len(t, assetGroups[2].Assets, 1)
			require.Equal(t, "file 3", assetGroups[2].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[2].Assets[0].CourseID)
			require.Equal(t, 3, int(assetGroups[2].Assets[0].Prefix.Int16))
			require.Empty(t, assetGroups[2].Assets[0].Module)
			require.True(t, assetGroups[2].Assets[0].Type.IsPDF())
			require.Equal(t, "c4ca2e438d8809f0e4459bde1f948de8fe6289f1c179d506da8720fb79859be6", assetGroups[2].Assets[0].Hash)
		}

		// Add file 1 under a chapter (create op)
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 Chapter 1/01 file 1.pdf", course.Path), []byte("hash 4"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 4)

			require.Len(t, assetGroups[3].Assets, 1)
			require.Equal(t, "file 1", assetGroups[3].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[3].Assets[0].CourseID)
			require.Equal(t, 1, int(assetGroups[3].Assets[0].Prefix.Int16))
			require.Equal(t, "01 Chapter 1", assetGroups[3].Assets[0].Module)
			require.True(t, assetGroups[3].Assets[0].Type.IsPDF())
		}

		// Delete file 1 in chapter (delete op)
		{
			scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 Chapter 1/01 file 1.pdf", course.Path))

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 3)

			require.Len(t, assetGroups[0].Assets, 1)
			require.Equal(t, fmt.Sprintf("%s/01 file 1.mkv", course.Path), assetGroups[0].Assets[0].Path)
			require.Len(t, assetGroups[1].Assets, 1)
			require.Equal(t, fmt.Sprintf("%s/02 file 2.html", course.Path), assetGroups[1].Assets[0].Path)
			require.Len(t, assetGroups[2].Assets, 1)
			require.Equal(t, fmt.Sprintf("%s/03 file 3.pdf", course.Path), assetGroups[2].Assets[0].Path)
		}

		// Rename file 3 to file 4 (update op)
		{
			existingAssetID := assetGroups[2].Assets[0].ID
			scanner.appFs.Fs.Rename(fmt.Sprintf("%s/03 file 3.pdf", course.Path), fmt.Sprintf("%s/04 file 4.pdf", course.Path))

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 3)

			require.Len(t, assetGroups[0].Assets, 1)
			require.Equal(t, fmt.Sprintf("%s/01 file 1.mkv", course.Path), assetGroups[0].Assets[0].Path)

			require.Len(t, assetGroups[1].Assets, 1)
			require.Equal(t, fmt.Sprintf("%s/02 file 2.html", course.Path), assetGroups[1].Assets[0].Path)

			require.Len(t, assetGroups[2].Assets, 1)
			require.Equal(t, fmt.Sprintf("%s/04 file 4.pdf", course.Path), assetGroups[2].Assets[0].Path)
			require.Equal(t, existingAssetID, assetGroups[2].Assets[0].ID)
		}

		// Replace file 4 with new content (replace op)
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/04 file 4.pdf", course.Path), []byte("hash 4"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 3)

			require.Len(t, assetGroups[0].Assets, 1)
			require.Equal(t, fmt.Sprintf("%s/01 file 1.mkv", course.Path), assetGroups[0].Assets[0].Path)

			require.Len(t, assetGroups[1].Assets, 1)
			require.Equal(t, fmt.Sprintf("%s/02 file 2.html", course.Path), assetGroups[1].Assets[0].Path)

			require.Len(t, assetGroups[2].Assets, 1)
			require.Equal(t, fmt.Sprintf("%s/04 file 4.pdf", course.Path), assetGroups[2].Assets[0].Path)
			require.Equal(t, "file 4", assetGroups[2].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[2].Assets[0].CourseID)
			require.Equal(t, 4, int(assetGroups[2].Assets[0].Prefix.Int16))
			require.Empty(t, assetGroups[2].Assets[0].Module)
			require.True(t, assetGroups[2].Assets[0].Type.IsPDF())
			require.Equal(t, "e72c82bb74988135e7b6c478fe3659a14b4941f867a93a23687ea172031e4e06", assetGroups[2].Assets[0].Hash)
		}

		// Swap file 1 and file 2 (swap op)
		{
			scanner.appFs.Fs.Rename(fmt.Sprintf("%s/01 file 1.mkv", course.Path), fmt.Sprintf("%s/02 file 2.html.temp", course.Path))
			scanner.appFs.Fs.Rename(fmt.Sprintf("%s/02 file 2.html", course.Path), fmt.Sprintf("%s/01 file 1.mkv", course.Path))
			scanner.appFs.Fs.Rename(fmt.Sprintf("%s/02 file 2.html.temp", course.Path), fmt.Sprintf("%s/02 file 2.html", course.Path))

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 3)

			require.Len(t, assetGroups[0].Assets, 1)
			require.Equal(t, fmt.Sprintf("%s/01 file 1.mkv", course.Path), assetGroups[0].Assets[0].Path)
			require.Len(t, assetGroups[1].Assets, 1)
			require.Equal(t, fmt.Sprintf("%s/02 file 2.html", course.Path), assetGroups[1].Assets[0].Path)
			require.Len(t, assetGroups[2].Assets, 1)
			require.Equal(t, fmt.Sprintf("%s/04 file 4.pdf", course.Path), assetGroups[2].Assets[0].Path)

			require.Equal(t, "ac4f5d7f5ca1f7b2a9e8107ca793b5ead43a1d04afdafabc9488e93b5d738b41", assetGroups[0].Assets[0].Hash)
			require.Equal(t, "0657190350cbea662b6c15d703d9c7482308e511504d3308306d0f1ede153a34", assetGroups[1].Assets[0].Hash)
		}

		// Delete file 1 and move file 2 to file 1 (overwrite op)
		{
			require.NoError(t, scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 file 1.mkv", course.Path)))

			require.NoError(t, scanner.appFs.Fs.Rename(
				fmt.Sprintf("%s/02 file 2.html", course.Path),
				fmt.Sprintf("%s/01 file 1.mkv", course.Path),
			))

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 2)

			require.Len(t, assetGroups[0].Assets, 1)
			require.Equal(t, fmt.Sprintf("%s/01 file 1.mkv", course.Path), assetGroups[0].Assets[0].Path)
			require.Len(t, assetGroups[1].Assets, 1)
			require.Equal(t, fmt.Sprintf("%s/04 file 4.pdf", course.Path), assetGroups[1].Assets[0].Path)

			require.Equal(t, "0657190350cbea662b6c15d703d9c7482308e511504d3308306d0f1ede153a34", assetGroups[0].Assets[0].Hash)
		}

		// Delete all files but keep the course directory
		{
			// Delete all files but keep the course directory
			scanner.appFs.Fs.RemoveAll(fmt.Sprintf("%s/01 file 1.mkv", course.Path))
			scanner.appFs.Fs.RemoveAll(fmt.Sprintf("%s/04 file 4.pdf", course.Path))

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 0)
		}

		// Add file 1 with a sub-prefix (create op)
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 group 1 {01 video 1}.mkv", course.Path), []byte("hash 1"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)
			require.Len(t, assetGroups[0].Assets, 1)

			require.Equal(t, "group 1", assetGroups[0].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[0].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[0].Prefix.Int16))
			require.Equal(t, 1, int(assetGroups[0].Assets[0].SubPrefix.Int16))
			require.Equal(t, "video 1", assetGroups[0].Assets[0].SubTitle)
		}

		// Add file 2 with a sub-prefix (no op)
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 group 1 {02 video 2}.mkv", course.Path), []byte("hash 2"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)
			require.Len(t, assetGroups[0].Assets, 2)

			require.Equal(t, "group 1", assetGroups[0].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[0].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[0].Prefix.Int16))
			require.Equal(t, 1, int(assetGroups[0].Assets[0].SubPrefix.Int16))
			require.Equal(t, "video 1", assetGroups[0].Assets[0].SubTitle)

			require.Equal(t, "group 1", assetGroups[0].Assets[1].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[1].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[1].Prefix.Int16))
			require.Equal(t, 2, int(assetGroups[0].Assets[1].SubPrefix.Int16))
			require.Equal(t, "video 2", assetGroups[0].Assets[1].SubTitle)
		}
	})

	t.Run("attachments", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		assetGroups := []*models.AssetGroup{}
		assetGroupOptions := &database.Options{
			OrderBy: []string{models.ASSET_GROUP_TABLE_MODULE + " asc", models.ASSET_GROUP_TABLE_PREFIX + " asc"},
			Where:   squirrel.Eq{models.ASSET_GROUP_TABLE_COURSE_ID: course.ID},
		}

		// Add asset (so the asset group is created)
		{
			scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 file 1.mkv", course.Path), []byte("hash 1"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Assets, 1)
			require.Equal(t, "file 1", assetGroups[0].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[0].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[0].Prefix.Int16))
			require.Empty(t, assetGroups[0].Assets[0].Module)
			require.True(t, assetGroups[0].Assets[0].Type.IsVideo())
			require.Equal(t, "0657190350cbea662b6c15d703d9c7482308e511504d3308306d0f1ede153a34", assetGroups[0].Assets[0].Hash)
		}

		// Add attachment 1
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 attachment 1.url", course.Path), []byte("attachment 1"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Attachments, 1)
			require.Equal(t, "attachment 1.url", assetGroups[0].Attachments[0].Title)
			require.Equal(t, filepath.Join(course.Path, "01 attachment 1.url"), assetGroups[0].Attachments[0].Path)
		}

		// Add another attachment
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 attachment 2.url", course.Path), []byte("attachment 2"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Attachments, 2)
			require.Equal(t, "attachment 1.url", assetGroups[0].Attachments[0].Title)
			require.Equal(t, filepath.Join(course.Path, "01 attachment 1.url"), assetGroups[0].Attachments[0].Path)

			require.Equal(t, "attachment 2.url", assetGroups[0].Attachments[1].Title)
			require.Equal(t, filepath.Join(course.Path, "01 attachment 2.url"), assetGroups[0].Attachments[1].Path)
		}

		// Delete attachment
		{
			scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 attachment 1.url", course.Path))

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Attachments, 1)
			require.Equal(t, "attachment 2.url", assetGroups[0].Attachments[0].Title)
			require.Equal(t, filepath.Join(course.Path, "01 attachment 2.url"), assetGroups[0].Attachments[0].Path)
		}
	})

	t.Run("asset priority", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		// Priority is VIDEO -> HTML -> PDF -> MARKDOWN -> TEXT

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		assetGroups := []*models.AssetGroup{}

		assetGroupOptions := &database.Options{
			OrderBy: []string{models.ASSET_GROUP_TABLE_MODULE + " asc", models.ASSET_GROUP_TABLE_PREFIX + " asc"},
			Where:   squirrel.Eq{models.ASSET_GROUP_TABLE_COURSE_ID: course.ID},
		}

		scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)

		// Add TEXT asset
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 text 1.txt", course.Path), []byte("text 1"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)
			require.Len(t, assetGroups[0].Assets, 1)
			require.Len(t, assetGroups[0].Attachments, 0)

			require.Equal(t, "text 1", assetGroups[0].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[0].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[0].Prefix.Int16))
			require.Empty(t, assetGroups[0].Assets[0].Module)
			require.True(t, assetGroups[0].Assets[0].Type.IsText())
			require.Equal(t, "900a4469df00ccbfd0c145c6d1e4b7953dd0afafadd7534e3a4019e8d38fc663", assetGroups[0].Assets[0].Hash)
		}

		// Add MARKDOWN asset
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 markdown 1.md", course.Path), []byte("markdown 1"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Assets, 1)
			require.Equal(t, "markdown 1", assetGroups[0].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[0].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[0].Prefix.Int16))
			require.Empty(t, assetGroups[0].Assets[0].Module)
			require.True(t, assetGroups[0].Assets[0].Type.IsMarkdown())
			require.Equal(t, "728cfbd456c4734229b7b545d69d182608eecc860c46081f51e3f1f108096eca", assetGroups[0].Assets[0].Hash)

			require.Len(t, assetGroups[0].Attachments, 1)
			require.Equal(t, filepath.Join(course.Path, "01 text 1.txt"), assetGroups[0].Attachments[0].Path)
		}

		// Add PDF asset
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 pdf 1.pdf", course.Path), []byte("pdf 1"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Assets, 1)
			require.Equal(t, "pdf 1", assetGroups[0].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[0].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[0].Prefix.Int16))
			require.Empty(t, assetGroups[0].Assets[0].Module)
			require.True(t, assetGroups[0].Assets[0].Type.IsPDF())
			require.Equal(t, "9c9bfc90d1a2738f701a22c1ef10d42d5f2c285998a221eba9b7953e202bcf1a", assetGroups[0].Assets[0].Hash)

			require.Len(t, assetGroups[0].Attachments, 2)
			require.Equal(t, filepath.Join(course.Path, "01 text 1.txt"), assetGroups[0].Attachments[0].Path)
			require.Equal(t, filepath.Join(course.Path, "01 markdown 1.md"), assetGroups[0].Attachments[1].Path)
		}

		// Add HTML asset
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 index.html", course.Path), []byte("index"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Assets, 1)
			require.Equal(t, "index", assetGroups[0].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[0].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[0].Prefix.Int16))
			require.Empty(t, assetGroups[0].Assets[0].Module)
			require.True(t, assetGroups[0].Assets[0].Type.IsHTML())
			require.Equal(t, "1bc04b5291c26a46d918139138b992d2de976d6851d0893b0476b85bfbdfc6e6", assetGroups[0].Assets[0].Hash)

			require.Len(t, assetGroups[0].Attachments, 3)
			require.Equal(t, filepath.Join(course.Path, "01 text 1.txt"), assetGroups[0].Attachments[0].Path)
			require.Equal(t, filepath.Join(course.Path, "01 markdown 1.md"), assetGroups[0].Attachments[1].Path)
			require.Equal(t, filepath.Join(course.Path, "01 pdf 1.pdf"), assetGroups[0].Attachments[2].Path)
		}

		// Add VIDEO asset
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 video.mp4", course.Path), []byte("video"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Assets, 1)
			require.Equal(t, "video", assetGroups[0].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[0].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[0].Prefix.Int16))
			require.Empty(t, assetGroups[0].Assets[0].Module)
			require.True(t, assetGroups[0].Assets[0].Type.IsVideo())
			require.Equal(t, "0cab1c9617404faf2b24e221e189ca5945813e14d3f766345b09ca13bbe28ffc", assetGroups[0].Assets[0].Hash)

			require.Len(t, assetGroups[0].Attachments, 4)
			require.Equal(t, filepath.Join(course.Path, "01 text 1.txt"), assetGroups[0].Attachments[0].Path)
			require.Equal(t, filepath.Join(course.Path, "01 markdown 1.md"), assetGroups[0].Attachments[1].Path)
			require.Equal(t, filepath.Join(course.Path, "01 pdf 1.pdf"), assetGroups[0].Attachments[2].Path)
			require.Equal(t, filepath.Join(course.Path, "01 index.html"), assetGroups[0].Attachments[3].Path)
		}

		// Add another PDF asset
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 pdf 2.pdf", course.Path), []byte("pdf 2"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Assets, 1)
			require.Equal(t, "video", assetGroups[0].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[0].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[0].Prefix.Int16))
			require.Empty(t, assetGroups[0].Assets[0].Module)
			require.True(t, assetGroups[0].Assets[0].Type.IsVideo())
			require.Equal(t, "0cab1c9617404faf2b24e221e189ca5945813e14d3f766345b09ca13bbe28ffc", assetGroups[0].Assets[0].Hash)

			require.Len(t, assetGroups[0].Attachments, 5)
			require.Equal(t, filepath.Join(course.Path, "01 text 1.txt"), assetGroups[0].Attachments[0].Path)
			require.Equal(t, filepath.Join(course.Path, "01 markdown 1.md"), assetGroups[0].Attachments[1].Path)
			require.Equal(t, filepath.Join(course.Path, "01 pdf 1.pdf"), assetGroups[0].Attachments[2].Path)
			require.Equal(t, filepath.Join(course.Path, "01 index.html"), assetGroups[0].Attachments[3].Path)
			require.Equal(t, filepath.Join(course.Path, "01 pdf 2.pdf"), assetGroups[0].Attachments[4].Path)
		}
	})

	t.Run("asset with sub-prefix and sub-title", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		assetGroups := []*models.AssetGroup{}

		assetGroupOptions := &database.Options{
			OrderBy: []string{models.ASSET_GROUP_TABLE_MODULE + " asc", models.ASSET_GROUP_TABLE_PREFIX + " asc"},
			Where:   squirrel.Eq{models.ASSET_GROUP_TABLE_COURSE_ID: course.ID},
		}

		// Add video 1 asset with sub-prefix of 1 and sub-title "Part 1"
		{
			scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 Group 1 {1 Video 1}.mp4", course.Path), []byte("video 1"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Assets, 1)
			require.Equal(t, "Group 1", assetGroups[0].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[0].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[0].Prefix.Int16))
			require.Equal(t, 1, int(assetGroups[0].Assets[0].SubPrefix.Int16))
			require.Equal(t, "Video 1", assetGroups[0].Assets[0].SubTitle)
			require.Empty(t, assetGroups[0].Assets[0].Module)
			require.True(t, assetGroups[0].Assets[0].Type.IsVideo())
			require.Equal(t, "3b857b8441d7c9e734535d6b82f69a34c6fcd63ed0ef989ff03808ecb29a2f1f", assetGroups[0].Assets[0].Hash)

			require.Len(t, assetGroups[0].Attachments, 0)
		}

		// Add video 2 asset with sub-prefix of 2 and sub-title "Part 2"
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 Group 1 {2 Video 2}.mp4", course.Path), []byte("video 2"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Assets, 2)
			require.Equal(t, "3b857b8441d7c9e734535d6b82f69a34c6fcd63ed0ef989ff03808ecb29a2f1f", assetGroups[0].Assets[0].Hash)

			require.Equal(t, "Group 1", assetGroups[0].Assets[1].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[1].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[1].Prefix.Int16))
			require.Equal(t, 2, int(assetGroups[0].Assets[1].SubPrefix.Int16))
			require.Equal(t, "Video 2", assetGroups[0].Assets[1].SubTitle)
			require.Empty(t, assetGroups[0].Assets[1].Module)
			require.True(t, assetGroups[0].Assets[1].Type.IsVideo())
			require.Equal(t, "614ef49d4a1ef39bc763b7c9665f6f30a0eea3ec5ec10e04b897bdad9b973f9c", assetGroups[0].Assets[1].Hash)

			require.Len(t, assetGroups[0].Attachments, 0)
		}

		// Add video 3 asset with sub-prefix of 3 and no sub-title
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 Group 1 {3}.mp4", course.Path), []byte("video 3"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Assets, 3)

			require.Equal(t, "3b857b8441d7c9e734535d6b82f69a34c6fcd63ed0ef989ff03808ecb29a2f1f", assetGroups[0].Assets[0].Hash)
			require.Equal(t, "614ef49d4a1ef39bc763b7c9665f6f30a0eea3ec5ec10e04b897bdad9b973f9c", assetGroups[0].Assets[1].Hash)

			require.Equal(t, "Group 1", assetGroups[0].Assets[2].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[2].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[2].Prefix.Int16))
			require.Equal(t, 3, int(assetGroups[0].Assets[2].SubPrefix.Int16))
			require.Empty(t, assetGroups[0].Assets[2].SubTitle)
			require.Empty(t, assetGroups[0].Assets[2].Module)
			require.True(t, assetGroups[0].Assets[2].Type.IsVideo())
			require.Equal(t, "36d9fa5c21ca58822f678a5e1cebbaefcbcff37894771089cc608e8fbe32121e", assetGroups[0].Assets[2].Hash)

			require.Len(t, assetGroups[0].Attachments, 0)
		}

		// Add video 4 with no sub-prefix and no sub-title
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 Group 1.mp4", course.Path), []byte("video 4"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Assets, 3)

			require.Equal(t, "3b857b8441d7c9e734535d6b82f69a34c6fcd63ed0ef989ff03808ecb29a2f1f", assetGroups[0].Assets[0].Hash)
			require.Equal(t, "614ef49d4a1ef39bc763b7c9665f6f30a0eea3ec5ec10e04b897bdad9b973f9c", assetGroups[0].Assets[1].Hash)
			require.Equal(t, "36d9fa5c21ca58822f678a5e1cebbaefcbcff37894771089cc608e8fbe32121e", assetGroups[0].Assets[2].Hash)

			require.Len(t, assetGroups[0].Attachments, 1)
			require.Equal(t, filepath.Join(course.Path, "01 Group 1.mp4"), assetGroups[0].Attachments[0].Path)
		}

		// Add attachment
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 attachment 1.txt", course.Path), []byte("attachment 1"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Assets, 3)
			require.Equal(t, "3b857b8441d7c9e734535d6b82f69a34c6fcd63ed0ef989ff03808ecb29a2f1f", assetGroups[0].Assets[0].Hash)
			require.Equal(t, "614ef49d4a1ef39bc763b7c9665f6f30a0eea3ec5ec10e04b897bdad9b973f9c", assetGroups[0].Assets[1].Hash)
			require.Equal(t, "36d9fa5c21ca58822f678a5e1cebbaefcbcff37894771089cc608e8fbe32121e", assetGroups[0].Assets[2].Hash)

			require.Len(t, assetGroups[0].Attachments, 2)
			require.Equal(t, filepath.Join(course.Path, "01 Group 1.mp4"), assetGroups[0].Attachments[0].Path)
			require.Equal(t, filepath.Join(course.Path, "01 attachment 1.txt"), assetGroups[0].Attachments[1].Path)
		}
	})

	t.Run("asset description", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		assetGroups := []*models.AssetGroup{}

		assetGroupOptions := &database.Options{
			OrderBy: []string{models.ASSET_GROUP_TABLE_MODULE + " asc", models.ASSET_GROUP_TABLE_PREFIX + " asc"},
			Where:   squirrel.Eq{models.ASSET_GROUP_TABLE_COURSE_ID: course.ID},
		}

		// Add video 1 asset
		{
			scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 video 1.mp4", course.Path), []byte("video 1"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Len(t, assetGroups[0].Assets, 1)
			require.Equal(t, "video 1", assetGroups[0].Assets[0].Title)
			require.Equal(t, course.ID, assetGroups[0].Assets[0].CourseID)
			require.Equal(t, 1, int(assetGroups[0].Assets[0].Prefix.Int16))
			require.Empty(t, assetGroups[0].Assets[0].Module)
			require.True(t, assetGroups[0].Assets[0].Type.IsVideo())
			require.Equal(t, "3b857b8441d7c9e734535d6b82f69a34c6fcd63ed0ef989ff03808ecb29a2f1f", assetGroups[0].Assets[0].Hash)
		}

		// Add description
		{
			scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 description.md", course.Path), []byte("description"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssetGroups(ctx, &assetGroups, assetGroupOptions)
			require.NoError(t, err)
			require.Len(t, assetGroups, 1)

			require.Equal(t, fmt.Sprintf("%s/01 description.md", course.Path), assetGroups[0].DescriptionPath)
			require.True(t, assetGroups[0].DescriptionType.IsMarkdown())
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanner_ParseFilename(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		var tests = []string{
			// No prefix
			"file",
			"file.file",
			"file.avi",
			" - file.avi",
			"- - file.avi",
			".avi",
			// Invalid prefix
			"-1 - file.avi",
			"a - file.avi",
			"1.1 - file.avi",
			"2.3-file.avi",
			"1file.avi",
		}

		for _, tt := range tests {
			parsed := parseFilename("xx", tt)
			require.Nil(t, parsed)
		}
	})

	t.Run("assets", func(t *testing.T) {
		var tests = []struct {
			in       string
			expected *parsedFile
		}{
			// Video (varied)
			{"0    file 0.avi", &parsedFile{Prefix: 0, Title: "file 0", SubPrefix: nil, SubTitle: "", Ext: "avi", AssetType: types.NewAsset("avi"), IsCard: false, Original: "0    file 0.avi", NormalizedPath: "0    file 0.avi"}},
			{"001 file 1.mp4", &parsedFile{Prefix: 1, Title: "file 1", SubPrefix: nil, SubTitle: "", Ext: "mp4", AssetType: types.NewAsset("mp4"), IsCard: false, Original: "001 file 1.mp4", NormalizedPath: "001 file 1.mp4"}},
			{"1-file.ogg", &parsedFile{Prefix: 1, Title: "file", SubPrefix: nil, SubTitle: "", Ext: "ogg", AssetType: types.NewAsset("ogg"), IsCard: false, Original: "1-file.ogg", NormalizedPath: "1-file.ogg"}},
			{"2 - file.webm", &parsedFile{Prefix: 2, Title: "file", SubPrefix: nil, SubTitle: "", Ext: "webm", AssetType: types.NewAsset("webm"), IsCard: false, Original: "2 - file.webm", NormalizedPath: "2 - file.webm"}},
			{"3 -file.m4a", &parsedFile{Prefix: 3, Title: "file", SubPrefix: nil, SubTitle: "", Ext: "m4a", AssetType: types.NewAsset("m4a"), IsCard: false, Original: "3 -file.m4a", NormalizedPath: "3 -file.m4a"}},
			{"4- file.opus", &parsedFile{Prefix: 4, Title: "file", SubPrefix: nil, SubTitle: "", Ext: "opus", AssetType: types.NewAsset("opus"), IsCard: false, Original: "4- file.opus", NormalizedPath: "4- file.opus"}},
			{"5000 --- file.wav", &parsedFile{Prefix: 5000, Title: "file", SubPrefix: nil, SubTitle: "", Ext: "wav", AssetType: types.NewAsset("wav"), IsCard: false, Original: "5000 --- file.wav", NormalizedPath: "5000 --- file.wav"}},
			{"0100 file.mp3", &parsedFile{Prefix: 100, Title: "file", SubPrefix: nil, SubTitle: "", Ext: "mp3", AssetType: types.NewAsset("mp3"), IsCard: false, Original: "0100 file.mp3", NormalizedPath: "0100 file.mp3"}},
			// PDF (including mixed case)
			{"1 - doc.pdf", &parsedFile{Prefix: 1, Title: "doc", SubPrefix: nil, SubTitle: "", Ext: "pdf", AssetType: types.NewAsset("pdf"), IsCard: false, Original: "1 - doc.pdf", NormalizedPath: "1 - doc.pdf"}},
			{"2 - REPORT.PDF", &parsedFile{Prefix: 2, Title: "REPORT", SubPrefix: nil, SubTitle: "", Ext: "pdf", AssetType: types.NewAsset("pdf"), IsCard: false, Original: "2 - REPORT.PDF", NormalizedPath: "2 - REPORT.PDF"}},
			// HTML
			{"1 index.html", &parsedFile{Prefix: 1, Title: "index", SubPrefix: nil, SubTitle: "", Ext: "html", AssetType: types.NewAsset("html"), IsCard: false, Original: "1 index.html", NormalizedPath: "1 index.html"}},
			// Markdown
			{"5 notes.md", &parsedFile{Prefix: 5, Title: "notes", SubPrefix: nil, SubTitle: "", Ext: "md", AssetType: types.NewAsset("md"), IsCard: false, Original: "5 notes.md", NormalizedPath: "5 notes.md"}},
			// Text
			{"6 readme.txt", &parsedFile{Prefix: 6, Title: "readme", SubPrefix: nil, SubTitle: "", Ext: "txt", AssetType: types.NewAsset("txt"), IsCard: false, Original: "6 readme.txt", NormalizedPath: "6 readme.txt"}},
			// With sub-prefix but no subtitle
			{"01 file 0 {1}.avi", &parsedFile{Prefix: 1, Title: "file 0", SubPrefix: intPtr(1), SubTitle: "", Ext: "avi", AssetType: types.NewAsset("avi"), IsCard: false, Original: "01 file 0 {1}.avi", NormalizedPath: "01 file 0 {1}.avi"}},
			// With sub-prefix and subtitle
			{"01 file 0 {1 Part 1}.avi", &parsedFile{Prefix: 1, Title: "file 0", SubPrefix: intPtr(1), SubTitle: "Part 1", Ext: "avi", AssetType: types.NewAsset("avi"), IsCard: false, Original: "01 file 0 {1 Part 1}.avi", NormalizedPath: "01 file 0 {1 Part 1}.avi"}},
			{"01 file 0 {2 -   Part 2}.mp4", &parsedFile{Prefix: 1, Title: "file 0", SubPrefix: intPtr(2), SubTitle: "Part 2", Ext: "mp4", AssetType: types.NewAsset("mp4"), IsCard: false, Original: "01 file 0 {2 -   Part 2}.mp4", NormalizedPath: "01 file 0 {2 -   Part 2}.mp4"}},
			{"01 file 0 {}.mp4", &parsedFile{Prefix: 1, Title: "file 0", SubPrefix: nil, SubTitle: "", Ext: "mp4", AssetType: types.NewAsset("mp4"), IsCard: false, Original: "01 file 0 {}.mp4", NormalizedPath: "01 file 0 {}.mp4"}},
			// Description-like filenames
			{"01 description.md", &parsedFile{Prefix: 1, Title: "description", SubPrefix: nil, SubTitle: "", Ext: "md", AssetType: types.NewAsset("md"), IsCard: false, Original: "01 description.md", NormalizedPath: "01 description.md"}},
			{"02 Description.TXT", &parsedFile{Prefix: 2, Title: "Description", SubPrefix: nil, SubTitle: "", Ext: "txt", AssetType: types.NewAsset("txt"), IsCard: false, Original: "02 Description.TXT", NormalizedPath: "02 Description.TXT"}},
		}

		for _, tt := range tests {
			parsed := parseFilename(tt.in, tt.in)
			require.Equal(t, tt.expected, parsed, fmt.Sprintf("error for [%s]", tt.in))

		}
	})

	t.Run("cards", func(t *testing.T) {
		var tests = []struct {
			in       string
			expected *parsedFile
		}{
			{"card.jpg", &parsedFile{Prefix: 0, Title: "card", SubPrefix: nil, SubTitle: "", Ext: "jpg", AssetType: nil, IsCard: true, Original: "card.jpg", NormalizedPath: "card.jpg"}},
			{"card.jpeg", &parsedFile{Prefix: 0, Title: "card", SubPrefix: nil, SubTitle: "", Ext: "jpeg", AssetType: nil, IsCard: true, Original: "card.jpeg", NormalizedPath: "card.jpeg"}},
			{"card.png", &parsedFile{Prefix: 0, Title: "card", SubPrefix: nil, SubTitle: "", Ext: "png", AssetType: nil, IsCard: true, Original: "card.png", NormalizedPath: "card.png"}},
		}

		for _, tt := range tests {
			parsed := parseFilename(tt.in, tt.in)
			require.Equal(t, tt.expected, parsed, fmt.Sprintf("error for [%s]", tt.in))
		}
	})

	t.Run("attachments", func(t *testing.T) {
		var tests = []struct {
			in       string
			expected *parsedFile
		}{
			// No title
			{"01", &parsedFile{Prefix: 1, Title: "", SubPrefix: nil, SubTitle: "", Ext: "", AssetType: nil, IsCard: false, Original: "01", NormalizedPath: "01"}},
			{"200.pdf", &parsedFile{Prefix: 200, Title: "", SubPrefix: nil, SubTitle: "", Ext: "pdf", AssetType: types.NewAsset("pdf"), IsCard: false, Original: "200.pdf", NormalizedPath: "200.pdf"}},
			{"1 -.txt", &parsedFile{Prefix: 1, Title: "", SubPrefix: nil, SubTitle: "", Ext: "txt", AssetType: types.NewAsset("txt"), IsCard: false, Original: "1 -.txt", NormalizedPath: "1 -.txt"}},
			{"1 - .txt", &parsedFile{Prefix: 1, Title: "", SubPrefix: nil, SubTitle: "", Ext: "txt", AssetType: types.NewAsset("txt"), IsCard: false, Original: "1 - .txt", NormalizedPath: "1 - .txt"}},
			{"1 .txt", &parsedFile{Prefix: 1, Title: "", SubPrefix: nil, SubTitle: "", Ext: "txt", AssetType: types.NewAsset("txt"), IsCard: false, Original: "1 .txt", NormalizedPath: "1 .txt"}},
			{"1     .pdf", &parsedFile{Prefix: 1, Title: "", SubPrefix: nil, SubTitle: "", Ext: "pdf", AssetType: types.NewAsset("pdf"), IsCard: false, Original: "1     .pdf", NormalizedPath: "1     .pdf"}},
			// No extension
			{"1 - file", &parsedFile{Prefix: 1, Title: "file", SubPrefix: nil, SubTitle: "", Ext: "", AssetType: nil, IsCard: false, Original: "1 - file", NormalizedPath: "1 - file"}},
			{"2 file", &parsedFile{Prefix: 2, Title: "file", SubPrefix: nil, SubTitle: "", Ext: "", AssetType: nil, IsCard: false, Original: "2 file", NormalizedPath: "2 file"}},
			{"3-file", &parsedFile{Prefix: 3, Title: "file", SubPrefix: nil, SubTitle: "", Ext: "", AssetType: nil, IsCard: false, Original: "3-file", NormalizedPath: "3-file"}},
			{"4 file", &parsedFile{Prefix: 4, Title: "file", SubPrefix: nil, SubTitle: "", Ext: "", AssetType: nil, IsCard: false, Original: "4 file", NormalizedPath: "4 file"}},
			{"6 --- file", &parsedFile{Prefix: 6, Title: "file", SubPrefix: nil, SubTitle: "", Ext: "", AssetType: nil, IsCard: false, Original: "6 --- file", NormalizedPath: "6 --- file"}},
			// Non-asset extensions
			{"1 - file.exe", &parsedFile{Prefix: 1, Title: "file", SubPrefix: nil, SubTitle: "", Ext: "exe", AssetType: types.NewAsset("exe"), IsCard: false, Original: "1 - file.exe", NormalizedPath: "1 - file.exe"}},
		}

		for _, tt := range tests {
			parsed := parseFilename(tt.in, tt.in)
			require.Equal(t, tt.expected, parsed, fmt.Sprintf("error for [%s]", tt.in))
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanner_CategorizeFile(t *testing.T) {
	var tests = []struct {
		in       string
		expected FileCategory
	}{
		{"file", Ignore},
		{"file.file", Ignore},
		{"file.avi", Ignore},
		{" - file.avi", Ignore},
		{"- - file.avi", Ignore},
		{".avi", Ignore},
		{"-1 - file.avi", Ignore},
		{"a - file.avi", Ignore},
		{"1.1 - file.avi", Ignore},
		{"2.3-file.avi", Ignore},
		{"1file.avi", Ignore},
		// Asset
		{"0    file 0.avi", Asset},
		{"001 file 1.mp4", Asset},
		{"1-file.ogg", Asset},
		{"2 - file.webm", Asset},
		{"3 -file.m4a", Asset},
		{"4- file.opus", Asset},
		{"5000 --- file.wav", Asset},
		{"0100 file.mp3", Asset},
		{"1 - doc.pdf", Asset},
		{"2 - REPORT.PDF", Asset},
		{"1 index.html", Asset},
		{"5 notes.md", Asset},
		{"6 readme.txt", Asset},
		{"01 file 0 {}.mp4", Asset},
		// GroupedAsset
		{"01 file 0 {1}.avi", GroupedAsset},
		{"01 file 0 {1 Part 1}.avi", GroupedAsset},
		{"01 file 0 {2 -   Part 2}.mp4", GroupedAsset},
		// Card
		{"card.jpg", Card},
		{"card.jpeg", Card},
		// Description
		{"01 description.md", Description},
		{"02 Description.TXT", Description},
		// Attachment
		{"01", Attachment},
		{"200.pdf", Attachment},
		{"1 -.txt", Attachment},
		{"1 - .txt", Attachment},
		{"1 .txt", Attachment},
		{"1     .pdf", Attachment},
		{"1 - file", Attachment},
		{"2 file", Attachment},
		{"3-file", Attachment},
		{"6 --- file", Attachment},
		{"1 - file.exe", Attachment},
	}

	for _, tt := range tests {
		parsed := parseFilename("xx", tt.in)
		category := categorizeFile(parsed)
		require.Equal(t, tt.expected, category, fmt.Sprintf("error for [%s]", tt.in))
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanner_IsCard(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		var tests = []string{
			"card",
			"1234",
			"1234.jpg",
			"jpg",
			"card.test.jpg",
			"card.txt",
		}

		for _, tt := range tests {
			require.False(t, isCard(tt))
		}
	})

	t.Run("valid", func(t *testing.T) {
		var tests = []string{
			"card.jpg",
			"card.jpeg",
			"card.png",
			"card.webp",
			"card.tiff",
		}

		for _, tt := range tests {
			require.True(t, isCard(tt))
		}
	})
}
