package api

import (
	"net/url"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/queryparser"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type tagsAPI struct {
	logger *logger.Logger
	dao    *dao.DAO
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initTagRoutes initializes the tag routes
func (r *Router) initTagRoutes() {
	tagsAPI := tagsAPI{
		logger: r.logger.WithAPI(),
		dao:    r.dao,
	}

	tagGroup := r.api.Group("/tags")
	tagGroup.Get("", tagsAPI.getTags)
	tagGroup.Get("/names", tagsAPI.getTagNames)
	tagGroup.Get("/:name", tagsAPI.getTag)
	tagGroup.Post("", protectedRoute, tagsAPI.createTag)
	tagGroup.Put("/:id", protectedRoute, tagsAPI.updateTag)
	tagGroup.Delete("/:id", protectedRoute, tagsAPI.deleteTag)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tagsAPI) getTags(c *fiber.Ctx) error {
	builderOpts := builderOptions{
		DefaultOrderBy: defaultTagsOrderBy,
		Paginate:       true,
		AfterParseHook: tagsAfterParseHook,
	}

	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts, err := optionsBuilder(c, builderOpts, principal.UserID)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing query", err)
	}

	tags, err := api.dao.ListTags(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up tags", err)
	}

	pResult, err := dbOpts.Pagination.BuildResult(tagResponseHelper(tags))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tagsAPI) getTagNames(c *fiber.Ctx) error {
	builderOpts := builderOptions{
		DefaultOrderBy: defaultTagsOrderBy,
		Paginate:       false,
		AfterParseHook: tagsAfterParseHook,
	}

	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts, err := optionsBuilder(c, builderOpts, principal.UserID)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing query", err)
	}

	tags, err := api.dao.ListTagNames(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up tags", err)
	}

	return c.Status(fiber.StatusOK).JSON(tags)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tagsAPI) getTag(c *fiber.Ctx) error {
	name := c.Params("name")

	var err error
	name, err = url.QueryUnescape(name)

	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error decoding name parameter", err)
	}

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.TAG_TABLE_TAG: name})

	tag, err := api.dao.GetTag(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up tag", err)
	}

	if tag == nil {
		return errorResponse(c, fiber.StatusNotFound, "Tag not found", nil)
	}

	return c.Status(fiber.StatusOK).JSON(tagResponseHelper([]*models.Tag{tag})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tagsAPI) createTag(c *fiber.Ctx) error {
	req := &tagRequest{}
	if err := c.BodyParser(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	if req.Tag == "" {
		return errorResponse(c, fiber.StatusBadRequest, "A tag is required", nil)
	}

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	tag := &models.Tag{Tag: req.Tag}
	if err := api.dao.CreateTag(ctx, tag); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "Tag already exists", err)
		}
		return errorResponse(c, fiber.StatusInternalServerError, "Error creating tag", err)
	}

	return c.Status(fiber.StatusCreated).JSON(tagResponseHelper([]*models.Tag{tag})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tagsAPI) updateTag(c *fiber.Ctx) error {
	id := c.Params("id")

	tagReq := &tagRequest{}
	if err := c.BodyParser(tagReq); err != nil {
		api.logger.Warn().Err(err).Str("tag_id", id).Msg("tags: invalid update tag payload")
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.TAG_TABLE_ID: id})
	tag, err := api.dao.GetTag(ctx, dbOpts)
	if err != nil {
		api.logger.Error().Err(err).Str("tag_id", id).Msg("tags: failed to get tag for update")
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up tag", err)
	}

	if tag == nil {
		return errorResponse(c, fiber.StatusNotFound, "Tag not found", nil)
	}

	tag.Tag = tagReq.Tag

	if err := api.dao.UpdateTag(ctx, tag); err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "Tag already exists", err)
		}
		return errorResponse(c, fiber.StatusInternalServerError, "Error updating tag", err)
	}

	return c.Status(fiber.StatusOK).JSON(tagResponseHelper([]*models.Tag{tag})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *tagsAPI) deleteTag(c *fiber.Ctx) error {
	id := c.Params("id")

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.TAG_TABLE_ID: id})
	if err = api.dao.DeleteTags(ctx, dbOpts); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting tag", err)
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// tagsAfterParseHook builds the database.Options.Where based on the query expression
func tagsAfterParseHook(parsed *queryparser.QueryResult, options *database.Options, _ string) {
	if len(parsed.FreeText) == 0 {
		return
	}

	if slices.Contains(parsed.Sort, "special") {
		// During special ordering, filter by the first filter (there should only be one) and
		// order by a case expression
		filter := strings.ToLower(parsed.FreeText[0])

		options.Where = squirrel.Like{models.TAG_TABLE_TAG: "%" + filter + "%"}

		caseExpr := squirrel.Case().
			When(squirrel.Eq{"LOWER(" + models.TAG_TABLE_TAG + ")": filter}, "0").
			When(squirrel.Like{"LOWER(" + models.TAG_TABLE_TAG + ")": filter + "%"}, "1").
			When(squirrel.Like{"LOWER(" + models.TAG_TABLE_TAG + ")": "%" + filter + "%"}, "2")

		sql, args, _ := caseExpr.ToSql()
		options.OrderByClause = squirrel.Expr(sql+", "+defaultTagsOrderBy[0], args...)

		options.OrderBy = []string{}
	} else {
		options.Where = tagsWhereBuilder(parsed.Expr)
	}

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// tagsWhereBuilder builds a squirrel.Sqlizer, for use in a WHERE clause
//
// TODO Support count filter (ex HAVING COUNT(courses_tags.id) > 1)
func tagsWhereBuilder(expr queryparser.QueryExpr) squirrel.Sqlizer {
	switch node := expr.(type) {
	case *queryparser.ValueExpr:
		return squirrel.Like{models.TAG_TABLE_TAG: "%" + node.Value + "%"}
	case *queryparser.AndExpr:
		var andSlice []squirrel.Sqlizer
		for _, child := range node.Children {
			andSlice = append(andSlice, tagsWhereBuilder(child))
		}

		return squirrel.And(andSlice)
	case *queryparser.OrExpr:
		var orSlice []squirrel.Sqlizer
		for _, child := range node.Children {
			orSlice = append(orSlice, tagsWhereBuilder(child))
		}

		return squirrel.Or(orSlice)
	default:
		return nil
	}
}
