package shared

import (
	"log/slog"
	"net/http"
	"github.com/dracory/funcstore"
)

// UiConfig holds the dependencies passed to subcontroller UI factories.
// The Layout function uses an anonymous struct for options to match
// blogadmin/shopadmin exactly, allowing consumers to reuse their
// existing layout function for funcadmin.
type UiConfig struct {
	Store  funcstore.StoreInterface
	Logger *slog.Logger
	Layout func(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
}
