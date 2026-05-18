package settings

import (
	"testing"
)

func TestDiffKeys(t *testing.T) {
	required := []string{"permissions", "hooks"}
	actual := map[string]any{
		"permissions": true,
		"theme":       "dark",
	}

	d := diffKeys(required, actual)

	if len(d.missing) != 1 || d.missing[0] != "hooks" {
		t.Errorf("expected missing=[hooks], got %v", d.missing)
	}
	if len(d.unexpected) != 1 || d.unexpected[0] != "theme" {
		t.Errorf("expected unexpected=[theme], got %v", d.unexpected)
	}
}

func TestDiffKeysEmpty(t *testing.T) {
	d := diffKeys([]string{}, map[string]any{"permissions": true})
	if len(d.missing) != 0 {
		t.Error("expected no missing keys")
	}
	// No required set means we don't flag unexpected keys
	if len(d.unexpected) != 0 {
		t.Error("expected no unexpected keys when required is empty")
	}
}
