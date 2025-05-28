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

		options := &database.Options{
			OrderBy:          []string{models.ASSET_TABLE_CHAPTER + " asc", models.ASSET_TABLE_PREFIX + " asc"},
			Where:            squirrel.Eq{models.ASSET_TABLE_COURSE_ID: course.ID},
			ExcludeRelations: []string{models.ASSET_RELATION_PROGRESS},
		}

		assets := []*models.Asset{}

		// Add file 1, file 2 and file 3
		{
			scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 file 1.mkv", course.Path), []byte("hash 1"), os.ModePerm)
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/02 file 2.html", course.Path), []byte("hash 2"), os.ModePerm)
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/03 file 3.pdf", course.Path), []byte("hash 3"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, options)
			require.NoError(t, err)
			require.Len(t, assets, 3)

			require.Equal(t, "file 1", assets[0].Title)
			require.Equal(t, course.ID, assets[0].CourseID)
			require.Equal(t, 1, int(assets[0].Prefix.Int16))
			require.Empty(t, assets[0].Chapter)
			require.True(t, assets[0].Type.IsVideo())
			require.Equal(t, "0657190350cbea662b6c15d703d9c7482308e511504d3308306d0f1ede153a34", assets[0].Hash)

			require.Equal(t, "file 2", assets[1].Title)
			require.Equal(t, course.ID, assets[1].CourseID)
			require.Equal(t, 2, int(assets[1].Prefix.Int16))
			require.Empty(t, assets[1].Chapter)
			require.True(t, assets[1].Type.IsHTML())
			require.Equal(t, "ac4f5d7f5ca1f7b2a9e8107ca793b5ead43a1d04afdafabc9488e93b5d738b41", assets[1].Hash)

			require.Equal(t, "file 3", assets[2].Title)
			require.Equal(t, course.ID, assets[2].CourseID)
			require.Equal(t, 3, int(assets[2].Prefix.Int16))
			require.Empty(t, assets[2].Chapter)
			require.True(t, assets[2].Type.IsPDF())
			require.Equal(t, "c4ca2e438d8809f0e4459bde1f948de8fe6289f1c179d506da8720fb79859be6", assets[2].Hash)
		}

		// Add file 1 under a chapter
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 Chapter 1/01 file 1.pdf", course.Path), []byte("hash 4"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, options)
			require.NoError(t, err)
			require.Len(t, assets, 4)

			require.Equal(t, "file 1", assets[3].Title)
			require.Equal(t, course.ID, assets[3].CourseID)
			require.Equal(t, 1, int(assets[3].Prefix.Int16))
			require.Equal(t, "01 Chapter 1", assets[3].Chapter)
			require.True(t, assets[3].Type.IsPDF())
			require.Equal(t, "e72c82bb74988135e7b6c478fe3659a14b4941f867a93a23687ea172031e4e06", assets[3].Hash)
		}

		// Delete file 1 in chapter
		{
			scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 Chapter 1/01 file 1.pdf", course.Path))

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, options)
			require.NoError(t, err)
			require.Len(t, assets, 3)

			require.Equal(t, fmt.Sprintf("%s/01 file 1.mkv", course.Path), assets[0].Path)
			require.Equal(t, fmt.Sprintf("%s/02 file 2.html", course.Path), assets[1].Path)
			require.Equal(t, fmt.Sprintf("%s/03 file 3.pdf", course.Path), assets[2].Path)
		}

		// Rename file 3 to file 4
		{
			existingAssetID := assets[2].ID
			scanner.appFs.Fs.Rename(fmt.Sprintf("%s/03 file 3.pdf", course.Path), fmt.Sprintf("%s/04 file 4.pdf", course.Path))

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, options)
			require.NoError(t, err)
			require.Len(t, assets, 3)

			require.Equal(t, fmt.Sprintf("%s/01 file 1.mkv", course.Path), assets[0].Path)
			require.Equal(t, fmt.Sprintf("%s/02 file 2.html", course.Path), assets[1].Path)
			require.Equal(t, fmt.Sprintf("%s/04 file 4.pdf", course.Path), assets[2].Path)
			require.Equal(t, existingAssetID, assets[2].ID)
		}

		// Replace file 4 with new content
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/04 file 4.pdf", course.Path), []byte("hash 4"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, options)
			require.NoError(t, err)
			require.Len(t, assets, 3)

			require.Equal(t, fmt.Sprintf("%s/01 file 1.mkv", course.Path), assets[0].Path)
			require.Equal(t, fmt.Sprintf("%s/02 file 2.html", course.Path), assets[1].Path)
			require.Equal(t, fmt.Sprintf("%s/04 file 4.pdf", course.Path), assets[2].Path)

			require.Equal(t, "file 4", assets[2].Title)
			require.Equal(t, course.ID, assets[2].CourseID)
			require.Equal(t, 4, int(assets[2].Prefix.Int16))
			require.Empty(t, assets[2].Chapter)
			require.True(t, assets[2].Type.IsPDF())
			require.Equal(t, "e72c82bb74988135e7b6c478fe3659a14b4941f867a93a23687ea172031e4e06", assets[2].Hash)
		}

		// Swap file 1 and file 2
		{
			scanner.appFs.Fs.Rename(fmt.Sprintf("%s/01 file 1.mkv", course.Path), fmt.Sprintf("%s/02 file 2.html.temp", course.Path))
			scanner.appFs.Fs.Rename(fmt.Sprintf("%s/02 file 2.html", course.Path), fmt.Sprintf("%s/01 file 1.mkv", course.Path))
			scanner.appFs.Fs.Rename(fmt.Sprintf("%s/02 file 2.html.temp", course.Path), fmt.Sprintf("%s/02 file 2.html", course.Path))

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, options)
			require.NoError(t, err)
			require.Len(t, assets, 3)

			require.Equal(t, fmt.Sprintf("%s/01 file 1.mkv", course.Path), assets[0].Path)
			require.Equal(t, fmt.Sprintf("%s/02 file 2.html", course.Path), assets[1].Path)
			require.Equal(t, fmt.Sprintf("%s/04 file 4.pdf", course.Path), assets[2].Path)

			require.Equal(t, "ac4f5d7f5ca1f7b2a9e8107ca793b5ead43a1d04afdafabc9488e93b5d738b41", assets[0].Hash)
			require.Equal(t, "0657190350cbea662b6c15d703d9c7482308e511504d3308306d0f1ede153a34", assets[1].Hash)
		}

		// Overwrite: delete file 1 and move file 2 to file 1
		{
			require.NoError(t, scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 file 1.mkv", course.Path)))

			require.NoError(t, scanner.appFs.Fs.Rename(
				fmt.Sprintf("%s/02 file 2.html", course.Path),
				fmt.Sprintf("%s/01 file 1.mkv", course.Path),
			))

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, options)
			require.NoError(t, err)
			require.Len(t, assets, 2)

			require.Equal(t, fmt.Sprintf("%s/01 file 1.mkv", course.Path), assets[0].Path)
			require.Equal(t, fmt.Sprintf("%s/04 file 4.pdf", course.Path), assets[1].Path)

			require.Equal(t, "0657190350cbea662b6c15d703d9c7482308e511504d3308306d0f1ede153a34", assets[0].Hash)
		}
	})

	t.Run("attachments", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		assets := []*models.Asset{}
		attachments := []*models.Attachment{}

		assetOptions := &database.Options{
			OrderBy:          []string{models.ASSET_TABLE_CHAPTER + " asc", models.ASSET_TABLE_PREFIX + " asc"},
			Where:            squirrel.Eq{models.ASSET_TABLE_COURSE_ID: course.ID},
			ExcludeRelations: []string{models.ASSET_RELATION_PROGRESS},
		}

		// Add file 1
		{
			scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 file 1.mkv", course.Path), []byte("hash 1"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, assetOptions)
			require.NoError(t, err)
			require.Len(t, assets, 1)

			require.Equal(t, "file 1", assets[0].Title)
			require.Equal(t, course.ID, assets[0].CourseID)
			require.Equal(t, 1, int(assets[0].Prefix.Int16))
			require.Empty(t, assets[0].Chapter)
			require.True(t, assets[0].Type.IsVideo())
			require.Equal(t, "0657190350cbea662b6c15d703d9c7482308e511504d3308306d0f1ede153a34", assets[0].Hash)
		}

		attachmentOptions := &database.Options{
			OrderBy: []string{models.ATTACHMENT_TABLE_CREATED_AT + " asc"},
			Where:   squirrel.Eq{models.ATTACHMENT_TABLE_ASSET_ID: assets[0].ID},
		}

		// Add attachment 1
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 attachment 1.txt", course.Path), []byte("attachment 1"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAttachments(ctx, &attachments, attachmentOptions)
			require.NoError(t, err)
			require.Len(t, attachments, 1)

			require.Equal(t, "attachment 1.txt", attachments[0].Title)
			require.Equal(t, assets[0].ID, attachments[0].AssetID)
			require.Equal(t, filepath.Join(course.Path, "01 attachment 1.txt"), attachments[0].Path)
		}

		// Add another attachment
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 attachment 2.txt", course.Path), []byte("attachment 2"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAttachments(ctx, &attachments, attachmentOptions)
			require.NoError(t, err)
			require.Len(t, attachments, 2)

			require.Equal(t, "attachment 1.txt", attachments[0].Title)
			require.Equal(t, assets[0].ID, attachments[0].AssetID)
			require.Equal(t, filepath.Join(course.Path, "01 attachment 1.txt"), attachments[0].Path)

			require.Equal(t, "attachment 2.txt", attachments[1].Title)
			require.Equal(t, assets[0].ID, attachments[0].AssetID)
			require.Equal(t, filepath.Join(course.Path, "01 attachment 2.txt"), attachments[1].Path)
		}

		// Delete attachment
		{
			scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 attachment 1.txt", course.Path))

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAttachments(ctx, &attachments, attachmentOptions)
			require.NoError(t, err)
			require.Len(t, attachments, 1)

			require.Equal(t, "attachment 2.txt", attachments[0].Title)
			require.Equal(t, assets[0].ID, attachments[0].AssetID)
			require.Equal(t, filepath.Join(course.Path, "01 attachment 2.txt"), attachments[0].Path)
		}
	})

	t.Run("asset priority", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		// Priority is VIDEO -> HTML -> PDF

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		assetOptions := &database.Options{
			OrderBy:          []string{models.ASSET_TABLE_CHAPTER + " asc", models.ASSET_TABLE_PREFIX + " asc"},
			Where:            squirrel.Eq{models.ASSET_TABLE_COURSE_ID: course.ID},
			ExcludeRelations: []string{models.ASSET_RELATION_PROGRESS},
		}

		attachmentOptions := &database.Options{
			OrderBy: []string{models.ATTACHMENT_TABLE_CREATED_AT + " asc"},
		}

		assets := []*models.Asset{}
		attachments := []*models.Attachment{}

		// Add PDF asset
		{
			scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 doc 1.pdf", course.Path), []byte("doc 1"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, assetOptions)
			require.NoError(t, err)
			require.Len(t, assets, 1)

			require.Equal(t, "doc 1", assets[0].Title)
			require.Equal(t, course.ID, assets[0].CourseID)
			require.Equal(t, 1, int(assets[0].Prefix.Int16))
			require.Empty(t, assets[0].Chapter)
			require.True(t, assets[0].Type.IsPDF())
			require.Equal(t, "a41a06f389aa3855fa07fa764b96cac08ff558978f10d9bb027299a85a6677c6", assets[0].Hash)
			require.Len(t, assets[0].Attachments, 0)
		}

		// Add HTML asset
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 index.html", course.Path), []byte("index"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, assetOptions)
			require.NoError(t, err)
			require.Len(t, assets, 1)

			require.Equal(t, "index", assets[0].Title)
			require.Equal(t, course.ID, assets[0].CourseID)
			require.Equal(t, 1, int(assets[0].Prefix.Int16))
			require.Empty(t, assets[0].Chapter)
			require.True(t, assets[0].Type.IsHTML())
			require.Equal(t, "1bc04b5291c26a46d918139138b992d2de976d6851d0893b0476b85bfbdfc6e6", assets[0].Hash)
			require.Len(t, assets[0].Attachments, 1)

			err = scanner.dao.ListAttachments(ctx, &attachments, attachmentOptions)
			require.NoError(t, err)
			require.Len(t, attachments, 1)
			require.Equal(t, filepath.Join(course.Path, "01 doc 1.pdf"), attachments[0].Path)
		}

		// Add VIDEO asset
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 video.mp4", course.Path), []byte("video"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, assetOptions)
			require.NoError(t, err)
			require.Len(t, assets, 1)

			require.Equal(t, "video", assets[0].Title)
			require.Equal(t, course.ID, assets[0].CourseID)
			require.Equal(t, 1, int(assets[0].Prefix.Int16))
			require.Empty(t, assets[0].Chapter)
			require.True(t, assets[0].Type.IsVideo())
			require.Equal(t, "0cab1c9617404faf2b24e221e189ca5945813e14d3f766345b09ca13bbe28ffc", assets[0].Hash)
			require.Len(t, assets[0].Attachments, 2)

			err = scanner.dao.ListAttachments(ctx, &attachments, attachmentOptions)
			require.NoError(t, err)
			require.Len(t, attachments, 2)
			require.Equal(t, filepath.Join(course.Path, "01 doc 1.pdf"), attachments[0].Path)
			require.Equal(t, filepath.Join(course.Path, "01 index.html"), attachments[1].Path)
		}

		// Add another PDF asset
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 doc 2.pdf", course.Path), []byte("doc 2"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, assetOptions)
			require.NoError(t, err)
			require.Len(t, assets, 1)

			require.Equal(t, "video", assets[0].Title)
			require.Equal(t, course.ID, assets[0].CourseID)
			require.Equal(t, 1, int(assets[0].Prefix.Int16))
			require.Empty(t, assets[0].Chapter)
			require.True(t, assets[0].Type.IsVideo())
			require.Equal(t, "0cab1c9617404faf2b24e221e189ca5945813e14d3f766345b09ca13bbe28ffc", assets[0].Hash)
			require.Len(t, assets[0].Attachments, 3)

			err = scanner.dao.ListAttachments(ctx, &attachments, attachmentOptions)
			require.NoError(t, err)
			require.Len(t, attachments, 3)
			require.Equal(t, filepath.Join(course.Path, "01 doc 1.pdf"), attachments[0].Path)
			require.Equal(t, filepath.Join(course.Path, "01 index.html"), attachments[1].Path)
			require.Equal(t, filepath.Join(course.Path, "01 doc 2.pdf"), attachments[2].Path)
		}
	})

	t.Run("asset with sub-prefix and sub-title", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID, Status: types.NewScanStatusWaiting()}
		require.NoError(t, scanner.dao.CreateScan(ctx, scan))

		assetOptions := &database.Options{
			OrderBy:          []string{models.ASSET_TABLE_CHAPTER + " asc", models.ASSET_TABLE_PREFIX + " asc"},
			Where:            squirrel.Eq{models.ASSET_TABLE_COURSE_ID: course.ID},
			ExcludeRelations: []string{models.ASSET_RELATION_PROGRESS},
		}

		assets := []*models.Asset{}

		// Add video 1 asset with sub-prefix of 1 and sub-title "Part 1"
		{
			scanner.appFs.Fs.Mkdir(course.Path, os.ModePerm)
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 video 1 {1 Part 1}.mp4", course.Path), []byte("video 1"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, assetOptions)
			require.NoError(t, err)
			require.Len(t, assets, 1)

			require.Equal(t, "video 1", assets[0].Title)
			require.Equal(t, course.ID, assets[0].CourseID)
			require.Equal(t, 1, int(assets[0].Prefix.Int16))
			require.Equal(t, 1, int(assets[0].SubPrefix.Int16))
			require.Equal(t, "Part 1", assets[0].SubTitle)
			require.Empty(t, assets[0].Chapter)
			require.True(t, assets[0].Type.IsVideo())
			require.Equal(t, "3b857b8441d7c9e734535d6b82f69a34c6fcd63ed0ef989ff03808ecb29a2f1f", assets[0].Hash)
			require.Len(t, assets[0].Attachments, 0)
		}

		// Add video 2 asset with sub-prefix of 2 and sub-title "Part 2"
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 video 2 {2 Part 2}.mp4", course.Path), []byte("video 2"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, assetOptions)
			require.NoError(t, err)
			require.Len(t, assets, 2)

			require.Equal(t, "video 1", assets[0].Title)

			require.Equal(t, "video 2", assets[1].Title)
			require.Equal(t, course.ID, assets[1].CourseID)
			require.Equal(t, 1, int(assets[1].Prefix.Int16))
			require.Equal(t, 2, int(assets[1].SubPrefix.Int16))
			require.Equal(t, "Part 2", assets[1].SubTitle)
			require.Empty(t, assets[1].Chapter)
			require.True(t, assets[1].Type.IsVideo())
			require.Equal(t, "614ef49d4a1ef39bc763b7c9665f6f30a0eea3ec5ec10e04b897bdad9b973f9c", assets[1].Hash)
			require.Len(t, assets[1].Attachments, 0)
		}

		// Add video 3 asset with sub-prefix of 3 and no sub-title
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 video 3 {3}.mp4", course.Path), []byte("video 3"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, assetOptions)
			require.NoError(t, err)
			require.Len(t, assets, 3)

			require.Equal(t, "video 1", assets[0].Title)
			require.Equal(t, "video 2", assets[1].Title)

			require.Equal(t, "video 3", assets[2].Title)
			require.Equal(t, course.ID, assets[2].CourseID)
			require.Equal(t, 1, int(assets[2].Prefix.Int16))
			require.Equal(t, 3, int(assets[2].SubPrefix.Int16))
			require.Empty(t, assets[2].SubTitle)
			require.Empty(t, assets[2].Chapter)
			require.True(t, assets[2].Type.IsVideo())
			require.Equal(t, "36d9fa5c21ca58822f678a5e1cebbaefcbcff37894771089cc608e8fbe32121e", assets[2].Hash)
			require.Len(t, assets[2].Attachments, 0)
		}

		// Add attachment (should get attached to asset 1)
		{
			afero.WriteFile(scanner.appFs.Fs, fmt.Sprintf("%s/01 attachment 1.txt", course.Path), []byte("attachment 1"), os.ModePerm)

			err := Processor(ctx, scanner, scan)
			require.NoError(t, err)

			err = scanner.dao.ListAssets(ctx, &assets, assetOptions)
			require.NoError(t, err)
			require.Len(t, assets, 3)

			require.Equal(t, "video 1", assets[0].Title)
			require.Equal(t, "3b857b8441d7c9e734535d6b82f69a34c6fcd63ed0ef989ff03808ecb29a2f1f", assets[0].Hash)

			require.Len(t, assets[0].Attachments, 1)
			require.Equal(t, filepath.Join(course.Path, "01 attachment 1.txt"), assets[0].Attachments[0].Path)
			require.Equal(t, "attachment 1.txt", assets[0].Attachments[0].Title)
			require.Equal(t, assets[0].ID, assets[0].Attachments[0].AssetID)

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
			fb := parseFilename(tt)
			require.Nil(t, fb)
		}
	})

	t.Run("assets", func(t *testing.T) {
		var tests = []struct {
			in       string
			expected *parsedFilename
		}{
			// Video (with varied filenames)
			{"0    file 0.avi", &parsedFilename{prefix: 0, title: "file 0", subPrefix: nil, subTitle: "", asset: types.NewAsset("avi")}},
			{"001 file 1.mp4", &parsedFilename{prefix: 1, title: "file 1", subPrefix: nil, subTitle: "", asset: types.NewAsset("mp4")}},
			{"1-file.ogg", &parsedFilename{prefix: 1, title: "file", subPrefix: nil, subTitle: "", asset: types.NewAsset("ogg")}},
			{"2 - file.webm", &parsedFilename{prefix: 2, title: "file", subPrefix: nil, subTitle: "", asset: types.NewAsset("webm")}},
			{"3 -file.m4a", &parsedFilename{prefix: 3, title: "file", subPrefix: nil, subTitle: "", asset: types.NewAsset("m4a")}},
			{"4- file.opus", &parsedFilename{prefix: 4, title: "file", subPrefix: nil, subTitle: "", asset: types.NewAsset("opus")}},
			{"5000 --- file.wav", &parsedFilename{prefix: 5000, title: "file", subPrefix: nil, subTitle: "", asset: types.NewAsset("wav")}},
			{"0100 file.mp3", &parsedFilename{prefix: 100, title: "file", subPrefix: nil, subTitle: "", asset: types.NewAsset("mp3")}},
			// PDF
			{"1 - doc.pdf", &parsedFilename{prefix: 1, title: "doc", subPrefix: nil, subTitle: "", asset: types.NewAsset("pdf")}},
			// HTML
			{"1 index.html", &parsedFilename{prefix: 1, title: "index", subPrefix: nil, subTitle: "", asset: types.NewAsset("html")}},
			// With sub-prefix
			{"01 file 0 {1}.avi", &parsedFilename{prefix: 1, title: "file 0", subPrefix: intPtr(1), subTitle: "", asset: types.NewAsset("avi")}},
			{"01 file 0 {2}.mp4", &parsedFilename{prefix: 1, title: "file 0", subPrefix: intPtr(2), subTitle: "", asset: types.NewAsset("mp4")}},
			// // With sub-prefix and sub-title
			{"01 file 0 {1 Part 1}.avi", &parsedFilename{prefix: 1, title: "file 0", subPrefix: intPtr(1), subTitle: "Part 1", asset: types.NewAsset("avi")}},
			{"01 file 0 {2 Part 2}.mp4", &parsedFilename{prefix: 1, title: "file 0", subPrefix: intPtr(2), subTitle: "Part 2", asset: types.NewAsset("mp4")}},
			{"01 file 0 {3 -   Part 3}.mp4", &parsedFilename{prefix: 1, title: "file 0", subPrefix: intPtr(3), subTitle: "Part 3", asset: types.NewAsset("mp4")}},
			{"01 file 0 {1 - }.mp4", &parsedFilename{prefix: 1, title: "file 0", subPrefix: intPtr(1), subTitle: "", asset: types.NewAsset("mp4")}},
			{"01 file 0 {1 - }.mp4", &parsedFilename{prefix: 1, title: "file 0", subPrefix: intPtr(1), subTitle: "", asset: types.NewAsset("mp4")}},
			{"01 file 0 {1 --- Part 1}.mp4", &parsedFilename{prefix: 1, title: "file 0", subPrefix: intPtr(1), subTitle: "Part 1", asset: types.NewAsset("mp4")}},
			{"01 file 0 {}.mp4", &parsedFilename{prefix: 1, title: "file 0", subPrefix: nil, subTitle: "", asset: types.NewAsset("mp4")}},
		}

		for _, tt := range tests {
			fb := parseFilename(tt.in)
			require.Equal(t, tt.expected, fb, fmt.Sprintf("error for [%s]", tt.in))
		}
	})

	t.Run("attachments", func(t *testing.T) {
		var tests = []struct {
			in       string
			expected *parsedFilename
		}{
			// No title
			{"01", &parsedFilename{prefix: 1, title: "01"}},
			{"200.pdf", &parsedFilename{prefix: 200, title: "200.pdf"}},
			{"1 -.txt", &parsedFilename{prefix: 1, title: "1 -.txt"}},
			{"1 - .txt", &parsedFilename{prefix: 1, title: "1 - .txt"}},
			{"1 .txt", &parsedFilename{prefix: 1, title: "1 .txt"}},
			{"1     .pdf", &parsedFilename{prefix: 1, title: "1     .pdf"}},
			// No extension (fileName should have no prefix)
			{"0    file 0", &parsedFilename{prefix: 0, title: "file 0"}},
			{"001    file 1", &parsedFilename{prefix: 1, title: "file 1"}},
			{"1001 - file", &parsedFilename{prefix: 1001, title: "file"}},
			{"0123-file", &parsedFilename{prefix: 123, title: "file"}},
			{"1 --- file", &parsedFilename{prefix: 1, title: "file"}},
			// Non-asset extension (fileName should have no prefix)
			{"1 file.txt", &parsedFilename{prefix: 1, title: "file.txt"}},
		}

		for _, tt := range tests {
			fb := parseFilename(tt.in)
			require.Equal(t, tt.expected, fb, fmt.Sprintf("error for [%s]", tt.in))
		}
	})
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
