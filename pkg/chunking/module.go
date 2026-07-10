package chunking

import (
	"fmt"
	"math"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/modules"
	"github.com/go-go-golems/go-go-goja/pkg/tsgen/spec"
)

type module struct{}

var _ modules.NativeModule = (*module)(nil)
var _ modules.TypeScriptDeclarer = (*module)(nil)

func (module) Name() string { return "chunking" }

func (module) Doc() string {
	return `
The chunking module provides source-preserving text segmentation and deterministic budgeted packing.

Functions:
  lines(source, options?): Partition LF or CRLF lines.
  paragraphs(source, options?): Partition paragraphs with explicit blank-line ownership.
  markdownBlocks(source, options?): Partition top-level Markdown structures.
  markdownSections(source, options?): Partition heading sections and expose heading paths.
  pack(spans, options): Greedily pack complete spans by byte, rune, or word budget.
  packWeighted(items, options): Pack complete spans using caller-provided weights.
  recursive(source, options): Refine oversized source ranges through ordered segmenters.

Results are Go-backed objects. Read exported fields with PascalCase, for example:
  result.Spans[0].Text, result.Spans[0].StartByte, result.Chunks[0].Weight
Options are plain JavaScript objects with lower-camel keys. Unknown keys are rejected.
`
}

func (module) TypeScriptModule() *spec.Module {
	return &spec.Module{
		Name: "chunking",
		RawDTS: []string{
			"export interface Diagnostic { Code: string; Severity: 'warning' | 'error' | string; Message: string; StartByte: number; EndByte: number; }",
			"export interface StrategySpec { Name: string; Version: string; Options: Record<string, unknown>; }",
			"export interface Span { Ordinal: number; Kind: string; Text: string; StartByte: number; EndByte: number; StartRune: number; EndRune: number; StartLine: number; StartColumn: number; EndLine: number; EndColumn: number; Atomic: boolean; HeadingLevel: number; HeadingPath: string[]; Language: string; Level: string; }",
			"export interface SegmentResult { Spec: StrategySpec; SourceBytes: number; SourceRunes: number; Spans: Span[]; Diagnostics: Diagnostic[]; }",
			"export interface PackedChunk { Ordinal: number; Text: string; StartByte: number; EndByte: number; StartRune: number; EndRune: number; SpanOrdinals: number[]; HeadingPath: string[]; Weight: number; Oversized: boolean; Level: string; Diagnostics: Diagnostic[]; }",
			"export interface PackResult { Spec: StrategySpec; Chunks: PackedChunk[]; Diagnostics: Diagnostic[]; }",
			"export interface LineOptions { keepTerminators?: boolean; }",
			"export interface ParagraphOptions { blankLines?: 'trailing' | 'separate' | 'leading'; }",
			"export interface MarkdownBlockOptions { atomic?: string[]; }",
			"export interface MarkdownSectionOptions { maxHeadingLevel?: number; }",
			"export interface OverlapOptions { unit?: 'spans'; value?: number; }",
			"export interface PackOptions { maxUnits: number; measure?: 'bytes' | 'runes' | 'words'; overlap?: OverlapOptions; oversized?: 'allow' | 'error'; }",
			"export interface WeightedSpan { span: Span; weight: number; }",
			"export interface WeightedPackOptions { maxWeight: number; overlapWeight?: number; oversized?: 'allow' | 'error'; }",
			"export type RecursiveLevel = 'markdownSections' | 'markdownBlocks' | 'paragraphs' | 'lines' | 'runes';",
			"export interface RecursiveOptions extends PackOptions { levels?: RecursiveLevel[]; atomic?: string[]; }",
		},
		Functions: []spec.Function{
			{Name: "lines", Params: []spec.Param{{Name: "source", Type: spec.String()}, {Name: "options", Type: spec.Named("LineOptions"), Optional: true}}, Returns: spec.Named("SegmentResult")},
			{Name: "paragraphs", Params: []spec.Param{{Name: "source", Type: spec.String()}, {Name: "options", Type: spec.Named("ParagraphOptions"), Optional: true}}, Returns: spec.Named("SegmentResult")},
			{Name: "markdownBlocks", Params: []spec.Param{{Name: "source", Type: spec.String()}, {Name: "options", Type: spec.Named("MarkdownBlockOptions"), Optional: true}}, Returns: spec.Named("SegmentResult")},
			{Name: "markdownSections", Params: []spec.Param{{Name: "source", Type: spec.String()}, {Name: "options", Type: spec.Named("MarkdownSectionOptions"), Optional: true}}, Returns: spec.Named("SegmentResult")},
			{Name: "pack", Params: []spec.Param{{Name: "spans", Type: spec.Array(spec.Named("Span"))}, {Name: "options", Type: spec.Named("PackOptions")}}, Returns: spec.Named("PackResult")},
			{Name: "packWeighted", Params: []spec.Param{{Name: "items", Type: spec.Array(spec.Named("WeightedSpan"))}, {Name: "options", Type: spec.Named("WeightedPackOptions")}}, Returns: spec.Named("PackResult")},
			{Name: "recursive", Params: []spec.Param{{Name: "source", Type: spec.String()}, {Name: "options", Type: spec.Named("RecursiveOptions")}}, Returns: spec.Named("PackResult")},
		},
	}
}

