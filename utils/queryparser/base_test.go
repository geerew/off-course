package queryparser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var allowed = []string{"available", "tag", "progress"}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestParse_Empty(t *testing.T) {
	q := ""
	result, err := Parse(q, allowed)
	require.NoError(t, err)
	require.Nil(t, result.Expr)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TODO: Test open ended brackets
// TODO: Test open ended quotes
func TestParse_EdgeCases(t *testing.T) {
	t.Run("missing value", func(t *testing.T) {
		q := "course 1 AND progress: OR progress:started"
		result, err := Parse(q, allowed)
		require.NoError(t, err)
		require.NotNil(t, result.Expr)

		require.Equal(t, "(course 1 OR progress:started)", result.Expr.String())
		require.Empty(t, result.Sort)
		require.False(t, result.FoundFilters["available"])
		require.False(t, result.FoundFilters["tag"])
		require.True(t, result.FoundFilters["progress"])
	})

	t.Run("multiple operators", func(t *testing.T) {
		q := `AND "" OR "   " AND AND course 1 OR OR course 2`
		result, err := Parse(q, allowed)
		require.NoError(t, err)
		require.NotNil(t, result)

		require.NotNil(t, result.Expr)
		require.Equal(t, "(course 1 OR course 2)", result.Expr.String())
		require.Empty(t, result.Sort)
		require.False(t, result.FoundFilters["available"])
		require.False(t, result.FoundFilters["tag"])
		require.False(t, result.FoundFilters["progress"])
	})

	t.Run("case", func(t *testing.T) {
		q := `course 1 or "course 2" CANDY OR BORE`
		result, err := Parse(q, allowed)
		require.NoError(t, err)
		require.NotNil(t, result)

		require.Equal(t, "((course 1 or AND course 2 AND CANDY) OR BORE)", result.Expr.String())
		require.Empty(t, result.Sort)
		require.False(t, result.FoundFilters["available"])
		require.False(t, result.FoundFilters["tag"])
		require.False(t, result.FoundFilters["progress"])
	})

	t.Run("unbalanced quotes", func(t *testing.T) {
		q := `"course 1 AND course 2`
		result, err := Parse(q, allowed)
		require.NoError(t, err)
		require.NotNil(t, result)

		require.NotNil(t, result.Expr)
		require.IsType(t, &ValueExpr{}, result.Expr)
		require.Equal(t, "course 1 AND course 2", result.Expr.String())
		require.Empty(t, result.Sort)
		require.False(t, result.FoundFilters["available"])
		require.False(t, result.FoundFilters["tag"])
		require.False(t, result.FoundFilters["progress"])
	})

	t.Run("unbalanced brackets", func(t *testing.T) {
		q := "(course 1 AND course 2"
		res, err := Parse(q, allowed)
		require.Error(t, err)
		require.EqualError(t, err, "expected ')'")
		require.Nil(t, res)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestParse_Sort(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		q := `sort:"created_at asc" sort:"id desc" sort:title`
		result, err := Parse(q, allowed)
		require.NoError(t, err)
		require.NotNil(t, result)

		require.Nil(t, result.Expr)
		require.Equal(t, []string{"created_at asc", "id desc", "title"}, result.Sort)
		require.False(t, result.FoundFilters["available"])
		require.False(t, result.FoundFilters["tag"])
		require.False(t, result.FoundFilters["progress"])
	})

	t.Run("empty", func(t *testing.T) {
		q := `sort:"    " sort:"" sort:`
		result, err := Parse(q, allowed)
		require.NoError(t, err)
		require.NotNil(t, result)

		require.Nil(t, result.Expr)
		require.Empty(t, result.Sort)
		require.False(t, result.FoundFilters["available"])
		require.False(t, result.FoundFilters["tag"])
		require.False(t, result.FoundFilters["progress"])
	})

	t.Run("mixed", func(t *testing.T) {
		q := `course 1 sort:"created_at asc" tag:test sort:"id desc" sort:title`
		result, err := Parse(q, allowed)
		require.NoError(t, err)
		require.NotNil(t, result)

		require.Equal(t, "(course 1 AND tag:test)", result.Expr.String())
		require.Equal(t, []string{"created_at asc", "id desc", "title"}, result.Sort)
		require.False(t, result.FoundFilters["available"])
		require.False(t, result.FoundFilters["progress"])
		require.True(t, result.FoundFilters["tag"])
	})

	t.Run("quoted", func(t *testing.T) {
		q := `sort:"created_at asc" "sort: test" sort:"title" course 1`
		result, err := Parse(q, allowed)
		require.NoError(t, err)
		require.NotNil(t, result)

		require.Equal(t, "(sort: test AND course 1)", result.Expr.String())
		require.Equal(t, []string{"created_at asc", "title"}, result.Sort)
		require.False(t, result.FoundFilters["available"])
		require.False(t, result.FoundFilters["tag"])
		require.False(t, result.FoundFilters["progress"])
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestParse_FreeText(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		q := `course 1 OR course 2 AND "course 3" OR course 4 "course 5"`
		result, err := Parse(q, allowed)
		require.NoError(t, err)
		require.NotNil(t, result)

		require.Equal(t, "((course 1 OR (course 2 AND course 3)) OR (course 4 AND course 5))", result.Expr.String())
		require.Empty(t, result.Sort)
		require.False(t, result.FoundFilters["available"])
		require.False(t, result.FoundFilters["tag"])
		require.False(t, result.FoundFilters["progress"])
	})

	t.Run("mixed", func(t *testing.T) {
		q := `course 1 AND course 2 OR "course 3" OR available:true`
		result, err := Parse(q, allowed)
		require.NoError(t, err)
		require.NotNil(t, result)

		require.Equal(t, "(((course 1 AND course 2) OR course 3) OR available:true)", result.Expr.String())
		require.Empty(t, result.Sort)
		require.True(t, result.FoundFilters["available"])
		require.False(t, result.FoundFilters["tag"])
		require.False(t, result.FoundFilters["progress"])
	})

	t.Run("pretend filter", func(t *testing.T) {
		q := `course:1 AND tag:a OR "course: a b"`
		result, err := Parse(q, allowed)
		require.NoError(t, err)
		require.NotNil(t, result)

		require.Equal(t, "((course:1 AND tag:a) OR course: a b)", result.Expr.String())
		require.Empty(t, result.Sort)
		require.False(t, result.FoundFilters["available"])
		require.True(t, result.FoundFilters["tag"])
		require.False(t, result.FoundFilters["progress"])
	})

	t.Run("empty", func(t *testing.T) {
		q := `"" AND "   " OR tag:1`
		result, err := Parse(q, allowed)
		require.NoError(t, err)
		require.NotNil(t, result)

		require.Equal(t, "tag:1", result.Expr.String())
		require.Empty(t, result.Sort)
		require.True(t, result.FoundFilters["tag"])
		require.False(t, result.FoundFilters["available"])
		require.False(t, result.FoundFilters["progress"])
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestParse_Filters(t *testing.T) {
	q := "available:true AND tag:go OR progress:completed"
	result, err := Parse(q, allowed)
	require.NoError(t, err)

	require.Equal(t, "((available:true AND tag:go) OR progress:completed)", result.Expr.String())
	require.Empty(t, result.Sort)
	require.True(t, result.FoundFilters["available"])
	require.True(t, result.FoundFilters["tag"])
	require.True(t, result.FoundFilters["progress"])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestParse_ComplexParentheses(t *testing.T) {
	q := "(course 1 AND (progress:started OR progress:completed)) OR (course 2 AND progress:completed)"
	result, err := Parse(q, allowed)
	require.NoError(t, err)

	expected := "((course 1 AND (progress:started OR progress:completed)) OR (course 2 AND progress:completed))"
	require.Empty(t, result.Sort)
	require.Equal(t, expected, result.Expr.String())
	require.False(t, result.FoundFilters["available"])
	require.False(t, result.FoundFilters["tag"])
	require.True(t, result.FoundFilters["progress"])
}
