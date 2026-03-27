package components

import (
	"fmt"
	"log/slog"
	"time"

	"autotests/pkg/elements"

	"github.com/playwright-community/playwright-go"
)

// Header provides methods for interacting with the site header component.
type Header struct {
	// page is the Playwright page for interactions.
	page playwright.Page
	// Log is the component-scoped logger.
	Log *slog.Logger
	// timeout is the default timeout for waits.
	timeout time.Duration
	// searchInput locates the visible search input field.
	searchInput *elements.Element
	// searchBtn locates the visible search submit button.
	searchBtn *elements.Element
	// wishlistBtn locates the visible wishlist indicator.
	wishlistBtn *elements.Element
	// cartBtn locates the visible cart counter.
	cartBtn *elements.Element
}

// NewHeader creates a new Header with pre-initialized element locators.
// All locators target visible elements only using :visible CSS pseudo-selector.
func NewHeader(
	page playwright.Page,
	timeout time.Duration,
	log *slog.Logger,
) *Header {
	return &Header{
		page:    page,
		timeout: timeout,
		Log:     log,

		searchInput: elements.NewCSS(
			page,
			"header .search .input-group input:visible",
			"Search Input",
			timeout,
			log,
		),
		searchBtn: elements.NewCSS(
			page,
			"header .search .input-group button.submit:visible",
			"Search Button",
			timeout,
			log,
		),
		wishlistBtn: elements.NewCSS(
			page,
			"header .container .favorite-informer:visible",
			"Wishlist Button",
			timeout,
			log,
		),
		cartBtn: elements.NewCSS(
			page,
			"header .cart-counter:visible",
			"Cart Button",
			timeout,
			log,
		),
	}
}

// Search fills the search input with the query and presses Enter.
// Returns error if the input is not visible or interaction fails.
func (h *Header) Search(query string) error {
	h.Log.Info("Searching for product via header", "query", query)

	if err := h.searchInput.WaitForVisible(); err != nil {
		return fmt.Errorf("search input not found: %w", err)
	}

	if err := h.searchInput.Fill(query); err != nil {
		return fmt.Errorf("failed to type search query: %w", err)
	}

	if err := h.searchInput.Press("Enter"); err != nil {
		return fmt.Errorf("failed to press Enter: %w", err)
	}

	return nil
}

// OpenWishlist clicks the wishlist button to navigate to the wishlist page.
func (h *Header) OpenWishlist() error {
	h.Log.Debug("Click button that opens wishlist")

	return h.wishlistBtn.Click()
}

// OpenCart clicks the cart button to navigate to the cart page.
func (h *Header) OpenCart() error {
	h.Log.Debug("Click button that opens cart")

	return h.cartBtn.Click()
}
