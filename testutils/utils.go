package testutils

import (
	"database/sql"
	"os"
	"strings"

	"github.com/dracory/funcstore"
)

// InitStore creates a funcstore for testing.
//
// Note: The SQLite driver (modernc.org/sqlite) must be imported with a
// side-effect import in _test.go files of the consuming package, NOT here.
// This avoids double driver registration panics when a host project that
// already imports modernc.org/sqlite also imports this testutils package.
//
// Example (in your _test.go file):
//
//	import (
//	    "github.com/dracory/funcadmin/testutils"
//	    _ "modernc.org/sqlite"
//	)
func InitStore(filepath string) (funcstore.StoreInterface, error) {
	db, err := initDB(filepath)
	if err != nil {
		return nil, err
	}

	store, err := funcstore.NewStore(funcstore.StoreOptions{
		DB:                 db,
		Driver:             "sqlite",
		AutomigrateEnabled: true,
		Runtimes:           []funcstore.Runtime{funcstore.NewGojaRuntime()},
	})
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func initDB(filepath string) (*sql.DB, error) {
	if filepath != ":memory:" {
		err := os.Remove(filepath)
		if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
	}

	// SQLite in-memory databases are connection-scoped. Pinning the pool to
	// one connection keeps migrations and test data visible to every query.
	db.SetMaxOpenConns(1)

	return db, nil
}
