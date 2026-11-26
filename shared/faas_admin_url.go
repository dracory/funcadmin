package shared

import "net/http"

// FaaSAdminURL returns the FaaS admin base URL from request context
func FaaSAdminURL(r *http.Request) string {
	value, ok := r.Context().Value(KeyFaaSAdminURL).(string)
	if !ok {
		return ""
	}
	return value
}
