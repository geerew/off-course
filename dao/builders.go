package dao

import (
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
)

func countBuilder(builderOpts builderOptions) (string, []interface{}, error) {
	if builderOpts.Table == "" {
		return "", nil, fmt.Errorf("builderOpts.Table cannot be empty")
	}

	builder := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question)

	commonBuilder := builder.
		Select().
		From(builderOpts.Table)

	// Builder joins
	commonBuilder = applyJoins(commonBuilder, builderOpts.Joins)

	if builderOpts.DbOpts != nil {
		if builderOpts.DbOpts.Where != nil {
			commonBuilder = commonBuilder.Where(builderOpts.DbOpts.Where)
		}

		// Additional joins
		commonBuilder = applyJoins(commonBuilder, builderOpts.DbOpts.Joins)
	}

	// Fast path when no GROUP BY/HAVING
	if len(builderOpts.GroupBy) == 0 || builderOpts.Having == nil {
		return commonBuilder.
			Columns("COUNT(DISTINCT " + builderOpts.Table + ".id)").
			ToSql()
	}

	inner := commonBuilder.Columns(builderOpts.Table + ".id")

	if len(builderOpts.GroupBy) > 0 {
		inner = inner.GroupBy(builderOpts.GroupBy...)
	}

	if builderOpts.Having != nil {
		inner = inner.Having(builderOpts.Having)
	}

	return builder.
		Select("COUNT(*)").
		FromSelect(inner, "sub").
		ToSql()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// // countBuilder builds a query for counting distinct records in a table
// func countBuilder(builderOpts builderOptions) (string, []interface{}, error) {
// 	if builderOpts.Table == "" {
// 		return "", nil, fmt.Errorf("builderOpts.Table cannot be empty")
// 	}

// 	builder := squirrel.
// 		StatementBuilder.
// 		Select("COUNT(DISTINCT " + builderOpts.Table + ".id)").
// 		From(builderOpts.Table)

// 	builder = applyJoins(builder, builderOpts.Joins)

// 	builder = builder.GroupBy(builderOpts.GroupBy...)

// 	if builderOpts.Having != nil {
// 		builder = builder.Having(builderOpts.Having)
// 	}

// 	// When there is a GroupBy and Having clause, we need to wrap the query in a subquery
// 	if len(builderOpts.GroupBy) > 0 && builderOpts.Having != nil {
// 		builder = squirrel.StatementBuilder.
// 			PlaceholderFormat(squirrel.Question).
// 			Select("COUNT(*)").
// 			FromSelect(builder, "sub")
// 	}

// 	if builderOpts.DbOpts != nil {
// 		builder = builder.Where(builderOpts.DbOpts.Where)

// 		builder = applyJoins(builder, builderOpts.DbOpts.Joins)

// 		if builderOpts.DbOpts.OrderByClause != nil {
// 			builder = builder.OrderByClause(builderOpts.DbOpts.OrderByClause)
// 		}
// 	}

// 	return builder.ToSql()
// }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// selectBuilder builds a query for selecting data from the database
func selectBuilder(builderOpts builderOptions) (string, []interface{}, error) {
	if builderOpts.Table == "" {
		return "", nil, fmt.Errorf("builderOpts.Table cannot be empty")
	}

	if len(builderOpts.Columns) == 0 {
		return "", nil, fmt.Errorf("builderOpts.Columns cannot be empty")
	}

	builder := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select(builderOpts.Columns...).
		From(builderOpts.Table)

	builder = applyJoins(builder, builderOpts.Joins)

	builder = builder.GroupBy(builderOpts.GroupBy...)

	if builderOpts.Having != nil {
		builder = builder.Having(builderOpts.Having)
	}

	// Database options
	if builderOpts.DbOpts != nil {
		builder = builder.Where(builderOpts.DbOpts.Where)

		builder = applyJoins(builder, builderOpts.DbOpts.Joins)

		builder = builder.OrderBy(builderOpts.DbOpts.OrderBy...)

		if builderOpts.DbOpts.OrderByClause != nil {
			builder = builder.OrderByClause(builderOpts.DbOpts.OrderByClause)
		}

		if builderOpts.DbOpts.Pagination != nil {
			builder = builder.
				Offset(uint64(builderOpts.DbOpts.Pagination.Offset())).
				Limit(uint64(builderOpts.DbOpts.Pagination.Limit()))
		}
	}

	return builder.ToSql()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// insertBuilder builds a query for inserting or replacing records in the database
func insertBuilder(builderOpts builderOptions) (string, []interface{}, error) {
	if builderOpts.Table == "" {
		return "", nil, fmt.Errorf("builderOpts.Table cannot be empty")
	}

	if len(builderOpts.Data) == 0 {
		return "", nil, fmt.Errorf("builderOpts.InsertData cannot be empty")
	}

	var builder squirrel.InsertBuilder
	if builderOpts.Replace {
		builder = squirrel.StatementBuilder.Replace(builderOpts.Table)
	} else {
		builder = squirrel.StatementBuilder.Insert(builderOpts.Table)
	}

	builder = builder.SetMap(builderOpts.Data)

	builder = builder.Suffix(builderOpts.Suffix)

	return builder.ToSql()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func updateBuilder(builderOpts builderOptions) (string, []interface{}, error) {
	if builderOpts.Table == "" {
		return "", nil, fmt.Errorf("builderOpts.Table cannot be empty")
	}

	if len(builderOpts.Data) == 0 {
		return "", nil, fmt.Errorf("builderOpts.Data cannot be empty")
	}

	if builderOpts.DbOpts == nil || builderOpts.DbOpts.Where == nil {
		return "", nil, fmt.Errorf("builderOpts.DbOpts.Where cannot be empty")
	}

	builder := squirrel.
		StatementBuilder.
		Update(builderOpts.Table).
		SetMap(builderOpts.Data).
		Where(builderOpts.DbOpts.Where)

	return builder.ToSql()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// deleteBuilder builds a query for deleting records from the database
func deleteBuilder(builderOpts builderOptions) (string, []interface{}, error) {
	if builderOpts.Table == "" {
		return "", nil, fmt.Errorf("builderOpts.Table cannot be empty")
	}

	builder := squirrel.
		StatementBuilder.
		Delete(builderOpts.Table)

	if builderOpts.DbOpts != nil && builderOpts.DbOpts.Where != nil {
		builder = builder.Where(builderOpts.DbOpts.Where)
	}

	return builder.ToSql()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func applyJoins(builder squirrel.SelectBuilder, joins []database.Join) squirrel.SelectBuilder {
	for _, join := range joins {
		clause := fmt.Sprintf("%s ON %s", join.Table, join.Condition)
		switch join.Type {
		case "LEFT JOIN":
			builder = builder.LeftJoin(clause)
		case "RIGHT JOIN":
			builder = builder.RightJoin(clause)
		case "JOIN", "INNER JOIN", "":
			builder = builder.Join(clause)
		}
	}
	return builder
}
