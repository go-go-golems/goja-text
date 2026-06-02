package extract

import "testing"

func TestRawStructuredJSON(t *testing.T) {
	got, err := RawStructured("  {\"ok\": true}\n", nil)
	if err != nil {
		t.Fatalf("RawStructured() error = %v", err)
	}
	if len(got) != 1 || got[0].Format != "json" || got[0].StartByte != 2 || got[0].Confidence < 0.9 {
		t.Fatalf("got = %#v, want strict JSON candidate", got)
	}
}

func TestRawStructuredRepairableJSON(t *testing.T) {
	got, err := RawStructured("{'ok': True,}\n", nil)
	if err != nil {
		t.Fatalf("RawStructured() error = %v", err)
	}
	if len(got) != 1 || got[0].Format != "json" || got[0].Confidence >= 0.9 {
		t.Fatalf("got = %#v, want repairable JSON candidate", got)
	}
}

func TestRawStructuredYAMLAndFalsePositive(t *testing.T) {
	got, err := RawStructured("name: Alice\nage: 30\n", nil)
	if err != nil {
		t.Fatalf("RawStructured() error = %v", err)
	}
	if len(got) != 1 || got[0].Format != "yaml" {
		t.Fatalf("got = %#v, want YAML candidate", got)
	}

	prose, err := RawStructured("Note: this is just a sentence with one colon.", nil)
	if err != nil {
		t.Fatalf("RawStructured() error = %v", err)
	}
	if len(prose) != 0 {
		t.Fatalf("prose candidates = %#v, want none", prose)
	}
}

func TestValidateCandidate(t *testing.T) {
	jsonCandidate := &ExtractionCandidate{Kind: "raw", Format: "json", Text: "{'ok': True,}"}
	jsonResult, err := Validate(jsonCandidate, nil)
	if err != nil {
		t.Fatalf("Validate(json) error = %v", err)
	}
	if !jsonResult.Valid || jsonResult.Sanitized == "" {
		t.Fatalf("json validation = %#v, want valid repaired JSON", jsonResult)
	}

	yamlCandidate := &ExtractionCandidate{Kind: "raw", Format: "yaml", Text: "name:Alice\n"}
	yamlResult, err := Validate(yamlCandidate, nil)
	if err != nil {
		t.Fatalf("Validate(yaml) error = %v", err)
	}
	if !yamlResult.Valid || yamlResult.Sanitized != "name: Alice\n" {
		t.Fatalf("yaml validation = %#v, want repaired YAML", yamlResult)
	}

	unknown, err := Validate(&ExtractionCandidate{Kind: "raw", Format: "unknown", Text: "hello"}, nil)
	if err != nil {
		t.Fatalf("Validate(unknown) error = %v", err)
	}
	if unknown.Valid || len(unknown.Errors) == 0 {
		t.Fatalf("unknown validation = %#v, want invalid", unknown)
	}
}

func TestAll(t *testing.T) {
	input := "---\ntitle: Demo\n---\n\n```json\n{\"ok\": true}\n```\n\n<yaml>name: Alice\nage: 30\n</yaml>\n"
	got, err := All(input, nil)
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	if len(got) < 3 {
		t.Fatalf("len(got) = %d, want at least 3: %#v", len(got), got)
	}
	kinds := make([]string, len(got))
	for i, c := range got {
		kinds[i] = c.Kind + ":" + c.Format
	}
	if got[0].Kind != "frontmatter" {
		t.Fatalf("ordering/kinds = %#v", kinds)
	}
	foundMarkdown := false
	foundXML := false
	for _, c := range got {
		if c.Kind == "markdownCodeBlock" && c.Format == "json" {
			foundMarkdown = true
		}
		if c.Kind == "xmlTagged" && c.Format == "yaml" {
			foundXML = true
		}
	}
	if !foundMarkdown || !foundXML {
		t.Fatalf("kinds = %#v, want markdown json and yaml xmlTagged candidates", kinds)
	}
}
