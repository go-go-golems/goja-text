package extract

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/modules"
)

type module struct{}

var _ modules.NativeModule = (*module)(nil)
var _ modules.TypeScriptDeclarer = (*module)(nil)

func (module) Name() string { return "extract" }

func (module) Doc() string {
	return `
The extract module locates structured-data candidates inside larger text.

Functions:
  options(): Create a Go-backed extraction options builder.
  markdownCodeBlocks(input, options?): Extract Markdown fenced code blocks.
  xmlTagged(input, options?): Extract XML-like tag wrappers.
  rawStructured(input, options?): Recognize raw JSON/YAML payloads.
  frontmatter(input, options?): Extract leading YAML frontmatter.
  all(input, options?): Run enabled extractors and merge candidates.
  validate(candidate, options?): Validate or sanitize a candidate payload.

Go-backed candidates expose exported Go field names in JavaScript:
  candidate.Kind, candidate.Format, candidate.Text, candidate.StartByte, candidate.PayloadStartByte, ...
`
}

func (mod module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").(*goja.Object)

	modules.SetExport(exports, mod.Name(), "options", func() *ExtractOptionsBuilder {
		return NewExtractOptionsBuilder()
	})
	modules.SetExport(exports, mod.Name(), "markdownCodeBlocks", func(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) {
		return MarkdownCodeBlocks(input, options)
	})
	modules.SetExport(exports, mod.Name(), "xmlTagged", func(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) {
		return XMLTagged(input, options)
	})
	modules.SetExport(exports, mod.Name(), "rawStructured", func(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) {
		return RawStructured(input, options)
	})
	modules.SetExport(exports, mod.Name(), "frontmatter", func(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) {
		return Frontmatter(input, options)
	})
	modules.SetExport(exports, mod.Name(), "all", func(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) {
		return All(input, options)
	})
	modules.SetExport(exports, mod.Name(), "validate", func(candidate *ExtractionCandidate, options *ExtractOptions) (*CandidateValidationResult, error) {
		result, err := Validate(candidate, options)
		if err != nil {
			return nil, fmt.Errorf("extract.validate: %w", err)
		}
		return result, nil
	})
}

func init() {
	modules.Register(&module{})
}
