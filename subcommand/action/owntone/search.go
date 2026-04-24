package owntone

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/johtani/smarthome/internal/resolver"
	"github.com/johtani/smarthome/subcommand/action"
	"golang.org/x/text/unicode/norm"
)

const defaultMusicIntentConfidenceThreshold = 0.75

// SearchAndPlayAction represents an action to search for music and play it on Owntone.
type SearchAndPlayAction struct {
	name                           string
	c                              *Client
	musicIntentResolver            MusicIntentResolver
	musicIntentConfidenceThreshold float64
}

// SearchAndPlayActionOption customizes SearchAndPlayAction behavior.
type SearchAndPlayActionOption func(*SearchAndPlayAction)

// WithMusicIntentResolver sets the music intent resolver.
func WithMusicIntentResolver(r MusicIntentResolver) SearchAndPlayActionOption {
	return func(a *SearchAndPlayAction) {
		a.musicIntentResolver = r
	}
}

// WithMusicIntentConfidenceThreshold sets the confidence threshold for auto play.
func WithMusicIntentConfidenceThreshold(threshold float64) SearchAndPlayActionOption {
	return func(a *SearchAndPlayAction) {
		if threshold >= 0 {
			a.musicIntentConfidenceThreshold = threshold
		}
	}
}

func appendMessage(items Items, label string, msg []string, uris []string, loopFunc func(item SearchItem, msg []string) ([]string, []string)) ([]string, []string) {
	if items.Total > 0 {
		msg = append(msg, fmt.Sprintf("# %s (%d)", label, items.Total))
		for _, item := range items.Items {
			msg, uris = loopFunc(item, msg)
		}
	}
	return msg, uris
}

// Run executes the SearchAndPlayAction.
func (a SearchAndPlayAction) Run(ctx context.Context, query string) (string, error) {
	ctx, span := action.StartRunSpan(ctx, "owntone", "SearchAndPlayAction.Run", query)
	defer span.End()

	searchQuery := Parse(query)
	originalKeyword := strings.Join(searchQuery.Terms, " ")
	searchKeyword := normalizeSearchKeyword(originalKeyword, a.c.config.SearchAliases)
	if strings.TrimSpace(searchKeyword) == "" {
		searchKeyword = originalKeyword
	}
	keywords := buildSearchKeywords(originalKeyword, searchKeyword)
	types := searchQuery.TypeArray()

	if intent, path, ok := a.resolveMusicIntent(ctx, originalKeyword); ok {
		strictKeywords := intent.StrictKeywords()
		if len(strictKeywords) > 0 {
			strictExpression := buildSearchExpressionAND(strictKeywords, types)
			if strictExpression != "" {
				strictResult, err := a.c.SearchByExpression(ctx, strictExpression, types, searchQuery.Limit)
				if err == nil && totalSearchResultCount(strictResult) > 0 {
					resolver.RecordDecision(ctx, resolver.DecisionRecord{
						InputTextHash:   hashInputText(originalKeyword),
						ResolverPath:    "music_intent_strict",
						ResolvedCommand: "search and play",
						ResolvedArgs:    query,
						LLMModel:        intent.Model,
					})
					if a.shouldSuggestCandidates(intent, strictResult) {
						resolver.RecordExecution(ctx, resolver.ExecutionRecord{
							ExecutionStatus:  "candidate_only",
							ResolvedCommand:  "search and play",
							ResolvedArgs:     query,
							ResolverPathHint: path,
						})
						return buildCandidateMessage(strictResult, intent), nil
					}
					resolver.RecordExecution(ctx, resolver.ExecutionRecord{
						ExecutionStatus:  "success_music_intent_strict",
						ResolvedCommand:  "search and play",
						ResolvedArgs:     query,
						ResolverPathHint: path,
					})
					return a.playAndBuildMessage(ctx, strictResult, genreKeywordFromIntent(intent, a.c.config.SearchAliases, searchKeyword))
				}
			}
		}
	}

	result, fallbackPath, err := a.searchWithFallback(ctx, keywords, searchKeyword, types, searchQuery.Limit)
	if err != nil {
		resolver.RecordExecution(ctx, resolver.ExecutionRecord{
			ExecutionStatus:  "search_error",
			ResolvedCommand:  "search and play",
			ResolvedArgs:     query,
			ResolverPathHint: fallbackPath,
		})
		return "Something wrong...", fmt.Errorf("error in SearchAndDisplayAction\n %v", err)
	}

	executionStatus := "success_alias"
	if fallbackPath == "legacy_query" {
		executionStatus = "fallback_legacy_success"
	}
	resolver.RecordExecution(ctx, resolver.ExecutionRecord{
		ExecutionStatus:  executionStatus,
		ResolvedCommand:  "search and play",
		ResolvedArgs:     query,
		ResolverPathHint: fallbackPath,
	})
	return a.playAndBuildMessage(ctx, result, searchKeyword)
}

