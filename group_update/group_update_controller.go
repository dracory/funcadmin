package group_update

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
var groupFiles embed.FS

const (
	actionLoadGroup = "load-group"
	actionSaveGroup = "save-group"
)

// UiInterface defines the group update controller's UI interface
type UiInterface interface {
	shared.UiInterface
	GroupUpdate(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new group update controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// GroupUpdate handles the group update controller requests
func (u *ui) GroupUpdate(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

// Handler processes the group update controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	action := req.GetStringTrimmed(r, "action")

	switch action {
	case actionLoadGroup:
		return u.handleLoadGroup(w, r)
	case actionSaveGroup:
		return u.handleSaveGroup(w, r)
	default:
		return u.renderPage(w, r)
	}
}

func (u *ui) renderPage(w http.ResponseWriter, r *http.Request) string {
	groupID := req.GetStringTrimmed(r, "group_id")
	if groupID == "" {
		return shared.ErrorAlert("Group ID is required")
	}

	breadcrumbs := shared.Breadcrumbs([]shared.Breadcrumb{
		{Name: "Home", URL: shared.AdminHomeURL(r)},
		{Name: "FaaS", URL: shared.URLR(r, shared.CONTROLLER_DASHBOARD, nil)},
		{Name: "Groups", URL: shared.URLR(r, shared.CONTROLLER_GROUP_MANAGER, nil)},
		{Name: "Edit Group", URL: ""},
	})

	linksHelper := shared.NewLinksFromRequest(r)

	heading := hb.Heading1().
		HTML("Edit Group").
		Child(
			hb.Hyperlink().
				Class("btn btn-secondary ms-2 float-end").
				HTML("Back").
				Href(linksHelper.GroupManager(nil)),
		)

	htmlContent, err := groupFiles.ReadFile("group.html")
	if err != nil {
		u.Logger().Error("Failed to read group HTML template", "error", err)
		return hb.Div().HTML("Error loading group component").ToHTML()
	}

	jsContent, err := groupFiles.ReadFile("group.js")
	if err != nil {
		u.Logger().Error("Failed to read group JavaScript file", "error", err)
		return hb.Div().HTML("Error loading group component").ToHTML()
	}

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())
	notiflixCSS := hb.Link().Rel("stylesheet").Href(cdn.Notiflix_3_2_8_CSS())
	notiflixJS := hb.Script("").Src(cdn.Notiflix_3_2_8())

	initScript := hb.Script(`
		const urlLoadGroup = '` + linksHelper.GroupUpdate(map[string]string{"action": actionLoadGroup, "group_id": groupID}) + `';
		const urlSaveGroup = '` + linksHelper.GroupUpdate(map[string]string{"action": actionSaveGroup, "group_id": groupID}) + `';
		const urlGroupManager = '` + linksHelper.GroupManager(nil) + `';
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

	return u.Layout(w, r, "Edit Group | FaaS", content.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{})
}

func (u *ui) handleLoadGroup(w http.ResponseWriter, r *http.Request) string {
	store := u.Store()
	if store == nil {
		return api.Error("FaaS store not available").ToString()
	}

	groupID := req.GetStringTrimmed(r, "group_id")
	if groupID == "" {
		return api.Error("Group ID is required").ToString()
	}

	g, err := store.GroupFindByID(r.Context(), groupID)
	if err != nil {
		return api.Error("Group not found").ToString()
	}

	return api.SuccessWithData("Group loaded successfully", map[string]any{
		"group": map[string]any{
			"id":          g.GetID(),
			"slug":        g.GetSlug(),
			"path":        g.GetPath(),
			"parent_id":   g.GetParentID(),
			"description": g.GetDescription(),
			"metadata":    g.GetMetadata(),
			"created_at":  g.GetCreatedAt(),
			"updated_at":  g.GetUpdatedAt(),
		},
	}).ToString()
}

func (u *ui) handleSaveGroup(w http.ResponseWriter, r *http.Request) string {
	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		GroupID     string `json:"group_id"`
		Description string `json:"description"`
		Metadata    string `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.GroupID == "" {
		return api.Error("Group ID is required").ToString()
	}

	store := u.Store()
	if store == nil {
		return api.Error("FaaS store not available").ToString()
	}

	g := funcstore.NewGroup()
	g.SetID(reqData.GroupID).
		SetDescription(reqData.Description).
		SetMetadata(reqData.Metadata)

	if err := store.GroupUpdate(r.Context(), g); err != nil {
		u.Logger().Error("Failed to update group", "error", err, "group_id", reqData.GroupID)
		return api.Error("Failed to save group: " + err.Error()).ToString()
	}

	return api.Success("Group saved successfully").ToString()
}
