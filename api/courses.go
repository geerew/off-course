package api

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/queryparser"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type coursesAPI struct {
	logger     *logger.Logger
	appFs      *appfs.AppFs
	courseScan *coursescan.CourseScan
	dao        *dao.DAO
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initCourseRoutes initializes the course routes
func (r *Router) initCourseRoutes() {
	coursesAPI := coursesAPI{
		logger:     r.logger.WithAPI(),
		appFs:      r.config.AppFs,
		courseScan: r.config.CourseScan,
		dao:        r.dao,
	}

	courseGroup := r.api.Group("/courses")

	// Course
	courseGroup.Get("", coursesAPI.getCourses)
	courseGroup.Get("/:id", coursesAPI.getCourse)
	courseGroup.Post("", protectedRoute, coursesAPI.createCourse)
	courseGroup.Delete("/:id", protectedRoute, coursesAPI.deleteCourse)

	// Progress
	courseGroup.Delete("/:id/progress", coursesAPI.deleteCourseProgress)

	// Card
	courseGroup.Head("/:id/card", coursesAPI.getCard)
	courseGroup.Get("/:id/card", coursesAPI.getCard)

	// Lessons
	courseGroup.Get("/:id/lessons", coursesAPI.getLessons)
	courseGroup.Get("/:id/lessons/:lesson", coursesAPI.getLesson)

	// Modules (chaptered lessons)
	courseGroup.Get("/:id/modules", coursesAPI.getModules)

	// lesson attachments
	courseGroup.Get("/:id/lessons/:lesson/attachments", coursesAPI.getAttachments)
	courseGroup.Get("/:id/lessons/:lesson/attachments/:attachment", coursesAPI.getAttachment)
	courseGroup.Get("/:id/lessons/:lesson/attachments/:attachment/serve", coursesAPI.serveAttachment)

	// Asset
	courseGroup.Get("/:id/lessons/:lesson/assets/:asset/serve", coursesAPI.serveAsset)
	courseGroup.Put("/:id/lessons/:lesson/assets/:asset/progress", coursesAPI.updateAssetProgress)
	courseGroup.Delete("/:id/lessons/:lesson/assets/:asset/progress", coursesAPI.deleteAssetProgress)

	// Tags
	courseGroup.Get("/:id/tags", coursesAPI.getTags)
	courseGroup.Post("/:id/tags", protectedRoute, coursesAPI.createTag)
	courseGroup.Delete("/:id/tags/:tagId", protectedRoute, coursesAPI.deleteTag)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TODO add tests when initial scan is false
func (api coursesAPI) getCourses(c *fiber.Ctx) error {
	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	allowedQueryFilters := []string{"available", "tag"}

	withUserProgress := false
	if raw := c.Query("withUserProgress"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil && v {
			withUserProgress = v
		}
	}

	if withUserProgress {
		allowedQueryFilters = append(allowedQueryFilters, "progress")
	}

	builderOpts := builderOptions{
		DefaultOrderBy: defaultCoursesOrderBy,
		AllowedFilters: allowedQueryFilters,
		Paginate:       true,
		AfterParseHook: coursesAfterParseHook,
	}

	dbOpts, err := optionsBuilder(c, builderOpts, principal.UserID)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing query", err)
	}

	if withUserProgress {
		dbOpts.WithUserProgress()
	}

	courses, err := api.dao.ListCourses(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up courses", err)
	}

	pResult, err := dbOpts.Pagination.BuildResult(courseResponseHelper(courses, principal.Role == types.UserRoleAdmin))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TODO add tests when initial scan is false
func (api coursesAPI) getCourse(c *fiber.Ctx) error {
	id := c.Params("id")

	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_TABLE_ID: id})

	if raw := c.Query("withUserProgress"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil && v {
			dbOpts.WithUserProgress()
		}
	}

	course, err := api.dao.GetCourse(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up course", err)
	}

	if course == nil {
		return errorResponse(c, fiber.StatusNotFound, "Course not found", fmt.Errorf("course not found"))
	}

	return c.Status(fiber.StatusOK).JSON(courseResponseHelper([]*models.Course{course}, principal.Role == types.UserRoleAdmin)[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) createCourse(c *fiber.Ctx) error {
	req := &courseRequest{}
	if err := c.BodyParser(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	// Ensure there is a title and path
	if req.Title == "" || req.Path == "" {
		return errorResponse(c, fiber.StatusBadRequest, "A title and path are required", nil)
	}

	course := &models.Course{
		Title: req.Title,
		Path:  utils.NormalizeWindowsDrive(req.Path),
	}

	// Validate the path
	if exists, err := afero.DirExists(api.appFs.Fs, course.Path); err != nil || !exists {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid course path", err)
	}

	// Set the course to available
	course.Available = true

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	if err := api.dao.CreateCourse(ctx, course); err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "A course with this path already exists", err)
		}
		return errorResponse(c, fiber.StatusInternalServerError, "Error creating course", err)
	}

	// Start a scan job
	if _, err := api.courseScan.Add(ctx, course.ID); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error creating scan job", err)
	}

	return c.Status(fiber.StatusCreated).JSON(courseResponseHelper([]*models.Course{course}, true)[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) deleteCourse(c *fiber.Ctx) error {
	id := c.Params("id")

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_TABLE_ID: id})
	if err := api.dao.DeleteCourses(ctx, dbOpts); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting course", err)
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TODO add tests
func (api coursesAPI) deleteCourseProgress(c *fiber.Ctx) error {
	courseId := c.Params("id")

	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	err = dao.RunInTransaction(ctx, api.dao, func(txCtx context.Context) error {
		// Delete the course progress for this user
		dbOpts := database.NewOptions().WithWhere(squirrel.And{
			squirrel.Eq{models.COURSE_PROGRESS_COURSE_ID: courseId},
			squirrel.Eq{models.COURSE_PROGRESS_USER_ID: principal.UserID},
		},
		)

		if err := api.dao.DeleteCourseProgress(txCtx, dbOpts); err != nil {
			return err
		}

		if err := api.dao.DeleteAssetProgressForCourse(txCtx, courseId, principal.UserID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting course progress", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getCard(c *fiber.Ctx) error {
	id := c.Params("id")

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_TABLE_ID: id})
	course, err := api.dao.GetCourse(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up course", err)
	}

	if course == nil {
		return errorResponse(c, fiber.StatusNotFound, "Course not found", nil)

	}
	if course.CardPath == "" {
		return errorResponse(c, fiber.StatusNotFound, "Course has no card", nil)
	}

	_, err = api.appFs.Fs.Stat(course.CardPath)
	if os.IsNotExist(err) {
		return errorResponse(c, fiber.StatusNotFound, "Course card not found", nil)
	}

	// The fiber function sendFile(...) does not support using a custom FS. Therefore, use
	// SendFile() from the filesystem middleware.
	return filesystem.SendFile(c, afero.NewHttpFs(api.appFs.Fs), course.CardPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TODO support chaptered query param
func (api coursesAPI) getLessons(c *fiber.Ctx) error {
	id := c.Params("id")

	builderOpts := builderOptions{
		DefaultOrderBy: defaultCourseLessonsOrderBy,
		Paginate:       true,
	}

	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts, err := optionsBuilder(c, builderOpts, principal.UserID)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing query", err)
	}

	dbOpts.WithAssetMetadata().WithWhere(squirrel.Eq{models.LESSON_TABLE_COURSE_ID: id})

	if raw := c.Query("withUserProgress"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil && v {
			dbOpts.WithUserProgress()
		}
	}

	lessons, err := api.dao.ListLessons(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up lessons", err)
	}

	pResult, err := dbOpts.Pagination.BuildResult(lessonResponseHelper(lessons))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getLesson(c *fiber.Ctx) error {
	id := c.Params("id")
	lessonId := c.Params("lesson")

	dbOpts := database.NewOptions().
		WithAssetMetadata().
		WithWhere(squirrel.And{
			squirrel.Eq{models.LESSON_TABLE_ID: lessonId},
			squirrel.Eq{models.LESSON_TABLE_COURSE_ID: id},
		})

	if raw := c.Query("withUserProgress"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil && v {
			dbOpts.WithUserProgress()
		}
	}

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	lesson, err := api.dao.GetLesson(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up lesson", err)
	}

	if lesson == nil {
		return errorResponse(c, fiber.StatusNotFound, "Lesson not found", nil)
	}

	return c.Status(fiber.StatusOK).JSON(lessonResponseHelper([]*models.Lesson{lesson})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getModules(c *fiber.Ctx) error {
	id := c.Params("id")

	builderOpts := builderOptions{
		DefaultOrderBy: defaultCourseLessonsOrderBy,
		Paginate:       false,
	}

	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts, err := optionsBuilder(c, builderOpts, principal.UserID)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing query", err)
	}

	dbOpts.WithAssetMetadata().WithWhere(squirrel.Eq{models.LESSON_TABLE_COURSE_ID: id})

	if raw := c.Query("withUserProgress"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil && v {
			dbOpts.WithUserProgress()
		}
	}

	lessons, err := api.dao.ListLessons(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up lessons", err)
	}

	return c.Status(fiber.StatusOK).JSON(modulesResponseHelper(lessons))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getAttachments(c *fiber.Ctx) error {
	id := c.Params("id")
	lessonId := c.Params("lesson")

	builderOpts := builderOptions{
		DefaultOrderBy: defaultCourseLessonAttachmentsOrderBy,
		Paginate:       true,
	}

	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts, err := optionsBuilder(c, builderOpts, principal.UserID)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing query", err)
	}

	dbOpts.WithCourse().
		WithLesson().
		WithWhere(squirrel.And{
			squirrel.Eq{models.LESSON_TABLE_ID: lessonId},
			squirrel.Eq{models.COURSE_TABLE_ID: id},
		})

	attachments, err := api.dao.ListAttachments(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up attachments", err)
	}

	pResult, err := dbOpts.Pagination.BuildResult(attachmentResponseHelper(attachments))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getAttachment(c *fiber.Ctx) error {
	id := c.Params("id")
	lessonId := c.Params("lesson")
	attachmentId := c.Params("attachment")

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts := database.NewOptions().
		WithCourse().
		WithLesson().
		WithWhere(squirrel.And{
			squirrel.Eq{models.ATTACHMENT_TABLE_ID: attachmentId},
			squirrel.Eq{models.LESSON_TABLE_ID: lessonId},
			squirrel.Eq{models.COURSE_TABLE_ID: id},
		})

	attachment, err := api.dao.GetAttachment(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up attachment", err)
	}

	if attachment == nil {
		return errorResponse(c, fiber.StatusNotFound, "Attachment not found", nil)
	}

	return c.Status(fiber.StatusOK).JSON(attachmentResponseHelper([]*models.Attachment{attachment})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) serveAttachment(c *fiber.Ctx) error {
	id := c.Params("id")
	lessonId := c.Params("lesson")
	attachmentId := c.Params("attachment")

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts := database.NewOptions().
		WithCourse().
		WithLesson().
		WithWhere(squirrel.And{
			squirrel.Eq{models.ATTACHMENT_TABLE_ID: attachmentId},
			squirrel.Eq{models.LESSON_TABLE_ID: lessonId},
			squirrel.Eq{models.COURSE_TABLE_ID: id},
		})

	attachment, err := api.dao.GetAttachment(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up attachment", err)
	}

	if attachment == nil {
		return errorResponse(c, fiber.StatusNotFound, "Attachment not found", nil)
	}

	if exists, err := afero.Exists(api.appFs.Fs, attachment.Path); err != nil || !exists {
		return errorResponse(c, fiber.StatusBadRequest, "Attachment does not exist", err)
	}

	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+attachment.Title+`"`)
	return filesystem.SendFile(c, afero.NewHttpFs(api.appFs.Fs), attachment.Path)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TODO Handle PDF
func (api coursesAPI) serveAsset(c *fiber.Ctx) error {
	id := c.Params("id")
	lessonId := c.Params("lesson")
	assetId := c.Params("asset")

	dbOpts := database.NewOptions().
		WithCourse().
		WithLesson().
		WithWhere(squirrel.And{
			squirrel.Eq{models.COURSE_TABLE_ID: id},
			squirrel.Eq{models.LESSON_TABLE_ID: lessonId},
			squirrel.Eq{models.ASSET_TABLE_ID: assetId},
		})

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	asset, err := api.dao.GetAsset(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset == nil {
		return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
	}

	// Check for invalid path
	if exists, err := afero.Exists(api.appFs.Fs, asset.Path); err != nil || !exists {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not exist", nil)
	}

	if asset.Type.IsVideo() {
		return handleVideo(c, api.appFs, asset)
	} else if asset.Type.IsText() || asset.Type.IsMarkdown() {
		return handleText(c, api.appFs, asset)
	}

	return c.Status(fiber.StatusOK).SendString("done")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) updateAssetProgress(c *fiber.Ctx) error {
	assetId := c.Params("asset")

	req := &assetProgressRequest{}
	if err := c.BodyParser(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	assetProgress := &models.AssetProgress{
		AssetID:   assetId,
		Position:  req.Position,
		Completed: req.Completed,
	}

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	if err := api.dao.UpsertAssetProgress(ctx, assetProgress); err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}
		return errorResponse(c, fiber.StatusInternalServerError, "Error updating asset", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TODO add tests
func (api coursesAPI) deleteAssetProgress(c *fiber.Ctx) error {
	courseId := c.Params("id")
	assetId := c.Params("asset")

	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	// First, verify the asset belongs to the specified course
	asset, err := api.dao.GetAsset(ctx, database.NewOptions().
		WithWhere(squirrel.And{
			squirrel.Eq{models.ASSET_TABLE_ID: assetId},
			squirrel.Eq{models.ASSET_TABLE_COURSE_ID: courseId},
		}))

	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset == nil {
		return errorResponse(c, fiber.StatusNotFound, "Asset not found for this course", nil)
	}

	dbOpts := database.NewOptions().
		WithWhere(squirrel.And{
			squirrel.Eq{models.ASSET_PROGRESS_ASSET_ID: assetId},
			squirrel.Eq{models.ASSET_PROGRESS_USER_ID: principal.UserID},
		})

	if err := api.dao.DeleteAssetProgress(ctx, dbOpts); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting asset progress", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getTags(c *fiber.Ctx) error {
	id := c.Params("id")

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts := database.NewOptions().
		WithOrderBy(defaultTagsOrderBy...).
		WithWhere(squirrel.Eq{models.COURSE_TAG_TABLE_COURSE_ID: id})

	courseTags, err := api.dao.ListCourseTags(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up course tags", err)
	}

	return c.Status(fiber.StatusOK).JSON(courseTagResponseHelper(courseTags))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) createTag(c *fiber.Ctx) error {
	courseId := c.Params("id")
	tagRequest := &tagRequest{}

	if err := c.BodyParser(tagRequest); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	if tagRequest.Tag == "" {
		return errorResponse(c, fiber.StatusBadRequest, "A tag is required", nil)
	}

	courseTag := &models.CourseTag{
		CourseID: courseId,
		Tag:      tagRequest.Tag,
	}

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	if err := api.dao.CreateCourseTag(ctx, courseTag); err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "A tag for this course already exists", err)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error creating course tag", err)
	}

	return c.Status(fiber.StatusCreated).JSON(courseTagResponseHelper([]*models.CourseTag{courseTag})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) deleteTag(c *fiber.Ctx) error {
	courseId := c.Params("id")
	tagId := c.Params("tagId")

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts := database.NewOptions().
		WithWhere(squirrel.And{
			squirrel.Eq{models.COURSE_TAG_TABLE_COURSE_ID: courseId},
			squirrel.Eq{models.COURSE_TAG_TABLE_ID: tagId},
		})

	if err := api.dao.DeleteCourseTags(ctx, dbOpts); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting course tag", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// coursesAfterParseHook runs after parsing the query expression and is used to build the
// WHERE/JOIN clauses
func coursesAfterParseHook(parsed *queryparser.QueryResult, dbOpts *database.Options, userID string) {
	dbOpts.WithWhere(coursesWhereBuilder(parsed.Expr))

	// if foundProgress, ok := parsed.FoundFilters["progress"]; ok && foundProgress {
	// 	dbOpts.WithLeftJoin(models.COURSE_PROGRESS_TABLE,
	// 		fmt.Sprintf("%s = %s AND %s = '%s'",
	// 			models.COURSE_PROGRESS_TABLE_COURSE_ID,
	// 			models.COURSE_TABLE_ID,
	// 			models.COURSE_PROGRESS_TABLE_USER_ID,
	// 			userID),
	// 	)
	// }
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// coursesWhereBuilder builds a squirrel.Sqlizer, for use in a WHERE clause
func coursesWhereBuilder(expr queryparser.QueryExpr) squirrel.Sqlizer {
	switch node := expr.(type) {
	case *queryparser.ValueExpr:
		return squirrel.Like{models.COURSE_TABLE_TITLE: "%" + node.Value + "%"}
	case *queryparser.FilterExpr:
		switch node.Key {
		case "available":
			value, err := cast.ToBoolE(node.Value)
			if err != nil {
				return squirrel.Expr("1=0")
			}
			return squirrel.Eq{models.COURSE_TABLE_AVAILABLE: value}
		case "tag":
			return courseTagsBuilder([]string{node.Value})
		case "progress":
			switch strings.ToLower(node.Value) {
			case "not started":
				// For "not started", we need:
				// 1. Either no progress record exists (IS NULL after LEFT JOIN)
				// 2. Or the progress record exists but started=false
				return squirrel.Or{
					// No progress record exists
					squirrel.Expr(models.COURSE_PROGRESS_TABLE_ID + " IS NULL"),
					// Or started is false
					squirrel.Eq{models.COURSE_PROGRESS_TABLE_STARTED: false},
				}
			case "started":
				return squirrel.And{
					// Must have a progress record
					squirrel.Expr(models.COURSE_PROGRESS_TABLE_ID + " IS NOT NULL"),
					// Started must be true
					squirrel.Eq{models.COURSE_PROGRESS_TABLE_STARTED: true},
					// But not 100% complete
					squirrel.NotEq{models.COURSE_PROGRESS_TABLE_PERCENT: 100},
				}
			case "completed":
				return squirrel.And{
					// Must have a progress record
					squirrel.Expr(models.COURSE_PROGRESS_TABLE_ID + " IS NOT NULL"),
					// And must be 100% complete
					squirrel.Eq{models.COURSE_PROGRESS_TABLE_PERCENT: 100},
				}
			default:
				return nil
			}
		default:
			return nil
		}
	case *queryparser.AndExpr:
		var andSlice []squirrel.Sqlizer
		var tags []string
		onlyTags := true

		// Loop through all children and separate tag filters from non-tag conditions
		for _, child := range node.Children {
			if queryparser.IsFilterWithKey(child, "tag") {
				tags = append(tags, child.(*queryparser.FilterExpr).Value)
			} else {
				onlyTags = false
				andSlice = append(andSlice, coursesWhereBuilder(child))
			}
		}

		// If we found tags, build the EXISTS subquery
		var tagCond squirrel.Sqlizer
		if len(tags) > 0 {
			tagCond = courseTagsBuilder(tags)

			if onlyTags {
				return tagCond
			} else if tagCond != nil {
				andSlice = append(andSlice, tagCond)
			}
		}

		return squirrel.And(andSlice)
	case *queryparser.OrExpr:
		var orSlice []squirrel.Sqlizer
		for _, child := range node.Children {
			orSlice = append(orSlice, coursesWhereBuilder(child))
		}

		return squirrel.Or(orSlice)
	default:
		return nil
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// courseTagsBuilder builds an EXISTS squirrel.Sqlizer subquery for a list of tags
func courseTagsBuilder(tags []string) squirrel.Sqlizer {
	if len(tags) == 0 {
		return squirrel.Expr("1=1")
	}

	baseQuery := squirrel.
		Select("1").
		From(models.COURSE_TAG_TABLE).
		Join(models.TAG_TABLE + " ON " + models.TAG_TABLE_ID + " = " + models.COURSE_TAG_TABLE_TAG_ID).
		Where(models.COURSE_TAG_TABLE_COURSE_ID + " = " + models.COURSE_TABLE_ID)

	if len(tags) == 1 {
		baseQuery = baseQuery.Where(squirrel.Eq{models.TAG_TABLE_TAG: tags[0]})
	} else if len(tags) > 1 {
		baseQuery = baseQuery.
			Where(squirrel.Eq{models.TAG_TABLE_TAG: tags}).
			GroupBy(models.COURSE_TAG_TABLE_COURSE_ID).
			Having("COUNT(DISTINCT "+models.TAG_TABLE_TAG+") = ?", len(tags))
	}

	return squirrel.Expr("EXISTS (?)", baseQuery)
}
