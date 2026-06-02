package extract

import (
	"encoding/json"
	"fmt"

	jsonsanitize "github.com/go-go-golems/sanitize/pkg/json"
	yamlsanitize "github.com/go-go-golems/sanitize/pkg/yaml"
)

// Validate validates or repairs a candidate according to its inferred format.
func Validate(candidate *ExtractionCandidate, options *ExtractOptions) (*CandidateValidationResult, error) {
	if candidate == nil {
		return nil, fmt.Errorf("extract.validate: candidate must be an ExtractionCandidate")
	}
	format := candidate.Format
	if format == "" || format == "unknown" {
		format = inferFormatFromPayload(candidate.Text)
	}
	result := &CandidateValidationResult{Candidate: candidate, Format: format}
	switch format {
	case "json":
		var v any
		if err := json.Unmarshal([]byte(candidate.Text), &v); err == nil {
			result.Valid = true
			result.Sanitized = candidate.Text
			return result, nil
		}
		repaired := jsonsanitize.Sanitize(candidate.Text)
		result.Valid = repaired.StrictParseClean
		result.Sanitized = repaired.Sanitized
		result.Fixes = repaired.Fixes
		result.Issues = repaired.LintIssues
		for _, issue := range repaired.LintIssues {
			result.Errors = append(result.Errors, issue.Description)
		}
		return result, nil
	case "yaml", "yml":
		repaired := yamlsanitize.Sanitize(candidate.Text)
		result.Valid = repaired.ParseClean && repaired.LintClean
		result.Sanitized = repaired.Sanitized
		result.Fixes = repaired.Fixes
		result.Issues = repaired.LintIssues
		for _, issue := range repaired.LintIssues {
			result.Errors = append(result.Errors, issue.Description)
		}
		return result, nil
	case "xml":
		result.Valid = candidate.Wrapper == "xmlTag" && candidate.Text != ""
		result.Sanitized = candidate.Text
		if !result.Valid {
			result.Errors = []string{"XML validation is limited to extracted XML-like wrappers in Phase 1"}
		}
		return result, nil
	default:
		result.Valid = false
		result.Errors = []string{fmt.Sprintf("unsupported or unknown candidate format %q", format)}
		return result, nil
	}
}
