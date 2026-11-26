package e2e

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestFunctionCreateFlow verifies the "New Function" modal creates a
// function that then appears in the manager.
func TestFunctionCreateFlow(t *testing.T) {
	ctx := newTab(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	name := "E2E Function " + suffix
	alias := "e2e-fn-" + suffix
	source := "return input + 1"

	var body string
	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("function-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
		chromedp.WaitReady("#btn-new-function", chromedp.ByID),
		chromedp.Click("#btn-new-function", chromedp.ByID),
		chromedp.WaitReady("#function-create-source", chromedp.ByID),
		chromedp.SendKeys("#function-create-name", name, chromedp.ByID),
		chromedp.SendKeys("#function-create-alias", alias, chromedp.ByID),
		chromedp.SendKeys("#function-create-source", source, chromedp.ByID),
		chromedp.Click("#btn-create-function", chromedp.ByID),
		chromedp.Poll(`document.body.innerText.includes(`+strconv.Quote(alias)+`)`, nil,
			chromedp.WithPollingInterval(200*time.Millisecond)),
		chromedp.Text("body", &body),
	)
	if err != nil {
		t.Fatalf("function create flow failed: %v", err)
	}
	if !strings.Contains(body, alias) {
		t.Fatalf("expected newly created function alias %q in manager", alias)
	}
}
