package testutils

import (
	"context"
	"log/slog"

	"github.com/dracory/funcstore"
)

// SeedDB populates the store with sample data so the admin panel has
// enough rows to exercise the manager tables and detail screens.
func SeedDB(store funcstore.StoreInterface, logger *slog.Logger) {
	ctx := context.Background()

	seedGroups(ctx, store, logger)
	seedFunctions(ctx, store, logger)

	logger.Info("seedDB complete", "groups", 2, "functions", 3)
}

func seedGroups(ctx context.Context, store funcstore.StoreInterface, logger *slog.Logger) {
	groups := []struct {
		slug        string
		description string
	}{
		{slug: "math", description: "Mathematical helper functions"},
		{slug: "utilities", description: "General utility functions"},
	}

	for _, seed := range groups {
		group := funcstore.NewGroup()
		group.SetSlug(seed.slug).
			SetDescription(seed.description)

		if _, err := store.GroupCreate(ctx, group); err != nil {
			logger.Error("seedGroups: failed to create group", "slug", seed.slug, "error", err)
		}
	}
}

func seedFunctions(ctx context.Context, store funcstore.StoreInterface, logger *slog.Logger) {
	functions := []struct {
		name    string
		alias   string
		group   string
		source  string
		timeout int64
	}{
		{
			name:    "Seed Double",
			alias:   "seed-double",
			group:   "math",
			source:  "return input * 2",
			timeout: 5000,
		},
		{
			name:    "Seed Greet",
			alias:   "seed-greet",
			group:   "utilities",
			source:  `return "Hello, " + input`,
			timeout: 5000,
		},
		{
			name:    "Seed Echo",
			alias:   "seed-echo",
			group:   "utilities",
			source:  "return input",
			timeout: 5000,
		},
	}

	for _, seed := range functions {
		fn := funcstore.NewFunction()
		fn.SetName(seed.name).
			SetAlias(seed.alias).
			SetGroup(seed.group).
			SetRuntime(funcstore.RUNTIME_JS).
			SetSource(seed.source).
			SetTimeoutMs(seed.timeout)

		if err := store.FunctionCreate(ctx, fn); err != nil {
			logger.Error("seedFunctions: failed to create function", "alias", seed.alias, "error", err)
		}
	}
}
