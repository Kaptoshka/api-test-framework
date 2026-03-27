package testpages

import (
	"log/slog"
	"time"

	"autotests/pkg/pages"

	"github.com/playwright-community/playwright-go"
)

// SearchResultsPage provides methods for interacting with search results.
type SearchResultsPage struct {
	// BasePage provides inherited navigation and element methods.
	*pages.BasePage
}

// NewSearchResultsPage creates a new SearchResultsPage instance.
func NewSearchResultsPage(
	page playwright.Page,
	baseURL string,
	timeout time.Duration,
	testLog *slog.Logger,
) *SearchResultsPage {
	return &SearchResultsPage{
		BasePage: pages.New(
			page,
			baseURL,
			timeout,
			"SearchResultsPage",
			testLog,
		),
	}
}

// CheckSearchResult returns the product name text from the first search result.
// The query parameter is currently unused for validation.
func (p *SearchResultsPage) CheckSearchResult(query string) (string, error) {
	p.Log.Debug("Checking search result", "query", query)
	return p.CSS(
		".content .product-card",
		"Check first search result",
	).First("First search result").FindCSS(
		".product-card__name a",
		"Product name",
	).GetText()
}
