package sanitize

import "testing"

func TestYamlOptionsBuilderBuildsConfig(t *testing.T) {
	cfg, err := NewYamlOptionsBuilder().
		MaxIterations(5).
		TabWidth(4).
		OnlyRules("tab_indent", "missing_space_after_colon").
		DisabledRules("duplicate_key").
		Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if cfg.MaxIterations != 5 || cfg.TabWidth != 4 {
		t.Fatalf("cfg = %#v, want maxIterations=5 tabWidth=4", cfg)
	}
	if len(cfg.OnlyRules) != 2 || cfg.OnlyRules[0] != "tab_indent" {
		t.Fatalf("OnlyRules = %#v", cfg.OnlyRules)
	}
	if len(cfg.DisabledRules) != 1 || cfg.DisabledRules[0] != "duplicate_key" {
		t.Fatalf("DisabledRules = %#v", cfg.DisabledRules)
	}
}

func TestBuilderRejectsUnknownOptionsByDefault(t *testing.T) {
	result := NewYamlOptionsBuilder().FromObject(map[string]any{"typo": true}).Validate()
	if result.Valid {
		t.Fatalf("Validate() = valid, want unknown option error")
	}
	if len(result.Unknown) != 1 || result.Unknown[0] != "typo" {
		t.Fatalf("Unknown = %#v, want [typo]", result.Unknown)
	}
}

func TestBuilderCanCollectUnknownOptions(t *testing.T) {
	cfg, err := NewJsonOptionsBuilder().
		CollectUnknownOptions().
		FromObject(map[string]any{"futureOption": 1, "maxIterations": float64(3)}).
		Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if cfg.MaxIterations != 3 {
		t.Fatalf("MaxIterations = %d, want 3", cfg.MaxIterations)
	}
	if len(cfg.Unknown) != 1 || cfg.Unknown[0] != "futureOption" {
		t.Fatalf("Unknown = %#v, want [futureOption]", cfg.Unknown)
	}
}

func TestBuilderRejectsUnknownRulesAndOverlap(t *testing.T) {
	unknown := NewJsonOptionsBuilder().OnlyRules("not_a_rule").Validate()
	if unknown.Valid || len(unknown.Errors) == 0 {
		t.Fatalf("unknown Validate() = %#v, want errors", unknown)
	}

	overlap := NewYamlOptionsBuilder().OnlyRules("tab_indent").DisabledRules("tab_indent").Validate()
	if overlap.Valid || len(overlap.Errors) == 0 {
		t.Fatalf("overlap Validate() = %#v, want errors", overlap)
	}
}
