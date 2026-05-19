package commandsearch

import (
	"fmt"
	"sort"
	"strings"
)

const DefaultLimit = 20

type Document struct {
	Path  string   `json:"path"`
	Use   string   `json:"use"`
	Short string   `json:"short,omitempty"`
	Flags []string `json:"flags,omitempty"`
}

type Options struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

type Result struct {
	Query   string  `json:"query"`
	Limit   int     `json:"limit"`
	Matches []Match `json:"matches"`
}

type Match struct {
	Path  string `json:"path"`
	Use   string `json:"use"`
	Short string `json:"short,omitempty"`
	Score int    `json:"score"`
}

func Search(documents []Document, options Options) (Result, error) {
	query := strings.TrimSpace(options.Query)
	if query == "" {
		return Result{}, fmt.Errorf("search query is required")
	}
	limit := options.Limit
	if limit < 0 {
		return Result{}, fmt.Errorf("limit cannot be negative")
	}
	terms := strings.Fields(strings.ToLower(query))
	matches := make([]Match, 0)
	for _, document := range documents {
		score := scoreDocument(document, strings.ToLower(query), terms)
		if score == 0 {
			continue
		}
		matches = append(matches, Match{
			Path:  document.Path,
			Use:   document.Use,
			Short: document.Short,
			Score: score,
		})
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score == matches[j].Score {
			return matches[i].Path < matches[j].Path
		}
		return matches[i].Score > matches[j].Score
	})
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}
	return Result{Query: query, Limit: limit, Matches: matches}, nil
}

func scoreDocument(document Document, query string, terms []string) int {
	path := strings.ToLower(document.Path)
	use := strings.ToLower(document.Use)
	short := strings.ToLower(document.Short)
	flags := strings.ToLower(strings.Join(document.Flags, " "))
	haystack := strings.Join([]string{path, use, short, flags}, " ")
	for _, term := range terms {
		if !strings.Contains(haystack, term) {
			return 0
		}
	}
	score := 0
	if strings.Contains(path, query) {
		score += 20
	}
	for _, term := range terms {
		if strings.Contains(path, term) {
			score += 8
		}
		if strings.Contains(use, term) {
			score += 5
		}
		if strings.Contains(short, term) {
			score += 3
		}
		if strings.Contains(flags, term) {
			score++
		}
	}
	return score
}
