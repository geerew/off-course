package api

import (
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
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/queryparser"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	defaultCoursesOrderBy = []string{models.COURSE_TABLE + "." + models.BASE_CREATED_AT + " desc"}
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

	// Course card
	courseGroup.Head("/:id/card", coursesAPI.getCard)
	courseGroup.Get("/:id/card", coursesAPI.getCard)

	// Course asset
	courseGroup.Get("/:id/assets", coursesAPI.getAssets)
	courseGroup.Get("/:id/assets/:asset", coursesAPI.getAsset)
	courseGroup.Get("/:id/assets/:asset/serve", coursesAPI.serveAsset)
	courseGroup.Put("/:id/assets/:asset/progress", protectedRoute, coursesAPI.updateAssetProgress)

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

func (api coursesAPI) getCourses(c *fiber.Ctx) error {
	options, err := courseOptionsBuilder(c, true)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing query", err)
	}

	courses := []*models.Course{}
	err = api.dao.List(c.UserContext(), &courses, options)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up courses", err)
	}

	pResult, err := options.Pagination.BuildResult(courseResponseHelper(courses))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getCourse(c *fiber.Ctx) error {
	id := c.Params("id")

	course := &models.Course{Base: models.Base{ID: id}}
	err := api.dao.GetById(c.UserContext(), course)

	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Course not found", fmt.Errorf("course not found"))
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up course", err)
	}

	return c.Status(fiber.StatusOK).JSON(courseResponseHelper([]*models.Course{course})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) createCourse(c *fiber.Ctx) error {
	course := new(models.Course)

	if err := c.BodyParser(course); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	// Ensure there is a title and path
	if course.Title == "" || course.Path == "" {
		return errorResponse(c, fiber.StatusBadRequest, "A title and path are required", nil)
	}

	// Empty stuff that should not be set
	course.ID = ""

	course.Path = utils.NormalizeWindowsDrive(course.Path)

	// Validate the path
	if exists, err := afero.DirExists(api.appFs.Fs, course.Path); err != nil || !exists {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid course path", err)
	}

	// Set the course to available
	course.Available = true

	if err := api.dao.CreateCourse(c.UserContext(), course); err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "A course with this path already exists", err)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error creating course", err)
	}

	// Start a scan job
	if scan, err := api.courseScan.Add(c.UserContext(), course.ID); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error creating scan job", err)
	} else {
		course.ScanStatus = scan.Status
	}

	return c.Status(fiber.StatusCreated).JSON(courseResponseHelper([]*models.Course{course})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) deleteCourse(c *fiber.Ctx) error {
	id := c.Params("id")

	course := &models.Course{Base: models.Base{ID: id}}
	err := api.dao.Delete(c.UserContext(), course, nil)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting course", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getCard(c *fiber.Ctx) error {
	id := c.Params("id")

	course := &models.Course{Base: models.Base{ID: id}}
	err := api.dao.GetById(c.UserContext(), course)

	if err != nil {
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
	orderBy := c.Query("orderBy", "chapter asc,prefix asc")

	options := &database.Options{
		OrderBy:    strings.Split(orderBy, ","),
		Where:      squirrel.Eq{models.ASSET_TABLE + ".course_id": id},
		Pagination: pagination.NewFromApi(c),
	}

	assets := []*models.Asset{}
	err := api.dao.List(c.UserContext(), &assets, options)
	if err != nil {
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

	asset := &models.Asset{Base: models.Base{ID: assetId}}
	err := api.dao.GetById(c.UserContext(), asset)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset.CourseID != id {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not belong to course", nil)
	}

	return c.Status(fiber.StatusOK).JSON(assetResponseHelper([]*models.Asset{asset})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) serveAsset(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")

	asset := &models.Asset{Base: models.Base{ID: assetId}}
	err := api.dao.GetById(c.UserContext(), asset)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset.CourseID != id {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not belong to course", nil)
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

	// TODO: Handle PDF
	return c.Status(fiber.StatusOK).SendString("done")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) updateAssetProgress(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")

	req := &assetProgressRequest{}
	if err := c.BodyParser(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	asset := &models.Asset{Base: models.Base{ID: assetId}}
	err := api.dao.GetById(c.UserContext(), asset)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset.CourseID != id {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not belong to course", nil)
	}

	assetProgress := &models.AssetProgress{
		AssetID:   assetId,
		VideoPos:  req.VideoPos,
		Completed: req.Completed,
	}

	err = api.dao.CreateOrUpdateAssetProgress(c.UserContext(), assetProgress)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error updating asset", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getAttachments(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")
	orderBy := c.Query("orderBy", "title asc")

	asset := &models.Asset{Base: models.Base{ID: assetId}}
	err := api.dao.GetById(c.UserContext(), asset)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset.CourseID != id {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not belong to course", nil)
	}

	options := &database.Options{
		OrderBy:    strings.Split(orderBy, ","),
		Where:      squirrel.Eq{models.ATTACHMENT_TABLE + ".asset_id": assetId},
		Pagination: pagination.NewFromApi(c),
	}

	attachments := []*models.Attachment{}
	err = api.dao.List(c.UserContext(), &attachments, options)
	if err != nil {
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

	asset := &models.Asset{Base: models.Base{ID: assetId}}
	err := api.dao.GetById(c.UserContext(), asset)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset.CourseID != id {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not belong to course", nil)
	}

	attachment := &models.Attachment{Base: models.Base{ID: attachmentId}}
	err = api.dao.GetById(c.UserContext(), attachment)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Attachment not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up attachment", err)
	}

	if attachment.AssetID != assetId {
		return errorResponse(c, fiber.StatusBadRequest, "Attachment does not belong to asset", nil)
	}

	return c.Status(fiber.StatusOK).JSON(attachmentResponseHelper([]*models.Attachment{attachment})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) serveAttachment(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")
	attachmentID := c.Params("attachment")

	asset := &models.Asset{Base: models.Base{ID: assetId}}
	err := api.dao.GetById(c.UserContext(), asset)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset.CourseID != id {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not belong to course", nil)
	}

	attachment := &models.Attachment{Base: models.Base{ID: attachmentID}}
	err = api.dao.GetById(c.UserContext(), attachment)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Attachment not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up attachment", err)
	}

	if attachment.AssetID != assetId {
		return errorResponse(c, fiber.StatusBadRequest, "Attachment does not belong to asset", nil)
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

	course := &models.Course{Base: models.Base{ID: id}}
	err := api.dao.GetById(c.UserContext(), course)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Course not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up course", err)
	}

	options := &database.Options{
		OrderBy: []string{models.TAG_TABLE + ".tag asc"},
		Where:   squirrel.Eq{models.COURSE_TAG_TABLE + ".course_id": id},
	}

	tags := []*models.CourseTag{}
	err = api.dao.List(c.UserContext(), &tags, options)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up course tags", err)
	}

	return c.Status(fiber.StatusOK).JSON(courseTagResponseHelper(tags))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) createTag(c *fiber.Ctx) error {
	courseId := c.Params("id")
	courseTag := new(models.CourseTag)

	if err := c.BodyParser(courseTag); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	courseTag.ID = ""
	courseTag.TagID = ""

	if courseTag.Tag == "" {
		return errorResponse(c, fiber.StatusBadRequest, "A tag is required", nil)
	}

	courseTag.CourseID = courseId

	err := api.dao.CreateCourseTag(c.UserContext(), courseTag)
	if err != nil {
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

	err := api.dao.Delete(
		c.UserContext(),
		&models.CourseTag{},
		&database.Options{Where: squirrel.And{squirrel.Eq{"course_id": courseId}, squirrel.Eq{"id": tagId}}},
	)

	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting course tag", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// courseOptionsBuilder builds the database.Options for a courses query
func courseOptionsBuilder(c *fiber.Ctx, paginate bool) (*database.Options, error) {
	options := &database.Options{
		OrderBy: defaultCoursesOrderBy,
	}

	if paginate {
		options.Pagination = pagination.NewFromApi(c)
	}

	q := c.Query("q", "")
	if q == "" {
		return options, nil
	}

	parsed, err := queryparser.Parse(q, []string{"available", "tag", "progress"})
	if err != nil {
		return nil, err
	}

	if parsed == nil {
		return options, nil
	}

	if len(parsed.Sort) > 0 {
		options.OrderBy = parsed.Sort
	}

	if parsed.Expr == nil {
		return options, nil
	}

	options.Where = courseWhereBuilder(parsed.Expr)

	if foundProgress, ok := parsed.FoundFilters["progress"]; ok && foundProgress {
		// TODO: Make a LEFT JOIN for when courses do not have progress
		options.AdditionalJoins = append(options.AdditionalJoins,
			models.COURSE_PROGRESS_TABLE+" ON "+models.COURSE_PROGRESS_TABLE+"."+models.COURSE_PROGRESS_COURSE_ID+" = "+models.COURSE_TABLE+"."+models.BASE_ID,
		)
	}

	return options, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// courseWhereBuilder builds a squirrel.Sqlizer, for use in a WHERE clause, based on a query expression
func courseWhereBuilder(expr queryparser.QueryExpr) squirrel.Sqlizer {
	switch node := expr.(type) {
	case *queryparser.ValueExpr:
		return squirrel.Like{models.COURSE_TABLE + "." + models.COURSE_TITLE: "%" + node.Value + "%"}
	case *queryparser.FilterExpr:
		switch node.Key {
		case "available":
			value, err := cast.ToBoolE(node.Value)
			if err != nil {
				return squirrel.Expr("1=0")
			}
			return squirrel.Eq{models.COURSE_TABLE + "." + models.COURSE_AVAILABLE: value}
		case "tag":
			return buildTagsQuery([]string{node.Value})
		case "progress":
			switch strings.ToLower(node.Value) {
			case "not started":
				return squirrel.Eq{models.COURSE_PROGRESS_TABLE + "." + models.COURSE_PROGRESS_STARTED: false}
			case "started":
				return squirrel.And{
					squirrel.Eq{models.COURSE_PROGRESS_TABLE + "." + models.COURSE_PROGRESS_STARTED: true},
					squirrel.NotEq{models.COURSE_PROGRESS_TABLE + "." + models.COURSE_PROGRESS_PERCENT: 100},
				}
			case "completed":
				return squirrel.Eq{models.COURSE_PROGRESS_TABLE + "." + models.COURSE_PROGRESS_PERCENT: 100}
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
				andSlice = append(andSlice, courseWhereBuilder(child))
			}
		}

		// If we found tags, build the EXISTS subquery
		var tagCond squirrel.Sqlizer
		if len(tags) > 0 {
			tagCond = buildTagsQuery(tags)

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
			orSlice = append(orSlice, courseWhereBuilder(child))
		}

		return squirrel.Or(orSlice)
	default:
		return nil
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// buildTagsQuery builds an EXISTS squirrel.Sqlizer subquery for a list of tags
func buildTagsQuery(tags []string) squirrel.Sqlizer {
	if len(tags) == 0 {
		return squirrel.Expr("1=1")
	}

	baseQuery := squirrel.
		Select("1").
		From(models.COURSE_TAG_TABLE).
		Join(fmt.Sprintf("%s ON %s.%s = %s.%s",
			models.TAG_TABLE,
			models.TAG_TABLE, models.BASE_ID,
			models.COURSE_TAG_TABLE, models.COURSE_TAG_TAG_ID)).
		Where(fmt.Sprintf("%s.%s = %s.%s",
			models.COURSE_TAG_TABLE, models.COURSE_TAG_COURSE_ID,
			models.COURSE_TABLE, models.BASE_ID))

	if len(tags) == 1 {
		baseQuery = baseQuery.Where(squirrel.Eq{models.TAG_TABLE + "." + models.TAG_TAG: tags[0]})
	} else if len(tags) > 1 {
		baseQuery = baseQuery.
			Where(squirrel.Eq{models.TAG_TABLE + ".tag": tags}).
			GroupBy(models.COURSE_TAG_TABLE+".course_id").
			Having("COUNT(DISTINCT "+models.TAG_TABLE+".tag) = ?", len(tags))
	}

	return squirrel.Expr("EXISTS (?)", baseQuery)
}
