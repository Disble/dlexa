package search

import (
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

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