func (mod module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").(*goja.Object)

	modules.SetExport(exports, mod.Name(), "lines", func(call goja.FunctionCall) goja.Value {
		source := requiredString(vm, call, 0, "lines", "source")
		options := decodeLineOptions(vm, call.Argument(1))
		result, err := Lines(source, options)
		return resultValue(vm, result, err)
	})
	modules.SetExport(exports, mod.Name(), "paragraphs", func(call goja.FunctionCall) goja.Value {
		source := requiredString(vm, call, 0, "paragraphs", "source")
		options := decodeParagraphOptions(vm, call.Argument(1))
		result, err := Paragraphs(source, options)
		return resultValue(vm, result, err)
	})
	modules.SetExport(exports, mod.Name(), "markdownBlocks", func(call goja.FunctionCall) goja.Value {
		source := requiredString(vm, call, 0, "markdownBlocks", "source")
		options := decodeMarkdownBlockOptions(vm, call.Argument(1))
		result, err := MarkdownBlocks(source, options)
		return resultValue(vm, result, err)
	})
	modules.SetExport(exports, mod.Name(), "markdownSections", func(call goja.FunctionCall) goja.Value {
		source := requiredString(vm, call, 0, "markdownSections", "source")
		options := decodeMarkdownSectionOptions(vm, call.Argument(1))
		result, err := MarkdownSections(source, options)
		return resultValue(vm, result, err)
	})
	modules.SetExport(exports, mod.Name(), "pack", func(call goja.FunctionCall) goja.Value {
		spans := decodeSpans(vm, call.Argument(0))
		options := decodePackOptions(vm, call.Argument(1), "pack")
		result, err := Pack(spans, options)
		return resultValue(vm, result, err)
	})
	modules.SetExport(exports, mod.Name(), "packWeighted", func(call goja.FunctionCall) goja.Value {
		items := decodeWeightedSpans(vm, call.Argument(0))
		options := decodeWeightedPackOptions(vm, call.Argument(1))
		result, err := PackWeighted(items, options)
		return resultValue(vm, result, err)
	})
	modules.SetExport(exports, mod.Name(), "recursive", func(call goja.FunctionCall) goja.Value {
		source := requiredString(vm, call, 0, "recursive", "source")
		options := decodeRecursiveOptions(vm, call.Argument(1))
		result, err := Recursive(source, options)
		return resultValue(vm, result, err)
	})
}

func requiredString(vm *goja.Runtime, call goja.FunctionCall, index int, operation, name string) string {
	value := call.Argument(index)
	if missingValue(value) {
		panic(vm.NewTypeError("chunking.%s: %s is required", operation, name))
	}
	if _, ok := value.Export().(string); !ok {
		panic(vm.NewTypeError("chunking.%s: %s must be a string", operation, name))
	}
	return value.String()
}

func resultValue[T any](vm *goja.Runtime, result *T, err error) goja.Value {
	if err != nil {
		panic(vm.NewGoError(err))
	}
	return vm.ToValue(result)
}

func optionObject(vm *goja.Runtime, value goja.Value, operation string, allowed ...string) *goja.Object {
	if missingValue(value) {
		return vm.NewObject()
	}
	object := value.ToObject(vm)
	allowedSet := stringSet(allowed)
	for _, key := range object.Keys() {
		if !allowedSet[key] {
			panic(vm.NewTypeError("chunking.%s: unknown option %q", operation, key))
		}
	}
	return object
}

func optionString(vm *goja.Runtime, object *goja.Object, key, fallback string) string {
	value := object.Get(key)
	if missingValue(value) {
		return fallback
	}
	if _, ok := value.Export().(string); !ok {
		panic(vm.NewTypeError("chunking: option %s must be a string", key))
	}
	return value.String()
}

func optionBool(vm *goja.Runtime, object *goja.Object, key string, fallback bool) bool {
	value := object.Get(key)
	if missingValue(value) {
		return fallback
	}
	ret, ok := value.Export().(bool)
	if !ok {
		panic(vm.NewTypeError("chunking: option %s must be a boolean", key))
	}
	return ret
}

func optionInt(vm *goja.Runtime, object *goja.Object, key string, fallback int) int {
	value := object.Get(key)
	if missingValue(value) {
		return fallback
	}
	switch value.Export().(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
	default:
		panic(vm.NewTypeError("chunking: option %s must be an integer", key))
	}
	number := value.ToFloat()
	if math.IsNaN(number) || math.IsInf(number, 0) || math.Trunc(number) != number || number > math.MaxInt || number < math.MinInt {
		panic(vm.NewTypeError("chunking: option %s must be an integer", key))
	}
	return int(number)
}

