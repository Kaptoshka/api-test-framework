package testpages

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"autotests/pkg/elements"
	"autotests/pkg/pages"

	"github.com/playwright-community/playwright-go"
)

// Product represents product data extracted from catalog cards.
type Product struct {
	// Name is the product display name.
	Name string
	// Price is the current price in rubles.
	Price int
	// Width is the width dimension in millimeters.
	Width int
	// Depth is the depth dimension in millimeters.
	Depth int
	// URL is the absolute product page URL.
	URL string
}

// CatalogPage provides methods for interacting with the product catalog page.
type CatalogPage struct {
	// BasePage provides inherited navigation and element methods.
	*pages.BasePage
	// productCards is the locator for visible product cards (excludes carousel items).
	productCards *elements.Element
}

// NewCatalogPage creates a new CatalogPage with product card locator.
// The locator targets .product-card:not(.owl-carousel .product-card) to exclude carousel items.
func NewCatalogPage(
	page playwright.Page,
	baseURL string,
	timeout time.Duration,
	testLog *slog.Logger,
) *CatalogPage {
	return &CatalogPage{
		BasePage: pages.New(
			page,
			baseURL,
			timeout,
			"CatalogPage",
			testLog,
		),
		productCards: elements.NewCSS(
			page,
			".content .container .product-card:not(.owl-carousel .product-card)",
			"Product Cards List",
			timeout,
			testLog,
		),
	}
}

// ClickFilterContainer expands the filter panel by clicking the filter title link.
// The filterName should match the visible text on the filter toggle link.
func (p *CatalogPage) ClickFilterContainer(filterName string) error {
	p.Log.Debug("Clicking filter container", "filter", filterName)
	return p.CSS(
		fmt.Sprintf("div.filter__title a:has-text('%s')", filterName),
		fmt.Sprintf("Filter title link [%s]", filterName),
	).Click()
}

// SetRangeSliderByDrag sets the price range by dragging the min and max handles.
// It calculates handle positions proportionally within the track bounds.
// Returns error if from/to values are outside the filter's absolute min/max range.
func (p *CatalogPage) SetRangeSliderByDrag(
	filterName string,
	from int,
	to int,
) error {
	p.Log.Debug(
		"Setting range filter",
		"filter", filterName,
		"from", from,
		"to", to,
	)

	selector := fmt.Sprintf(
		`.filter__item:has(.filter__title:has-text("%s"))`,
		filterName,
	)

	filterContainer := p.CSS(
		selector,
		"Filter title",
	)

	minHandle := filterContainer.FindCSS(
		".slider-handle.min-slider-handle",
		"Minslider handle",
	)
	maxHandle := filterContainer.FindCSS(
		".slider-handle.max-slider-handle",
		"Maxslider handle",
	)
	track := filterContainer.FindCSS(
		".slider-track",
		"Slider track",
	)

	absMinStr, err := minHandle.GetAttribute("aria-valuemin")
	if err != nil {
		return fmt.Errorf("cannot read aria-valuemin: %w", err)
	}

	absMaxStr, err := maxHandle.GetAttribute("aria-valuemax")
	if err != nil {
		return fmt.Errorf("cannot read aria-valuemax: %w", err)
	}

	absMin, err := p.ParseInt(absMinStr)
	if err != nil {
		return fmt.Errorf("cannot parse aria-valuemin: %w", err)
	}

	absMax, err := p.ParseInt(absMaxStr)
	if err != nil {
		return fmt.Errorf("cannot parse aria-valuemax: %w", err)
	}

	if from < absMin || to > absMax {
		return fmt.Errorf(
			"invalid range [%d, %d] for [%s]: allowed [%d, %d]",
			from, to, filterName, absMin, absMax,
		)
	}

	trackBox, err := track.GetBoundingBox()
	if err != nil {
		return fmt.Errorf("cannot get track bounding box: %w", err)
	}

	rangeSize := float64(absMax - absMin)
	fromX := trackBox.X + (float64(from-absMin)/rangeSize)*trackBox.Width
	toX := trackBox.X + (float64(to-absMin)/rangeSize)*trackBox.Width

	const centerDivisor = 2

	centerY := trackBox.Y + trackBox.Height/centerDivisor

	if err = p.dragHandleTo(minHandle, fromX, centerY); err != nil {
		return fmt.Errorf("failed to drag min handle: %w", err)
	}

	if err = p.dragHandleTo(maxHandle, toX, centerY); err != nil {
		return fmt.Errorf("failed to drag max handle: %w", err)
	}

	p.Log.Info("Range filter set", "filter", filterName, "from", from, "to", to)
	return nil
}

// dragHandleTo is an internal helper that drags an element to the specified coordinates.
// It performs the drag in 10 steps for smooth movement.
func (p *CatalogPage) dragHandleTo(
	handle *elements.Element,
	targetX float64,
	targetY float64,
) error {
	p.Log.Debug(
		"Dragging handle to coordinates",
		"x", targetX,
		"y", targetY,
	)

	if err := handle.Hover(); err != nil {
		return fmt.Errorf("failed to hover handle: %w", err)
	}

	mouse := p.Page.Mouse()

	if err := mouse.Down(); err != nil {
		return fmt.Errorf("failed to press mouse button: %w", err)
	}

	steps := 10

	if err := p.Page.Mouse().Move(targetX, targetY, playwright.MouseMoveOptions{
		Steps: new(steps),
	}); err != nil {
		return fmt.Errorf("failed to move mouse to target: %w", err)
	}

	return mouse.Up()
}

