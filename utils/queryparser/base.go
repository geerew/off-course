package queryparser

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Parse a query string into an AST of sort tokens, free-text and allowed filters
func Parse(q string, allowedFilters []string) (*QueryResult, error) {
	allTokens := tokenize(q)
	remainingTokens, sortTokens := extractSortTokens(allTokens)

	ast := newASTParser(remainingTokens, allowedFilters)
	expr, err := ast.parseOr()
	if err != nil {
		return nil, err
	}

	return &QueryResult{
		Expr:         expr,
		Sort:         sortTokens,
		FoundFilters: ast.FoundFilters,
	}, nil
}
