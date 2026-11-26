package e2e

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestFunctionInvokeFlow verifies the test-invoke screen executes a
// seeded JavaScript function and displays its result.
func TestFunctionInvokeFlow(t *testing.T) {
	ctx := newTab(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	fn := createTestFunction(t, "E2E Invoke "+suffix, "e2e-invoke-"+suffix, "return input * 2")
	url := adminURL("function-invoke") + "&function_id=" + fn.GetID()

	var body string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Poll(`!!document.querySelector('#invokeApp[data-v-app]')`, nil,
			chromedp.WithPollingInterval(200*time.Millisecond)),
		chromedp.WaitReady("#invoke-input", chromedp.ByID),
		chromedp.Text("body", &body),
	)
	if err != nil {
		t.Fatalf("invoke page did not load: %v", err)
	}
	if !strings.Contains(body, "Test Invoke") {
		t.Fatal("expected 'Test Invoke' header on invoke page")
	}
	if !strings.Contains(body, "Source Preview") {
		t.Fatal("expected 'Source Preview' panel on invoke page")
	}

	var result string
	err = chromedp.Run(ctx,
		chromedp.SendKeys("#invoke-input", "21", chromedp.ByID),
		chromedp.Click("#btn-invoke-run", chromedp.ByID),
		chromedp.WaitReady("#invoke-result", chromedp.ByID),
		chromedp.Text("#invoke-result", &result, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("function invoke failed: %v", err)
	}
	if strings.TrimSpace(result) != "42" {
		t.Fatalf("expected invoke result 42, got %q", result)
	}
}
