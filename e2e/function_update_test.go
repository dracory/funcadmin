package e2e

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestFunctionUpdateSourceFlow verifies the source editor loads the
// existing source and saving persists the change.
func TestFunctionUpdateSourceFlow(t *testing.T) {
	ctx := newTab(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	fn := createTestFunction(t, "E2E Update Source "+suffix, "e2e-update-source-"+suffix, "return input")
	url := adminURL("function-update") + "&function_id=" + fn.GetID()
	newSource := "return input * 3"

	var body string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("#function-source", chromedp.ByID),
		chromedp.Text("body", &body),
	)
	if err != nil {
		t.Fatalf("function update page did not load: %v", err)
	}
	if !strings.Contains(body, "Edit Function") {
		t.Fatal("expected 'Edit Function' header on function update page")
	}
	if !strings.Contains(body, "Source Code") {
		t.Fatal("expected 'Source Code' panel on function update page")
	}

	err = chromedp.Run(ctx,
		chromedp.SetValue("#function-source", newSource, chromedp.ByID),
		chromedp.Click("#btn-save-source", chromedp.ByID),
		chromedp.Poll(`document.body.innerText.includes('Source saved successfully')`, nil,
			chromedp.WithPollingInterval(200*time.Millisecond)),
	)
	if err != nil {
		t.Fatalf("function source save failed: %v", err)
	}

	var persisted string
	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("#function-source", chromedp.ByID),
		chromedp.Value("#function-source", &persisted, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("reload of function update failed: %v", err)
	}
	if persisted != newSource {
		t.Fatalf("expected persisted source %q, got %q", newSource, persisted)
	}
}

// TestFunctionVersionsLoad verifies saved source versions render in the
// versions tab after a function source change.
func TestFunctionVersionsLoad(t *testing.T) {
	ctx := newTab(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	fn := createTestFunction(t, "E2E Versions "+suffix, "e2e-versions-"+suffix, "return input")
	fn.SetSource("return input + 7")
	if err := testStore.FunctionUpdate(context.Background(), fn); err != nil {
		t.Fatalf("create source version: %v", err)
	}

	url := adminURL("function-update") + "&function_id=" + fn.GetID() + "&view=versions"
	var body string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Poll(`!!document.querySelector('#versionsApp[data-v-app]')`, nil,
			chromedp.WithPollingInterval(200*time.Millisecond)),
		chromedp.Poll(`document.body.innerText.includes('return input')`, nil,
			chromedp.WithPollingInterval(200*time.Millisecond)),
		chromedp.Text("body", &body),
	)
	if err != nil {
		t.Fatalf("versions page failed: %v", err)
	}
	if !strings.Contains(body, "Source Versions") {
		t.Fatal("expected 'Source Versions' header on versions page")
	}
	if !strings.Contains(body, "return input") {
		t.Fatal("expected saved version source preview on versions page")
	}
}
