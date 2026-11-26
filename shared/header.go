package shared

import (
	"log/slog"
	"net/http"
	"github.com/dracory/funcstore"

	"github.com/dracory/hb"
)

// Header renders the FaaS admin navigation header.
// It is nil-safe for both store and logger.
func Header(store funcstore.StoreInterface, logger *slog.Logger, r *http.Request) hb.TagInterface {
	if store == nil {
		if logger != nil {
			logger.Error("FaaS store is nil")
		}
		return nil
	}

	linkHome := hb.NewHyperlink().
		HTML("Dashboard").
		Href(AdminHomeURL(r)).
		Class("nav-link")

	linkFaaS := hb.NewHyperlink().
		HTML("FaaS").
		Href(URLR(r, CONTROLLER_DASHBOARD, nil)).
		Class("nav-link")

	linkFunctions := hb.Hyperlink().
		HTML("Functions ").
		Href(URLR(r, CONTROLLER_FUNCTION_MANAGER, nil)).
		Class("nav-link")

	functionCount := int64(-1)
	fns, err := store.FunctionList(r.Context())
	if err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
	} else {
		functionCount = int64(len(fns))
	}

	ulNav := hb.NewUL().
		Class("nav  nav-pills justify-content-center").
		Child(hb.NewLI().
			Class("nav-item").Child(linkHome)).
		Child(hb.NewLI().
			Class("nav-item").Child(linkFaaS)).
		Child(hb.LI().
			Class("nav-item").
			Child(linkFunctions.
				Child(hb.Span().
					Class("badge bg-secondary ms-2").
					HTML(formatInt64(functionCount)))))

	fileManagerURL := FileManagerURL(r)
	if fileManagerURL != "" {
		linkFileManager := hb.Hyperlink().
			HTML("File Manager").
			Href(fileManagerURL).
			Class("nav-link")
		ulNav.Child(hb.LI().
			Class("nav-item").
			Child(linkFileManager))
	}

	divCard := hb.NewDiv().Class("card card-default mt-3 mb-3")
	divCardBody := hb.NewDiv().Class("card-body").Style("padding: 2px;")
	return divCard.AddChild(divCardBody.AddChild(ulNav))
}
