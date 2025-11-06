package api

import (
	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/queryparser"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type logsAPI struct {
	r *Router
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initLogRoutes initializes the log routes
func (r *Router) initLogRoutes() {
	logsAPI := logsAPI{
		r: r,
	}

	g := r.apiGroup("logs")
	g.Get("/", protectedRoute, logsAPI.getLogs)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *logsAPI) getLogs(c *fiber.Ctx) error {
	builderOpts := builderOptions{
		DefaultOrderBy: defaultLogsOrderBy,
		Paginate:       true,
		AllowedFilters: []string{"level", "type", "component"},
		AfterParseHook: logsAfterParseHook,
	}

	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts, err := optionsBuilder(c, builderOpts, principal.UserID)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing query", err)
	}

	logs, err := api.r.logDao.ListLogs(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up logs", err)
	}

	pResult, err := dbOpts.Pagination.BuildResult(logsResponseHelper(logs))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// tagsAfterParseHook builds the database.Options.Where based on the query expression
func logsAfterParseHook(parsed *queryparser.QueryResult, options *database.Options, _ string) {
	options.Where = logsWhereBuilder(parsed.Expr)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// logsWhereBuilder builds a squirrel.Sqlizer, for use in a WHERE clause, based on a query expression
func logsWhereBuilder(expr queryparser.QueryExpr) squirrel.Sqlizer {
	switch node := expr.(type) {
	case *queryparser.ValueExpr:
		return squirrel.Like{models.LOG_TABLE_MESSAGE: "%" + node.Value + "%"}
	case *queryparser.FilterExpr:
		switch node.Key {
		case "level":
			return squirrel.Eq{models.LOG_TABLE_LEVEL: node.Value}
		case "type":
			return squirrel.Eq{"JSON_EXTRACT(" + models.LOG_TABLE_DATA + ", '$.type')": node.Value}
		case "component":
			return squirrel.Eq{"JSON_EXTRACT(" + models.LOG_TABLE_DATA + ", '$.component')": node.Value}
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
