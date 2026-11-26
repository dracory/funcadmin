// Package e2e contains end-to-end tests that drive the funcadmin panel
// with a real headless Chrome via chromedp (CDP). The admin is served
// in-process by httptest on a random port — no binary build needed.
package e2e

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/dracory/funcadmin"
	"github.com/dracory/funcadmin/testutils"
	"github.com/dracory/funcstore"
	_ "modernc.org/sqlite"
)

var (
	baseURL           string
	allocCtx          context.Context
	testStore         funcstore.StoreInterface
	seedFunctionID    string
	seedFunctionName  string
	seedFunctionAlias string
	seedGroupID       string
	seedGroupSlug     string
)

const adminPath = "/admin/FaaS"

func TestMain(m *testing.M) {
	// In-process admin server backed by a seeded in-memory SQLite DB.
	store, err := testutils.InitStore(":memory:")
	if err != nil {
		panic("e2e: failed to init store: " + err.Error())
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	testutils.SeedDB(store, logger)
	testStore = store

	functions, err := store.FunctionList(context.Background())
	if err != nil || len(functions) < 2 {
		panic("e2e: not enough seeded functions: " + fmt.Sprint(err))
	}
	for _, fn := range functions {
		if fn.GetAlias() == "seed-double" {
			seedFunctionID = fn.GetID()
			seedFunctionName = fn.GetName()
			seedFunctionAlias = fn.GetAlias()
		}
	}
	if seedFunctionID == "" {
		panic("e2e: seeded seed-double function not found")
	}

	groups, err := store.GroupList(context.Background(), funcstore.GroupFilter{})
	if err != nil || len(groups) < 2 {
		panic("e2e: not enough seeded groups: " + fmt.Sprint(err))
	}
	for _, group := range groups {
		if group.GetSlug() == "math" {
			seedGroupID = group.GetID()
			seedGroupSlug = group.GetSlug()
		}
	}
	if seedGroupID == "" {
		panic("e2e: seeded math group not found")
	}

	admin, err := funcadmin.New(funcadmin.AdminOptions{
		Store:        store,
		Logger:       logger,
		AdminHomeURL: "/admin",
		FaaSAdminURL: adminPath,
		// AuthUserID intentionally nil — the panel is open.
	})
	if err != nil {
		panic("e2e: failed to create funcadmin: " + err.Error())
	}

	mux := http.NewServeMux()
	mux.Handle(adminPath, http.HandlerFunc(admin.Handle))
	mux.Handle(adminPath+"/", http.HandlerFunc(admin.Handle))
	srv := httptest.NewServer(mux)
	baseURL = srv.URL

	// One shared browser process for the entire test run.
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Headless)
	if bin := chromeBin(); bin != "" {
		opts = append(opts, chromedp.ExecPath(bin))
	}
	var cancel context.CancelFunc
	allocCtx, cancel = chromedp.NewExecAllocator(context.Background(), opts...)

	code := m.Run()

	cancel()
	srv.Close()
	os.Exit(code)
}

// newTab opens a fresh tab in the shared browser and returns its
// context. Each tab gets a 30s deadline; cleanup closes the tab.
func newTab(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := chromedp.NewContext(allocCtx)
	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	t.Cleanup(func() { cancelTimeout(); cancel() })
	return ctx
}

// chromeBin resolves a headless Chromium binary. Order: CHROME_BIN env,
// Playwright's downloaded headless shell, empty (chromedp finds Chrome
// on PATH / standard install locations).
func chromeBin() string {
	if b := os.Getenv("CHROME_BIN"); b != "" {
		return b
	}
	patterns := []string{
		// Windows
		filepath.Join(os.Getenv("LOCALAPPDATA"),
			"ms-playwright/chromium_headless_shell-*/chrome-headless-shell-win64/chrome-headless-shell.exe"),
		// WSL: Windows-side Playwright shell via interop
		"/mnt/*/Users/*/AppData/Local/ms-playwright/chromium_headless_shell-*/chrome-headless-shell-win64/chrome-headless-shell.exe",
		// Linux
		filepath.Join(os.Getenv("HOME"),
			".cache/ms-playwright/chromium_headless_shell-*/chrome-headless-shell-linux64/chrome-headless-shell"),
	}
	for _, pattern := range patterns {
		if matches, _ := filepath.Glob(pattern); len(matches) > 0 {
			return matches[len(matches)-1]
		}
	}
	return ""
}
