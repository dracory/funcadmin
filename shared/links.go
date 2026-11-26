package shared

import "net/http"

// Links provides URL helpers for FaaSadmin controllers.
// The base URL is read from request context (injected by Handle()),
// not hardcoded. This follows the blogadmin/shopadmin pattern.
type Links struct {
	baseURL string
}

// NewLinks creates a Links helper with the given base URL.
// If baseURL is empty, defaults to "/admin/faas".
func NewLinks(baseURL string) *Links {
	if baseURL == "" {
		baseURL = "/admin/FaaS"
	}
	return &Links{baseURL: baseURL}
}

// NewLinksFromRequest creates a Links helper using the FaaS admin URL
// from the request context.
func NewLinksFromRequest(r *http.Request) *Links {
	return NewLinks(FaaSAdminURL(r))
}

// Dashboard builds the URL for the dashboard controller
func (l *Links) Dashboard(params map[string]string) string {
	return l.url(CONTROLLER_DASHBOARD, params)
}

// FunctionManager builds the URL for the function manager controller
func (l *Links) FunctionManager(params map[string]string) string {
	return l.url(CONTROLLER_FUNCTION_MANAGER, params)
}

// FunctionUpdate builds the URL for the function update controller
func (l *Links) FunctionUpdate(params map[string]string) string {
	return l.url(CONTROLLER_FUNCTION_UPDATE, params)
}

// FunctionInvoke builds the URL for the function invoke controller
func (l *Links) FunctionInvoke(params map[string]string) string {
	return l.url(CONTROLLER_FUNCTION_INVOKE, params)
}

// GroupManager builds the URL for the group manager controller
func (l *Links) GroupManager(params map[string]string) string {
	return l.url(CONTROLLER_GROUP_MANAGER, params)
}

// GroupUpdate builds the URL for the group update controller
func (l *Links) GroupUpdate(params map[string]string) string {
	return l.url(CONTROLLER_GROUP_UPDATE, params)
}

// url builds a URL for the given controller. The params map is copied
// before mutation (does not modify caller's map).
func (l *Links) url(controller string, params map[string]string) string {
	return URL(l.baseURL, controller, params)
}
