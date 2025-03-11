package api

import (
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/queryparser"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type logsAPI struct {
	logger *slog.Logger
	dao    *dao.DAO
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initLogRoutes initializes the log routes
func (r *Router) initLogRoutes() {
	logsAPI := logsAPI{
		logger: r.config.Logger,
		dao:    r.logDao,
	}

	logGroup := r.api.Group("/logs")
	logGroup.Get("/", protectedRoute, logsAPI.getLogs)
	logGroup.Get("/types", protectedRoute, logsAPI.getLogTypes)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *logsAPI) getLogs(c *fiber.Ctx) error {
	builderOptions := builderOptions{
		DefaultOrderBy: defaultLogsOrderBy,
		Paginate:       true,
		AllowedFilters: []string{"level", "type"},
		AfterParseHook: logsAfterParseHook,
	}

	options, err := optionsBuilder(c, builderOptions)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing query", err)
	}

	logs := []*models.Log{}
	err = api.dao.List(c.UserContext(), &logs, options)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up logs", err)
	}

	pResult, err := options.Pagination.BuildResult(logsResponseHelper(logs))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *logsAPI) getLogTypes(c *fiber.Ctx) error {
	types := types.AllLogTypes()
	return c.Status(fiber.StatusOK).JSON(types)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// tagsAfterParseHook builds the database.Options.Where based on the query expression
func logsAfterParseHook(parsed *queryparser.QueryResult, options *database.Options) {
	options.Where = logsWhereBuilder(parsed.Expr)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// logsWhereBuilder builds a squirrel.Sqlizer, for use in a WHERE clause, based on a query expression
func logsWhereBuilder(expr queryparser.QueryExpr) squirrel.Sqlizer {
	switch node := expr.(type) {
	case *queryparser.ValueExpr:
		return squirrel.Like{models.LOG_TABLE + "." + models.LOG_MESSAGE: "%" + node.Value + "%"}
	case *queryparser.FilterExpr:
		switch node.Key {
		case "level":
			return squirrel.Eq{models.LOG_TABLE + "." + models.LOG_LEVEL: node.Value}
		case "type":
			return squirrel.Eq{"JSON_EXTRACT(" + models.LOG_TABLE + "." + models.LOG_DATA + ", '$.type')": node.Value}
		default:
			return nil
		}
	case *queryparser.AndExpr:
		var andSlice []squirrel.Sqlizer
		for _, child := range node.Children {
			andSlice = append(andSlice, logsWhereBuilder(child))
		}

		return squirrel.And(andSlice)
	case *queryparser.OrExpr:
		var orSlice []squirrel.Sqlizer
		for _, child := range node.Children {
			orSlice = append(orSlice, logsWhereBuilder(child))
		}

		return squirrel.Or(orSlice)
	default:
		return nil
	}
}
