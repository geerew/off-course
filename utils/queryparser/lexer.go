package queryparser

import (
	"strings"
	"unicode"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Token represents a token with its text and whether it was quoted
type token struct {
	Text   string
	Quoted bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// tokenize tokenizes an input string while respecting quoted substrings
func tokenize(input string) []token {
	var tokens []token
	var current strings.Builder
	inQuotes := false
	for i, r := range input {
		if r == '"' {
			if inQuotes {
				tokens = append(tokens, token{Text: current.String(), Quoted: true})
				current.Reset()
				inQuotes = false
			} else {
				if current.Len() > 0 {
					tokens = append(tokens, token{Text: current.String(), Quoted: false})
					current.Reset()
				}
				inQuotes = true
			}
		} else if unicode.IsSpace(r) && !inQuotes {
			if current.Len() > 0 {
				tokens = append(tokens, token{Text: current.String(), Quoted: false})
				current.Reset()
			}
		} else if (r == '(' || r == ')') && !inQuotes {
			if current.Len() > 0 {
				tokens = append(tokens, token{Text: current.String(), Quoted: false})
				current.Reset()
			}
			tokens = append(tokens, token{Text: string(r), Quoted: false})
		} else {
			current.WriteRune(r)
		}

		if i == len(input)-1 && current.Len() > 0 {
			tokens = append(tokens, token{Text: current.String(), Quoted: inQuotes})
		}
	}

	return tokens
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// extractSortTokens extracts sort tokens, which is a token with in the format
// "sort:created_at asc"
func extractSortTokens(tokens []token) ([]token, []string) {
	var sortTokens []string
	var remaining []token
	i := 0

	for i < len(tokens) {
		tok := tokens[i]
		if strings.Contains(tok.Text, ":") {
			parts := strings.SplitN(tok.Text, ":", 2)
			key := parts[0]

			if strings.EqualFold(key, "sort") && !tok.Quoted {
				val := strings.TrimSpace(parts[1])

				if val == "" && i+1 < len(tokens) && tokens[i+1].Quoted {
					val = strings.TrimSpace(tokens[i+1].Text)
					i++
				}

				if val != "" {
					sortTokens = append(sortTokens, val)
				}

				i++
				continue
			}
		}

		remaining = append(remaining, tok)
		i++
	}

	return remaining, sortTokens
}
