package e2e

import (
	"strings"
	"testing"
)

func TestGroupManagerLoadsSeededGroups(t *testing.T) {
	body := waitTable(t, newTab(t), adminURL("group-manager"))
	if !strings.Contains(body, "faas. Group Manager") {
		t.Fatal("expected 'faas. Group Manager' header on group manager")
	}
	if !strings.Contains(body, seedGroupSlug) {
		t.Fatalf("expected seeded group %q in group manager", seedGroupSlug)
	}
	if !strings.Contains(body, "utilities") {
		t.Fatal("expected seeded 'utilities' group in group manager")
	}
}