func optionStrings(vm *goja.Runtime, object *goja.Object, key string, fallback []string) []string {
	value := object.Get(key)
	if missingValue(value) {
		return append([]string(nil), fallback...)
	}
	var result []string
	if err := vm.ExportTo(value, &result); err != nil {
		panic(vm.NewTypeError("chunking: option %s must be an array of strings", key))
	}
	return result
}

func missingValue(value goja.Value) bool {
	return value == nil || goja.IsUndefined(value) || goja.IsNull(value)
}

func decodeLineOptions(vm *goja.Runtime, value goja.Value) LineOptions {
	object := optionObject(vm, value, "lines", "keepTerminators")
	return LineOptions{KeepTerminators: optionBool(vm, object, "keepTerminators", true)}
}

func decodeParagraphOptions(vm *goja.Runtime, value goja.Value) ParagraphOptions {
	object := optionObject(vm, value, "paragraphs", "blankLines")
	return ParagraphOptions{BlankLines: optionString(vm, object, "blankLines", "trailing")}
}

func decodeMarkdownBlockOptions(vm *goja.Runtime, value goja.Value) MarkdownBlockOptions {
	object := optionObject(vm, value, "markdownBlocks", "atomic")
	return MarkdownBlockOptions{Atomic: optionStrings(vm, object, "atomic", []string{"fencedCodeBlock", "codeBlock", "htmlBlock"})}
}

func decodeMarkdownSectionOptions(vm *goja.Runtime, value goja.Value) MarkdownSectionOptions {
	object := optionObject(vm, value, "markdownSections", "maxHeadingLevel")
	return MarkdownSectionOptions{MaxHeadingLevel: optionInt(vm, object, "maxHeadingLevel", 6)}
}

func decodePackOptions(vm *goja.Runtime, value goja.Value, operation string) PackOptions {
	object := optionObject(vm, value, operation, "maxUnits", "measure", "overlap", "oversized")
	overlapObject := optionObject(vm, object.Get("overlap"), operation+".overlap", "unit", "value")
	return PackOptions{
		MaxUnits:  optionInt(vm, object, "maxUnits", 0),
		Measure:   optionString(vm, object, "measure", "runes"),
		Overlap:   OverlapOptions{Unit: optionString(vm, overlapObject, "unit", "spans"), Value: optionInt(vm, overlapObject, "value", 0)},
		Oversized: optionString(vm, object, "oversized", "allow"),
	}
}

func decodeWeightedPackOptions(vm *goja.Runtime, value goja.Value) WeightedPackOptions {
	object := optionObject(vm, value, "packWeighted", "maxWeight", "overlapWeight", "oversized")
	return WeightedPackOptions{
		MaxWeight:     optionInt(vm, object, "maxWeight", 0),
		OverlapWeight: optionInt(vm, object, "overlapWeight", 0),
		Oversized:     optionString(vm, object, "oversized", "allow"),
	}
}

func decodeRecursiveOptions(vm *goja.Runtime, value goja.Value) RecursiveOptions {
	object := optionObject(vm, value, "recursive", "maxUnits", "measure", "levels", "overlap", "oversized", "atomic")
	overlapObject := optionObject(vm, object.Get("overlap"), "recursive.overlap", "unit", "value")
	return RecursiveOptions{
		MaxUnits:  optionInt(vm, object, "maxUnits", 0),
		Measure:   optionString(vm, object, "measure", "runes"),
		Levels:    optionStrings(vm, object, "levels", nil),
		Overlap:   OverlapOptions{Unit: optionString(vm, overlapObject, "unit", "spans"), Value: optionInt(vm, overlapObject, "value", 0)},
		Oversized: optionString(vm, object, "oversized", "allow"),
		Atomic:    optionStrings(vm, object, "atomic", []string{"fencedCodeBlock", "codeBlock", "htmlBlock"}),
	}
}

func decodeSpans(vm *goja.Runtime, value goja.Value) []Span {
	var spans []Span
	if err := vm.ExportTo(value, &spans); err != nil {
		panic(vm.NewTypeError("chunking.pack: spans must be an array of Span values"))
	}
	return spans
}

func decodeWeightedSpans(vm *goja.Runtime, value goja.Value) []WeightedSpan {
	object := value.ToObject(vm)
	length := int(object.Get("length").ToInteger())
	items := make([]WeightedSpan, 0, length)
	for i := 0; i < length; i++ {
		item := object.Get(fmt.Sprintf("%d", i)).ToObject(vm)
		var span Span
		if err := vm.ExportTo(item.Get("span"), &span); err != nil {
			panic(vm.NewTypeError("chunking.packWeighted: item %d span is invalid", i))
		}
		items = append(items, WeightedSpan{Span: span, Weight: optionInt(vm, item, "weight", -1)})
	}
	return items
}

func init() {
	modules.Register(&module{})
}
