package e2e

import (
	"context"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/dracory/funcstore"
)

// adminURL builds a controller URL, e.g. /admin/FaaS?controller=function-manager.
func adminURL(controller string) string {
	return baseURL + adminPath + "?controller=" + controller
}

// waitTable navigates to a manager page and waits until Vue has loaded
// data and rendered the results table. Returns the page's text content.
//
// Waits for the SECOND table row: the raw v-for template row exists in
// the DOM before Vue mounts, so `table tr` alone can resolve while the
// table is still a template. A second row only exists after v-for has
// expanded with real data.
func waitTable(t *testing.T, ctx context.Context, url string) string {
	t.Helper()
	var body string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
		chromedp.Text("body", &body),
	)
	if err != nil {
		t.Fatalf("page %s failed: %v", url, err)
	}
	return body
}

func createTestFunction(t *testing.T, name, alias, source string) funcstore.FunctionInterface {
	t.Helper()
	fn := funcstore.NewFunction()
	fn.SetName(name).
		SetAlias(alias).
		SetRuntime(funcstore.RUNTIME_JS).
		SetSource(source)
	if err := testStore.FunctionCreate(context.Background(), fn); err != nil {
		t.Fatalf("create test function: %v", err)
	}
	return fn
}

func createTestGroup(t *testing.T, slug, description string) funcstore.GroupInterface {
	t.Helper()
	group := funcstore.NewGroup()
	group.SetSlug(slug).SetDescription(description)
	created, err := testStore.GroupCreate(context.Background(), group)
	if err != nil {
		t.Fatalf("create test group: %v", err)
	}
	return created
}
