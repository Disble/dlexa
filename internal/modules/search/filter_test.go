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

func TestCurateCandidatesPrefersDPDIndexHitForEquivalentDPDDestination(t *testing.T) {
	candidates := []model.SearchCandidate{
		{Title: "Abu Dhabi", Snippet: "resultado del buscador general", URL: "https://www.rae.es/dpd/Abu_Dabi", SourceHint: "Búsqueda general RAE"},
		{DisplayText: "Abu Dhabi", ArticleKey: "Abu Dabi", SourceHint: "Diccionario panhispánico de dudas"},
		{Title: "Ruta rara", URL: "https://www.rae.es/archivo/ruta-rara"},
	}

	got := curateCandidates("Abu Dhabi", candidates)

	if len(got) != 2 {
		t.Fatalf("Candidates len = %d, want 2 after semantic DPD deduplication", len(got))
	}
	if got[0].NextCommand != "dlexa dpd Abu Dabi" {
		t.Fatalf("top candidate next command = %q, want DPD entry command", got[0].NextCommand)
	}
	if got[0].SourceHint != "Diccionario panhispánico de dudas" {
		t.Fatalf("top candidate source hint = %q, want specialized DPD provider", got[0].SourceHint)
	}
}

func TestCurateCandidatesBoostsQueryAffinityAcrossComplementaryHits(t *testing.T) {
	candidates := []model.SearchCandidate{
		{Title: "Preguntas frecuentes: sobre la tilde", Snippet: "faq rescatable pero genérica", URL: "https://www.rae.es/noticia/preguntas-frecuentes-sobre-la-tilde", SourceHint: "Búsqueda general RAE"},
		{Title: "Alícuota", Snippet: "artículo complementario exacto", URL: "https://www.rae.es/espanol-al-dia/alicuota", SourceHint: "Búsqueda general RAE"},
		{Title: "Ruta rara", URL: "https://www.rae.es/archivo/ruta-rara"},
	}

	got := curateCandidates("alicuota", candidates)

	if want := []string{"Alícuota", "Preguntas frecuentes: sobre la tilde", "Ruta rara"}; !reflect.DeepEqual(candidateTitles(got), want) {
		t.Fatalf("candidate order = %v, want %v", candidateTitles(got), want)
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
