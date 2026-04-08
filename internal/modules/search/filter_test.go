package search

import (
	"reflect"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

func TestCurateCandidatesRanksAndDeduplicatesCrossProviderResults(t *testing.T) {
	candidates := []model.SearchCandidate{
		{Title: "Ruta rara", URL: "https://www.rae.es/archivo/ruta-rara"},
		{Title: "La conjuncion o", URL: "https://www.rae.es/espanol-al-dia/la-conjuncion-o"},
		{Title: "Solo", ArticleKey: "solo", URL: "https://www.rae.es/dpd/solo"},
		{Title: "Preguntas frecuentes: tilde en solo", URL: "https://www.rae.es/noticia/preguntas-frecuentes-sobre-la-tilde"},
		{Title: "Solo", ArticleKey: "solo", URL: "https://www.rae.es/dpd/solo", Snippet: "entrada con snippet mas rico"},
	}

	got := curateCandidates("solo o solo", candidates)

	if len(got) != 4 {
		t.Fatalf("Candidates len = %d, want 4 after deduplication", len(got))
	}
	if want := []string{"Solo", "Preguntas frecuentes: tilde en solo", "La conjuncion o", "Ruta rara"}; !reflect.DeepEqual(candidateTitles(got), want) {
		t.Fatalf("candidate order = %v, want %v", candidateTitles(got), want)
	}
	if got[0].Snippet != "entrada con snippet mas rico" {
		t.Fatalf("top ranked duplicate snippet = %q, want richer duplicate retained", got[0].Snippet)
	}
}

func TestEnrichCandidateMarksDeferredOnlyForNonDPDDestinations(t *testing.T) {
	tests := []struct {
		name      string
		candidate model.SearchCandidate
		want      bool
	}{
		{
			name:      "dpd article key stays executable",
			candidate: model.SearchCandidate{ArticleKey: "solo"},
			want:      false,
		},
		{
			name:      "espanol al dia is deferred",
			candidate: model.SearchCandidate{URL: "https://www.rae.es/espanol-al-dia/solo"},
			want:      true,
		},
		{
			name:      "rescued noticia is deferred",
			candidate: model.SearchCandidate{Title: "Preguntas frecuentes: tildes", URL: "https://www.rae.es/noticia/tildes"},
			want:      true,
		},
		{
			name:      "duda linguistica is deferred",
			candidate: model.SearchCandidate{URL: "https://www.rae.es/duda-linguistica/solo"},
			want:      true,
		},
		{
			name:      "unknown stays non deferred",
			candidate: model.SearchCandidate{URL: "https://www.rae.es/archivo/ruta-rara"},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := enrichCandidate("solo", tt.candidate)

			if got.Deferred != tt.want {
				t.Fatalf("Deferred = %v, want %v (candidate=%#v)", got.Deferred, tt.want, got)
			}
		})
	}
}

func candidateTitles(candidates []model.SearchCandidate) []string {
	titles := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		titles = append(titles, candidate.Title)
	}
	return titles
}