func (a SearchAndPlayAction) resolveMusicIntent(ctx context.Context, keyword string) (MusicIntent, string, bool) {
	if a.musicIntentResolver == nil {
		return MusicIntent{}, "music_intent_disabled", false
	}
	if strings.TrimSpace(keyword) == "" {
		return MusicIntent{}, a.musicIntentResolver.Path(), false
	}

	intent, err := a.musicIntentResolver.Resolve(ctx, keyword)
	if err != nil {
		slog.WarnContext(ctx, "music intent resolver failed", "path", a.musicIntentResolver.Path(), "error", err)
		resolver.RecordDecision(ctx, resolver.DecisionRecord{
			InputTextHash:   hashInputText(keyword),
			ResolverPath:    "music_intent_error",
			ResolvedCommand: "search and play",
			ResolvedArgs:    keyword,
		})
		return MusicIntent{}, a.musicIntentResolver.Path(), false
	}

	if intent.IsEmpty() {
		resolver.RecordDecision(ctx, resolver.DecisionRecord{
			InputTextHash:   hashInputText(keyword),
			ResolverPath:    "music_intent_empty",
			ResolvedCommand: "search and play",
			ResolvedArgs:    keyword,
			LLMModel:        intent.Model,
		})
		return MusicIntent{}, a.musicIntentResolver.Path(), false
	}

	resolver.RecordDecision(ctx, resolver.DecisionRecord{
		InputTextHash:   hashInputText(keyword),
		ResolverPath:    a.musicIntentResolver.Path(),
		ResolvedCommand: "search and play",
		ResolvedArgs:    keyword,
		LLMModel:        intent.Model,
	})
	return intent, a.musicIntentResolver.Path(), true
}

func (a SearchAndPlayAction) shouldSuggestCandidates(intent MusicIntent, result *SearchResult) bool {
	if intent.Confidence > 0 && intent.Confidence < a.musicIntentConfidenceThreshold {
		return true
	}

	// Tie/ambiguity should be explicit from resolver side to avoid over-triggering.
	return intent.Ambiguous && result != nil && totalSearchResultCount(result) > 0
}

func buildCandidateMessage(result *SearchResult, intent MusicIntent) string {
	msg := []string{"Search Results...", "Candidate results only (no autoplay)."}
	if intent.Confidence > 0 {
		msg = append(msg, fmt.Sprintf("Reason: low confidence (%.2f)", intent.Confidence))
	}
	if strings.TrimSpace(intent.Reason) != "" {
		msg = append(msg, "Resolver note: "+intent.Reason)
	}
	msg, _ = appendMessage(result.Artists, "Artists", msg, nil, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v", item.Name))
		return msg, nil
	})
	msg, _ = appendMessage(result.Albums, "Albums", msg, nil, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v / %v", item.Name, item.Artist))
		return msg, nil
	})
	msg, _ = appendMessage(result.Tracks, "Tracks", msg, nil, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v / %v ", item.Title, item.Artist))
		return msg, nil
	})
	msg, _ = appendMessage(result.Genres, "Genres", msg, nil, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v ", item.Name))
		return msg, nil
	})
	return strings.Join(msg, "\n")
}

func (a SearchAndPlayAction) playAndBuildMessage(ctx context.Context, result *SearchResult, genreKeyword string) (string, error) {
	msg := []string{"Search Results..."}
	var uris []string
	msg, uris = appendMessage(result.Artists, "Artists", msg, uris, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v", item.Name))
		uris = append(uris, item.URI)
		return msg, uris
	})
	msg, uris = appendMessage(result.Albums, "Albums", msg, uris, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v / %v", item.Name, item.Artist))
		uris = append(uris, item.URI)
		return msg, uris
	})
	msg, uris = appendMessage(result.Tracks, "Tracks", msg, uris, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v / %v ", item.Title, item.Artist))
		uris = append(uris, item.URI)
		return msg, uris
	})
	msg, uris = appendMessage(result.Genres, "Genres", msg, uris, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v ", item.Name))
		return msg, uris
	})

	if len(uris) > 0 || len(result.Genres.Items) > 0 {
		err := a.c.ClearQueue(ctx)
		if err != nil {
			return "", fmt.Errorf("error in ClearQueue\n %v", err)
		}
	}

	if len(uris) > 0 {
		err := a.c.AddItem2QueueAndPlay(ctx, strings.Join(uris, ","), "")
		if err != nil {
			return "", fmt.Errorf("error calling AddItem2QueueAndPlay\n %v", err)
		}
	}

	if len(result.Genres.Items) > 0 {
		err := a.c.AddItem2QueueAndPlay(ctx, "", fmt.Sprintf("genre is \"%s\"", genreKeyword))
		if err != nil {
			return "", fmt.Errorf("error calling AddItem2QueueAndPlay with expression\n %v", err)
		}
	}

	if len(msg) > 1 {
		msg = append(msg, "And play these items")
	} else {
		msg = append(msg, "And no play items...")
	}
	return strings.Join(msg, "\n"), nil
}

