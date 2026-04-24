package owntone

import "strings"

// MusicIntent is a structured representation of music search intent.
type MusicIntent struct {
	ArtistCandidates []string `json:"artist_candidates"`
	TrackCandidates  []string `json:"track_candidates"`
	GenreCandidates  []string `json:"genre_candidates"`
	MustTerms        []string `json:"must_terms"`
	Confidence       float64  `json:"confidence"`
	Ambiguous        bool     `json:"ambiguous,omitempty"`
	Reason           string   `json:"reason"`
	Model            string   `json:"model,omitempty"`
}

// IsEmpty returns true when the intent does not contain any actionable tokens.
func (m MusicIntent) IsEmpty() bool {
	return len(m.StrictKeywords()) == 0
}

// StrictKeywords returns de-duplicated keywords used for strict AND search.
func (m MusicIntent) StrictKeywords() []string {
	seen := map[string]struct{}{}
	keywords := make([]string, 0, len(m.ArtistCandidates)+len(m.TrackCandidates)+len(m.GenreCandidates)+len(m.MustTerms))
	appendUnique := func(values []string) {
		for _, value := range values {
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			keywords = append(keywords, trimmed)
		}
	}

	appendUnique(m.ArtistCandidates)
	appendUnique(m.TrackCandidates)
	appendUnique(m.GenreCandidates)
	appendUnique(m.MustTerms)
	return keywords
}
