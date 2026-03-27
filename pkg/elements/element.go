package elements

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/playwright-community/playwright-go"
)

// LocatorType defines the element location strategy.
type LocatorType string

// Supported locator types.
const (
	CSS   LocatorType = "css"   // CSS selector
	XPath LocatorType = "xpath" // XPath expression
)

// Element wraps a Playwright locator with explicit waits and descriptive logging.
// It provides a fluent API for element interactions with automatic timeout handling.
type Element struct {
	// page is the Playwright page for creating child elements.
	page playwright.Page
	// locator is the Playwright locator for this element.
	locator playwright.Locator
	// description is the human-readable name for logging.
	description string
	// timeout is the default wait timeout in milliseconds.
	timeout time.Duration
	// log is the element-scoped logger.
	log *slog.Logger
}

// NewCSS creates an Element with the specified CSS selector.
func NewCSS(
	page playwright.Page,
	selector string,
	description string,
	timeout time.Duration,
	log *slog.Logger,
) *Element {
	return newElement(page, selector, description, CSS, timeout, log)
}

// NewXPath creates an Element with the specified XPath expression.
func NewXPath(
	page playwright.Page,
	xpath string,
	description string,
	timeout time.Duration,
	log *slog.Logger,
) *Element {
	return newElement(page, "xpath="+xpath, description, XPath, timeout, log)
}

// newElement is the internal constructor that creates an Element with the given locator.
func newElement(
	page playwright.Page,
	selector string,
	description string,
	lt LocatorType,
	timeout time.Duration,
	log *slog.Logger,
) *Element {
	log.Debug("Creating element", "element", description, "type", lt, "selector", selector)

	return &Element{
		page:        page,
		locator:     page.Locator(selector),
		description: description,
		timeout:     timeout,
		log:         log,
	}
}

// WaitForVisible waits for the element to be visible in the DOM.
// Returns error on timeout.
func (e *Element) WaitForVisible() error {
	e.log.Debug("Waiting for element to be visible", "element", e.description)

	if err := e.locator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: new(float64(e.timeout)),
	}); err != nil {
		return fmt.Errorf("element [%s] not visible after %v: %w", e.description, e.timeout, err)
	}

	return nil
}

// WaitForHidden waits for the element to be hidden or detached from DOM.
// Returns error on timeout.
func (e *Element) WaitForHidden() error {
	e.log.Debug("Waiting for element to be hidden", "element", e.description)

	if err := e.locator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: new(float64(e.timeout)),
	}); err != nil {
		return fmt.Errorf("element [%s] not visible after %v: %w", e.description, e.timeout, err)
	}

	return nil
}

// Click clicks the element after waiting for it to be actionable.
// It handles scrolling into view and waits for visibility.
func (e *Element) Click() error {
	e.log.Debug("Clicking element", "element", e.description)

	if err := e.locator.Click(); err != nil {
		return fmt.Errorf("failed to click [%s]: %w", e.description, err)
	}

	e.log.Debug("Clicked element", "element", e.description)

	return nil
}

// Fill clears the input and fills it with the specified text.
// It does not dispatch key events (use Press for keyboard shortcuts).
func (e *Element) Fill(text string) error {
	e.log.Debug("Filling element", "element", e.description, "text", text)

	if err := e.locator.Fill(text); err != nil {
		return fmt.Errorf("failed to fill [%s]: %w", e.description, err)
	}

	return nil
}

// Clear clears the input field content.
// It waits for the element to be visible before clearing.
func (e *Element) Clear() error {
	e.log.Debug("Clearing element", "element", e.description)

	if err := e.WaitForVisible(); err != nil {
		return err
	}

	if err := e.locator.Clear(); err != nil {
		return fmt.Errorf("failed to clear [%s]: %w", e.description, err)
	}

	return nil
}

// GetText returns the visible text content of the element including all children.
// Returns empty string if the element is detached from DOM.
func (e *Element) GetText() (string, error) {
	e.log.Debug("Getting text from element", "element", e.description)

	text, err := e.locator.TextContent()
	if err != nil {
		return "", fmt.Errorf("failed to get text from [%s]: %w", e.description, err)
	}

	e.log.Debug("Got text from element", "element", e.description, "text", text)

	return text, nil
}

// GetAttribute returns the value of the specified attribute.
// Returns empty string if the attribute does not exist.
func (e *Element) GetAttribute(attr string) (string, error) {
	e.log.Debug(
		"Getting attribute",
		"element", e.description,
		"attribute", attr,
	)

	value, err := e.locator.GetAttribute(attr)
	if err != nil {
		return "", fmt.Errorf(
			"failed to get attribute [%s] from [%s]: %w",
			attr,
			e.description,
			err,
		)
	}

	return value, nil
}

// IsVisible checks if the element is visible without waiting.
// Returns false for hidden or detached elements.
func (e *Element) IsVisible() (bool, error) {
	visible, err := e.locator.IsVisible()
	if err != nil {
		return false, fmt.Errorf("failed to check visibility of [%s]: %w", e.description, err)
	}

	return visible, nil
}

// IsEnabled checks if the element is enabled (not disabled).
// Returns false for disabled elements.
func (e *Element) IsEnabled() (bool, error) {
	enabled, err := e.locator.IsEnabled()
	if err != nil {
		return false, fmt.Errorf("failed to check enabled state of [%s]: %w", e.description, err)
	}

	return enabled, nil
}

