package e2e

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestGroupCreateFlow verifies the "New Group" modal creates a group
// that then appears in the manager.
func TestGroupCreateFlow(t *testing.T) {
	ctx := newTab(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	slug := "e2e-group-" + suffix
	description := "Created by e2e test"

	var body string
	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("group-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
		chromedp.WaitReady("#btn-new-group", chromedp.ByID),
		chromedp.Click("#btn-new-group", chromedp.ByID),
		chromedp.WaitReady("#group-create-slug", chromedp.ByID),
		chromedp.SendKeys("#group-create-slug", slug, chromedp.ByID),
		chromedp.SendKeys("#group-create-description", description, chromedp.ByID),
		chromedp.Click("#btn-create-group", chromedp.ByID),
		chromedp.Poll(`document.body.innerText.includes(`+strconv.Quote(slug)+`)`, nil,
			chromedp.WithPollingInterval(200*time.Millisecond)),
		chromedp.Text("body", &body),
	)
	if err != nil {
		t.Fatalf("group create flow failed: %v", err)
	}
	if !strings.Contains(body, slug) {
		t.Fatalf("expected newly created group slug %q in manager", slug)
	}
}
