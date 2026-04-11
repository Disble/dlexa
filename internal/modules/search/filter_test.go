package search

import (
	"reflect"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

var (
	soloRankingCandidates = []model.SearchCandidate{
		{Title: "solo", Classification: classificationDPDEntry, ArticleKey: "solo"},
		{Title: "El uso de tilde en solo", Classification: classificationFAQ},
	}
	alicuotaRankingCandidates = []model.SearchCandidate{
		{Title: alicuotaTitle, Classification: classificationDPDEntry, ArticleKey: "alicuota"},
		{Title: "Dudas sobre tildes en esdrújulas", Classification: classificationFAQ},
	}
	asimismoRankingCandidates = []model.SearchCandidate{
		{Title: "asimismo", Classification: classificationDPDEntry, ArticleKey: "asimismo"},
		{Title: "así mismo", Classification: classificationDPDEntry, ArticleKey: "asimismo"},
		{Title: "a sí mismo", Classification: classificationDPDEntry, ArticleKey: "asimismo"},
	}
	exRankingCandidates = []model.SearchCandidate{
		{Title: "ex", Classification: classificationDPDEntry, ArticleKey: "ex"},
		{Title: "Texto de ejemplo", Classification: classificationFAQ},
		{Title: "expresión", Classification: classificationLinguisticArticle},
	}
	guionRankingCandidates = []model.SearchCandidate{
		{Title: "guion", Classification: classificationDPDEntry, ArticleKey: "guion"},
		{Title: "guión", Classification: classificationDPDEntry, ArticleKey: "guion"},
		{Title: "Usos del guion en la escritura", Classification: classificationFAQ},
	}
)

func TestCurateCandidatesRanksAndDeduplicatesCrossProviderResults(t *testing.T) {
	candidates := []model.SearchCandidate{
		{Title: rutaRaraTitle, URL: rutaRaraURL},
		{Title: "La conjuncion o", URL: "https://www.rae.es/espanol-al-dia/la-conjuncion-o"},
		{Title: "Solo", ArticleKey: "solo", URL: "https://www.rae.es/dpd/solo"},
		{Title: "Preguntas frecuentes: tilde en solo", URL: faqSobreTildeURL},
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
		{Title: abuDhabiTitle, Snippet: "resultado del buscador general", URL: "https://www.rae.es/dpd/Abu_Dabi", SourceHint: sourceBusquedaGeneral},
		{DisplayText: abuDhabiTitle, ArticleKey: "Abu Dabi", SourceHint: "Diccionario panhispánico de dudas"},
		{Title: rutaRaraTitle, URL: rutaRaraURL},
	}

	got := curateCandidates(abuDhabiTitle, candidates)

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
		{Title: faqSobreTildeTitle, Snippet: "faq rescatable pero genérica", URL: faqSobreTildeURL, SourceHint: sourceBusquedaGeneral},
		{Title: alicuotaTitle, Snippet: "artículo complementario exacto", URL: "https://www.rae.es/espanol-al-dia/alicuota", SourceHint: sourceBusquedaGeneral},
		{Title: rutaRaraTitle, URL: rutaRaraURL},
	}

	got := curateCandidates("alicuota", candidates)

	if want := []string{alicuotaTitle, faqSobreTildeTitle, rutaRaraTitle}; !reflect.DeepEqual(candidateTitles(got), want) {
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
			wantTopTitle: alicuotaTitle,
		},
		{
			name:         "alicuota accented",
			query:        "alícuota",
			candidates:   alicuotaRankingCandidates,
			wantTopTitle: alicuotaTitle,
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
			candidate: model.SearchCandidate{Title: faqSobreTildeTitle, Snippet: "Respuesta normativa breve.", URL: faqSobreTildeURL},
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
		{Title: faqSobreTildeTitle, Snippet: "FAQ rescatable.", URL: faqSobreTildeURL},
	}

	got := curateCandidates("tilde", candidates)
	if len(got) != 1 {
		t.Fatalf("Candidates len = %d, want 1 rescued noticia only", len(got))
	}
	if got[0].Title != faqSobreTildeTitle {
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
