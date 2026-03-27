package testpages

import (
	"fmt"
	"log/slog"
	"time"

	"autotests/pkg/pages"

	"github.com/playwright-community/playwright-go"
)

// WishlistPage provides methods for interacting with the wishlist/favorites page.
type WishlistPage struct {
	// BasePage provides inherited navigation and element methods.
	*pages.BasePage
}

// NewWishlistPage creates a new WishlistPage instance with page-scoped logger.
func NewWishlistPage(
	page playwright.Page,
	baseURL string,
	timeout time.Duration,
	testLog *slog.Logger,
) *WishlistPage {
	return &WishlistPage{
		BasePage: pages.New(
			page,
			baseURL,
			timeout,
			"WishlistPage",
			testLog.With("page", "WishlistPage"),
		),
	}
}

// GetItemsCount returns the number of product cards in the wishlist.
func (p *WishlistPage) GetItemsCount() (int, error) {
	return p.CSS(
		".page-favorite .product-card",
		"Get wishlist items count",
	).Count()
}

// Clear removes all items from the wishlist by repeatedly clicking delete buttons.
// It loops until the wishlist is empty, waiting for network idle after each removal.
// Navigates to /favorite first.
func (p *WishlistPage) Clear() error {
	p.Log.Debug("Clearing wishlist")

	if err := p.Navigate("/favorite"); err != nil {
		return fmt.Errorf(
			"cannot navigate to wishlist for clearing: %w",
			err,
		)
	}

	for {
		count, err := p.GetItemsCount()
		if err != nil {
			return fmt.Errorf("cannot count wishlist items: %w", err)
		}
		if count == 0 {
			p.Log.Info("Wishlist is empty")
			return nil
		}

		p.Log.Debug("Items remaining", "count", count)

		removeBtn := p.CSS(
			".product-card__favorite-delete",
			"Get remove from wishlist button",
		).First("First remove from wishlist button")

		if err = removeBtn.Click(); err != nil {
			return fmt.Errorf("cannot click remove from wishlist button: %w", err)
		}

		if err = p.WaitForNetworkIdle(); err != nil {
			return fmt.Errorf("page did not update after remove: %w", err)
		}
	}
}

// IsEmpty returns true if the wishlist contains no items.
func (p *WishlistPage) IsEmpty() (bool, error) {
	count, err := p.GetItemsCount()
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// FindProductURL returns the href from the product name link in the wishlist.
// Returns error if the product is not found in the wishlist.
func (p *WishlistPage) FindProductURL(name string) (string, error) {
	p.Log.Debug("Checking product in wishlist", "name", name)

	productURL, err := p.CSS(
		".page-favorite .product-card__name a",
		"Get product url in wishlist",
	).GetAttribute("href")
	if err != nil {
		p.Log.Error(
			"failed to find product url in wishlist",
			"name",
			name,
			"error",
			err,
		)
		return "", fmt.Errorf("failed to find product url in wishlist: %w", err)
	}

	return productURL, nil
}
