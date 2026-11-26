// Package funcadmin provides a standalone admin interface for the
// funcstore package, following the
// folder-per-controller pattern. Each controller is in its own
// subfolder and handles its own views and AJAX data.
//
// This module is modeled on github.com/dracory/blogadmin.
package funcadmin

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dracory/funcadmin/dashboard"
	"github.com/dracory/funcadmin/function_invoke"
	"github.com/dracory/funcadmin/function_manager"
	"github.com/dracory/funcadmin/function_update"
	"github.com/dracory/funcadmin/group_manager"
	"github.com/dracory/funcadmin/group_update"
	"github.com/dracory/funcadmin/shared"
	"github.com/dracory/funcstore"

	"github.com/dracory/req"
)

// AdminOptions contains all dependencies and configuration for the
// FaaS admin.
//
// Store and Logger are required. FuncLayout is optional — if nil, a
// default bare-bones HTML page is used (Bootstrap + Vue CDN).
type AdminOptions struct {
	// Store is the funcstore.StoreInterface (required)
	Store funcstore.StoreInterface

	// Logger is required
	Logger *slog.Logger

	// FuncLayout is an optional function to render the admin interface
	// inside your own layout (branding, menus, etc.). It receives the
	// request and response writer so the host project can access request
	// context (auth user, locale, etc.) when rendering the layout.
	FuncLayout func(w http.ResponseWriter, r *http.Request, title string, body string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string

	// AdminHomeURL is the URL for the admin home page (default: "/admin")
	AdminHomeURL string

	// FaaSAdminURL is the base URL for FaaS admin (default: "/admin/FaaS")
	FaaSAdminURL string

	// FileManagerURL is the URL for the file manager (optional)
	FileManagerURL string

	// AuthUserID returns the authenticated user ID from the request.
	// If it returns "", the user is treated as unauthenticated.
	AuthUserID func(r *http.Request) string
}

// AdminInterface defines the interface for the FaaS admin
type AdminInterface interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

// admin implements AdminInterface
type admin struct {
	store      funcstore.StoreInterface
	logger     *slog.Logger
	funcLayout func(w http.ResponseWriter, r *http.Request, title string, body string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
	adminHomeURL   string
	FaaSAdminURL   string
	fileManagerURL string
	authUserID     func(r *http.Request) string
	routes         map[string]func(w http.ResponseWriter, r *http.Request)
}

// New creates a new FaaS admin instance.
// Returns ErrStoreRequired if Store is nil, ErrLoggerRequired if Logger is nil.
func New(opts AdminOptions) (AdminInterface, error) {
	if opts.Store == nil {
		return nil, ErrStoreRequired
	}
	if opts.Logger == nil {
		return nil, ErrLoggerRequired
	}

	// Set defaults
	if opts.AdminHomeURL == "" {
		opts.AdminHomeURL = "/admin"
	}
	if opts.FaaSAdminURL == "" {
		opts.FaaSAdminURL = "/admin/FaaS"
	}

	a := &admin{
		store:          opts.Store,
		logger:         opts.Logger,
		funcLayout:     opts.FuncLayout,
		adminHomeURL:   opts.AdminHomeURL,
		FaaSAdminURL:   opts.FaaSAdminURL,
		fileManagerURL: opts.FileManagerURL,
		authUserID:     opts.AuthUserID,
	}

	// Build routes once at construction time
	a.routes = a.buildRoutes()

	return a, nil
}

// Handle processes all FaaS admin requests.
// Config values are injected into the request context.
// Route lookup is map-based.
func (a *admin) Handle(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	if a.authUserID != nil && a.authUserID(r) == "" {
		http.Redirect(w, r, a.adminHomeURL, http.StatusSeeOther)
		return
	}

	// Inject config into request context
	ctx := context.WithValue(r.Context(), shared.KeyEndpoint, r.URL.Path)
	ctx = context.WithValue(ctx, shared.KeyAdminHomeURL, a.adminHomeURL)
	ctx = context.WithValue(ctx, shared.KeyFaaSAdminURL, a.FaaSAdminURL)
	ctx = context.WithValue(ctx, shared.KeyFileManagerURL, a.fileManagerURL)
	r = r.WithContext(ctx)

	// Map-based route lookup
	controller := req.GetStringTrimmed(r, "controller")
	if controller == "" {
		controller = shared.CONTROLLER_DASHBOARD
	}

	handler, ok := a.routes[controller]
	if !ok {
		handler = a.routes[shared.CONTROLLER_DASHBOARD]
	}

	handler(w, r)
}

// buildRoutes creates the handler dispatch map once at construction time.
func (a *admin) buildRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	uiConfig := shared.UiConfig{
		Store:  a.store,
		Logger: a.logger,
		Layout: a.render,
	}

	return map[string]func(w http.ResponseWriter, r *http.Request){
		shared.CONTROLLER_DASHBOARD:        func(w http.ResponseWriter, r *http.Request) { dashboard.UI(uiConfig).Dashboard(w, r) },
		shared.CONTROLLER_FUNCTION_MANAGER: func(w http.ResponseWriter, r *http.Request) { function_manager.UI(uiConfig).FunctionManager(w, r) },
		shared.CONTROLLER_FUNCTION_UPDATE:  func(w http.ResponseWriter, r *http.Request) { function_update.UI(uiConfig).FunctionUpdate(w, r) },
		shared.CONTROLLER_FUNCTION_INVOKE:  func(w http.ResponseWriter, r *http.Request) { function_invoke.UI(uiConfig).FunctionInvoke(w, r) },
		shared.CONTROLLER_GROUP_MANAGER:    func(w http.ResponseWriter, r *http.Request) { group_manager.UI(uiConfig).GroupManager(w, r) },
		shared.CONTROLLER_GROUP_UPDATE:     func(w http.ResponseWriter, r *http.Request) { group_update.UI(uiConfig).GroupUpdate(w, r) },
	}
}

// render wraps content in the layout. If FuncLayout is provided and
// returns non-empty, it is used; otherwise the default shared.Layout
// is used.
func (a *admin) render(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
	Styles     []string
	StyleURLs  []string
	Scripts    []string
	ScriptURLs []string
}) string {
	// If a custom layout is provided, try it first
	if a.funcLayout != nil {
		custom := a.funcLayout(w, r, webpageTitle, webpageHtml, options)
		if custom != "" {
			if w != nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(custom))
				return ""
			}
			return custom
		}
	}

	webpage := shared.Layout(w, r, webpageTitle, webpageHtml, options)

	if w != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(webpage))
		return ""
	}

	return webpage
}
