package e2e

import (
	"strconv"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestGroupUpdateFlow verifies the group editor loads the existing group
// and saving persists description changes.
func TestGroupUpdateFlow(t *testing.T) {
	ctx := newTab(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	group := createTestGroup(t, "e2e-update-group-"+suffix, "initial description")
	url := adminURL("group-update") + "&group_id=" + group.GetID()
	newDescription := "Updated by e2e " + suffix

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Poll(`!!document.querySelector('#groupEditApp[data-v-app]')`, nil,
			chromedp.WithPollingInterval(200*time.Millisecond)),
		chromedp.WaitReady("#group-edit-description", chromedp.ByID),
		chromedp.Poll(`document.querySelector('#group-edit-description').value === 'initial description'`, nil,
			chromedp.WithPollingInterval(200*time.Millisecond)),
	)
	if err != nil {
		t.Fatalf("group update form did not prefill: %v", err)
	}

	err = chromedp.Run(ctx,
		chromedp.SetValue("#group-edit-description", newDescription, chromedp.ByID),
		chromedp.Click("#btn-save-group", chromedp.ByID),
		chromedp.Poll(`document.body.innerText.includes('Group saved successfully')`, nil,
			chromedp.WithPollingInterval(200*time.Millisecond)),
	)
	if err != nil {
		t.Fatalf("group update submit failed: %v", err)
	}

	var persisted string
	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Poll(`!!document.querySelector('#groupEditApp[data-v-app]')`, nil,
			chromedp.WithPollingInterval(200*time.Millisecond)),
		chromedp.WaitReady("#group-edit-description", chromedp.ByID),
		chromedp.Poll(`document.querySelector('#group-edit-description').value !== ''`, nil,
			chromedp.WithPollingInterval(200*time.Millisecond)),
		chromedp.Value("#group-edit-description", &persisted, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("reload of group update failed: %v", err)
	}
	if persisted != newDescription {
		t.Fatalf("expected persisted description %q, got %q", newDescription, persisted)
	}
}
