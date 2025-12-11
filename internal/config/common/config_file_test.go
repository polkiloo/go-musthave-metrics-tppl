package commoncfg

import "testing"

func TestLoadConfigFile(t *testing.T) {
	t.Parallel()

	t.Run("empty path", func(t *testing.T) {
		t.Parallel()
		type sample struct {
			Value string `json:"value"`
		}
		var cfg sample
		if err := LoadConfigFile("", &cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Value != "" {
			t.Fatalf("expected zero value for empty path, got %q", cfg.Value)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		var cfg struct{}
		if err := LoadConfigFile("missing.json", &cfg); err == nil {
			t.Fatalf("expected error for missing file")
		}
	})
}
