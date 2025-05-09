package database

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

	// GROUP BY
	//
	// Example: []string{table1.id}
	GroupBy []string

	// HAVING (used with GROUP BY)
	//
	// Example: squirrel.Eq{"COUNT(table1.id)": 1}
	Having squirrel.Sqlizer

	// Joins to use in SELECT queries
	Joins []Join

	// Additional filters to use in SELECT queries for relations
	//
	// Example: map[string]map[string]interface{}{
	//   "table1": {
	//     "user": "test-user",
	//   },
	RelationFilters map[string]map[string]interface{}

	// Used to paginate the results
	Pagination *pagination.Pagination

	// Use REPLACE instead of INSERT
	Replace bool
}

// Helper methods to add different types of joins
func (o *Options) AddJoin(table, condition string) *Options {
	o.Joins = append(o.Joins, Join{Type: "JOIN", Table: table, Condition: condition})
	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (o *Options) AddLeftJoin(table, condition string) *Options {
	o.Joins = append(o.Joins, Join{Type: "LEFT JOIN", Table: table, Condition: condition})
	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (o *Options) AddRightJoin(table, condition string) *Options {
	o.Joins = append(o.Joins, Join{Type: "RIGHT JOIN", Table: table, Condition: condition})
	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AddRelationFilter adds a filter for a specific relation field
func (o *Options) AddRelationFilter(relation, field string, value interface{}) *Options {
	if o.RelationFilters == nil {
		o.RelationFilters = make(map[string]map[string]interface{})
	}

	if o.RelationFilters[relation] == nil {
		o.RelationFilters[relation] = make(map[string]interface{})
	}

	o.RelationFilters[relation][field] = value
	return o
}
