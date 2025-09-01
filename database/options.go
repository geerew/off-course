package database

// TODO Tidy to to make this more consistent. Use the builder pattern for all options
import (
	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/utils/pagination"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Join defines a join in a SQL query
type Join struct {
	Type      string // "JOIN", "LEFT JOIN", "RIGHT JOIN", etc.
	Table     string // Table to join with
	Condition string // ON condition
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Options defines optional params for a database query
type Options struct {
	// ORDER BY
	//
	// Example: []string{"id DESC", "title ASC"}
	OrderBy []string

	// ORDER BY (clause)
	//
	// Example: []string{"id DESC", "title ASC"}
	OrderByClause squirrel.Sqlizer

	// Any valid squirrel WHERE expression
	//
	// Examples:
	//
	//   EQ:   squirrel.Eq{"id": "123"}
	//   IN:   squirrel.Eq{"id": []string{"123", "456"}}
	//   OR:   squirrel.Or{squirrel.Expr("id = ?", "123"), squirrel.Expr("id = ?", "456")}
	//   AND:  squirrel.And{squirrel.Eq{"id": "123"}, squirrel.Eq{"title": "devops"}}
	//   LIKE: squirrel.Like{"title": "%dev%"}
	//   NOT:  squirrel.NotEq{"id": "123"}
	Where squirrel.Sqlizer

	// IncludeProgress indicates whether to include course/asset progress
	// when performing a query
	IncludeProgress bool

	// IncludeAssetMetadata indicates whether to include asset metadata
	// when performing a query
	IncludeAssetVideoMetadata bool

	// GROUP BY
	//
	// Example: []string{table1.id}
	// TODO REMOVE
	GroupBy []string

	// HAVING (used with GROUP BY)
	//
	// Example: squirrel.Eq{"COUNT(table1.id)": 1}
	// TODO REMOVE
	Having squirrel.Sqlizer

	// Joins to use in SELECT queries
	Joins []Join

	// Additional filters to use in SELECT queries for relations
	//
	// Example: map[string]map[string]interface{}{
	//   "table1": {
	//     "user": "test-user",
	//   },
	// TODO REMOVE
	RelationFilters map[string]map[string]interface{}

	// Exclude relations from the query
	// TODO REMOVE
	ExcludeRelations []string

	// Used to paginate the results
	Pagination *pagination.Pagination

	// Use REPLACE instead of INSERT
	// TODO REMOVE
	Replace bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewOptions creates an empty Options builder
func NewOptions() *Options {
	return &Options{
		RelationFilters: make(map[string]map[string]interface{}),
	}
}

// WithOrderBy appends ORDER BY fields
func (o *Options) WithOrderBy(fields ...string) *Options {
	o.OrderBy = append(o.OrderBy, fields...)
	return o
}

func (o *Options) OverrideOrderBy(fields ...string) *Options {
	o.OrderBy = fields
	return o
}

// WithOrderByClause sets a custom ORDER BY clause
//
// Use only if you need a complex ORDER BY that cannot be expressed with WithOrderBy
func (o *Options) WithOrderByClause(clause squirrel.Sqlizer) *Options {
	o.OrderByClause = clause
	return o
}

// WithWhere sets the WHERE clause using a squirrel.Sqlizer
func (o *Options) WithWhere(pred squirrel.Sqlizer) *Options {
	o.Where = pred
	return o
}

// WithGroupBy appends GROUP BY fields
func (o *Options) WithGroupBy(fields ...string) *Options {
	o.GroupBy = append(o.GroupBy, fields...)
	return o
}

// WithHaving sets the HAVING clause using a squirrel.Sqlizer
func (o *Options) WithHaving(pred squirrel.Sqlizer) *Options {
	o.Having = pred
	return o
}

// WithJoin appends a join clause
func (o *Options) WithJoin(table, condition string) *Options {
	o.Joins = append(o.Joins, Join{Type: "JOIN", Table: table, Condition: condition})
	return o
}

// WithLeftJoin appends a LEFT JOIN clause
func (o *Options) WithLeftJoin(table, condition string) *Options {
	o.Joins = append(o.Joins, Join{Type: "LEFT JOIN", Table: table, Condition: condition})
	return o
}

// WithRightJoin appends a RIGHT JOIN clause
func (o *Options) WithRightJoin(table, condition string) *Options {
	o.Joins = append(o.Joins, Join{Type: "RIGHT JOIN", Table: table, Condition: condition})
	return o
}

// WithRelationFilter adds a filter for a specific relation and field
// TODO rework
func (o *Options) WithRelationFilter(relation, field string, value interface{}) *Options {
	if o.RelationFilters == nil {
		o.RelationFilters = make(map[string]map[string]interface{})
	}

	if o.RelationFilters[relation] == nil {
		o.RelationFilters[relation] = make(map[string]interface{})
	}

	o.RelationFilters[relation][field] = value
	return o
}

// WithExcludeRelations appends relation names to exclude from queries
func (o *Options) WithExcludeRelations(relations ...string) *Options {
	o.ExcludeRelations = append(o.ExcludeRelations, relations...)
	return o
}

// WithPagination sets the pagination options
func (o *Options) WithPagination(p *pagination.Pagination) *Options {
	o.Pagination = p
	return o
}

// UseReplace toggles INSERT to use REPLACE instead
//
// TODO REMOVE
func (o *Options) UseReplace() *Options {
	o.Replace = true
	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithProgress enables progress inclusion in queries
func (o *Options) WithProgress() *Options {
	o.IncludeProgress = true
	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithAssetVideoMetadata enables asset video metadata inclusion in queries
func (o *Options) WithAssetVideoMetadata() *Options {
	o.IncludeAssetVideoMetadata = true
	return o
}
