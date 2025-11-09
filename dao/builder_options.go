package dao

// TODO Tidy to to make this more consistent. Use the builder pattern for all options
import (
	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/utils/pagination"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// join defines a join in a SQL query (internal to DAO)
type join struct {
	Type      string // "JOIN", "LEFT JOIN", "RIGHT JOIN", etc.
	Table     string // Table to join with
	Condition string // ON condition
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// builderOptions defines builder options for a database query
type builderOptions struct {
	// The name of the table to query
	//
	// Example: "table1"
	Table string

	// Columns to select
	//
	// Example: []string{"id", "title", "created_at"}
	Columns []string

	// Data is a key/value map of data to insert into the table during an INSERT or UPDATE operation
	//
	// Example: map[string]interface{}{"id": "123", "title": "Test", "created_at": time.Now()}
	Data map[string]interface{}

	// Columns to group by
	//
	// Example: []string{table1.id}
	GroupBy []string

	// Having clause (used with GroupBy)
	//
	// Example: squirrel.Eq{"COUNT(table1.id)": 1}
	Having squirrel.Sqlizer

	// Joins to use in SELECT queries
	Joins []join

	// Suffix is raw SQL to append to the query
	//
	Suffix string

	// Used to paginate the results
	Pagination *pagination.Pagination

	// Limit for the number of results to return
	Limit int

	// Whether to use REPLACE INTO instead of INSERT INTO
	Replace bool

	DbOpts *Options
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// newBuilderOptions creates an new builderOptions instance with the table name set
func newBuilderOptions(table string) *builderOptions {
	return &builderOptions{Table: table}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithColumns sets the columns to select
func (o *builderOptions) WithColumns(columns ...string) *builderOptions {
	o.Columns = append(o.Columns, columns...)
	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithData sets the data to insert into the table
func (o *builderOptions) WithData(data map[string]interface{}) *builderOptions {
	// Merge the new data with existing data
	if o.Data == nil {
		o.Data = make(map[string]interface{})
	}

	for key, value := range data {
		o.Data[key] = value
	}

	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithGroupBy appends GROUP BY fields
func (o *builderOptions) WithGroupBy(fields ...string) *builderOptions {
	o.GroupBy = append(o.GroupBy, fields...)
	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithHaving sets the HAVING clause using a squirrel.Sqlizer
func (o *builderOptions) WithHaving(pred squirrel.Sqlizer) *builderOptions {
	o.Having = pred
	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AddJoin appends a join clause
func (o *builderOptions) WithJoin(table, condition string) *builderOptions {
	o.Joins = append(o.Joins, join{Type: "JOIN", Table: table, Condition: condition})
	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithLeftJoin appends a LEFT JOIN clause
func (o *builderOptions) WithLeftJoin(table, onCondition string) *builderOptions {
	o.Joins = append(o.Joins, join{Type: "LEFT JOIN", Table: table, Condition: onCondition})
	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AddRightJoin appends a RIGHT JOIN clause
func (o *builderOptions) WithRightJoin(table, condition string) *builderOptions {
	o.Joins = append(o.Joins, join{Type: "RIGHT JOIN", Table: table, Condition: condition})
	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithSuffix registers raw SQL to append via squirrel.Suffix(...)
func (o *builderOptions) WithSuffix(sql string) *builderOptions {
	o.Suffix = sql
	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithPagination sets the pagination options
func (o *builderOptions) WithPagination(p *pagination.Pagination) *builderOptions {
	o.Pagination = p
	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithLimit sets the limit for the number of results to return
func (o *builderOptions) WithLimit(limit int) *builderOptions {
	o.Limit = limit
	return o
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithReplace sets the builder to use REPLACE INTO instead of INSERT INTO
func (o *builderOptions) WithReplace() *builderOptions {
	o.Replace = true
	return o
}

func (o *builderOptions) SetDbOpts(opts *Options) *builderOptions {
	if opts == nil {
		o.DbOpts = NewOptions()
	} else {
		o.DbOpts = opts
	}

	return o
}
