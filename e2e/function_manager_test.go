package e2e

import (
	"strings"
	"testing"
)

func TestFunctionManagerLoadsSeededFunctions(t *testing.T) {
	body := waitTable(t, newTab(t), adminURL("function-manager"))
	if !strings.Contains(body, "faas. Function Manager") {
		t.Fatal("expected 'faas. Function Manager' header on function manager")
	}
	if !strings.Contains(body, seedFunctionName) {
		t.Fatalf("expected seeded function %q in function manager", seedFunctionName)
	}
	if !strings.Contains(body, seedFunctionAlias) {
		t.Fatalf("expected seeded alias %q in function manager", seedFunctionAlias)
	}
}