// SelectOption selects an option in a dropdown by its value attribute.
// Use this for <select> elements.
func (e *Element) SelectOption(value string) error {
	e.log.Debug("Selecting option", "value", value, "element", e.description)

	_, err := e.locator.SelectOption(playwright.SelectOptionValues{
		Values: &[]string{value},
	})
	if err != nil {
		return fmt.Errorf(
			"failed to select option [%s] in [%s]: %w",
			value,
			e.description,
			err,
		)
	}

	return nil
}

// Hover hovers the mouse over the element center.
// Does not scroll the element into view automatically.
func (e *Element) Hover() error {
	e.log.Debug("Hovering over element", "element", e.description)

	if err := e.locator.Hover(); err != nil {
		return fmt.Errorf("failed to hover over [%s]: %w", e.description, err)
	}

	return nil
}

// ScrollIntoView scrolls the element into the viewport if not already visible.
// Use this for elements that are hidden below the fold.
func (e *Element) ScrollIntoView() error {
	e.log.Debug("Scrolling element into view", "element", e.description)

	if err := e.locator.ScrollIntoViewIfNeeded(); err != nil {
		return fmt.Errorf(
			"failed to scroll [%s] into view: %w",
			e.description,
			err,
		)
	}

	return nil
}

// FilterByText returns a new Element filtered to match the specified text.
// Useful for narrowing down elements in dynamic lists (e.g., finding a specific product card).
func (e *Element) FilterByText(text string, description string) *Element {
	e.log.Debug(
		"Filtering element by text",
		"element", e.description,
		"text", text,
	)

	return &Element{
		page: e.page,
		locator: e.locator.Filter(playwright.LocatorFilterOptions{
			HasText: text,
		}),
		description: fmt.Sprintf("%s [Text: %s]", e.description, text),
		timeout:     e.timeout,
		log:         e.log,
	}
}

// FindCSS finds a child element using the specified CSS selector.
// The search is scoped within this element's locator.
func (e *Element) FindCSS(subSelector string, description string) *Element {
	e.log.Debug(
		"Finding sub-element by CSS",
		"parent", e.description,
		"child", description,
	)

	return &Element{
		page:        e.page,
		locator:     e.locator.Locator(subSelector),
		description: fmt.Sprintf("%s -> %s", e.description, description),
		timeout:     e.timeout,
		log:         e.log,
	}
}

// FindXPath finds a child element using the specified XPath expression.
// The search is scoped within this element's locator.
func (e *Element) FindXPath(xpath string, description string) *Element {
	e.log.Debug(
		"Finding sub-element by XPath",
		"parent", e.description,
		"child", description,
	)

	return &Element{
		page:        e.page,
		locator:     e.locator.Locator("xpath=" + xpath),
		description: fmt.Sprintf("%s -> %s", e.description, description),
		timeout:     e.timeout,
		log:         e.log,
	}
}

// First returns the first element from the locator's matched elements.
// Use when the locator matches multiple elements but only the first is needed.
func (e *Element) First(description string) *Element {
	return &Element{
		page:        e.page,
		locator:     e.locator.First(),
		description: fmt.Sprintf("%s [First]", e.description),
		timeout:     e.timeout,
		log:         e.log,
	}
}

// Nth returns the element at the specified zero-based index.
// Use for accessing specific elements in a list.
func (e *Element) Nth(index int, description string) *Element {
	return &Element{
		page:        e.page,
		locator:     e.locator.Nth(index),
		description: fmt.Sprintf("%s[Index: %d]", e.description, index),
		timeout:     e.timeout,
		log:         e.log,
	}
}

// Count returns the number of elements matching the locator.
func (e *Element) Count() (int, error) {
	e.log.Debug("Counting element", "element", e.description)

	count, err := e.locator.Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count elements [%s]: %w", e.description, err)
	}

	return count, nil
}

// Blur removes focus from the element by dispatching a blur event.
// Use this to trigger validation or state changes that occur on focus loss.
func (e *Element) Blur() error {
	e.log.Debug("Removing focus from element", "element", e.description)

	if err := e.locator.Blur(); err != nil {
		return fmt.Errorf("failed to blur element [%s]: %w", e.description, err)
	}

	return nil
}

// Press simulates a keyboard key press on the element.
// Use for keyboard shortcuts (e.g., Enter, Tab, Escape) or special keys.
func (e *Element) Press(key string) error {
	e.log.Debug(
		"Pressing key on element",
		"element", e.description,
		"key", key,
	)

	if err := e.locator.Press(key); err != nil {
		return fmt.Errorf("failed to press key [%s] on [%s]: %w", key, e.description, err)
	}

	return nil
}

// GetBoundingBox returns the element's position and size as {X, Y, Width, Height}.
// It waits for the element to be visible before measuring.
// Returns nil if the element is detached from DOM.
func (e *Element) GetBoundingBox() (*playwright.Rect, error) {
	e.log.Debug(
		"Getting bounding box",
		"element",
		e.description,
	)

	if err := e.WaitForVisible(); err != nil {
		return nil, fmt.Errorf(
			"cannot get bounding box for [%s]: %w",
			e.description,
			err,
		)
	}

	box, err := e.locator.BoundingBox()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get bounding box for [%s]: %w",
			e.description,
			err,
		)
	}

	if box == nil {
		return nil, fmt.Errorf(
			"bounding box for [%s] is nil (element is not visible or detached)",
			e.description,
		)
	}

	return box, nil
}
