// Package testutil provides shared deterministic test constants.
package testutil

// Live-search fixture constants shared across package tests.
const (
	LiveSearchQuery          = "solo o solo"
	LiveSearchDPDTitle       = "bien"
	LiveSearchDPDSnippet     = "Entrada normativa sobre el adverbio bien."
	LiveSearchDPDURL         = "https://www.rae.es/dpd/bien"
	LiveSearchUnknownTitle   = "Ruta desconocida pero quizá útil"
	LiveSearchUnknownSnippet = "No debe romper el mapeo."
	LiveSearchUnknownURL     = "https://www.rae.es/archivo/ruta-rara"
	LiveSearchResultsFixture = "live_search_results.html"
	LiveSearchEmptyFixture   = "live_search_empty.html"
	LiveSearchBrokenFixture  = "live_search_broken.html"
)
