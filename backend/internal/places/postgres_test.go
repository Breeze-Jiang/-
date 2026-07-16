package places

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestUnclassifiedCategoryDatabaseBoundary(t *testing.T) {
	if got := materializedCategory(CategoryUnclassified); got != "" {
		t.Fatalf("unclassified category must be stored as NULL input, got %q", got)
	}
	if got := storedCategory(pgtype.Text{}); got != CategoryUnclassified {
		t.Fatalf("NULL category must be exposed as unclassified, got %q", got)
	}
	if got := materializedCategory(CategoryToilet); got != string(CategoryToilet) {
		t.Fatalf("known category changed at database boundary: %q", got)
	}
}
