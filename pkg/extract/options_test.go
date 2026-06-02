package extract

import "testing"

func TestExtractOptionsBuilderDefaults(t *testing.T) {
	cfg, err := NewExtractOptionsBuilder().Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !cfg.InferFormat {
		t.Fatalf("InferFormat = false, want true")
	}
	if len(cfg.Tags) == 0 || len(cfg.Extractors) == 0 {
		t.Fatalf("cfg = %#v, want default tags and extractors", cfg)
	}
}

func TestExtractOptionsBuilderValidation(t *testing.T) {
	bad := NewExtractOptionsBuilder().Formats("json", "bogus").Extractors("rawStructured", "bad").MinConfidence(2).Validate()
	if bad.Valid || len(bad.Errors) < 3 {
		t.Fatalf("Validate() = %#v, want multiple errors", bad)
	}

	cfg, err := NewExtractOptionsBuilder().Formats("json", "yaml").Tags("json", "payload").Extractors("markdownCodeBlocks").MinConfidence(0.5).MaxCandidates(3).Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(cfg.Formats) != 2 || cfg.Formats[0] != "json" || cfg.MinConfidence != 0.5 || cfg.MaxCandidates != 3 {
		t.Fatalf("cfg = %#v", cfg)
	}
}
