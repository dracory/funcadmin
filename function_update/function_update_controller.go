package function_update

import (
	"embed"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dracory/funcadmin/shared"
	"github.com/dracory/funcstore"

	"github.com/dracory/api"
	"github.com/dracory/bs"
	"github.com/dracory/cdn"
	"github.com/dracory/hb"
	"github.com/dracory/req"
)

//go:embed versions.html
//go:embed versions.js
var versionFiles embed.FS

// UiInterface defines the function update controller's UI interface
type UiInterface interface {
	shared.UiInterface
	FunctionUpdate(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new function update controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// FunctionUpdate handles the function update controller requests
func (u *ui) FunctionUpdate(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

const (
	actionSaveSource     = "save-source"
	actionSaveSettings   = "save-settings"
	actionLoadVersions   = "load-versions"
	actionRestoreVersion = "restore-version"
	actionDeleteVersion  = "delete-version"
)

// Handler processes the function update controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	action := req.GetStringTrimmed(r, "action")
	functionID := req.GetStringTrimmed(r, "function_id")

	if action != "" {
		switch action {
		case actionSaveSource:
			return u.handleSaveSource(w, r)
		case actionSaveSettings:
			return u.handleSaveSettings(w, r)
		case actionLoadVersions:
			return u.handleLoadVersions(w, r)
		case actionRestoreVersion:
			return u.handleRestoreVersion(w, r)
		case actionDeleteVersion:
			return u.handleDeleteVersion(w, r)
		}
	}

	if functionID == "" {
		return shared.ErrorAlert("Function ID is required")
	}

	fn, err := u.Store().FunctionFindByID(r.Context(), functionID)
	if err != nil {
		u.Logger().Error("functionUpdateController: FunctionFindByID", "error", err, "function_id", functionID)
		return shared.ErrorAlert("Function not found")
	}

	if fn == nil {
		return shared.ErrorAlert("Function not found")
	}

	view := req.GetStringTrimmedOr(r, "view", "source")
	pageContent := u.page(r, fn, view)

	return u.Layout(w, r, "Edit Function | FaaS", pageContent.ToHTML(), struct {
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

func (u *ui) page(r *http.Request, fn funcstore.FunctionInterface, view string) hb.TagInterface {
	linksHelper := shared.NewLinksFromRequest(r)

	breadcrumbs := shared.Breadcrumbs([]shared.Breadcrumb{
		{Name: "Home", URL: shared.AdminHomeURL(r)},
		{Name: "FaaS", URL: shared.URLR(r, shared.CONTROLLER_DASHBOARD, nil)},
		{Name: "Function Manager", URL: linksHelper.FunctionManager(nil)},
		{Name: "Edit Function", URL: linksHelper.FunctionUpdate(map[string]string{"function_id": fn.GetID()})},
	})

	buttonBack := hb.Hyperlink().
		Class("btn btn-secondary ms-2 float-end").
		Child(hb.I().Class("bi bi-chevron-left me-2")).
		HTML("Back").
		Href(linksHelper.FunctionManager(nil))

	buttonInvoke := hb.Hyperlink().
		Class("btn btn-success ms-2 float-end").
		Child(hb.I().Class("bi bi-play me-2")).
		HTML("Test Invoke").
		Href(linksHelper.FunctionInvoke(map[string]string{"function_id": fn.GetID()}))

	heading := hb.Heading1().
		HTML("Edit Function").
		Child(buttonBack).
		Child(buttonInvoke)

	tabs := bs.NavTabs().
		Class("mb-3").
		Child(bs.NavItem().
			Child(bs.NavLink().
				ClassIf(view == "source", "active").
				Href(linksHelper.FunctionUpdate(map[string]string{"function_id": fn.GetID(), "view": "source"})).
				HTML("Source"))).
		Child(bs.NavItem().
			Child(bs.NavLink().
				ClassIf(view == "settings", "active").
				Href(linksHelper.FunctionUpdate(map[string]string{"function_id": fn.GetID(), "view": "settings"})).
				HTML("Settings"))).
		Child(bs.NavItem().
			Child(bs.NavLink().
				ClassIf(view == "versions", "active").
				Href(linksHelper.FunctionUpdate(map[string]string{"function_id": fn.GetID(), "view": "versions"})).
				HTML("Versions")))

	fnTitle := hb.Heading2().
		Class("mb-3").
		HTML("Function: ").
		Text(fn.GetName())

	var body hb.TagInterface

	switch view {
	case "source":
		body = u.renderSourceView(r, fn)
	case "settings":
		body = u.renderSettingsView(r, fn)
	case "versions":
		body = u.renderVersionsView(r, fn)
	default:
		body = u.renderSourceView(r, fn)
	}

	card := hb.Div().
		Class("card").
		Child(
			hb.Div().
				Class("card-header").
				Child(hb.Heading4().
					HTMLIf(view == "source", "Source Code").
					HTMLIf(view == "settings", "Function Settings").
					HTMLIf(view == "versions", "Source Versions").
					Style("margin-bottom:0;display:inline-block;")),
		).
		Child(
			hb.Div().
				Class("card-body").
				Child(body),
		)

	return hb.Div().
		Class("container").
		Child(heading).
		Child(breadcrumbs).
		Child(fnTitle).
		Child(tabs).
		Child(card)
}

func (u *ui) renderSourceView(r *http.Request, fn funcstore.FunctionInterface) hb.TagInterface {
	linksHelper := shared.NewLinksFromRequest(r)
	submitUrl := linksHelper.FunctionUpdate(map[string]string{
		"function_id": fn.GetID(),
		"action":      actionSaveSource,
	})

	formGroupSource := bs.FormGroup().
		Class("mb-3").
		Child(bs.FormLabel("Source Code (JavaScript)")).
		Child(hb.Textarea().
			Class("form-control").
			Name("source").
			Attr("rows", "20").
			Style("font-family: monospace; font-size: 14px;").
			Text(fn.GetSource()))

	buttonSave := hb.Button().
		Class("btn btn-primary").
		Child(hb.I().Class("bi bi-check me-2")).
		HTML("Save Source").
		HxPost(submitUrl).
		HxInclude("#formSource").
		HxTarget("body").
		HxSwap("beforeend")

	return hb.Form().
		ID("formSource").
		Child(formGroupSource).
		Child(buttonSave)
}

func (u *ui) renderSettingsView(r *http.Request, fn funcstore.FunctionInterface) hb.TagInterface {
	linksHelper := shared.NewLinksFromRequest(r)
	submitUrl := linksHelper.FunctionUpdate(map[string]string{
		"function_id": fn.GetID(),
		"action":      actionSaveSettings,
	})

	formGroupName := bs.FormGroup().
		Class("mb-3").
		Child(bs.FormLabel("Name (display label)")).
		Child(bs.FormInput().Name("name").Value(fn.GetName()))

	formGroupAlias := bs.FormGroup().
		Class("mb-3").
		Child(bs.FormLabel("Alias (unique callable key)")).
		Child(bs.FormInput().Name("alias").Value(fn.GetAlias()))

	formGroupRuntime := bs.FormGroup().
		Class("mb-3").
		Child(bs.FormLabel("Runtime")).
		Child(hb.Select().
			Class("form-select").
			Name("runtime").
			Child(hb.Option().Value("js").Text("JavaScript (Goja)").AttrIf(fn.GetRuntime() == "js", "selected", "selected")))

	formGroupTimeout := bs.FormGroup().
		Class("mb-3").
		Child(bs.FormLabel("Timeout (ms, 0 = default 5000)")).
		Child(bs.FormInput().Type("number").Name("timeout_ms").Value(strconv.FormatInt(fn.GetTimeoutMs(), 10)))

	formGroupMetadata := bs.FormGroup().
		Class("mb-3").
		Child(bs.FormLabel("Metadata (JSON, optional)")).
		Child(hb.Textarea().
			Class("form-control").
			Name("metadata").
			Attr("rows", "4").
			Style("font-family: monospace;").
			Text(fn.GetMetadata()))

	buttonSave := hb.Button().
		Class("btn btn-primary").
		Child(hb.I().Class("bi bi-check me-2")).
		HTML("Save Settings").
		HxPost(submitUrl).
		HxInclude("#formSettings").
		HxTarget("body").
		HxSwap("beforeend")

	return hb.Form().
		ID("formSettings").
		Child(formGroupName).
		Child(formGroupAlias).
		Child(formGroupRuntime).
		Child(formGroupTimeout).
		Child(formGroupMetadata).
		Child(buttonSave)
}

func (u *ui) handleSaveSource(w http.ResponseWriter, r *http.Request) string {
	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		FunctionID string `json:"function_id"`
		Source     string `json:"source"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		// Fallback to form values
		reqData.FunctionID = req.GetStringTrimmed(r, "function_id")
		reqData.Source = req.GetStringTrimmed(r, "source")
	}

	if reqData.FunctionID == "" {
		return api.Error("Function ID is required").ToString()
	}

	if reqData.Source == "" {
		return api.Error("Source is required").ToString()
	}

	fn, err := u.Store().FunctionFindByID(r.Context(), reqData.FunctionID)
	if err != nil {
		return api.Error("Function not found").ToString()
	}

	fn.SetSource(reqData.Source)

	if err := u.Store().FunctionUpdate(r.Context(), fn); err != nil {
		u.Logger().Error("Failed to update function source", "error", err)
		return api.Error("Failed to save source: " + err.Error()).ToString()
	}

	return api.Success("Source saved successfully").ToString()
}

func (u *ui) handleSaveSettings(w http.ResponseWriter, r *http.Request) string {
	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		FunctionID string `json:"function_id"`
		Name       string `json:"name"`
		Alias      string `json:"alias"`
		Runtime    string `json:"runtime"`
		TimeoutMs  int64  `json:"timeout_ms"`
		Metadata   string `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		reqData.FunctionID = req.GetStringTrimmed(r, "function_id")
		reqData.Name = req.GetStringTrimmed(r, "name")
		reqData.Alias = req.GetStringTrimmed(r, "alias")
		reqData.Runtime = req.GetStringTrimmedOr(r, "runtime", "js")
		reqData.TimeoutMs = req.GetInt64(r, "timeout_ms")
		reqData.Metadata = req.GetStringTrimmed(r, "metadata")
	}

	if reqData.FunctionID == "" {
		return api.Error("Function ID is required").ToString()
	}

	fn, err := u.Store().FunctionFindByID(r.Context(), reqData.FunctionID)
	if err != nil {
		return api.Error("Function not found").ToString()
	}

	fn.SetName(reqData.Name)
	fn.SetAlias(reqData.Alias)
	fn.SetRuntime(reqData.Runtime)
	fn.SetTimeoutMs(reqData.TimeoutMs)
	fn.SetMetadata(reqData.Metadata)

	if err := u.Store().FunctionUpdate(r.Context(), fn); err != nil {
		u.Logger().Error("Failed to update function settings", "error", err)
		return api.Error("Failed to save settings: " + err.Error()).ToString()
	}

	return api.Success("Settings saved successfully").ToString()
}

// === Versions ================================================================

func (u *ui) renderVersionsView(r *http.Request, fn funcstore.FunctionInterface) hb.TagInterface {
	linksHelper := shared.NewLinksFromRequest(r)

	htmlContent, err := versionFiles.ReadFile("versions.html")
	if err != nil {
		u.Logger().Error("Failed to read versions HTML template", "error", err)
		return hb.Div().HTML("Error loading versions component")
	}

	jsContent, err := versionFiles.ReadFile("versions.js")
	if err != nil {
		u.Logger().Error("Failed to read versions JavaScript file", "error", err)
		return hb.Div().HTML("Error loading versions component")
	}

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())
	notiflixCSS := hb.Link().Rel("stylesheet").Href(cdn.Notiflix_3_2_8_CSS())
	notiflixJS := hb.Script("").Src(cdn.Notiflix_3_2_8())

	initScript := hb.Script(`
		const urlVersionsLoad = '` + linksHelper.FunctionUpdate(map[string]string{"function_id": fn.GetID(), "action": actionLoadVersions}) + `';
		const urlVersionRestore = '` + linksHelper.FunctionUpdate(map[string]string{"function_id": fn.GetID(), "action": actionRestoreVersion}) + `';
		const urlVersionDelete = '` + linksHelper.FunctionUpdate(map[string]string{"function_id": fn.GetID(), "action": actionDeleteVersion}) + `';
		const versionFunctionId = '` + fn.GetID() + `';
	`)

	htmlTemplate := hb.Wrap().HTML(string(htmlContent))
	componentScript := hb.Script(string(jsContent))

	return hb.Div().
		Child(vueCDN).
		Child(notiflixCSS).
		Child(notiflixJS).
		Child(htmlTemplate).
		Child(initScript).
		Child(componentScript)
}

func (u *ui) handleLoadVersions(w http.ResponseWriter, r *http.Request) string {
	functionID := req.GetStringTrimmed(r, "function_id")
	if functionID == "" {
		return api.Error("Function ID is required").ToString()
	}

	store := u.Store()
	if store == nil {
		return api.Error("FaaS store not available").ToString()
	}

	versions, err := store.FunctionVersionList(r.Context(), functionID)
	if err != nil {
		u.Logger().Error("Failed to load versions", "error", err, "function_id", functionID)
		return api.Error("Failed to load versions").ToString()
	}

	versionList := []map[string]any{}
	for _, v := range versions {
		versionList = append(versionList, map[string]any{
			"id":          v.GetID(),
			"function_id": v.GetFunctionID(),
			"source":      v.GetSource(),
			"created_at":  v.GetCreatedAt(),
		})
	}

	return api.SuccessWithData("Versions loaded successfully", map[string]any{
		"versions": versionList,
		"total":    len(versionList),
	}).ToString()
}

func (u *ui) handleRestoreVersion(w http.ResponseWriter, r *http.Request) string {
	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		FunctionID string `json:"function_id"`
		VersionID  string `json:"version_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		reqData.FunctionID = req.GetStringTrimmed(r, "function_id")
		reqData.VersionID = req.GetStringTrimmed(r, "version_id")
	}

	if reqData.FunctionID == "" {
		return api.Error("Function ID is required").ToString()
	}
	if reqData.VersionID == "" {
		return api.Error("Version ID is required").ToString()
	}

	store := u.Store()
	if store == nil {
		return api.Error("FaaS store not available").ToString()
	}

	if err := store.FunctionRestore(r.Context(), reqData.FunctionID, reqData.VersionID); err != nil {
		u.Logger().Error("Failed to restore version", "error", err, "function_id", reqData.FunctionID, "version_id", reqData.VersionID)
		return api.Error("Failed to restore version: " + err.Error()).ToString()
	}

	return api.Success("Version restored successfully").ToString()
}

func (u *ui) handleDeleteVersion(w http.ResponseWriter, r *http.Request) string {
	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		VersionID string `json:"version_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		reqData.VersionID = req.GetStringTrimmed(r, "version_id")
	}

	if reqData.VersionID == "" {
		return api.Error("Version ID is required").ToString()
	}

	store := u.Store()
	if store == nil {
		return api.Error("FaaS store not available").ToString()
	}

	if err := store.FunctionVersionDelete(r.Context(), reqData.VersionID); err != nil {
		u.Logger().Error("Failed to delete version", "error", err, "version_id", reqData.VersionID)
		return api.Error("Failed to delete version").ToString()
	}

	return api.Success("Version deleted successfully").ToString()
}
