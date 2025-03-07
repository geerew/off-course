package queryparser

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import "strings"

// QueryResult represents the result of parsing.=
type QueryResult struct {
	Expr         QueryExpr
	Sort         []string
	FoundFilters map[string]bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// QueryExpr is the interface for a boolean expression
type QueryExpr interface {
	String() string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ValueExpr represents free-text
type ValueExpr struct {
	Value string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// String implements the Stringer interface for ValueExpr
func (v *ValueExpr) String() string {
	return v.Value
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// FilterExpr represents a filter token (e.g. tag:test or progress:started)
type FilterExpr struct {
	Key   string
	Value string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// String implements the Stringer interface for FilterExpr
func (f *FilterExpr) String() string {
	return f.Key + ":" + f.Value
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AndExpr represents an AND expression.
type AndExpr struct {
	Children []QueryExpr
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// String implements the Stringer interface for AndExpr
func (a *AndExpr) String() string {
	var parts []string
	for _, child := range a.Children {
		parts = append(parts, child.String())
	}
	return "(" + strings.Join(parts, " AND ") + ")"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// OrExpr represents an OR expression
type OrExpr struct {
	Children []QueryExpr
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// String implements the Stringer interface for OrExpr
func (o *OrExpr) String() string {
	var parts []string
	for _, child := range o.Children {
		parts = append(parts, child.String())
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}
