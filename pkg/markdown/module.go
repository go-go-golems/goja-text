package markdown

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/modules"
	"github.com/go-go-golems/go-go-goja/pkg/tsgen/spec"
)

type module struct{}

var _ modules.NativeModule = (*module)(nil)
var _ modules.TypeScriptDeclarer = (*module)(nil)

func (module) Name() string { return "markdown" }

func (module) Doc() string {
	return `
The markdown module provides Markdown parsing, rendering, traversal, and validation.

Functions:
  parse(input): Parse Markdown into a Go-backed MarkdownNode tree.
  renderHTML(input): Parse and render Markdown to HTML.
  walk(root, visitor): Traverse a Go-backed AST using a JavaScript callback.
  textContent(node): Extract plain text from a MarkdownNode subtree.
  validate(value): Validate a string input or Go-backed MarkdownNode object.
  builder(): Create a Go-backed fluent Markdown document builder.
  inline(): Create helpers for explicit inline nodes such as Code, Strong, Link, and Raw.

Go-backed MarkdownNode and builder objects expose exported Go field and method names in JavaScript:
  node.Type, node.Children, node.Text, node.Level, node.Destination, ...
  markdown.builder().Title("Report").Table().Columns("Name", "Status").Row("Parser", "done").End().Render().Text
`
}

func (module) TypeScriptModule() *spec.Module {
	return &spec.Module{
		Name: "markdown",
		RawDTS: []string{
			"export interface MarkdownNode {",
			"  Type: string;",
			"  Children?: MarkdownNode[];",
			"  Text?: string;",
			"  Level?: number;",
			"  Language?: string;",
			"  Destination?: string;",
			"  Title?: string;",
			"  Alt?: string;",
			"  Ordered?: boolean;",
			"  Start?: number;",
			"  Marker?: string;",
			"  Info?: string;",
			"  Raw?: string;",
			"  StartLine?: number;",
			"  StartColumn?: number;",
			"  SourcePos?: [number, number];",
			"}",
			"export interface WalkContext {",
			"  Parent?: MarkdownNode;",
			"  Depth: number;",
			"  Index: number;",
			"  Path: number[];",
			"}",
			"export type WalkResult = void | boolean | 'skip' | 'stop';",
			"export interface ValidationResult {",
			"  Valid: boolean;",
			"  Errors?: string[];",
			"}",
			"export interface MarkdownRenderResult {",
			"  Text: string;",
			"  Bytes: number;",
			"  Blocks: number;",
			"}",
			"export type TableAlignment = 'default' | 'left' | 'center' | 'right';",
			"export type InlineInput = string | number | boolean | TextInline | RawInline | CodeInline | EmphasisInline | StrongInline | LinkInline | InlineInput[];",
			"export type ColumnInput = string | { label?: InlineInput; Label?: InlineInput; align?: TableAlignment; Align?: TableAlignment };",
			"export type ChecklistInput = string | { text?: InlineInput; Text?: InlineInput; checked?: boolean; Checked?: boolean };",
			"export interface TextInline { Text: string; }",
			"export interface RawInline { Markdown: string; }",
			"export interface CodeInline { Code: string; }",
			"export interface EmphasisInline { Children: InlineInput[]; }",
			"export interface StrongInline { Children: InlineInput[]; }",
			"export interface LinkInline { Text: InlineInput[]; URL: string; Title?: string; }",
			"export interface InlineFactory {",
			"  Text(text: string): TextInline;",
			"  Raw(markdown: string): RawInline;",
			"  Code(code: string): CodeInline;",
			"  Em(...parts: InlineInput[]): EmphasisInline;",
			"  Strong(...parts: InlineInput[]): StrongInline;",
			"  Link(text: InlineInput, url: string, title?: string): LinkInline;",
			"}",
			"export interface TableBuilder {",
			"  Columns(...columns: ColumnInput[]): TableBuilder;",
			"  Align(...alignments: TableAlignment[]): TableBuilder;",
			"  Row(...cells: InlineInput[]): TableBuilder;",
			"  Rows(rows: InlineInput[][]): TableBuilder;",
			"  End(): MarkdownBuilder;",
			"}",
			"export interface MarkdownBuilder {",
			"  Title(text: InlineInput): MarkdownBuilder;",
			"  Heading(level: number, text: InlineInput): MarkdownBuilder;",
			"  Paragraph(...parts: InlineInput[]): MarkdownBuilder;",
			"  Text(text: string): MarkdownBuilder;",
			"  Raw(markdown: string): MarkdownBuilder;",
			"  ThematicBreak(): MarkdownBuilder;",
			"  Blockquote(body: unknown): MarkdownBuilder;",
			"  Callout(kind: string, title: string, body?: unknown): MarkdownBuilder;",
			"  BulletList(items: InlineInput[]): MarkdownBuilder;",
			"  OrderedList(items: InlineInput[], start?: number): MarkdownBuilder;",
			"  Checklist(items: ChecklistInput[]): MarkdownBuilder;",
			"  CodeBlock(language: string, code: string): MarkdownBuilder;",
			"  Table(): TableBuilder;",
			"  Validate(): ValidationResult;",
			"  Render(): MarkdownRenderResult;",
			"  RenderString(): string;",
			"  RenderHTML(): string;",
			"}",
		},
		Functions: []spec.Function{
			{Name: "parse", Params: []spec.Param{{Name: "input", Type: spec.String()}}, Returns: spec.Named("MarkdownNode")},
			{Name: "renderHTML", Params: []spec.Param{{Name: "input", Type: spec.String()}}, Returns: spec.String()},
			{Name: "walk", Params: []spec.Param{{Name: "root", Type: spec.Named("MarkdownNode")}, {Name: "visitor", Type: spec.Any()}}, Returns: spec.Void()},
			{Name: "textContent", Params: []spec.Param{{Name: "node", Type: spec.Named("MarkdownNode")}}, Returns: spec.String()},
			{Name: "validate", Params: []spec.Param{{Name: "value", Type: spec.Any()}}, Returns: spec.Named("ValidationResult")},
			{Name: "builder", Returns: spec.Named("MarkdownBuilder")},
			{Name: "inline", Returns: spec.Named("InlineFactory")},
		},
	}
}