// ClickApplyButton clicks the "Применить фильтр" button to apply selected filters.
// Call this after setting filter values.
func (p *CatalogPage) ClickApplyButton() error {
	p.Log.Debug("Clicking apply button")
	return p.CSS(
		".filter__link div.btn:has-text('Применить фильтр')",
		"Apply button",
	).Click()
}

// WaitForResults waits for at least one product card to appear after filter application.
// Use this after applying filters to ensure results are loaded.
func (p *CatalogPage) WaitForResults() error {
	p.Log.Debug("Waiting for results to update")

	if err := p.productCards.First("First product card").WaitForVisible(); err != nil {
		return fmt.Errorf("no product cards appeared after filter: %w", err)
	}

	return nil
}

// GetResultsCount returns the number of visible product cards matching the locator.
func (p *CatalogPage) GetResultsCount() (int, error) {
	p.Log.Debug("Getting results count")
	return p.productCards.Count()
}

// ClickSortButton clicks the sort button with the specified name.
// Common sort options include "цене", "названию", "новизне".
func (p *CatalogPage) ClickSortButton(sortName string) error {
	p.Log.Debug("Clicking sort button", "SortName", sortName)
	return p.CSS(
		fmt.Sprintf(
			".sorting-bar .sorting-bar__text b:has-text('%s')",
			sortName,
		),
		"Sort button",
	).Click()
}

// FindProduct verifies that a product with the specified name is visible in the catalog.
// Returns error if the product is not found or not visible.
func (p *CatalogPage) FindProduct(
	name string,
) error {
	p.Log.Debug("Checking product visibility", "name", name)

	product := p.productCards.FilterByText(name, "Card for "+name)

	count, err := product.Count()
	if err != nil || count == 0 {
		return fmt.Errorf("failed to find product '%s': %w", name, err)
	}

	p.Log.Debug("Product is visible", "name", name, "count", count)
	return nil
}

// GetProductCard extracts complete product data from the catalog card.
// It reads the name, price, dimensions (Ширина/Глубина), and URL.
// Returns Product struct with all extracted fields.
func (p *CatalogPage) GetProductCard(name string) (*Product, error) {
	p.Log.Debug("Getting product card", "name", name)

	card := p.productCards.FilterByText(name, "Card for "+name).First(
		"First product card",
	)

	cardName, err := card.FindCSS(
		".product-card__name",
		"Product card name",
	).GetText()
	if err != nil {
		return nil, fmt.Errorf("cannot get card name: %w", err)
	}

	cardName = strings.TrimSpace(cardName)

	cardWidthStr, err := p.GetParamCSS(
		card,
		"Ширина",
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get card width: %w", err)
	}

	width, err := p.ParseInt(cardWidthStr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse card width to int: %w", err)
	}

	cardDepthStr, err := p.GetParamCSS(
		card,
		"Глубина",
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get card depth: %w", err)
	}

	depth, err := p.ParseInt(cardDepthStr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse card depth to int: %w", err)
	}

	cardPriceStr, err := card.FindCSS(
		".product-card__now_price:not(.product-card__old_price) span b",
		fmt.Sprintf(
			"Retrieve current price of product [%s]",
			name,
		),
	).GetText()
	if err != nil {
		return nil, fmt.Errorf("cannot get card price: %w", err)
	}

	price, err := p.ParseInt(cardPriceStr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse card price to int: %w", err)
	}

	url, err := p.GetURL(card, ".product-card__name a")
	if err != nil {
		return nil, fmt.Errorf("cannot get card URL: %w", err)
	}

	return &Product{
		Name:  cardName,
		Price: price,
		Width: width,
		Depth: depth,
		URL:   url,
	}, nil
}

// GetParamCSS is an internal helper that extracts a parameter value from a card element.
// It finds the small text matching paramName within the element.
func (p *CatalogPage) GetParamCSS(
	elem *elements.Element,
	paramName string,
) (string, error) {
	return elem.FindCSS(
		".text-center small",
		fmt.Sprintf("Parameter '%s'", paramName),
	).FilterByText(paramName, paramName).GetText()
}

// GetURL is an internal helper that extracts the href attribute from a child element.
func (p *CatalogPage) GetURL(
	elem *elements.Element,
	selector string,
) (string, error) {
	return elem.FindCSS(
		selector,
		"Product card URL retrieving",
	).GetAttribute("href")
}

// GetProductCardURL returns the absolute URL for the product card with the specified name.
func (p *CatalogPage) GetProductCardURL(name string) (string, error) {
	return p.GetURL(
		p.productCards.FilterByText(
			name,
			"Product card with name "+name,
		).First("Product card URL"),
		".product-card__name a",
	)
}

// AddToWishlist clicks the favorite icon on the product card to add it to wishlist.
func (p *CatalogPage) AddToWishlist(name string) error {
	p.Log.Debug("Click button that adds product to wishlist", "name", name)
	return p.productCards.FilterByText(
		name,
		"Product card with name "+name,
	).First(
		"First product card",
	).FindCSS(
		".product-card__favorites .favorite-icon",
		"Favorite icon",
	).Click()
}

// IsActiveIcon verifies that the favorite icon has the active class for the product.
// Returns error if the icon is not active, indicating the add to wishlist failed.
func (p *CatalogPage) IsActiveIcon(name string) error {
	p.Log.Debug(
		"Checking if favorite icon for product [%s] is active",
		"product",
		name,
	)

	productCard := p.productCards.FilterByText(
		name,
		"Card for "+name,
	)

	activeIcon := productCard.FindCSS(
		".product-card__favorites .favorite-icon.active",
		"Active favorite icon",
	)

	err := activeIcon.WaitForVisible()
	if err != nil {
		return fmt.Errorf("favorite icon for product [%s] is not active: %w", name, err)
	}

	return nil
}
