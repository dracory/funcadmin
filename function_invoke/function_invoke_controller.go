package function_invoke

import (
	"embed"
	"encoding/json"
	"net/http"

	"github.com/dracory/funcadmin/shared"

	"github.com/dracory/api"
	"github.com/dracory/cdn"
	"github.com/dracory/hb"
	"github.com/dracory/req"
)

//go:embed *.html
//go:embed *.js
var invokeFiles embed.FS

const (
	actionInvoke = "invoke"
)

// UiInterface defines the function invoke controller's UI interface
type UiInterface interface {
	shared.UiInterface
	FunctionInvoke(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new function invoke controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// FunctionInvoke handles the function invoke controller requests
func (u *ui) FunctionInvoke(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

// Handler processes the function invoke controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	action := req.GetStringTrimmed(r, "action")

	switch action {
	case actionInvoke:
		return u.handleInvoke(w, r)
	default:
		return u.renderPage(w, r)
	}
}

func (u *ui) renderPage(w http.ResponseWriter, r *http.Request) string {
	functionID := req.GetStringTrimmed(r, "function_id")

	if functionID == "" {
		return shared.ErrorAlert("Function ID is required")
	}

	fn, err := u.Store().FunctionFindByID(r.Context(), functionID)
	if err != nil {
		return shared.ErrorAlert("Function not found")
	}

	breadcrumbs := shared.Breadcrumbs([]shared.Breadcrumb{
		{Name: "Home", URL: shared.AdminHomeURL(r)},
		{Name: "FaaS", URL: shared.URLR(r, shared.CONTROLLER_DASHBOARD, nil)},
		{Name: "Function Manager", URL: shared.NewLinksFromRequest(r).FunctionManager(nil)},
		{Name: "Test Invoke", URL: ""},
	})

	linksHelper := shared.NewLinksFromRequest(r)

	buttonBack := hb.Hyperlink().
		Class("btn btn-secondary ms-2 float-end").
		Child(hb.I().Class("bi bi-chevron-left me-2")).
		HTML("Back").
		Href(linksHelper.FunctionUpdate(map[string]string{"function_id": functionID}))

	heading := hb.Heading1().
		HTML("Test Invoke").
		Child(buttonBack)

	fnTitle := hb.Heading2().
		Class("mb-3").
		HTML("Function: ").
		Text(fn.GetName())

	htmlContent, err := invokeFiles.ReadFile("invoke.html")
	if err != nil {
		u.Logger().Error("Failed to read invoke HTML template", "error", err)
		return hb.Div().HTML("Error loading invoke component").ToHTML()
	}

	jsContent, err := invokeFiles.ReadFile("invoke.js")
	if err != nil {
		u.Logger().Error("Failed to read invoke JavaScript file", "error", err)
		return hb.Div().HTML("Error loading invoke component").ToHTML()
	}

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())
	notiflixCSS := hb.Link().Rel("stylesheet").Href(cdn.Notiflix_3_2_8_CSS())
	notiflixJS := hb.Script("").Src(cdn.Notiflix_3_2_8())

	initScript := hb.Script(`
		const urlInvoke = '` + linksHelper.FunctionInvoke(map[string]string{"function_id": functionID, "action": actionInvoke}) + `';
		const functionId = '` + functionID + `';
		const functionRuntime = '` + fn.GetRuntime() + `';
		const functionSource = ` + jsonSource(fn.GetSource()) + `;
	`)

	htmlTemplate := hb.Wrap().HTML(string(htmlContent))
	componentScript := hb.Script(string(jsContent))

	vueContainer := hb.Div().
		Child(vueCDN).
		Child(notiflixCSS).
		Child(notiflixJS).
		Child(htmlTemplate).
		Child(initScript).
		Child(componentScript)

	content := hb.Div().
		Class("container").
		Child(heading).
		Child(breadcrumbs).
		Child(fnTitle).
		Child(vueContainer)

	return u.Layout(w, r, "Test Invoke | FaaS", content.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{})
}

func (u *ui) handleInvoke(w http.ResponseWriter, r *http.Request) string {
	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		FunctionID string `json:"function_id"`
		Input      any    `json:"input"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body: " + err.Error()).ToString()
	}

	if reqData.FunctionID == "" {
		return api.Error("Function ID is required").ToString()
	}

	result, err := u.Store().FunctionInvoke(r.Context(), reqData.FunctionID, reqData.Input)
	if err != nil {
		u.Logger().Error("Invoke failed", "error", err, "function_id", reqData.FunctionID)
		return api.Error("Invoke failed: " + err.Error()).ToString()
	}

	return api.SuccessWithData("Function invoked successfully", map[string]any{
		"result": result,
	}).ToString()
}

// jsonSource safely encodes the source as a JSON string for embedding
// in a JavaScript init script.
func jsonSource(source string) string {
	b, _ := json.Marshal(source)
	return string(b)
}
