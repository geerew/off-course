package jobs

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Add(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, false, 0, 0)

		scan, err := scanner.Add(workingData[0].ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].ID, scan.CourseID)
	})

	t.Run("duplicate", func(t *testing.T) {
		scanner, db, lh := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, false, 0, 0)

		scan, err := scanner.Add(workingData[0].ID)
		assert.Nil(t, err)
		assert.Equal(t, workingData[0].ID, scan.CourseID)

		// Add the same course again
		scan, err = scanner.Add(workingData[0].ID)
		require.Nil(t, err)
		require.NotNil(t, lh.LastEntry())
		require.Nil(t, scan)
		lh.LastEntry().ExpMsg("scan already in progress")
		lh.LastEntry().ExpLevel(zerolog.DebugLevel)
	})

	t.Run("invalid course", func(t *testing.T) {
		scanner, _, _ := setupCourseScanner(t)

		scan, err := scanner.Add("test")
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, scan)
	})

	t.Run("not blocked", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 2, false, 0, 0)

		scan1, err := scanner.Add(workingData[0].ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].ID, scan1.CourseID)

		scan2, err := scanner.Add(workingData[1].ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[1].ID, scan2.CourseID)
	})

	t.Run("db error", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, false, 0, 0)

		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.TableScans())
		require.Nil(t, err)

		scan, err := scanner.Add(workingData[0].ID)
		require.ErrorContains(t, err, fmt.Sprintf("no such table: %s", daos.TableScans()))
		assert.Nil(t, scan)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Worker(t *testing.T) {
	t.Run("single job", func(t *testing.T) {
		scanner, db, lh := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, false, 0, 0)

		// Start the worker
		go scanner.Worker(func(*CourseScanner, *models.Scan) error {
			return nil
		})

		// Add the job
		scan, err := scanner.Add(workingData[0].ID)
		require.Nil(t, err)
		assert.Equal(t, scan.CourseID, workingData[0].ID)

		// Give time for the worker to finish
		time.Sleep(time.Millisecond * 3)

		// Assert the scan job was deleted from the DB
		s, err := scanner.scanDao.Get(workingData[0].ID)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, s)

		// Validate the logs
		require.NotNil(t, lh.LastEntry())
		lh.Entries().Get()[lh.Len()-2].ExpMsg("finished processing scan job")
		lh.Entries().Get()[lh.Len()-2].ExpStr("path", workingData[0].Path)
		lh.LastEntry().ExpLevel(zerolog.InfoLevel)
		lh.LastEntry().ExpMsg("finished processing all scan jobs")
	})

	t.Run("several jobs", func(t *testing.T) {
		scanner, db, lh := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 3, false, 0, 0)

		for _, course := range workingData {
			_, err := scanner.Add(course.ID)
			require.Nil(t, err)
		}

		// Start the worker
		go scanner.Worker(func(*CourseScanner, *models.Scan) error {
			return nil
		})

		// Give time for the worker to finish
		time.Sleep(time.Millisecond * 5)

		// Assert the scan job was deleted from the DB
		for _, course := range workingData {
			s, err := scanner.scanDao.Get(course.ID)
			require.ErrorIs(t, err, sql.ErrNoRows)
			assert.Nil(t, s)
		}

		// Validate the logs
		require.NotNil(t, lh.LastEntry())
		lh.Entries().Get()[lh.Len()-2].ExpMsg("finished processing scan job")
		lh.Entries().Get()[lh.Len()-2].ExpStr("path", workingData[2].Path)
		lh.LastEntry().ExpLevel(zerolog.InfoLevel)
		lh.LastEntry().ExpMsg("finished processing all scan jobs")
	})

	t.Run("error processing", func(t *testing.T) {
		scanner, db, lh := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, false, 0, 0)

		// Start the worker
		go scanner.Worker(func(*CourseScanner, *models.Scan) error {
			return errors.New("processing error")
		})

		scan, err := scanner.Add(workingData[0].ID)
		require.Nil(t, err)
		assert.Equal(t, scan.CourseID, workingData[0].ID)

		// Give time for the worker to finish
		time.Sleep(time.Millisecond * 3)

		// Validate the logs
		require.NotNil(t, lh.LastEntry())
		lh.Entries().Get()[lh.Len()-2].ExpMsg("error processing scan job")
		lh.Entries().Get()[lh.Len()-2].ExpErr(errors.New("processing error"))
		lh.Entries().Get()[lh.Len()-2].ExpLevel(zerolog.ErrorLevel)
		lh.LastEntry().ExpLevel(zerolog.InfoLevel)
		lh.LastEntry().ExpMsg("finished processing all scan jobs")
	})

	t.Run("scan error", func(t *testing.T) {
		scanner, db, lh := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, false, 0, 0)

		// Add the job
		scan, err := scanner.Add(workingData[0].ID)
		require.Nil(t, err)
		assert.Equal(t, scan.CourseID, workingData[0].ID)

		// Drop the DB
		_, err = db.Exec("DROP TABLE IF EXISTS " + daos.TableScans())
		require.Nil(t, err)

		// Start the worker
		go scanner.Worker(func(*CourseScanner, *models.Scan) error {
			return nil
		})

		// Give time for the worker to finish
		time.Sleep(time.Millisecond * 3)

		// Validate the logs
		require.NotNil(t, lh.LastEntry())
		lh.LastEntry().ExpMsg("error looking up next scan job")
		lh.LastEntry().ExpLevel(zerolog.ErrorLevel)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CourseProcessor(t *testing.T) {
	t.Run("scan nil", func(t *testing.T) {
		scanner, _, _ := setupCourseScanner(t)

		err := CourseProcessor(scanner, nil)
		require.EqualError(t, err, "scan cannot be empty")
	})

	t.Run("error getting course", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 0, 0)

		// Drop the table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.TableCourses())
		require.Nil(t, err)

		err = CourseProcessor(scanner, workingData[0].Scan)
		require.ErrorContains(t, err, fmt.Sprintf("no such table: %s", daos.TableCourses()))
	})

	t.Run("course unavailable", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 0, 0)

		// Mark the course as available
		workingData[0].Available = true
		err := scanner.courseDao.Update(workingData[0].Course)
		require.Nil(t, err)

		err = CourseProcessor(scanner, workingData[0].Scan)
		require.EqualError(t, err, "course unavailable")
	})

	t.Run("card", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		// ----------------------------
		// Found card
		// ----------------------------
		workingData := daos.NewTestData(t, db, 1, true, 0, 0)
		require.Empty(t, workingData[0].CardPath)

		scanner.appFs.Fs.Mkdir(workingData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/card.jpg", workingData[0].Path))

		err := CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		c, err := scanner.courseDao.Get(workingData[0].ID)
		require.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("%s/card.jpg", workingData[0].Path), c.CardPath)

		// ----------------------------
		// Ignore card in chapter
		// ----------------------------
		workingData = daos.NewTestData(t, db, 1, true, 0, 0)
		require.Empty(t, workingData[0].CardPath)

		scanner.appFs.Fs.Mkdir(workingData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/chapter 1/card.jpg", workingData[0].Path))

		err = CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		c, err = scanner.courseDao.Get(workingData[0].ID)
		require.Nil(t, err)
		assert.Empty(t, c.CardPath)
		assert.Empty(t, workingData[0].CardPath)

		// ----------------------------
		// Multiple cards types
		// ----------------------------
		workingData = daos.NewTestData(t, db, 1, true, 0, 0)
		require.Empty(t, workingData[0].CardPath)

		scanner.appFs.Fs.Mkdir(workingData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/card.jpg", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/card.png", workingData[0].Path))

		err = CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		c, err = scanner.courseDao.Get(workingData[0].ID)
		require.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("%s/card.jpg", workingData[0].Path), c.CardPath)
	})

	t.Run("card error", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 0, 0)
		require.Empty(t, workingData[0].CardPath)

		scanner.appFs.Fs.Mkdir(workingData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/card.jpg", workingData[0].Path))

		// Rename the card_path column
		_, err := db.Exec(fmt.Sprintf("ALTER TABLE %s RENAME COLUMN card_path TO ignore_card_path", daos.TableCourses()))
		require.Nil(t, err)

		err = CourseProcessor(scanner, workingData[0].Scan)
		require.ErrorContains(t, err, "no such column: card_path")
	})

	t.Run("ignore files", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 0, 0)

		scanner.appFs.Fs.Mkdir(workingData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/file 1", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/file.file", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/file.avi", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/ - file.avi", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/- - file.avi", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/.avi", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/-1 - file.avi", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/a - file.avi", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/1.1 - file.avi", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/2.3-file.avi", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/1file.avi", workingData[0].Path))

		err := CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		assets, err := scanner.assetDao.List(&database.DatabaseParams{Where: squirrel.Eq{daos.TableAssets() + ".course_id": workingData[0].ID}})
		require.Nil(t, err)
		require.Len(t, assets, 0)
	})

	t.Run("assets", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 0, 0)

		dbParams := &database.DatabaseParams{
			OrderBy: []string{daos.TableAssets() + ".chapter asc", daos.TableAssets() + ".prefix asc"},
			Where:   squirrel.Eq{daos.TableAssets() + ".course_id": workingData[0].ID},
		}

		// ----------------------------
		// Add 2 assets
		// ----------------------------
		scanner.appFs.Fs.Mkdir(workingData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 file 1.mkv", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/02 file 2.html", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/should ignore", workingData[0].Path))

		err := CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		assets, err := scanner.assetDao.List(dbParams)
		require.Nil(t, err)
		require.Len(t, assets, 2)

		assert.Equal(t, "file 1", assets[0].Title)
		assert.Equal(t, workingData[0].ID, assets[0].CourseID)
		assert.Equal(t, 1, int(assets[0].Prefix.Int16))
		assert.Empty(t, assets[0].Chapter)
		assert.True(t, assets[0].Type.IsVideo())

		assert.Equal(t, "file 2", assets[1].Title)
		assert.Equal(t, workingData[0].ID, assets[1].CourseID)
		assert.Equal(t, 2, int(assets[1].Prefix.Int16))
		assert.Empty(t, assets[1].Chapter)
		assert.True(t, assets[1].Type.IsHTML())

		// ----------------------------
		// Delete asset
		// ----------------------------
		scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 file 1.mkv", workingData[0].Path))

		err = CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		assets, err = scanner.assetDao.List(dbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)

		assert.Equal(t, "file 2", assets[0].Title)
		assert.Equal(t, workingData[0].ID, assets[0].CourseID)
		assert.Equal(t, 2, int(assets[0].Prefix.Int16))
		assert.Empty(t, assets[0].Chapter)
		assert.True(t, assets[0].Type.IsHTML())

		// ----------------------------
		// Add chapter asset
		// ----------------------------
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 Chapter 1/03 file 3.pdf", workingData[0].Path))

		err = CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		assets, err = scanner.assetDao.List(dbParams)
		require.Nil(t, err)
		require.Len(t, assets, 2)

		assert.Equal(t, "file 2", assets[0].Title)
		assert.Equal(t, workingData[0].ID, assets[0].CourseID)
		assert.Equal(t, 2, int(assets[0].Prefix.Int16))
		assert.Empty(t, assets[0].Chapter)
		assert.True(t, assets[0].Type.IsHTML())

		assert.Equal(t, "file 3", assets[1].Title)
		assert.Equal(t, workingData[0].ID, assets[1].CourseID)
		assert.Equal(t, 3, int(assets[1].Prefix.Int16))
		assert.Equal(t, "01 Chapter 1", assets[1].Chapter)
		assert.True(t, assets[1].Type.IsPDF())
	})

	t.Run("assets error", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 0, 0)

		scanner.appFs.Fs.Mkdir(workingData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 video.mkv", workingData[0].Path))

		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.TableAssets())
		require.Nil(t, err)

		err = CourseProcessor(scanner, workingData[0].Scan)
		require.ErrorContains(t, err, "no such table: "+daos.TableAssets())
	})

	t.Run("attachments", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 0, 0)

		scanner.appFs.Fs.Mkdir(workingData[0].Path, os.ModePerm)

		assetDbParams := &database.DatabaseParams{Where: squirrel.Eq{daos.TableAssets() + ".course_id": workingData[0].ID}}
		attachmentDbParams := &database.DatabaseParams{
			OrderBy: []string{"created_at asc"},
			Where:   squirrel.Eq{daos.TableAttachments() + ".course_id": workingData[0].ID},
		}

		// ----------------------------
		// Add 1 asset with 1 attachment
		// ----------------------------
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 video.mp4", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 info.txt", workingData[0].Path))

		err := CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		assets, err := scanner.assetDao.List(assetDbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, "video", assets[0].Title)
		assert.Equal(t, workingData[0].ID, assets[0].CourseID)
		assert.Equal(t, 1, int(assets[0].Prefix.Int16))
		assert.Equal(t, fmt.Sprintf("%s/01 video.mp4", workingData[0].Path), assets[0].Path)
		assert.True(t, assets[0].Type.IsVideo())

		attachments, err := scanner.attachmentDao.List(attachmentDbParams)
		require.Nil(t, err)
		require.Len(t, attachments, 1)
		assert.Equal(t, "info.txt", attachments[0].Title)
		assert.Equal(t, fmt.Sprintf("%s/01 info.txt", workingData[0].Path), attachments[0].Path)

		// ----------------------------
		// Add another attachment
		// ----------------------------
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 code.zip", workingData[0].Path))

		err = CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		// Assert the asset
		assets, err = scanner.assetDao.List(assetDbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		require.Len(t, assets[0].Attachments, 2)

		attachments, err = scanner.attachmentDao.List(attachmentDbParams)
		require.Nil(t, err)
		require.Len(t, attachments, 2)
		assert.Equal(t, "info.txt", attachments[0].Title)
		assert.Equal(t, fmt.Sprintf("%s/01 info.txt", workingData[0].Path), attachments[0].Path)
		assert.Equal(t, "code.zip", attachments[1].Title)
		assert.Equal(t, fmt.Sprintf("%s/01 code.zip", workingData[0].Path), attachments[1].Path)

		// ----------------------------
		// Delete first attachment
		// ----------------------------
		scanner.appFs.Fs.Remove(fmt.Sprintf("%s/01 info.txt", workingData[0].Path))

		err = CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		// Assert the asset
		assets, err = scanner.assetDao.List(assetDbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, "video", assets[0].Title)
		require.Len(t, assets[0].Attachments, 1)

		attachments, err = scanner.attachmentDao.List(attachmentDbParams)
		require.Nil(t, err)
		require.Len(t, attachments, 1)
		assert.Equal(t, "code.zip", attachments[0].Title)
		assert.Equal(t, fmt.Sprintf("%s/01 code.zip", workingData[0].Path), attachments[0].Path)
	})

	t.Run("attachments error", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 0, 0)

		scanner.appFs.Fs.Mkdir(workingData[0].Path, os.ModePerm)
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 video.mkv", workingData[0].Path))
		scanner.appFs.Fs.Create(fmt.Sprintf("%s/01 info", workingData[0].Path))

		// Drop the attachments table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.TableAttachments())
		require.Nil(t, err)

		err = CourseProcessor(scanner, workingData[0].Scan)
		require.ErrorContains(t, err, "no such table: "+daos.TableAttachments())
	})

	t.Run("asset priority", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		// ----------------------------
		// Priority is VIDEO -> HTML -> PDF
		// ----------------------------

		workingData := daos.NewTestData(t, db, 1, true, 0, 0)

		scanner.appFs.Fs.Mkdir(workingData[0].Path, os.ModePerm)

		assetDbParams := &database.DatabaseParams{Where: squirrel.Eq{daos.TableAssets() + ".course_id": workingData[0].ID}}
		attachmentDbParams := &database.DatabaseParams{
			OrderBy: []string{"created_at asc"},
			Where:   squirrel.Eq{daos.TableAttachments() + ".course_id": workingData[0].ID},
		}

		// ----------------------------
		// Add PDF (asset)
		// ----------------------------
		pdfFile := fmt.Sprintf("%s/01 doc 1.pdf", workingData[0].Path)
		scanner.appFs.Fs.Create(pdfFile)

		err := CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		assets, err := scanner.assetDao.List(assetDbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, pdfFile, assets[0].Path)
		assert.True(t, assets[0].Type.IsPDF())
		require.Empty(t, assets[0].Attachments)

		// ----------------------------
		// Add HTML (asset)
		// ----------------------------
		htmlFile := fmt.Sprintf("%s/01 example.html", workingData[0].Path)
		scanner.appFs.Fs.Create(htmlFile)

		err = CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		assets, err = scanner.assetDao.List(assetDbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, htmlFile, assets[0].Path)
		assert.True(t, assets[0].Type.IsHTML())
		require.Len(t, assets[0].Attachments, 1)

		attachments, err := scanner.attachmentDao.List(attachmentDbParams)
		require.Nil(t, err)
		require.Len(t, attachments, 1)
		assert.Equal(t, pdfFile, attachments[0].Path)

		// ----------------------------
		// Add VIDEO (asset)
		// ----------------------------
		videoFile := fmt.Sprintf("%s/01 video.mp4", workingData[0].Path)
		scanner.appFs.Fs.Create(videoFile)

		err = CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		assets, err = scanner.assetDao.List(assetDbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, videoFile, assets[0].Path)
		assert.True(t, assets[0].Type.IsVideo())
		require.Len(t, assets[0].Attachments, 2)

		attachments, err = scanner.attachmentDao.List(attachmentDbParams)
		require.Nil(t, err)
		require.Len(t, attachments, 2)
		assert.Equal(t, pdfFile, attachments[0].Path)
		assert.Equal(t, htmlFile, attachments[1].Path)

		// ----------------------------
		// Add PDF file (attachment)
		// ----------------------------
		pdfFile2 := fmt.Sprintf("%s/01 - e.pdf", workingData[0].Path)
		scanner.appFs.Fs.Create(pdfFile2)

		err = CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		assets, err = scanner.assetDao.List(assetDbParams)
		require.Nil(t, err)
		require.Len(t, assets, 1)
		assert.Equal(t, videoFile, assets[0].Path)
		assert.True(t, assets[0].Type.IsVideo())
		require.Len(t, assets[0].Attachments, 3)

		attachments, err = scanner.attachmentDao.List(attachmentDbParams)
		require.Nil(t, err)
		require.Len(t, attachments, 3)
		assert.Equal(t, pdfFile, attachments[0].Path)
		assert.Equal(t, htmlFile, attachments[1].Path)
		assert.Equal(t, pdfFile2, attachments[2].Path)
	})

	t.Run("course updated", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 1, 10)

		scanner.appFs.Fs.Mkdir(workingData[0].Path, os.ModePerm)

		err := CourseProcessor(scanner, workingData[0].Scan)
		require.Nil(t, err)

		updatedCourse, err := scanner.courseDao.Get(workingData[0].ID)
		require.Nil(t, err)
		assert.NotEqual(t, workingData[0].UpdatedAt, updatedCourse.UpdatedAt)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_BuildFileInfo(t *testing.T) {
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
			fb := parseFileName(tt)
			assert.Nil(t, fb)
		}
	})

	t.Run("assets", func(t *testing.T) {
		var tests = []struct {
			in       string
			expected *parsedFileName
		}{
			// Video (with varied filenames)
			{"0    file 0.avi", &parsedFileName{prefix: 0, title: "file 0", ext: "avi", attachmentTitle: "file 0.avi", asset: types.NewAsset("avi")}},
			{"001 file 1.mp4", &parsedFileName{prefix: 1, title: "file 1", ext: "mp4", attachmentTitle: "file 1.mp4", asset: types.NewAsset("mp4")}},
			{"1-file.ogg", &parsedFileName{prefix: 1, title: "file", ext: "ogg", attachmentTitle: "file.ogg", asset: types.NewAsset("ogg")}},
			{"2 - file.webm", &parsedFileName{prefix: 2, title: "file", ext: "webm", attachmentTitle: "file.webm", asset: types.NewAsset("webm")}},
			{"3 -file.m4a", &parsedFileName{prefix: 3, title: "file", ext: "m4a", attachmentTitle: "file.m4a", asset: types.NewAsset("m4a")}},
			{"4- file.opus", &parsedFileName{prefix: 4, title: "file", ext: "opus", attachmentTitle: "file.opus", asset: types.NewAsset("opus")}},
			{"5000 --- file.wav", &parsedFileName{prefix: 5000, title: "file", ext: "wav", attachmentTitle: "file.wav", asset: types.NewAsset("wav")}},
			{"0100 file.mp3", &parsedFileName{prefix: 100, title: "file", ext: "mp3", attachmentTitle: "file.mp3", asset: types.NewAsset("mp3")}},
			// PDF
			{"1 - doc.pdf", &parsedFileName{prefix: 1, title: "doc", ext: "pdf", attachmentTitle: "doc.pdf", asset: types.NewAsset("pdf")}},
			// HTML
			{"1 index.html", &parsedFileName{prefix: 1, title: "index", ext: "html", attachmentTitle: "index.html", asset: types.NewAsset("html")}},
		}

		for _, tt := range tests {
			fb := parseFileName(tt.in)
			assert.Equal(t, tt.expected, fb, fmt.Sprintf("error for [%s]", tt.in))
		}
	})

	t.Run("attachments", func(t *testing.T) {
		var tests = []struct {
			in       string
			expected *parsedFileName
		}{
			// No title
			{"01", &parsedFileName{prefix: 1, title: "", attachmentTitle: "01"}},
			{"200.pdf", &parsedFileName{prefix: 200, title: "", attachmentTitle: "200.pdf"}},
			{"1 -.txt", &parsedFileName{prefix: 1, title: "", attachmentTitle: "1 -.txt"}},
			{"1 .txt", &parsedFileName{prefix: 1, title: "", attachmentTitle: "1 .txt"}},
			{"1     .pdf", &parsedFileName{prefix: 1, title: "", attachmentTitle: "1     .pdf"}},
			// No extension (fileName should have no prefix)
			{"0    file 0", &parsedFileName{prefix: 0, title: "file 0", attachmentTitle: "file 0"}},
			{"001    file 1", &parsedFileName{prefix: 1, title: "file 1", attachmentTitle: "file 1"}},
			{"1001 - file", &parsedFileName{prefix: 1001, title: "file", attachmentTitle: "file"}},
			{"0123-file", &parsedFileName{prefix: 123, title: "file", attachmentTitle: "file"}},
			{"1 --- file", &parsedFileName{prefix: 1, title: "file", attachmentTitle: "file"}},
			// Non-asset extension (fileName should have no prefix)
			{"1 file.txt", &parsedFileName{prefix: 1, title: "file", ext: "txt", attachmentTitle: "file.txt"}},
		}

		for _, tt := range tests {
			fb := parseFileName(tt.in)
			assert.Equal(t, tt.expected, fb, fmt.Sprintf("error for [%s]", tt.in))
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_IsCard(t *testing.T) {
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
			assert.False(t, isCard(tt))
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
			assert.True(t, isCard(tt))
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateAssets(t *testing.T) {
	t.Run("nothing added or deleted", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 10, 0)

		err := updateAssets(scanner.assetDao, workingData[0].ID, workingData[0].Assets)
		require.Nil(t, err)

		dbParams := &database.DatabaseParams{Where: squirrel.Eq{daos.TableAssets() + ".course_id": workingData[0].ID}}
		count, err := scanner.assetDao.Count(dbParams)
		require.Nil(t, err)
		require.Equal(t, 10, count)
	})

	t.Run("add", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 12, 0)

		// Delete the assets (so we can add them again)
		for _, a := range workingData[0].Assets {
			require.Nil(t, scanner.assetDao.Delete(a.ID))
		}

		dbParams := &database.DatabaseParams{Where: squirrel.Eq{daos.TableAssets() + ".course_id": workingData[0].ID}}

		// ----------------------------
		// Add 10 assets
		// ----------------------------
		err := updateAssets(scanner.assetDao, workingData[0].ID, workingData[0].Assets[:10])
		require.Nil(t, err)

		count, err := scanner.assetDao.Count(dbParams)
		require.Nil(t, err)
		require.Equal(t, 10, count)

		// ----------------------------
		// Add another 2 assets
		// ----------------------------
		workingData[0].Assets[10].ID = ""
		workingData[0].Assets[11].ID = ""

		err = updateAssets(scanner.assetDao, workingData[0].ID, workingData[0].Assets)
		require.Nil(t, err)

		count, err = scanner.assetDao.Count(dbParams)
		require.Nil(t, err)
		require.Equal(t, 12, count)

		// Ensure all assets have an ID
		for _, a := range workingData[0].Assets {
			assert.NotEmpty(t, a.ID)
		}
	})

	t.Run("delete", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 12, 0)

		dbParams := &database.DatabaseParams{Where: squirrel.Eq{daos.TableAssets() + ".course_id": workingData[0].ID}}

		// ----------------------------
		// Remove 2 assets
		// ----------------------------
		workingData[0].Assets = workingData[0].Assets[2:]

		err := updateAssets(scanner.assetDao, workingData[0].ID, workingData[0].Assets)
		require.Nil(t, err)

		count, err := scanner.assetDao.Count(dbParams)
		require.Nil(t, err)
		require.Equal(t, 10, count)

		// ----------------------------
		// Remove another 2 assets
		// ----------------------------
		workingData[0].Assets = workingData[0].Assets[2:]

		err = updateAssets(scanner.assetDao, workingData[0].ID, workingData[0].Assets)
		require.Nil(t, err)

		count, err = scanner.assetDao.Count(dbParams)
		require.Nil(t, err)
		require.Equal(t, 8, count)
	})

	t.Run("db error", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		// Drop the table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.TableAssets())
		require.Nil(t, err)

		err = updateAssets(scanner.assetDao, "1234", []*models.Asset{})
		require.ErrorContains(t, err, fmt.Sprintf("no such table: %s", daos.TableAssets()))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateAttachments(t *testing.T) {
	t.Run("nothing added or delete)", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 1, 10)

		err := updateAttachments(scanner.attachmentDao, workingData[0].ID, workingData[0].Assets[0].Attachments)
		require.Nil(t, err)

		count, err := scanner.attachmentDao.Count(&database.DatabaseParams{Where: squirrel.Eq{daos.TableAttachments() + ".course_id": workingData[0].ID}})
		require.Nil(t, err)
		require.Equal(t, 10, count)
	})

	t.Run("add", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 1, 12)

		// Delete the attachments (so we can add them again)
		for _, a := range workingData[0].Assets[0].Attachments {
			require.Nil(t, scanner.attachmentDao.Delete(a.ID))
		}

		dbParams := &database.DatabaseParams{Where: squirrel.Eq{daos.TableAttachments() + ".course_id": workingData[0].ID}}

		// ----------------------------
		// Add 10 attachments
		// ----------------------------
		err := updateAttachments(scanner.attachmentDao, workingData[0].ID, workingData[0].Assets[0].Attachments[:10])
		require.Nil(t, err)

		count, err := scanner.attachmentDao.Count(dbParams)
		require.Nil(t, err)
		require.Equal(t, 10, count)

		// ----------------------------
		// Add another 2 attachments
		// ----------------------------
		err = updateAttachments(scanner.attachmentDao, workingData[0].ID, workingData[0].Assets[0].Attachments)
		require.Nil(t, err)

		count, err = scanner.attachmentDao.Count(dbParams)
		require.Nil(t, err)
		require.Equal(t, 12, count)
	})

	t.Run("delete", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		workingData := daos.NewTestData(t, db, 1, true, 1, 12)

		dbParams := &database.DatabaseParams{Where: squirrel.Eq{daos.TableAttachments() + ".course_id": workingData[0].ID}}

		// ----------------------------
		// Remove 2 attachments
		// ----------------------------
		workingData[0].Assets[0].Attachments = workingData[0].Assets[0].Attachments[2:]

		err := updateAttachments(scanner.attachmentDao, workingData[0].ID, workingData[0].Assets[0].Attachments)
		require.Nil(t, err)

		count, err := scanner.attachmentDao.Count(dbParams)
		require.Nil(t, err)
		require.Equal(t, 10, count)

		// ----------------------------
		// Remove another 2 attachments
		// ----------------------------
		workingData[0].Assets[0].Attachments = workingData[0].Assets[0].Attachments[2:]

		err = updateAttachments(scanner.attachmentDao, workingData[0].ID, workingData[0].Assets[0].Attachments)
		require.Nil(t, err)

		count, err = scanner.attachmentDao.Count(dbParams)
		require.Nil(t, err)
		require.Equal(t, 8, count)

	})

	t.Run("db error", func(t *testing.T) {
		scanner, db, _ := setupCourseScanner(t)

		// Drop the table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.TableAttachments())
		require.Nil(t, err)

		err = updateAttachments(scanner.attachmentDao, "1234", []*models.Attachment{})
		require.ErrorContains(t, err, fmt.Sprintf("no such table: %s", daos.TableAttachments()))
	})
}
