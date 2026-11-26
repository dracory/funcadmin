package e2e

import (
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
)

func TestDashboardLoads(t *testing.T) {
	ctx := newTab(t)
	var body string
	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL+adminPath),
		chromedp.WaitReady("body"),
		chromedp.Text("body", &body),
	)
	if err != nil {
		t.Fatalf("dashboard failed: %v", err)
	}
	if strings.TrimSpace(body) == "" {
		t.Fatal("expected non-empty dashboard page")
	}
	if !strings.Contains(body, "faas. Dashboard") {
		t.Fatal("expected 'faas. Dashboard' header on dashboard")
	}
	if !strings.Contains(body, "Total Functions") {
		t.Fatal("expected 'Total Functions' card on dashboard")
	}
}