func (a SearchAndPlayAction) searchWithFallback(ctx context.Context, keywords []string, fallbackKeyword string, types []SearchType, limit int) (*SearchResult, string, error) {
	expression := buildSearchExpression(keywords, types)
	if expression == "" {
		result, err := a.c.Search(ctx, fallbackKeyword, types, limit)
		return result, "legacy_query", err
	}

	result, err := a.c.SearchByExpression(ctx, expression, types, limit)
	if err != nil || totalSearchResultCount(result) == 0 {
		fallback, fallbackErr := a.c.Search(ctx, fallbackKeyword, types, limit)
		return fallback, "legacy_query", fallbackErr
	}
	return result, "alias_expression", nil
}

// NewSearchAndPlayAction creates a new SearchAndPlayAction.
func NewSearchAndPlayAction(client *Client, opts ...SearchAndPlayActionOption) SearchAndPlayAction {
	a := SearchAndPlayAction{
		name:                           "Search and Play music on Owntone by keyword",
		c:                              client,
		musicIntentConfidenceThreshold: defaultMusicIntentConfidenceThreshold,
	}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}

// SearchAndDisplayAction represents an action to search for music and display the results from Owntone.
type SearchAndDisplayAction struct {
	name string
	c    *Client
}

// Run executes the SearchAndDisplayAction.
func (a SearchAndDisplayAction) Run(ctx context.Context, query string) (string, error) {
	ctx, span := action.StartRunSpan(ctx, "owntone", "SearchAndDisplayAction.Run", query)
	defer span.End()
	msg := []string{"Search Results..."}
	// fmt.Println("original query... " + query)
	searchQuery := Parse(query)
	// fmt.Println("Terms... " + strings.Join(searchQuery.Terms, " "))
	result, err := a.c.Search(ctx, strings.Join(searchQuery.Terms, " "), searchQuery.TypeArray(), searchQuery.Limit)
	if err != nil {
		return "Something wrong...", fmt.Errorf("error in SearchAndDisplayAction(terms=%v)\n %v", strings.Join(searchQuery.Terms, " "), err)
	}
	msg, _ = appendMessage(result.Artists, "Artists", msg, nil, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v", item.Name))
		return msg, nil
	})
	msg, _ = appendMessage(result.Albums, "Albums", msg, nil, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v / %v", item.Name, item.Artist))
		return msg, nil
	})
	msg, _ = appendMessage(result.Tracks, "Tracks", msg, nil, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v / %v ", item.Title, item.Artist))
		return msg, nil
	})
	msg, _ = appendMessage(result.Genres, "Genres", msg, nil, func(item SearchItem, msg []string) ([]string, []string) {
		msg = append(msg, fmt.Sprintf(" %v ", item.Name))
		return msg, nil
	})

	return strings.Join(msg, "\n"), nil
}

// NewSearchAndDisplayAction creates a new SearchAndDisplayAction.
func NewSearchAndDisplayAction(client *Client) SearchAndDisplayAction {
	return SearchAndDisplayAction{
		name: "Search music by keyword on Owntone and display results",
		c:    client,
	}

}

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

// Parse parses a search query string into a SearchQuery struct.
func Parse(target string) *SearchQuery {
	split := strings.Fields(target)
	var queries []string
	var types []SearchType
	limit := -1
	offset := -1
	for _, term := range split {
		switch {
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

func totalSearchResultCount(result *SearchResult) int {
	if result == nil {
		return 0
	}
	return result.Artists.Total + result.Albums.Total + result.Tracks.Total + result.Genres.Total + result.Playlists.Total
}

func normalizeSearchKeyword(keyword string, aliases map[string]string) string {
	normalized := normalizeText(keyword)
	if normalized == "" {
		return ""
	}
	return applySearchAliases(normalized, aliases)
}

func applySearchAliases(keyword string, aliases map[string]string) string {
	if len(aliases) == 0 {
		return keyword
	}

	normalizedAliases := make(map[string]string, len(aliases))
	for from, to := range aliases {
		normalizedFrom := normalizeText(from)
		normalizedTo := normalizeText(to)
		if normalizedFrom == "" || normalizedTo == "" {
			continue
		}
		normalizedAliases[normalizedFrom] = normalizedTo
	}

	if replaced, ok := normalizedAliases[keyword]; ok {
		return replaced
	}

	terms := strings.Fields(keyword)
	for i, term := range terms {
		if replaced, ok := normalizedAliases[term]; ok {
			terms[i] = replaced
		}
	}
	return strings.Join(terms, " ")
}

func normalizeText(s string) string {
	normalized := norm.NFKC.String(s)
	var b strings.Builder
	b.Grow(len(normalized))

	for _, r := range normalized {
		r = katakanaToHiragana(r)
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
		case unicode.IsSpace(r):
			b.WriteByte(' ')
		default:
			// Treat punctuation/symbols as separators.
			b.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func katakanaToHiragana(r rune) rune {
	if r >= 'ァ' && r <= 'ヶ' {
		return r - 0x60
	}
	return r
}

func hashInputText(text string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(text)))
	return hex.EncodeToString(sum[:12])
}

func genreKeywordFromIntent(intent MusicIntent, aliases map[string]string, fallback string) string {
	for _, candidate := range intent.GenreCandidates {
		normalized := normalizeSearchKeyword(candidate, aliases)
		if strings.TrimSpace(normalized) != "" {
			return normalized
		}
	}
	return fallback
}
