# funcadmin

[![Go Report Card](https://goreportcard.com/badge/github.com/dracory/funcadmin)](https://goreportcard.com/report/github.com/dracory/funcadmin)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dracory/funcadmin)](https://pkg.go.dev/github.com/dracory/funcadmin)

## License

This project is licensed under the MIT License. You can find a copy of the
license at [LICENSE](LICENSE).

## Introduction

Admin interface for [`github.com/dracory/funcstore`](https://github.com/dracory/funcstore).
Provides a ready-to-use admin panel for managing serverless functions and
groups: dashboard, function manager, function editor (source + versions),
function invoke tester, and group management.

Modeled after [`github.com/dracory/shopadmin`](https://github.com/dracory/shopadmin)
— same folder-per-controller pattern, same `UiConfig`/`UiBase` conventions,
same `AdminOptions`/`New()`/`Handle()` entry point.

## Features

- **Dashboard** — overview of functions and groups
- **Function management** — create, update, delete functions with alias,
  runtime, timeout, metadata
- **Source versions** — view saved versions and restore previous sources
- **Function invoke** — test-run functions with JSON input from the UI
- **Group management** — hierarchical groups with create/update/delete
- **Custom layouts** — bring your own layout via `FuncLayout`
- **Bootstrap + Vue CDN** — default UI works out of the box

## Installation

```bash
go get github.com/dracory/funcadmin
```

## Quick Start

```go
package main

import (
    "log/slog"
    "net/http"
    "os"

    "github.com/dracory/funcadmin"
    "github.com/dracory/funcstore"
)

func main() {
    store, err := funcstore.NewStore(funcstore.StoreOptions{
        DB:                 yourDB,
        Driver:             "sqlite",
        Runtimes:           []funcstore.Runtime{funcstore.NewGojaRuntime()},
        AutomigrateEnabled: true,
    })
    if err != nil {
        log.Fatal(err)
    }

    admin, err := funcadmin.New(funcadmin.AdminOptions{
        Store:         store,
        Logger:        slog.New(slog.NewTextHandler(os.Stderr, nil)),
        AdminHomeURL:  "/admin",
        FaaSAdminURL:  "/admin/faas",
    })
    if err != nil {
        log.Fatal(err)
    }

    http.Handle("/admin/faas", admin)
    http.ListenAndServe(":8080", nil)
}
```

## Integration with a Router

`funcadmin.AdminInterface` exposes `Handle(w, r)`, which is an
`http.HandlerFunc`-compatible method. Wire it into any router that
accepts standard `http.Handler`:

```go
// stdlib
mux.Handle("/admin/faas", http.HandlerFunc(admin.Handle))
```

Requests are dispatched by the `?controller=` query parameter
(`dashboard`, `function-manager`, `function-update`, `function-invoke`,
`group-manager`, `group-update`). An empty or unknown controller falls
back to the dashboard.

## Custom Layout

By default, funcadmin renders a bare-bones HTML page with Bootstrap and
Vue from CDN. To embed the admin inside your own layout (branding, menus,
etc.), provide `FuncLayout`:

```go
admin, _ := funcadmin.New(funcadmin.AdminOptions{
    Store:  store,
    Logger: logger,
    FuncLayout: func(w http.ResponseWriter, r *http.Request, title, body string, opts struct {
        Styles     []string
        StyleURLs  []string
        Scripts    []string
        ScriptURLs []string
    }) string {
        return myLayout(w, r, title, body, opts)
    },
})
```

## Authentication

Provide `AuthUserID` to gate the whole admin behind your auth system:

```go
AuthUserID: func(r *http.Request) string {
    user := session.GetAuthUser(r)
    if user == nil {
        return ""
    }
    return user.GetID()
},
```

If it returns an empty string, the request is redirected to
`AdminHomeURL`. If `AuthUserID` is nil, no auth check is performed.

## Testing

```bash
go test ./...
```
