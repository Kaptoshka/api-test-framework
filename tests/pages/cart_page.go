package testpages

import (
	"fmt"
	"log/slog"
	"time"

	"autotests/pkg/pages"

	"github.com/playwright-community/playwright-go"
)

// CartPage provides methods for interacting with the shopping cart page.
type CartPage struct {
	// BasePage provides inherited navigation and element methods.
	*pages.BasePage
}

// NewCartPage creates a new CartPage instance.
func NewCartPage(
	page playwright.Page,
	baseURL string,
	timeout time.Duration,
	testLog *slog.Logger,
) *CartPage {
	return &CartPage{
		BasePage: pages.New(
			page,
			baseURL,
			timeout,
			"CartPage",
			testLog,
		),
	}
}

// Clear removes all items from the cart by repeatedly clicking "Удалить" button.
// It auto-accepts any confirmation dialogs that appear.
// Navigates to /cart first and loops until cart is empty.
func (p *CartPage) Clear() error {
	p.Log.Debug("Clearing cart")

	if err := p.Navigate("/cart"); err != nil {
		return fmt.Errorf("cannot navigate to cart for clearing: %w", err)
	}

	for {
		productList := p.CSS(".list-group-item", "Cart product list")

		count, err := productList.Count()
		if err != nil {
			return fmt.Errorf("failed to count cart items: %w", err)
		}
		if count == 0 {
			p.Log.Debug("Cart is empty")
			return nil
		}

		p.Log.Debug("Items remaining", "count", count)

		p.Page.On("dialog", func(dialog playwright.Dialog) {
			p.Log.Debug("Dialog appeared. Auto-accepting...", "message", dialog.Message())
			_ = dialog.Accept()
		})

		removeBtn := productList.FindCSS(
			"a.btn",
			"Remove button",
		).FilterByText(
			"Удалить",
			"Remove product from cart button",
		).First("First remove button")

		if err = removeBtn.Click(); err != nil {
			return fmt.Errorf("failed to click remove button: %w", err)
		}

		if err = p.WaitForNetworkIdle(); err != nil {
			return fmt.Errorf("page did not update after removing item: %w", err)
		}
	}
}

// CheckCartForAndValidatePrice finds the cart row matching the product name,
// extracts the price, and compares it with the expected price.
// Returns error if the product is not found or prices do not match.
func (p *CartPage) CheckCartForAndValidatePrice(
	expectedProduct *Product,
) error {
	p.Log.Debug("Check cart for product", "product", expectedProduct.Name)

	cartProductRow := p.CSS(
		fmt.Sprintf(
			".list-group-item .row:has(a:has-text('%s'))",
			expectedProduct.Name,
		),
		fmt.Sprintf(
			"Cart product item with name [%s]",
			expectedProduct.Name,
		),
	)

	p.Log.Debug(
		"Check cart product price",
		"product", expectedProduct.Name,
	)

	cartProductPriceStr, err := cartProductRow.FindCSS(
		"div:nth-child(2):has-text('₽')",
		"Price column (second column in cart row)",
	).GetText()
	if err != nil {
		return fmt.Errorf(
			"failed to get cart product price for [%s]: %w",
			expectedProduct.Name,
			err,
		)
	}

	cartProductPrice, err := p.ParseInt(cartProductPriceStr)
	if err != nil {
		return fmt.Errorf(
			"failed to parse cart product price for [%s]: %w",
			expectedProduct.Name,
			err,
		)
	}

	p.Log.Debug(
		"Compare cart product price with catalog product price",
		"cartPrice", cartProductPrice,
		"expectedPrice", expectedProduct.Price,
	)

	if expectedProduct.Price != cartProductPrice {
		return fmt.Errorf(
			"cart product price for [%s] is not equal to expected product price",
			expectedProduct.Name,
		)
	}

	return nil
}