func (mod module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").(*goja.Object)

	modules.SetExport(exports, mod.Name(), "parse", func(input string) (*MarkdownNode, error) {
		node, err := Parse(input)
		if err != nil {
			return nil, fmt.Errorf("markdown.parse: %w", err)
		}
		return node, nil
	})

	modules.SetExport(exports, mod.Name(), "renderHTML", func(input string) (string, error) {
		return RenderHTML(input)
	})

	modules.SetExport(exports, mod.Name(), "walk", func(root *MarkdownNode, visitor goja.Value) error {
		fn, ok := goja.AssertFunction(visitor)
		if !ok {
			return fmt.Errorf("markdown.walk: visitor must be a function")
		}
		state := &walkState{}
		return walkMarkdownNode(vm, root, nil, fn, 0, 0, nil, state)
	})

	modules.SetExport(exports, mod.Name(), "textContent", func(node *MarkdownNode) (string, error) {
		return TextContent(node)
	})

	modules.SetExport(exports, mod.Name(), "validate", func(value any) ValidationResult {
		switch v := value.(type) {
		case string:
			return ValidateInput(v)
		case *MarkdownNode:
			return ValidateNode(v)
		default:
			return ValidationResult{Valid: false, Errors: []string{fmt.Sprintf("markdown.validate: expected string or MarkdownNode, got %T", value)}}
		}
	})

	modules.SetExport(exports, mod.Name(), "builder", func() *MarkdownBuilder {
		return NewMarkdownBuilder()
	})

	modules.SetExport(exports, mod.Name(), "inline", func() InlineFactory {
		return NewInlineFactory()
	})
}

func walkMarkdownNode(vm *goja.Runtime, node *MarkdownNode, parent *MarkdownNode, fn goja.Callable, depth int, index int, path []int, state *walkState) error {
	if state == nil || state.Stopped || node == nil {
		return nil
	}

	ctx := WalkContext{
		Parent: parent,
		Depth:  depth,
		Index:  index,
		Path:   append([]int(nil), path...),
	}
	result, err := fn(goja.Undefined(), vm.ToValue(node), vm.ToValue(ctx))
	if err != nil {
		return err
	}

	skipChildren := false
	if result != nil && !goja.IsUndefined(result) && !goja.IsNull(result) {
		switch v := result.Export().(type) {
		case bool:
			skipChildren = !v
		case string:
			switch v {
			case "skip":
				skipChildren = true
			case "stop":
				state.Stopped = true
				return nil
			default:
				return fmt.Errorf("markdown.walk: unsupported visitor return string %q", v)
			}
		}
	}

	if skipChildren {
		return nil
	}
	for i, child := range node.Children {
		childPath := append(append([]int(nil), path...), i)
		if err := walkMarkdownNode(vm, child, node, fn, depth+1, i, childPath, state); err != nil {
			return err
		}
		if state.Stopped {
			return nil
		}
	}
	return nil
}

func init() {
	modules.Register(&module{})
}
