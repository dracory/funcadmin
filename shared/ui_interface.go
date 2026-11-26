package shared

import (
	"log/slog"
	"net/http"
	"github.com/dracory/funcstore"
)

// UiInterface defines the methods every subcontroller UI must implement.
// This follows the blogadmin/shopadmin pattern.
//
// Core controllers only use Store(), Logger(), and Layout().
type UiInterface interface {
	Store() funcstore.StoreInterface
	Logger() *slog.Logger

	Layout(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
}
