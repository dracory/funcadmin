package dashboard

import (
	"net/http"
	"strconv"

	"github.com/dracory/funcadmin/shared"

	"github.com/dracory/cdn"
	"github.com/dracory/hb"
)

// UiInterface defines the dashboard controller's UI interface
type UiInterface interface {
	shared.UiInterface
	Dashboard(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new dashboard controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// Dashboard handles the dashboard controller requests
func (u *ui) Dashboard(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

// Handler processes the dashboard controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	data, errorMessage := u.prepareData(r)

	if errorMessage != "" {
		return shared.ErrorAlert(errorMessage)
	}

	content := u.page(r, data)

	return u.Layout(w, r, "FaaS | Dashboard", content.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{
		ScriptURLs: []string{
			cdn.Htmx_2_0_0(),
		},
	})
}

func (u *ui) page(r *http.Request, data dashboardData) hb.TagInterface {
	breadcrumbs := shared.Breadcrumbs([]shared.Breadcrumb{
		{Name: "Home", URL: shared.AdminHomeURL(r)},
		{Name: "FaaS", URL: shared.URLR(r, shared.CONTROLLER_DASHBOARD, nil)},
		{Name: "Dashboard", URL: ""},
	})

	navTabs := u.navTabs(r, data)

	title := hb.Heading1().HTML("faas. Dashboard")

	content := hb.Wrap().
		Child(navTabs).
		Child(hb.BR()).
		Child(u.dashboardCards(r, data))

	return hb.Div().
		Class("container").
		Class("py-4").
		Child(title).
		Child(breadcrumbs).
		Child(content)
}

func (u *ui) navTabs(r *http.Request, data dashboardData) hb.TagInterface {
	linksHelper := shared.NewLinksFromRequest(r)

	children := []hb.TagInterface{
		hb.Hyperlink().
			Class("text-decoration-none").
			HTML("Dashboard").
			Href(linksHelper.Dashboard(nil)),
		hb.Hyperlink().
			Class("text-decoration-none").
			Child(
				hb.Wrap().
					Text("Functions ").
					Child(
						hb.Span().
							Class("badge bg-secondary ms-1").
							Text(strconv.FormatInt(data.functionCount, 10)),
					),
			).
			Href(linksHelper.FunctionManager(nil)),
		hb.Hyperlink().
			Class("text-decoration-none").
			HTML("Groups").
			Href(linksHelper.GroupManager(nil)),
	}

	return hb.Div().
		Class("card mb-4").
		Child(
			hb.Div().
				Class("card-body d-flex justify-content-center gap-4").
				Children(children),
		)
}

func (u *ui) dashboardCards(r *http.Request, data dashboardData) hb.TagInterface {
	cards := []hb.TagInterface{
		hb.Div().Class("col-md-4").Child(u.cardFunctions(r, data)),
	}

	return hb.Div().
		Class("row").
		Children(cards)
}

func (u *ui) cardFunctions(r *http.Request, data dashboardData) hb.TagInterface {
	linksHelper := shared.NewLinksFromRequest(r)

	return hb.Hyperlink().
		Class("text-decoration-none").
		Href(linksHelper.FunctionManager(nil)).
		Child(
			hb.Div().
				Class("card mb-4").
				Style("background-color: #a8d5ba; border: none;").
				Child(
					hb.Div().
						Class("card-body d-flex justify-content-between align-items-center").
						Children([]hb.TagInterface{
							hb.Div().
								Children([]hb.TagInterface{
									hb.Heading3().Class("mb-0").Text(strconv.FormatInt(data.functionCount, 10)).Style("color: #2c5f2d;"),
									hb.Paragraph().Class("mb-0").Text("Total Functions").Style("color: #2c5f2d;"),
								}),
							hb.I().Class("bi bi-lightning-charge fs-1").Style("color: rgba(44, 95, 45, 0.3);"),
						}),
				).
				Child(
					hb.Div().
						Class("card-footer bg-transparent border-0 text-center pb-3").
						Child(
							hb.Span().
								Class("small fw-medium").
								Text("More info ").
								Style("color: #2c5f2d;").
								Child(hb.I().Class("bi bi-arrow-right-circle")),
						),
				),
		)
}

func (u *ui) prepareData(r *http.Request) (data dashboardData, errorMessage string) {
	store := u.Store()
	if store == nil {
		return data, "FaaS store not available"
	}

	fns, err := store.FunctionList(r.Context())
	if err != nil {
		u.Logger().Error("FaaS dashboard: error listing functions", "error", err)
		data.functionCount = -1
	} else {
		data.functionCount = int64(len(fns))
	}

	return data, ""
}

type dashboardData struct {
	functionCount int64
}
