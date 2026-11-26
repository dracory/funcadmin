package function_manager

import (
	"embed"
	"encoding/json"
	"net/http"

	"github.com/dracory/funcadmin/shared"
	"github.com/dracory/funcstore"

	"github.com/dracory/api"
	"github.com/dracory/cdn"
	"github.com/dracory/hb"
	"github.com/dracory/req"
)

//go:embed *.html
//go:embed *.js
var functionsFiles embed.FS

const (
	actionLoadFunctions  = "load-functions"
	actionDeleteFunction = "delete-function"
	actionCreateFunction = "create-function"
)

// UiInterface defines the function manager controller's UI interface
type UiInterface interface {
	shared.UiInterface
	FunctionManager(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new function manager controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// FunctionManager handles the function manager controller requests
func (u *ui) FunctionManager(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

// Handler processes the function manager controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	action := req.GetStringTrimmed(r, "action")

	switch action {
	case actionLoadFunctions:
		return u.handleLoadFunctions(w, r)
	case actionDeleteFunction:
		return u.handleDeleteFunction(w, r)
	case actionCreateFunction:
		return u.handleCreateFunction(w, r)
	default:
		return u.renderPage(w, r)
	}
}

func (u *ui) renderPage(w http.ResponseWriter, r *http.Request) string {
	breadcrumbs := shared.Breadcrumbs([]shared.Breadcrumb{
		{Name: "Home", URL: shared.AdminHomeURL(r)},
		{Name: "FaaS", URL: shared.URLR(r, shared.CONTROLLER_DASHBOARD, nil)},
		{Name: "Function Manager", URL: ""},
	})

	linksHelper := shared.NewLinksFromRequest(r)

	heading := hb.Heading1().HTML("faas. Function Manager")

	htmlContent, err := functionsFiles.ReadFile("functions.html")
	if err != nil {
		u.Logger().Error("Failed to read functions HTML template", "error", err)
		return hb.Div().HTML("Error loading functions component").ToHTML()
	}

	jsContent, err := functionsFiles.ReadFile("functions.js")
	if err != nil {
		u.Logger().Error("Failed to read functions JavaScript file", "error", err)
		return hb.Div().HTML("Error loading functions component").ToHTML()
	}

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())
	notiflixCSS := hb.Link().Rel("stylesheet").Href(cdn.Notiflix_3_2_8_CSS())
	notiflixJS := hb.Script("").Src(cdn.Notiflix_3_2_8())

	initScript := hb.Script(`
		const urlFunctionsLoad = '` + linksHelper.FunctionManager(map[string]string{"action": actionLoadFunctions}) + `';
		const urlFunctionDelete = '` + linksHelper.FunctionManager(map[string]string{"action": actionDeleteFunction}) + `';
		const urlFunctionCreate = '` + linksHelper.FunctionManager(map[string]string{"action": actionCreateFunction}) + `';
		const urlFunctionUpdate = '` + linksHelper.FunctionUpdate(map[string]string{"function_id": "FUNCTION_ID_PLACEHOLDER"}) + `';
		const urlFunctionInvoke = '` + linksHelper.FunctionInvoke(map[string]string{"function_id": "FUNCTION_ID_PLACEHOLDER"}) + `';
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
		Child(hb.HR()).
		Child(vueContainer)

	return u.Layout(w, r, "FaaS | Function Manager", content.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{})
}

func (u *ui) handleLoadFunctions(w http.ResponseWriter, r *http.Request) string {
	store := u.Store()
	if store == nil {
		return api.Error("FaaS store not available").ToString()
	}

	fns, err := store.FunctionList(r.Context())
	if err != nil {
		u.Logger().Error("Failed to load functions", "error", err)
		return api.Error("Failed to load functions").ToString()
	}

	functionList := []map[string]any{}
	for _, fn := range fns {
		functionList = append(functionList, map[string]any{
			"id":         fn.GetID(),
			"alias":      fn.GetAlias(),
			"name":       fn.GetName(),
			"runtime":    fn.GetRuntime(),
			"source":     fn.GetSource(),
			"timeout_ms": fn.GetTimeoutMs(),
			"created_at": fn.GetCreatedAt(),
			"updated_at": fn.GetUpdatedAt(),
		})
	}

	return api.SuccessWithData("Functions loaded successfully", map[string]any{
		"functions": functionList,
		"total":     len(functionList),
	}).ToString()
}

func (u *ui) handleDeleteFunction(w http.ResponseWriter, r *http.Request) string {
	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		FunctionID string `json:"function_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.FunctionID == "" {
		return api.Error("Function ID is required").ToString()
	}

	store := u.Store()
	if store == nil {
		return api.Error("FaaS store not available").ToString()
	}

	if err := store.FunctionDelete(r.Context(), reqData.FunctionID); err != nil {
		u.Logger().Error("Failed to delete function", "error", err)
		return api.Error("Failed to delete function").ToString()
	}

	return api.SuccessWithData("Function deleted successfully", map[string]any{}).ToString()
}

func (u *ui) handleCreateFunction(w http.ResponseWriter, r *http.Request) string {
	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		Name      string `json:"name"`
		Alias     string `json:"alias"`
		Runtime   string `json:"runtime"`
		Source    string `json:"source"`
		TimeoutMs int64  `json:"timeout_ms"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.Source == "" {
		return api.Error("Source is required").ToString()
	}

	if reqData.Runtime == "" {
		reqData.Runtime = funcstore.RUNTIME_JS
	}

	store := u.Store()
	if store == nil {
		return api.Error("FaaS store not available").ToString()
	}

	fn := funcstore.NewFunction()
	fn.SetName(reqData.Name)
	fn.SetAlias(reqData.Alias)
	fn.SetRuntime(reqData.Runtime)
	fn.SetSource(reqData.Source)
	if reqData.TimeoutMs > 0 {
		fn.SetTimeoutMs(reqData.TimeoutMs)
	}

	if err := store.FunctionCreate(r.Context(), fn); err != nil {
		u.Logger().Error("Failed to create function", "error", err)
		return api.Error("Failed to create function: " + err.Error()).ToString()
	}

	return api.SuccessWithData("Function created successfully", map[string]any{
		"id": fn.GetID(),
	}).ToString()
}

// Ensure FaaS import is used
var _ funcstore.FunctionInterface
