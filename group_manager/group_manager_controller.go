package group_manager

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
var groupsFiles embed.FS

const (
	actionLoadGroups  = "load-groups"
	actionDeleteGroup = "delete-group"
	actionCreateGroup = "create-group"
)

// UiInterface defines the group manager controller's UI interface
type UiInterface interface {
	shared.UiInterface
	GroupManager(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new group manager controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// GroupManager handles the group manager controller requests
func (u *ui) GroupManager(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

// Handler processes the group manager controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	action := req.GetStringTrimmed(r, "action")

	switch action {
	case actionLoadGroups:
		return u.handleLoadGroups(w, r)
	case actionDeleteGroup:
		return u.handleDeleteGroup(w, r)
	case actionCreateGroup:
		return u.handleCreateGroup(w, r)
	default:
		return u.renderPage(w, r)
	}
}

func (u *ui) renderPage(w http.ResponseWriter, r *http.Request) string {
	breadcrumbs := shared.Breadcrumbs([]shared.Breadcrumb{
		{Name: "Home", URL: shared.AdminHomeURL(r)},
		{Name: "FaaS", URL: shared.URLR(r, shared.CONTROLLER_DASHBOARD, nil)},
		{Name: "Group Manager", URL: ""},
	})

	linksHelper := shared.NewLinksFromRequest(r)

	heading := hb.Heading1().HTML("faas. Group Manager")

	htmlContent, err := groupsFiles.ReadFile("groups.html")
	if err != nil {
		u.Logger().Error("Failed to read groups HTML template", "error", err)
		return hb.Div().HTML("Error loading groups component").ToHTML()
	}

	jsContent, err := groupsFiles.ReadFile("groups.js")
	if err != nil {
		u.Logger().Error("Failed to read groups JavaScript file", "error", err)
		return hb.Div().HTML("Error loading groups component").ToHTML()
	}

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())
	notiflixCSS := hb.Link().Rel("stylesheet").Href(cdn.Notiflix_3_2_8_CSS())
	notiflixJS := hb.Script("").Src(cdn.Notiflix_3_2_8())

	initScript := hb.Script(`
		const urlGroupsLoad = '` + linksHelper.GroupManager(map[string]string{"action": actionLoadGroups}) + `';
		const urlGroupDelete = '` + linksHelper.GroupManager(map[string]string{"action": actionDeleteGroup}) + `';
		const urlGroupCreate = '` + linksHelper.GroupManager(map[string]string{"action": actionCreateGroup}) + `';
		const urlGroupUpdate = '` + linksHelper.GroupUpdate(map[string]string{"group_id": "GROUP_ID_PLACEHOLDER"}) + `';
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

	return u.Layout(w, r, "FaaS | Group Manager", content.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{})
}

func (u *ui) handleLoadGroups(w http.ResponseWriter, r *http.Request) string {
	store := u.Store()
	if store == nil {
		return api.Error("FaaS store not available").ToString()
	}

	groups, err := store.GroupList(r.Context(), funcstore.GroupFilter{})
	if err != nil {
		u.Logger().Error("Failed to load groups", "error", err)
		return api.Error("Failed to load groups").ToString()
	}

	groupList := []map[string]any{}
	for _, g := range groups {
		groupList = append(groupList, map[string]any{
			"id":          g.GetID(),
			"slug":        g.GetSlug(),
			"path":        g.GetPath(),
			"parent_id":   g.GetParentID(),
			"description": g.GetDescription(),
			"metadata":    g.GetMetadata(),
			"created_at":  g.GetCreatedAt(),
			"updated_at":  g.GetUpdatedAt(),
		})
	}

	return api.SuccessWithData("Groups loaded successfully", map[string]any{
		"groups": groupList,
		"total":  len(groupList),
	}).ToString()
}

func (u *ui) handleDeleteGroup(w http.ResponseWriter, r *http.Request) string {
	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		GroupID string `json:"group_id"`
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

	if err := store.GroupDelete(r.Context(), reqData.GroupID, funcstore.DeleteOpts{}); err != nil {
		u.Logger().Error("Failed to delete group", "error", err)
		return api.Error("Failed to delete group: " + err.Error()).ToString()
	}

	return api.SuccessWithData("Group deleted successfully", map[string]any{}).ToString()
}

func (u *ui) handleCreateGroup(w http.ResponseWriter, r *http.Request) string {
	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		Slug        string `json:"slug"`
		Parent      string `json:"parent"`
		Description string `json:"description"`
		Metadata    string `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.Slug == "" {
		return api.Error("Slug is required").ToString()
	}

	store := u.Store()
	if store == nil {
		return api.Error("FaaS store not available").ToString()
	}

	g := funcstore.NewGroup()
	g.SetSlug(reqData.Slug).
		SetParent(reqData.Parent).
		SetDescription(reqData.Description).
		SetMetadata(reqData.Metadata)

	if _, err := store.GroupCreate(r.Context(), g); err != nil {
		u.Logger().Error("Failed to create group", "error", err)
		return api.Error("Failed to create group: " + err.Error()).ToString()
	}

	return api.SuccessWithData("Group created successfully", map[string]any{
		"id": g.GetID(),
	}).ToString()
}
