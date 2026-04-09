package search

import (
	"reflect"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

var (
	soloRankingCandidates = []model.SearchCandidate{
		{Title: "solo", Classification: "dpd-entry", ArticleKey: "solo"},
		{Title: "El uso de tilde en solo", Classification: "faq"},
	}
	alicuotaRankingCandidates = []model.SearchCandidate{
		{Title: "Alícuota", Classification: "dpd-entry", ArticleKey: "alicuota"},
		{Title: "Dudas sobre tildes en esdrújulas", Classification: "faq"},
	}
	asimismoRankingCandidates = []model.SearchCandidate{
		{Title: "asimismo", Classification: "dpd-entry", ArticleKey: "asimismo"},
		{Title: "así mismo", Classification: "dpd-entry", ArticleKey: "asimismo"},
		{Title: "a sí mismo", Classification: "dpd-entry", ArticleKey: "asimismo"},
	}
	exRankingCandidates = []model.SearchCandidate{
		{Title: "ex", Classification: "dpd-entry", ArticleKey: "ex"},
		{Title: "Texto de ejemplo", Classification: "faq"},
		{Title: "expresión", Classification: "linguistic-article"},
	}
	guionRankingCandidates = []model.SearchCandidate{
		{Title: "guion", Classification: "dpd-entry", ArticleKey: "guion"},
		{Title: "guión", Classification: "dpd-entry", ArticleKey: "guion"},
		{Title: "Usos del guion en la escritura", Classification: "faq"},
	}
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

func TestCurateCandidatesCalibratesRankingForExactVariantsAndShortQueries(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		candidates   []model.SearchCandidate
		wantTopTitle string
	}{
		{
			name:         "solo ranking",
			query:        "solo",
			candidates:   soloRankingCandidates,
			wantTopTitle: "solo",
		},
		{
			name:         "alicuota unaccented",
			query:        "alicuota",
			candidates:   alicuotaRankingCandidates,
			wantTopTitle: "Alícuota",
		},
		{
			name:         "alicuota accented",
			query:        "alícuota",
			candidates:   alicuotaRankingCandidates,
			wantTopTitle: "Alícuota",
		},
		{
			name:         "asimismo disambiguation",
			query:        "asimismo",
			candidates:   asimismoRankingCandidates,
			wantTopTitle: "asimismo",
		},
		{
			name:         "ex short query",
			query:        "ex",
			candidates:   exRankingCandidates,
			wantTopTitle: "ex",
		},
		{
			name:         "guion accent variant",
			query:        "guion",
			candidates:   guionRankingCandidates,
			wantTopTitle: "guion",
		},
		{
			name:         "guion accented query",
			query:        "guión",
			candidates:   guionRankingCandidates,
			wantTopTitle: "guion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := curateCandidates(tt.query, tt.candidates)
			if len(got) == 0 {
				t.Fatalf("curateCandidates(%q) returned no candidates", tt.query)
			}

			if got[0].Title != tt.wantTopTitle {
				t.Fatalf("top title = %q, want %q (order=%v)", got[0].Title, tt.wantTopTitle, candidateTitles(got))
			}
		})
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
			name:      "espanol al dia is executable",
			candidate: model.SearchCandidate{URL: "https://www.rae.es/espanol-al-dia/solo"},
			want:      false,
		},
		{
			name:      "rescued noticia is executable",
			candidate: model.SearchCandidate{Title: "Preguntas frecuentes: tildes", URL: "https://www.rae.es/noticia/tildes"},
			want:      false,
		},
		{
			name:      "duda linguistica is executable",
			candidate: model.SearchCandidate{URL: "https://www.rae.es/duda-linguistica/solo"},
			want:      false,
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

func TestIsRescuedNoticiaRequiresFAQPrefixOnly(t *testing.T) {
	tests := []struct {
		name      string
		candidate model.SearchCandidate
		want      bool
	}{
		{
			name:      "faq title is rescued",
			candidate: model.SearchCandidate{Title: "Preguntas frecuentes: sobre la tilde", Snippet: "Respuesta normativa breve.", URL: "https://www.rae.es/noticia/preguntas-frecuentes-sobre-la-tilde"},
			want:      true,
		},
		{
			name:      "faq title without linguistic signal is still rescued",
			candidate: model.SearchCandidate{Title: "Preguntas frecuentes: agenda institucional", Snippet: "Horarios y acceso al acto.", URL: "https://www.rae.es/noticia/agenda"},
			want:      true,
		},
		{
			name:      "non faq noticia is rejected",
			candidate: model.SearchCandidate{Title: "Nueva normativa del español se presenta en la RAE", Snippet: "Presentación institucional del libro.", URL: "https://www.rae.es/noticia/geopolitica"},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRescuedNoticia(tt.candidate)
			if got != tt.want {
				t.Fatalf("isRescuedNoticia(%#v) = %v, want %v", tt.candidate, got, tt.want)
			}
		})
	}
}

func TestCurateCandidatesRejectsNonFAQNoticiasEvenWithLinguisticWords(t *testing.T) {
	candidates := []model.SearchCandidate{
		{Title: "La obra Geopolítica del español se presenta en la RAE", Snippet: "Libro institucional sobre el español.", URL: "https://www.rae.es/noticia/la-obra-geopolitica-del-espanol-se-presenta-en-la-rae"},
		{Title: "Preguntas frecuentes: sobre la tilde", Snippet: "FAQ rescatable.", URL: "https://www.rae.es/noticia/preguntas-frecuentes-sobre-la-tilde"},
	}

	got := curateCandidates("tilde", candidates)
	if len(got) != 1 {
		t.Fatalf("Candidates len = %d, want 1 rescued noticia only", len(got))
	}
	if got[0].Title != "Preguntas frecuentes: sobre la tilde" {
		t.Fatalf("candidate titles = %v, want rescued FAQ only", candidateTitles(got))
	}
}

func candidateTitles(candidates []model.SearchCandidate) []string {
	titles := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		titles = append(titles, candidate.Title)
	}
	return titles
}
