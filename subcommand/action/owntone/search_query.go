package owntone

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// SearchQuery represents a parsed music search query.
type SearchQuery struct {
	Terms  []string
	Types  []SearchType
	Limit  int
	Offset int
}

// TypeArray returns the list of search types to use, defaulting to all types if none are specified.
func (sq SearchQuery) TypeArray() []SearchType {
	if sq.Types == nil {
		return []SearchType{artist, album, track, genre}
	}
	return sq.Types
}

const limitPrefix = "limit:"
const offsetPrefix = "offset:"
const typePrefix = "type:"
const keywordPrefix = "keyword:"

// Parse parses a search query string into a SearchQuery struct.
// Handles comma-separated input and strips "keyword:" prefix that DSPy may prepend erroneously.
func Parse(target string) *SearchQuery {
	normalized := strings.ReplaceAll(target, ",", " ")
	split := strings.Fields(normalized)
	var queries []string
	var types []SearchType
	limit := -1
	offset := -1
	for _, term := range split {
		switch {
		case strings.HasPrefix(term, keywordPrefix):
			if value := term[len(keywordPrefix):]; value != "" {
				queries = append(queries, value)
			}
		case strings.HasPrefix(term, limitPrefix):
			value := term[len(limitPrefix):]
			i, err := strconv.Atoi(value)
			if err == nil {
				limit = i
			} else {
				queries = append(queries, term)
			}
		case strings.HasPrefix(term, offsetPrefix):
			value := term[len(offsetPrefix):]
			i, err := strconv.Atoi(value)
			if err == nil {
				offset = i
			} else {
				queries = append(queries, term)
			}
		case strings.HasPrefix(term, typePrefix):
			value := term[len(typePrefix):]
			st, err := SearchTypeFromString(value)
			if err == nil {
				types = append(types, st)
			} else {
				queries = append(queries, term)
			}
		default:
			queries = append(queries, term)
		}
	}
	return &SearchQuery{Terms: queries, Limit: limit, Offset: offset, Types: types}
}

func buildSearchKeywords(originalKeyword string, normalizedKeyword string) []string {
	var keywords []string
	for _, keyword := range []string{originalKeyword, normalizedKeyword} {
		trimmed := strings.TrimSpace(keyword)
		if trimmed == "" || slices.Contains(keywords, trimmed) {
			continue
		}
		keywords = append(keywords, trimmed)
	}
	return keywords
}

func buildSearchExpression(keywords []string, types []SearchType) string {
	fields := expressionFields(types)
	if len(fields) == 0 || len(keywords) == 0 {
		return ""
	}

	keywordClauses := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		fieldClauses := make([]string, 0, len(fields))
		escaped := escapeExpressionValue(keyword)
		for _, field := range fields {
			fieldClauses = append(fieldClauses, fmt.Sprintf("%s includes \"%s\"", field, escaped))
		}
		keywordClauses = append(keywordClauses, fmt.Sprintf("(%s)", strings.Join(fieldClauses, " or ")))
	}
	return strings.Join(keywordClauses, " or ")
}

func buildSearchExpressionAND(keywords []string, types []SearchType) string {
	fields := expressionFields(types)
	if len(fields) == 0 || len(keywords) == 0 {
		return ""
	}

	keywordClauses := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		fieldClauses := make([]string, 0, len(fields))
		escaped := escapeExpressionValue(keyword)
		for _, field := range fields {
			fieldClauses = append(fieldClauses, fmt.Sprintf("%s includes \"%s\"", field, escaped))
		}
		keywordClauses = append(keywordClauses, fmt.Sprintf("(%s)", strings.Join(fieldClauses, " or ")))
	}
	return strings.Join(keywordClauses, " and ")
}

func expressionFields(types []SearchType) []string {
	if len(types) == 0 {
		types = []SearchType{artist, album, track, genre}
	}

	seen := map[string]struct{}{}
	fields := []string{}
	appendFields := func(candidates ...string) {
		for _, candidate := range candidates {
			if _, ok := seen[candidate]; ok {
				continue
			}
			seen[candidate] = struct{}{}
			fields = append(fields, candidate)
		}
	}

	for _, searchType := range types {
		switch searchType {
		case artist:
			appendFields("artist")
		case album:
			appendFields("album")
		case track:
			appendFields("title", "artist", "album")
		case genre:
			appendFields("genre")
		}
	}
	return fields
}

func escapeExpressionValue(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
	)
	return replacer.Replace(value)
}
