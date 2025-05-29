package api

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/queryparser"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type coursesAPI struct {
	logger     *slog.Logger
	appFs      *appfs.AppFs
	courseScan *coursescan.CourseScan
	dao        *dao.DAO
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initCourseRoutes initializes the course routes
func (r *Router) initCourseRoutes() {
	coursesAPI := coursesAPI{
		logger:     r.config.Logger,
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

	// Course progress
	courseGroup.Delete("/:id/progress", coursesAPI.deleteCourseProgress)

	// Course card
	courseGroup.Head("/:id/card", coursesAPI.getCard)
	courseGroup.Get("/:id/card", coursesAPI.getCard)

	// Course asset
	courseGroup.Get("/:id/assets", coursesAPI.getAssets)
	courseGroup.Get("/:id/assets/:asset", coursesAPI.getAsset)
	courseGroup.Get("/:id/assets/:asset/serve", coursesAPI.serveAsset)
	courseGroup.Get("/:id/assets/:asset/description", coursesAPI.serveAssetDescription)
	courseGroup.Put("/:id/assets/:asset/progress", coursesAPI.updateAssetProgress)
	courseGroup.Delete("/:id/assets/:asset/progress", coursesAPI.deleteAssetProgress)

	// Course asset attachments
	courseGroup.Get("/:id/assets/:asset/attachments", coursesAPI.getAttachments)
	courseGroup.Get("/:id/assets/:asset/attachments/:attachment", coursesAPI.getAttachment)
	courseGroup.Get("/:id/assets/:asset/attachments/:attachment/serve", coursesAPI.serveAttachment)

	// Course tags
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

	builderOptions := builderOptions{
		DefaultOrderBy: defaultCoursesOrderBy,
		AllowedFilters: []string{"available", "tag", "progress"},
		Paginate:       true,
		AfterParseHook: coursesAfterParseHook,
	}

	options, err := optionsBuilder(c, builderOptions, principal.UserID)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing query", err)
	}

	courses := []*models.Course{}
	if err = api.dao.ListCourses(ctx, &courses, options); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up courses", err)
	}

	pResult, err := options.Pagination.BuildResult(courseResponseHelper(courses, principal.Role == types.UserRoleAdmin))
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

	course := &models.Course{}
	if err := api.dao.GetCourse(ctx, course, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_ID: id}}); err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Course not found", fmt.Errorf("course not found"))
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up course", err)
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

	course := &models.Course{Base: models.Base{ID: id}}
	if err := dao.Delete(ctx, api.dao, course, nil); err != nil {
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
		options := &database.Options{
			Where: squirrel.And{
				squirrel.Eq{models.COURSE_PROGRESS_COURSE_ID: courseId},
				squirrel.Eq{models.COURSE_PROGRESS_USER_ID: principal.UserID},
			},
		}

		if err := dao.Delete(txCtx, api.dao, &models.CourseProgress{}, options); err != nil {
			return err
		}

		// Pluck the asset progress IDs for this asset progress associated with this course and user
		options = &database.Options{}
		options.AddJoin(models.ASSET_TABLE, models.ASSET_PROGRESS_TABLE_ASSET_ID+" = "+models.ASSET_TABLE_ID)
		options.AddJoin(models.COURSE_TABLE, models.ASSET_TABLE_COURSE_ID+" = "+models.COURSE_TABLE_ID)
		options.Where = squirrel.And{
			squirrel.Eq{models.ASSET_PROGRESS_USER_ID: principal.UserID},
			squirrel.Eq{models.COURSE_TABLE_ID: courseId},
		}

		assetProgressIDs, err := dao.ListPluck[[]string](txCtx, api.dao, &models.AssetProgress{}, options, models.BASE_ID)
		if err != nil {
			return err
		}

		if len(assetProgressIDs) == 0 {
			return nil
		}

		// Delete the asset progress for this user and course
		options = &database.Options{Where: squirrel.And{squirrel.Eq{models.ASSET_PROGRESS_TABLE_ID: assetProgressIDs}}}
		if err = dao.Delete(txCtx, api.dao, &models.AssetProgress{}, options); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting asset progress", err)
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

	course := &models.Course{Base: models.Base{ID: id}}
	if err := api.dao.GetCourse(ctx, course, nil); err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Course not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up course", err)
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

func (api coursesAPI) getAssets(c *fiber.Ctx) error {
	id := c.Params("id")

	builderOptions := builderOptions{
		DefaultOrderBy: defaultCourseAssetsOrderBy,
		Paginate:       true,
	}

	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	options, err := optionsBuilder(c, builderOptions, principal.UserID)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing query", err)
	}

	options.Where = squirrel.Eq{models.ASSET_TABLE_COURSE_ID: id}

	assets := []*models.Asset{}
	if err = api.dao.ListAssets(ctx, &assets, options); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up assets", err)
	}

	pResult, err := options.Pagination.BuildResult(assetResponseHelper(assets))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getAsset(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")

	options := &database.Options{}

	// Join the course table
	options.AddJoin(models.COURSE_TABLE, models.ASSET_TABLE_COURSE_ID+" = "+models.COURSE_TABLE_ID)

	options.Where = squirrel.And{
		squirrel.Eq{models.ASSET_TABLE_ID: assetId},
		squirrel.Eq{models.COURSE_TABLE_ID: id},
	}

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	asset := &models.Asset{}
	if err := api.dao.GetAsset(ctx, asset, options); err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	return c.Status(fiber.StatusOK).JSON(assetResponseHelper([]*models.Asset{asset})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) serveAsset(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")

	options := &database.Options{}

	// Join the course table
	options.AddJoin(models.COURSE_TABLE, models.ASSET_TABLE_COURSE_ID+" = "+models.COURSE_TABLE_ID)

	options.Where = squirrel.And{
		squirrel.Eq{models.ASSET_TABLE_ID: assetId},
		squirrel.Eq{models.COURSE_TABLE_ID: id},
	}

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	asset := &models.Asset{}
	if err := api.dao.GetAsset(ctx, asset, options); err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	// Check for invalid path
	if exists, err := afero.Exists(api.appFs.Fs, asset.Path); err != nil || !exists {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not exist", nil)
	}

	if asset.Type.IsVideo() {
		return handleVideo(c, api.appFs, asset)
	} else if asset.Type.IsHTML() {
		return handleHtml(c, api.appFs, asset)
	}

	// TODO Handle PDF and HTML
	return c.Status(fiber.StatusOK).SendString("done")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) serveAssetDescription(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")

	options := &database.Options{}

	// Join the course table
	options.AddJoin(models.COURSE_TABLE, models.ASSET_TABLE_COURSE_ID+" = "+models.COURSE_TABLE_ID)

	options.Where = squirrel.And{
		squirrel.Eq{models.ASSET_TABLE_ID: assetId},
		squirrel.Eq{models.COURSE_TABLE_ID: id},
	}

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	asset := &models.Asset{}
	if err := api.dao.GetAsset(ctx, asset, options); err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset.DescriptionPath == "" {
		return errorResponse(c, fiber.StatusNotFound, "Asset has no description", nil)
	}

	if exists, err := afero.Exists(api.appFs.Fs, asset.DescriptionPath); err != nil || !exists {
		return errorResponse(c, fiber.StatusBadRequest, "Asset description does not exist", err)
	}

	return filesystem.SendFile(c, afero.NewHttpFs(api.appFs.Fs), asset.DescriptionPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) updateAssetProgress(c *fiber.Ctx) error {
	courseId := c.Params("id")
	assetId := c.Params("asset")

	req := &assetProgressRequest{}
	if err := c.BodyParser(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	assetProgress := &models.AssetProgress{
		AssetID:   assetId,
		VideoPos:  req.VideoPos,
		Completed: req.Completed,
	}

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	if err := api.dao.CreateOrUpdateAssetProgress(ctx, courseId, assetProgress); err != nil {
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

	options := &database.Options{}
	options.AddJoin(models.COURSE_TABLE, models.ASSET_TABLE_COURSE_ID+" = "+models.COURSE_TABLE_ID)
	options.Where = squirrel.And{
		squirrel.Eq{models.ASSET_PROGRESS_ASSET_ID: assetId},
		squirrel.Eq{models.ASSET_PROGRESS_USER_ID: principal.UserID},
		squirrel.Eq{models.COURSE_TABLE_ID: courseId},
	}

	if err := dao.Delete(ctx, api.dao, &models.AssetProgress{}, options); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting asset progress", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getAttachments(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")

	builderOptions := builderOptions{
		DefaultOrderBy: defaultCourseAssetAttachmentsOrderBy,
		Paginate:       true,
	}

	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	options, err := optionsBuilder(c, builderOptions, principal.UserID)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing query", err)
	}

	// Join the asset and course tables to ensure the asset belongs to the course
	// Join the asset and course tables
	options.AddJoin(models.ASSET_TABLE, models.ATTACHMENT_TABLE_ASSET_ID+" = "+models.ASSET_TABLE_ID)
	options.AddJoin(models.COURSE_TABLE, models.ASSET_TABLE_COURSE_ID+" = "+models.COURSE_TABLE_ID)

	options.Where = squirrel.And{
		squirrel.Eq{models.ATTACHMENT_TABLE_ASSET_ID: assetId},
		squirrel.Eq{models.COURSE_TABLE_ID: id},
	}

	attachments := []*models.Attachment{}
	if err = api.dao.ListAttachments(ctx, &attachments, options); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up attachments", err)
	}

	pResult, err := options.Pagination.BuildResult(attachmentResponseHelper(attachments))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getAttachment(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")
	attachmentId := c.Params("attachment")

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	options := &database.Options{}

	// Join the asset and course tables
	options.AddJoin(models.ASSET_TABLE, models.ATTACHMENT_TABLE_ASSET_ID+" = "+models.ASSET_TABLE_ID)
	options.AddJoin(models.COURSE_TABLE, models.ASSET_TABLE_COURSE_ID+" = "+models.COURSE_TABLE_ID)

	options.Where = squirrel.And{
		squirrel.Eq{models.ATTACHMENT_TABLE_ID: attachmentId},
		squirrel.Eq{models.ASSET_TABLE_ID: assetId},
		squirrel.Eq{models.COURSE_TABLE_ID: id},
	}

	attachment := &models.Attachment{}
	if err := api.dao.GetAttachment(ctx, attachment, options); err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Attachment not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up attachment", err)
	}

	return c.Status(fiber.StatusOK).JSON(attachmentResponseHelper([]*models.Attachment{attachment})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) serveAttachment(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")
	attachmentId := c.Params("attachment")

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	options := &database.Options{}

	// Join the asset and course tables
	options.AddJoin(models.ASSET_TABLE, models.ATTACHMENT_TABLE_ASSET_ID+" = "+models.ASSET_TABLE_ID)
	options.AddJoin(models.COURSE_TABLE, models.ASSET_TABLE_COURSE_ID+" = "+models.COURSE_TABLE_ID)

	options.Where = squirrel.And{
		squirrel.Eq{models.ATTACHMENT_TABLE_ID: attachmentId},
		squirrel.Eq{models.ASSET_TABLE_ID: assetId},
		squirrel.Eq{models.COURSE_TABLE_ID: id},
	}

	attachment := &models.Attachment{}
	if err := api.dao.GetAttachment(ctx, attachment, options); err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Attachment not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up attachment", err)
	}

	if exists, err := afero.Exists(api.appFs.Fs, attachment.Path); err != nil || !exists {
		return errorResponse(c, fiber.StatusBadRequest, "Attachment does not exist", err)
	}

	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+attachment.Title+`"`)
	return filesystem.SendFile(c, afero.NewHttpFs(api.appFs.Fs), attachment.Path)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getTags(c *fiber.Ctx) error {
	id := c.Params("id")

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	options := &database.Options{
		OrderBy: defaultTagsOrderBy,
		Where:   squirrel.Eq{models.COURSE_TAG_TABLE_COURSE_ID: id},
	}

	tags := []*models.CourseTag{}
	if err := api.dao.ListCourseTags(ctx, &tags, options); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up course tags", err)
	}

	return c.Status(fiber.StatusOK).JSON(courseTagResponseHelper(tags))
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

	options := &database.Options{
		Where: squirrel.And{
			squirrel.Eq{models.COURSE_TAG_TABLE_COURSE_ID: courseId},
			squirrel.Eq{models.COURSE_TAG_TABLE_ID: tagId},
		},
	}

	if err := dao.Delete(ctx, api.dao, &models.CourseTag{}, options); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting course tag", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// coursesAfterParseHook runs after parsing the query expression and is used to build the
// WHERE/JOIN clauses
func coursesAfterParseHook(parsed *queryparser.QueryResult, options *database.Options, userID string) {
	options.Where = coursesWhereBuilder(parsed.Expr)

	if foundProgress, ok := parsed.FoundFilters["progress"]; ok && foundProgress {
		options.AddLeftJoin(
			models.COURSE_PROGRESS_TABLE,
			fmt.Sprintf("%s = %s AND %s = '%s'",
				models.COURSE_PROGRESS_TABLE_COURSE_ID,
				models.COURSE_TABLE_ID,
				models.COURSE_PROGRESS_TABLE_USER_ID,
				userID),
		)
	}
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
