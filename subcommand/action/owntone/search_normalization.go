package owntone

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func normalizeSearchKeyword(keyword string, aliases map[string]string) string {
	normalized := normalizeText(keyword)
	if normalized == "" {
		return ""
	}
	return applySearchAliases(normalized, aliases)
}

func normalizeSearchKeywordForExternalSearch(keyword string, aliases map[string]string) string {
	normalized := normalizeTextNoKana(keyword)
	if normalized == "" {
		return ""
	}
	return applySearchAliasesNoKana(normalized, aliases)
}

func applySearchAliases(keyword string, aliases map[string]string) string {
	return applyAliasesWithNormalizer(keyword, aliases, normalizeText)
}

func applySearchAliasesNoKana(keyword string, aliases map[string]string) string {
	return applyAliasesWithNormalizer(keyword, aliases, normalizeTextNoKana)
}

func applyAliasesWithNormalizer(keyword string, aliases map[string]string, normalizer func(string) string) string {
	if len(aliases) == 0 {
		return keyword
	}

	normalizedAliases := make(map[string]string, len(aliases))
	for from, to := range aliases {
		normalizedFrom := normalizer(from)
		normalizedTo := normalizer(to)
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
	return buildNormalizedText(s, true)
}

func normalizeTextNoKana(s string) string {
	return buildNormalizedText(s, false)
}

func buildNormalizedText(s string, convertKana bool) string {
	normalized := norm.NFKC.String(s)
	var b strings.Builder
	b.Grow(len(normalized))

	for _, r := range normalized {
		if convertKana {
			r = katakanaToHiragana(r)
		}
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
		case unicode.IsSpace(r):
			b.WriteByte(' ')
		default:
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
