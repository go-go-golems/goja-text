---
Title: Markdown Builder Analysis Design and Implementation Guide
Ticket: GOJA-TEXT-005
Status: active
Topics:
    - goja
    - goja-bindings
    - native-modules
    - markdown
    - text-algorithms
    - xgoja
    - jsverbs
    - cli
    - templating
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/goja-text/jsverbs/markdown.js
      Note: Phase 4 builder-examples and builder-example jsverbs for generated Markdown
    - Path: cmd/goja-text/jsverbs/template.js
      Note: CLI jsverb and embedded-example structure to reuse
    - Path: cmd/goja-text/markdown-builder-assets/api-table.yaml
      Note: Embedded API-table data fixture for builder example
    - Path: cmd/goja-text/markdown-builder-assets/report.yaml
      Note: Embedded sprint-report data fixture for builder example
    - Path: cmd/goja-text/xgoja.yaml
      Note: |-
        xgoja module
        Phase 4 embedded asset mount for markdown-builder examples
    - Path: pkg/markdown/builder.go
      Note: Phase 1 fluent MarkdownBuilder methods
    - Path: pkg/markdown/builder_render.go
      Note: Phase 1 Markdown serialization
    - Path: pkg/markdown/builder_table.go
      Note: Phase 1 TableBuilder lifecycle
    - Path: pkg/markdown/builder_test.go
      Note: Phase 1 service tests for document rendering
    - Path: pkg/markdown/builder_types.go
      Note: Phase 1 typed Markdown builder document
    - Path: pkg/markdown/convert.go
      Note: Existing Goldmark AST conversion evidence for current markdown model
    - Path: pkg/markdown/module.go
      Note: |-
        Current markdown NativeModule exports; proposed builder() should extend this API
        Phase 2 exports markdown.builder and markdown.inline from the existing NativeModule
    - Path: pkg/markdown/module_test.go
      Note: Phase 2 goja runtime tests for builder chains
    - Path: pkg/markdown/parser.go
      Note: Existing Markdown parse/renderHTML/validate functions that the builder can reuse
    - Path: pkg/markdown/types.go
      Note: Existing Go-backed MarkdownNode pattern to mirror for builder result objects
    - Path: pkg/template/builder.go
      Note: Completed fluent Go-backed builder pattern used as implementation model
    - Path: pkg/template/module.go
      Note: NativeModule adapter and goja value conversion reference
    - Path: pkg/template/render.go
      Note: Structured render-result pattern for builder output
    - Path: pkg/xgoja/providers/text/doc/markdown-api-reference.md
      Note: Phase 3 API reference for builder
    - Path: pkg/xgoja/providers/text/doc/markdown-user-guide.md
      Note: Phase 3 user guide updates for generating Markdown from structured data
    - Path: pkg/xgoja/providers/text/text.go
      Note: Provider registration boundary for module visibility
ExternalSources: []
Summary: Intern-ready design and implementation guide for adding a fluent Go-backed Markdown builder API to goja-text.
LastUpdated: 2026-06-07T18:35:00-04:00
WhatFor: Use this as the implementation plan for a fluent Markdown document builder in goja-text.
WhenToUse: Before coding GOJA-TEXT-005, reviewing its API shape, or extending goja-text document-generation support.
---






# Markdown Builder Analysis Design and Implementation Guide

## Executive summary

`goja-text` already lets JavaScript parse Markdown, render Markdown to HTML, sanitize structured text, extract structured data, and render Go templates. What it does not yet provide is a convenient way for scripts to *create* clean Markdown documents without manually concatenating strings or writing a template for every document shape.

This ticket proposes a new fluent, Go-backed Markdown builder API. JavaScript code will call methods such as `Heading`, `Paragraph`, `List`, `Table`, `CodeBlock`, and `Callout`, while Go owns the underlying document model, normalization, escaping, table formatting, blank-line rules, and final serialization. The result should feel simple from JavaScript, but it should behave predictably enough for generated reports, prompt packs, release notes, API references, ticket summaries, and tabular output.

A representative JavaScript script should look like this:

```js
const md = require("markdown");

const doc = md.builder()
  .Title("Sprint report")
  .Paragraph("Generated from structured runtime data.")
  .Table()
    .Columns("Name", "Status", "Owner")
    .Row("Parser", "done", "Ada")
    .Row("Builder", "planned", "Intern")
    .End()
  .Heading(2, "Next steps")
  .Checklist([
    { text: "Implement service layer", checked: false },
    { text: "Add goja runtime tests", checked: false },
  ])
  .RenderString();

console.log(doc);
```

Expected output:

```md
# Sprint report

Generated from structured runtime data.

| Name | Status | Owner |
| --- | --- | --- |
| Parser | done | Ada |
| Builder | planned | Intern |

## Next steps

- [ ] Implement service layer
- [ ] Add goja runtime tests
```

The design intentionally mirrors the completed `template` module. The template module proves that `goja-text` can expose fluent Go-backed builders through goja, validate builder state in Go, provide TypeScript declarations, ship jsverbs, include embedded examples, and wire everything into the generated xgoja binary. The Markdown builder should reuse that pattern instead of becoming a JavaScript-only utility.

## Problem statement and scope

### The problem

JavaScript scripts in `goja-text` can currently produce Markdown in three common ways, each with drawbacks:

1. **String concatenation:** Fast to start, but brittle. The caller must manage blank lines, escaping, list indentation, table pipes, and code fences.
2. **Template rendering:** Useful for fixed document layouts, but awkward when the document is assembled conditionally or when many rows/sections are added programmatically.
3. **Manual AST construction:** Not currently exposed as a builder, and too low-level for ordinary report generation.

The user request specifically calls out avoiding "a lot of string formatting or templating" and producing clean Markdown documents, including tables. That means the API should be structured, fluent, and purpose-built for Markdown output.

### In scope

The first implementation should include:

- A Go service package for building Markdown documents as typed blocks.
- A goja `NativeModule` adapter, preferably as additional exports on the existing `markdown` module.
- A fluent builder API using Go-backed objects, not JS maps as the source of truth.
- Markdown block support for:
  - title/heading
  - paragraph
  - blank line / thematic break
  - bullet list
  - ordered list
  - checklist
  - blockquote
  - callout
  - fenced code block
  - raw Markdown escape hatch
  - table with alignment and escaping
- Inline helpers for text, emphasis, strong, code spans, links, and raw inline fragments.
- `RenderString()`, `Render()`, `Validate()`, and optional `RenderHTML()` integration through the existing Markdown renderer.
- TypeScript declarations.
- Runtime tests using `go-go-goja/pkg/engine`.
- CLI jsverbs for practical document generation from YAML/JSON data.
- Embedded examples demonstrating reports and tables.
- Help pages similar to the template module documentation.

### Out of scope for phase 1

Avoid these until the basic builder works:

- A full CommonMark AST editing system.
- Round-tripping arbitrary parsed Markdown while preserving formatting.
- Asynchronous JavaScript callbacks during rendering.
- WYSIWYG or rich text editing.
- Custom Markdown extension parsing.
- Perfect table width alignment for CJK/wide Unicode cells. The first version should be deterministic and readable, not terminal-layout perfect.

## Current-state architecture with evidence

### Existing Markdown module

The existing `markdown` module already exposes parsing, HTML rendering, traversal, text extraction, and validation. Its goja module declares functions in `pkg/markdown/module.go`:

- `parse(input)` returns a Go-backed `MarkdownNode` tree.
- `renderHTML(input)` converts Markdown to HTML.
- `walk(root, visitor)` traverses the Go-backed tree with a JavaScript callback.
- `textContent(node)` extracts plain text.
- `validate(value)` validates either a string or a node.

Evidence:

- `pkg/markdown/module.go:16` names the module `markdown`.
- `pkg/markdown/module.go:35-72` wires exports with `modules.SetExport`.
- `pkg/markdown/module.go:74-123` shows callback traversal using the goja runtime and `goja.Callable`.
- `pkg/markdown/types.go:3-24` defines `MarkdownNode` as a Go-backed AST node whose exported Go fields are visible in JavaScript.
- `pkg/markdown/parser.go:11-20` parses Markdown through Goldmark and converts it to the project-owned `MarkdownNode` type.
- `pkg/markdown/parser.go:22-29` renders Markdown to HTML with Goldmark.
- `pkg/markdown/convert.go:10-69` converts Goldmark AST nodes into `MarkdownNode` values.

This module gives the builder two reusable capabilities:

1. Serialize builder output to Markdown, then validate it with `ValidateInput`.
2. Offer `RenderHTML()` by passing the rendered Markdown string into `RenderHTML`.

### Existing template module as the builder pattern reference

The completed `template` module is the strongest model for this work. It has a Go-backed builder, service layer, module adapter, TypeScript declarations, jsverbs, embedded examples, help docs, and tests.

Evidence:

- `pkg/template/builder.go:13-18` defines `TemplateBuilder` with Go-owned config, custom functions, runtime pointer, and accumulated errors.
- `pkg/template/builder.go:27-52` implements fluent builder methods returning `*TemplateBuilder`.
- `pkg/template/builder.go:54-85` validates and freezes configuration in Go.
- `pkg/template/builder.go:88-100` parses into a Go-backed `TemplateSet`.
- `pkg/template/render.go:11-16` defines the parsed template set wrapper.
- `pkg/template/render.go:50-87` renders and returns a structured `RenderResult`.
- `pkg/template/types.go:24-44` defines config, render result, and template metadata structs.
- `pkg/template/typescript.go:5-42` publishes TypeScript declarations for the goja-facing API.
- `pkg/template/module.go:35-53` exports `text`, `html`, `renderText`, and `renderHTML`.
- `pkg/template/module.go:64-93` implements `JSFunc` and shows how to safely capture runtime-backed callbacks.
- `pkg/template/module_test.go:14-190` tests the JavaScript API through an actual goja runtime.

The Markdown builder should imitate this shape:

```go
type MarkdownBuilder struct {
    doc    *Document
    cfg    BuilderConfig
    errors []string
}

func (b *MarkdownBuilder) Heading(level int, text any) *MarkdownBuilder { ... }
func (b *MarkdownBuilder) Table() *TableBuilder { ... }
func (b *MarkdownBuilder) Validate() ValidationResult { ... }
func (b *MarkdownBuilder) RenderString() (string, error) { ... }
```

### xgoja provider and generated binary wiring

The generated `goja-text` binary only sees modules that are registered with the provider and selected in the buildspec.

Evidence:

- `pkg/xgoja/providers/text/text.go:9-13` blank-imports module packages so their `init()` registrations run.
- `pkg/xgoja/providers/text/text.go:18-23` lists the module names exposed by the provider.
- `pkg/xgoja/providers/text/text.go:26-42` registers provider entries and help docs.
- `cmd/goja-text/xgoja.yaml:25-37` selects `markdown`, `sanitize`, `extract`, and `template` modules for the generated binary.
- `cmd/goja-text/xgoja.yaml:62-65` embeds jsverbs.
- `cmd/goja-text/xgoja.yaml:66-70` wires runtime API help sources.

If the builder is added to the existing `markdown` module, no new module entry is needed. If it becomes a new module name, both `text.go` and `xgoja.yaml` must be updated.

### jsverb and embedded-example pattern

The template jsverbs demonstrate practical CLI commands and one important scanner constraint.

Evidence:

- `cmd/goja-text/jsverbs/template.js:27-79` stores helper functions inside a `helpers` object, avoiding accidental exposure as CLI commands.
- `cmd/goja-text/jsverbs/template.js:81-127` implements `template text` and `template html` commands.
- `cmd/goja-text/jsverbs/template.js:177-214` lists and renders embedded examples.
- `cmd/goja-text/jsverbs/template.js:216-229` demonstrates a JS callback helper.

For the Markdown builder, jsverb helper functions should also stay inside a helper object. Top-level helper functions may be detected as commands by the jsverb scanner.

## Gap analysis

### What exists today

`goja-text` can:

- Parse Markdown into Go-backed nodes.
- Render Markdown to HTML.
- Walk parsed Markdown with JavaScript callbacks.
- Validate Markdown node invariants.
- Render fixed template files with data.
- Generate documentation via templates and jsverbs.

### What is missing

`goja-text` cannot yet:

- Build Markdown documents through a fluent structured API.
- Generate tables without manual pipe escaping.
- Normalize blank lines and block boundaries automatically.
- Add Markdown blocks conditionally without mixing business logic and string formatting.
- Provide a reusable API for report-like output from scripts.
- Offer a safe raw-Markdown escape hatch with clear boundaries.

### Why templates are not enough

Templates are excellent when the document shape is known up front:

```gotemplate
# {{ .Title }}

{{ range .Items }}- {{ . }}
{{ end }}
```

But a builder is better when the script decides what sections exist:

```js
const doc = md.builder().Title(report.title);

if (report.failures.length > 0) {
  doc.Callout("warning", "Failures", report.failures.join("\n"));
}

if (report.metrics.length > 0) {
  doc.Table().Columns("Metric", "Value").Rows(report.metrics).End();
}

return doc.RenderString();
```

The builder reduces incidental complexity: the script describes the document, while Go handles Markdown syntax.

## Proposed architecture

### Package layout

Add a builder subpackage or files under the existing markdown package. Prefer keeping it in `pkg/markdown` unless it grows large enough to justify `pkg/markdownbuilder`.

Recommended first layout:

```text
pkg/markdown/
  builder_types.go        — Document, block, inline, table, config/result types
  builder.go              — MarkdownBuilder fluent methods and validation
  builder_table.go        — TableBuilder and table formatting
  builder_render.go       — Markdown serialization and escaping helpers
  builder_module.go       — goja export wiring, if kept separate from module.go
  builder_typescript.go   — TypeScript declarations additions
  builder_test.go         — service tests
  builder_module_test.go  — runtime integration tests
```

Keep builder files separate from parser files. This makes review easier and avoids mixing "Markdown in" code with "Markdown out" code.

### Runtime architecture diagram

```text
┌─────────────────────────────────────────────────────────┐
│                    JavaScript script                    │
│  const md = require("markdown");                       │
│  md.builder().Heading(1, "Report").Table()...          │
└──────────────────────────┬──────────────────────────────┘
                           │ fluent goja calls
                           ▼
┌─────────────────────────────────────────────────────────┐
│               markdown NativeModule layer               │
│  Loader() exports builder(), render(), parse(), ...     │
│  Go-backed MarkdownBuilder / TableBuilder / Result      │
└──────────────────────────┬──────────────────────────────┘
                           │ appends typed blocks
                           ▼
┌─────────────────────────────────────────────────────────┐
│                   Go builder service                    │
│  Document []Block                                       │
│  Block variants: Heading, Paragraph, Table, Code, ...   │
│  Inline values: text, raw, code, link, emphasis         │
│  Validate() + RenderMarkdown()                         │
└──────────────────────────┬──────────────────────────────┘
                           │ serialized Markdown
                           ▼
┌─────────────────────────────────────────────────────────┐
│              Existing markdown package utilities        │
│  ValidateInput(markdown)                                │
│  RenderHTML(markdown) via Goldmark                      │
└─────────────────────────────────────────────────────────┘
```

### Data model

Use typed Go structs. Avoid making JavaScript maps the internal representation.

```go
type Document struct {
    Blocks []Block
}

type Block interface {
    blockKind() string
}

type HeadingBlock struct {
    Level int
    Text  []Inline
}

type ParagraphBlock struct {
    Inlines []Inline
}

type ListBlock struct {
    Ordered bool
    Start   int
    Items   []ListItem
}

type ChecklistBlock struct {
    Items []ChecklistItem
}

type CodeBlock struct {
    Language string
    Code     string
}

type TableBlock struct {
    Columns []TableColumn
    Rows    [][]InlineCell
}

type CalloutBlock struct {
    Kind  string // note, tip, warning, danger, info
    Title string
    Body  []Block
}
```

Inline values should also be typed:

```go
type Inline interface { inlineKind() string }

type TextInline struct { Text string }
type RawInline struct { Markdown string }
type CodeInline struct { Code string }
type EmphasisInline struct { Children []Inline }
type StrongInline struct { Children []Inline }
type LinkInline struct {
    Text []Inline
    URL  string
    Title string
}
```

A typed model lets the renderer enforce rules:

- Paragraph text escapes Markdown-sensitive characters.
- Code spans choose an appropriate number of backticks.
- Tables escape pipes and normalize newlines inside cells.
- Fenced code blocks choose fences longer than any run inside the code.
- Raw Markdown is allowed only when explicitly requested.

### Public JavaScript API sketch

The JavaScript API should be fluent and PascalCase-compatible with existing Go-backed APIs.

```ts
export interface MarkdownBuilder {
  Title(text: InlineInput): MarkdownBuilder;
  Heading(level: number, text: InlineInput): MarkdownBuilder;
  Paragraph(...parts: InlineInput[]): MarkdownBuilder;
  Text(text: string): MarkdownBuilder;
  Raw(markdown: string): MarkdownBuilder;
  Blockquote(textOrBuilder: unknown): MarkdownBuilder;
  Callout(kind: string, title: string, body?: unknown): MarkdownBuilder;
  BulletList(items: ListInput[]): MarkdownBuilder;
  OrderedList(items: ListInput[], start?: number): MarkdownBuilder;
  Checklist(items: ChecklistInput[]): MarkdownBuilder;
  CodeBlock(language: string, code: string): MarkdownBuilder;
  ThematicBreak(): MarkdownBuilder;
  Table(): TableBuilder;
  Validate(): ValidationResult;
  Render(): MarkdownRenderResult;
  RenderString(): string;
  RenderHTML(): string;
}

export interface TableBuilder {
  Columns(...columns: ColumnInput[]): TableBuilder;
  Align(...alignments: Alignment[]): TableBuilder;
  Row(...cells: CellInput[]): TableBuilder;
  Rows(rows: CellInput[][]): TableBuilder;
  End(): MarkdownBuilder;
}
```

Top-level functions:

```ts
export function builder(): MarkdownBuilder;
export function document(): MarkdownBuilder; // optional alias
export function render(build: (doc: MarkdownBuilder) => unknown): MarkdownRenderResult;
export function inline(): InlineFactory; // optional phase 2
```

Recommended minimal first example:

```js
const md = require("markdown");

const result = md.builder()
  .Title("API Reference")
  .Paragraph("Generated from module metadata.")
  .Table()
    .Columns("Function", "Description")
    .Row("parse(input)", "Parse Markdown into a Go-backed node tree.")
    .Row("renderHTML(input)", "Render Markdown to HTML.")
    .End()
  .Render();

console.log(result.Text);
console.log(result.Bytes);
console.log(result.Blocks);
```

### Render result

Use a structured result similar to `template.RenderResult`.

```go
type MarkdownRenderResult struct {
    Text   string `json:"text"`
    Bytes  int    `json:"bytes"`
    Blocks int    `json:"blocks"`
}
```

Optional phase-1 additions:

```go
type MarkdownRenderResult struct {
    Text         string   `json:"text"`
    Bytes        int      `json:"bytes"`
    Blocks       int      `json:"blocks"`
    Warnings     []string `json:"warnings,omitempty"`
    ValidMarkdown bool    `json:"validMarkdown"`
}
```

Keep result fields PascalCase in JavaScript because goja exposes exported Go fields, as the existing template and markdown modules already document.

## Markdown serialization rules

### Block spacing

The renderer should own blank lines. Every block renderer returns text without leading/trailing document-level blank lines; the document renderer joins blocks with exactly one blank line.

Pseudocode:

```go
func RenderDocument(doc *Document) string {
    parts := []string{}
    for _, block := range doc.Blocks {
        rendered := strings.Trim(renderBlock(block), "\n")
        if rendered != "" {
            parts = append(parts, rendered)
        }
    }
    return strings.Join(parts, "\n\n") + "\n"
}
```

### Heading rendering

```go
func renderHeading(h HeadingBlock) string {
    level := clamp(h.Level, 1, 6)
    return strings.Repeat("#", level) + " " + renderInline(h.Text, inlineContextNormal)
}
```

Validation should reject invalid levels rather than silently clamping in phase 1. Clamping is shown only to explain the serializer shape.

### Paragraph rendering

Paragraphs should collapse internal newlines to spaces unless the caller explicitly uses raw Markdown or line-break APIs.

```go
func renderParagraph(p ParagraphBlock) string {
    return normalizeParagraph(renderInline(p.Inlines, inlineContextNormal))
}
```

### Table rendering

Tables are the main reason to build this feature. The first version should render GitHub-Flavored-Markdown-style pipe tables.

```go
type Alignment string
const (
    AlignDefault Alignment = "default"
    AlignLeft    Alignment = "left"
    AlignCenter  Alignment = "center"
    AlignRight   Alignment = "right"
)
```

Rules:

- Header count defines the column count.
- Each row must have the same number of cells, or validation fails.
- Cell text escapes `|` as `\|`.
- Cell newlines become `<br>` by default.
- Inline code inside cells must still escape pipes if the Markdown target requires it.
- Alignment row uses `---`, `:---`, `:---:`, or `---:`.

Pseudocode:

```go
func renderTable(t TableBlock) (string, error) {
    if len(t.Columns) == 0 { return "", errors.New("table requires columns") }
    widths := computeWidths(t)

    lines := []string{
        renderTableRow(columnLabels(t.Columns), widths),
        renderAlignmentRow(t.Columns, widths),
    }
    for _, row := range t.Rows {
        if len(row) != len(t.Columns) { return "", errors.New("row width mismatch") }
        lines = append(lines, renderTableRow(row, widths))
    }
    return strings.Join(lines, "\n"), nil
}

func escapeTableCell(s string) string {
    s = strings.ReplaceAll(s, "\n", "<br>")
    s = strings.ReplaceAll(s, "|", `\|`)
    return strings.TrimSpace(s)
}
```

Example:

```js
md.builder()
  .Table()
    .Columns(
      { label: "Name", align: "left" },
      { label: "Score", align: "right" }
    )
    .Row("Ada", 42)
    .Row("Linus | Kernel", 100)
    .End()
  .RenderString();
```

Output:

```md
| Name | Score |
| :--- | ---: |
| Ada | 42 |
| Linus \| Kernel | 100 |
```

### Fenced code block rendering

The renderer should choose a fence that is longer than any backtick sequence in the code.

```go
func fenceFor(code string) string {
    maxRun := longestRun(code, '`')
    return strings.Repeat("`", max(3, maxRun+1))
}
```

This avoids broken code blocks when code contains triple backticks.

### Raw Markdown escape hatch

The builder should include `Raw(markdown string)` but document it clearly:

- Use it for advanced Markdown that the builder does not support yet.
- Do not use it for untrusted user data.
- Raw blocks are inserted as block text and still separated by document-level blank lines.

Raw inline should be separate from raw block:

```js
doc.Paragraph(md.inline().Text("Status: ").Raw("<mark>experimental</mark>"));
```

This distinction makes review easier and prevents accidental block injection from inline helper calls.

## Detailed API design

### Module placement

The preferred phase-1 API is to add `builder()` to `require("markdown")`:

```js
const markdown = require("markdown");
const text = markdown.builder().Title("Report").RenderString();
```

Rationale:

- The existing module already owns Markdown parse/render/validate concerns.
- Users will search for Markdown functionality under `markdown`.
- No new xgoja module selection is needed if the existing module remains named `markdown`.

A separate `require("markdown-builder")` can be considered later if the API becomes large enough to justify a module split.

### Builder methods for phase 1

```go
func NewMarkdownBuilder() *MarkdownBuilder

func (b *MarkdownBuilder) Title(text any) *MarkdownBuilder
func (b *MarkdownBuilder) Heading(level int, text any) *MarkdownBuilder
func (b *MarkdownBuilder) Paragraph(parts ...any) *MarkdownBuilder
func (b *MarkdownBuilder) Raw(markdown string) *MarkdownBuilder
func (b *MarkdownBuilder) Blockquote(body any) *MarkdownBuilder
func (b *MarkdownBuilder) Callout(kind, title string, body any) *MarkdownBuilder
func (b *MarkdownBuilder) BulletList(items any) *MarkdownBuilder
func (b *MarkdownBuilder) OrderedList(items any) *MarkdownBuilder
func (b *MarkdownBuilder) Checklist(items any) *MarkdownBuilder
func (b *MarkdownBuilder) CodeBlock(language, code string) *MarkdownBuilder
func (b *MarkdownBuilder) ThematicBreak() *MarkdownBuilder
func (b *MarkdownBuilder) Table() *TableBuilder
func (b *MarkdownBuilder) Validate() ValidationResult
func (b *MarkdownBuilder) Render() (*MarkdownRenderResult, error)
func (b *MarkdownBuilder) RenderString() (string, error)
func (b *MarkdownBuilder) RenderHTML() (string, error)
```

### TableBuilder methods

```go
type TableBuilder struct {
    parent *MarkdownBuilder
    table  *TableBlock
    closed bool
}

func (t *TableBuilder) Columns(columns ...any) *TableBuilder
func (t *TableBuilder) Align(alignments ...string) *TableBuilder
func (t *TableBuilder) Row(cells ...any) *TableBuilder
func (t *TableBuilder) Rows(rows any) *TableBuilder
func (t *TableBuilder) End() *MarkdownBuilder
```

Important behavior:

- `End()` appends the table to the parent exactly once.
- Calling methods after `End()` records a validation error.
- If the user forgets `End()`, the table is not appended. Consider a later `Render()` guard that detects open child builders, but that requires parent tracking.

### Inline factory (phase 2 or late phase 1)

For the first implementation, strings can be plain escaped text. To support richer inline content, add exported helper constructors:

```js
const md = require("markdown");
const i = md.inline();

doc.Paragraph(
  "See ",
  i.Link("the docs", "https://example.com"),
  " and run ",
  i.Code("go test ./...")
);
```

Go types:

```go
type InlineFactory struct{}
func (InlineFactory) Text(text string) TextInline
func (InlineFactory) Raw(markdown string) RawInline
func (InlineFactory) Code(code string) CodeInline
func (InlineFactory) Em(text any) EmphasisInline
func (InlineFactory) Strong(text any) StrongInline
func (InlineFactory) Link(text any, url string) LinkInline
```

## Data conversion strategy

The goja adapter should accept JavaScript values at the API boundary and convert them into Go types immediately. The internal `Document` should not store `goja.Value`.

Examples:

- `Paragraph("hello")` converts to `[]Inline{TextInline{"hello"}}`.
- `BulletList(["a", "b"])` converts to `ListBlock{Items: []ListItem{...}}`.
- `Checklist([{ text: "done", checked: true }])` converts to typed checklist items.
- `Table().Rows(data)` converts JS arrays into Go rows.

Pseudocode:

```go
func normalizeInlineInput(value any) ([]Inline, error) {
    switch v := value.(type) {
    case string:
        return []Inline{TextInline{Text: v}}, nil
    case TextInline, RawInline, CodeInline, LinkInline:
        return []Inline{v}, nil
    case []any:
        var out []Inline
        for _, item := range v { out = append(out, normalizeInlineInput(item)...) }
        return out, nil
    default:
        return []Inline{TextInline{Text: fmt.Sprint(v)}}, nil
    }
}
```

In `Loader`, exported functions can accept `goja.Value` where needed and call `Export()` in the adapter, following the template module's `exportTemplateData` pattern in `pkg/template/module.go:95-100`.

## Validation strategy

Validation should accumulate errors during builder calls and report them before rendering.

```go
type ValidationResult struct {
    Valid  bool     `json:"valid"`
    Errors []string `json:"errors,omitempty"`
}
```

Validate:

- Heading level is 1 through 6.
- Callout kind is non-empty and normalized.
- Table has at least one column.
- Every table row has exactly the column count.
- Table alignments are recognized.
- Ordered list start is positive.
- Code block language has no spaces or newlines.
- No child builder is appended twice.
- Raw blocks may be flagged with warnings if desired, but should not be errors.

Rendering should call validation and return a namespaced error:

```go
func (b *MarkdownBuilder) Render() (*MarkdownRenderResult, error) {
    result := b.Validate()
    if !result.Valid {
        return nil, fmt.Errorf("markdown.builder.validate: %s", joinErrors(result.Errors))
    }
    text := RenderDocument(b.doc)
    return &MarkdownRenderResult{Text: text, Bytes: len([]byte(text)), Blocks: len(b.doc.Blocks)}, nil
}
```

## Implementation phases

### Phase 1: Service-layer document builder

Files:

- `pkg/markdown/builder_types.go`
- `pkg/markdown/builder.go`
- `pkg/markdown/builder_table.go`
- `pkg/markdown/builder_render.go`
- `pkg/markdown/builder_test.go`

Implement:

- Document and block types.
- MarkdownBuilder and TableBuilder.
- Escaping helpers.
- Table rendering.
- Validation.
- Service tests only, no goja runtime yet.

Validation commands:

```bash
cd /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text
go test ./pkg/markdown -count=1
```

### Phase 2: goja module exports

Files:

- `pkg/markdown/module.go`
- `pkg/markdown/builder_module_test.go`

Add exports:

```go
modules.SetExport(exports, mod.Name(), "builder", func() *MarkdownBuilder {
    return NewMarkdownBuilder()
})
modules.SetExport(exports, mod.Name(), "render", func(build goja.Value) (*MarkdownRenderResult, error) {
    // optional callback-based sugar
})
```

Runtime test pattern should follow `pkg/template/module_test.go:178-190`, which builds a runtime with `engine.MiddlewareOnly("template")`. For this feature use `engine.MiddlewareOnly("markdown")`.

Example runtime test:

```go
func TestRequireMarkdownBuilderTable(t *testing.T) {
    rt := newMarkdownRuntime(t)
    ret, err := rt.Owner.Call(ctx, "markdown.builder.table", func(ctx context.Context, vm *goja.Runtime) (any, error) {
        value, runErr := vm.RunString(`
            const markdown = require("markdown");
            markdown.builder()
              .Title("Report")
              .Table().Columns("Name", "Status").Row("Parser", "done").End()
              .Render().Text;
        `)
        if runErr != nil { return nil, runErr }
        return value.Export(), nil
    })
    // assert output contains table
}
```

### Phase 3: TypeScript declarations and help docs

Files:

- `pkg/markdown/typescript.go` or equivalent existing declarations.
- `pkg/xgoja/providers/text/doc/markdown-builder-api-reference.md`
- `pkg/xgoja/providers/text/doc/markdown-builder-user-guide.md`

The template module's help pages provide the style and frontmatter pattern:

- `pkg/xgoja/providers/text/doc/template-api-reference.md:23-26` explains Go-backed naming.
- `pkg/xgoja/providers/text/doc/template-user-guide.md:22-25` introduces the user-facing purpose.
- `pkg/xgoja/providers/text/doc/template-writing-documentation.md:28-43` shows documentation-generation examples and embedded examples.

### Phase 4: jsverbs and embedded examples

Files:

- `cmd/goja-text/jsverbs/markdown-builder.js` or extend `markdown.js`.
- `cmd/goja-text/markdown-builder-assets/report.yaml`
- `cmd/goja-text/markdown-builder-assets/report.js` or data fixtures.
- `cmd/goja-text/xgoja.yaml` if new assets are added.

Suggested commands:

```bash
goja-text markdown build-table data.yaml --output-path report.md
goja-text markdown builder-example report
goja-text markdown builder-examples
```

Keep helper functions inside an object, following `cmd/goja-text/jsverbs/template.js:27-79`.

### Phase 5: xgoja generated binary validation

If only the existing `markdown` module changes, provider wiring may not require new module entries. Still regenerate and build the binary because TypeScript/help/jsverbs/assets may change.

Commands modeled after the template work:

```bash
cd /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/cmd/goja-text
GOTOOLCHAIN=go1.26.4 GOWORK=off go generate
GOTOOLCHAIN=go1.26.4 GOWORK=off go build -o ../../dist/goja-text .
```

The template project diary observed that both flags may be necessary when working inside the nested generated module.

## Decision records

### Decision: Add the fluent API to `require("markdown")`

- **Context:** The feature creates Markdown, while the existing `markdown` module parses, renders, walks, and validates Markdown.
- **Options considered:** Add `markdown.builder()` to the existing module; create a new `markdown-builder` module; create a generic `doc` module.
- **Decision:** Prefer `markdown.builder()` for phase 1.
- **Rationale:** It is discoverable, reuses existing Markdown utilities, and avoids additional xgoja module selection. It also keeps all Markdown operations in one API namespace.
- **Consequences:** The `markdown` module grows larger. If it becomes too broad, builder files can still remain physically separate and a later alias can be added.
- **Status:** proposed

### Decision: Use a Go-owned document tree, not direct string concatenation

- **Context:** The user wants to avoid brittle string formatting and templates.
- **Options considered:** Append raw strings; generate templates; construct a typed document model; expose Goldmark AST construction directly.
- **Decision:** Construct a small project-owned document model and serialize it to Markdown.
- **Rationale:** A typed model lets Go own escaping, table formatting, and validation while keeping JavaScript fluent.
- **Consequences:** The project must maintain a renderer, but only for a controlled subset of Markdown blocks.
- **Status:** proposed

### Decision: Make tables a first-class child builder

- **Context:** Tables are explicitly requested and are one of the easiest Markdown structures to get wrong by hand.
- **Options considered:** `Table(columns, rows)` one-shot method; chainable `Table().Columns().Row().End()` child builder; raw Markdown table helper.
- **Decision:** Provide a chainable `TableBuilder`, with optional one-shot helpers later.
- **Rationale:** The child builder supports incremental row construction from loops and keeps the parent chain readable.
- **Consequences:** The implementation must handle `End()` idempotence and validation for unfinished or reused child builders.
- **Status:** proposed

### Decision: Escape inline text by default and require explicit raw Markdown

- **Context:** Generated Markdown often includes user or tool data that may contain pipes, brackets, backticks, and angle brackets.
- **Options considered:** Trust all strings as Markdown; escape all strings; add per-method escape flags; separate `Text` and `Raw` inline types.
- **Decision:** Treat ordinary strings as text and add explicit `Raw` escape hatches.
- **Rationale:** This is safer and matches the goal of clean generated documents. Reviewers can grep for `Raw(` to find risky injection points.
- **Consequences:** Some advanced Markdown requires explicit raw fragments or future inline constructors.
- **Status:** proposed

### Decision: Provide `RenderHTML()` as convenience, not as the primary output

- **Context:** Existing `markdown.RenderHTML` already uses Goldmark.
- **Options considered:** Only render Markdown; also render HTML through Goldmark; add an HTML builder.
- **Decision:** Add `RenderHTML()` as a convenience method after `RenderString()` is correct.
- **Rationale:** The feature is about Markdown output, but HTML preview is useful and already available.
- **Consequences:** HTML rendering inherits Goldmark defaults and should not be confused with the separate `template.html()` API.
- **Status:** proposed

## Testing strategy

### Service tests

Add tests for:

- Empty document rendering.
- Heading levels and validation errors.
- Paragraph escaping.
- Bullet and ordered lists.
- Checklist items.
- Code block fence selection.
- Raw Markdown block insertion.
- Table rendering with alignment.
- Table pipe/newline escaping.
- Validation error for row width mismatch.
- RenderHTML bridge.

Example assertions:

```go
func TestMarkdownBuilderTableEscapesPipes(t *testing.T) {
    out, err := NewMarkdownBuilder().
        Table().Columns("Name", "Note").Row("Ada", "uses | pipe").End().
        RenderString()
    if err != nil { t.Fatal(err) }
    if !strings.Contains(out, `uses \| pipe`) { t.Fatalf("out = %s", out) }
}
```

### Runtime tests

Runtime tests should verify what JavaScript sees:

- `require("markdown").builder()` exists.
- Fluent calls return the right Go-backed objects.
- Result fields are available as `Text`, `Bytes`, and `Blocks`.
- Table child builder returns to the parent on `End()`.
- Invalid table rows throw or return validation errors consistently.

### Generated binary smoke tests

After jsverbs and help docs:

```bash
./dist/goja-text help goja-text-markdown-builder-api-reference
./dist/goja-text markdown builder-examples
./dist/goja-text markdown builder-example report
./dist/goja-text eval 'const md=require("markdown"); md.builder().Title("X").RenderString()'
```

## Risks and mitigations

### Risk: The API becomes too broad

Markdown has many constructs. Keep phase 1 small and focus on report-generation blocks. Add advanced inline features only after basic block rendering is tested.

### Risk: Escaping surprises users

Generated Markdown may look over-escaped if users expect raw Markdown strings. Mitigate with clear docs: ordinary strings are text; `Raw` is an explicit escape hatch.

### Risk: Table formatting handles ASCII width only

The first implementation can align based on byte or rune counts. Document that display-width-perfect tables are not guaranteed for CJK/wide glyphs. The Markdown remains valid even if columns are not visually perfect in a monospace terminal.

### Risk: Child builder lifecycle bugs

`Table().End()` introduces state. Make `End()` idempotent or explicitly error on double use. Add tests for double `End()` and row calls after `End()`.

### Risk: Module bloat

Adding builder APIs to `markdown` increases module size. Keep files modular and declarations organized. Revisit a separate module alias only after implementation pressure appears.

## Intern implementation checklist

1. Read `pkg/template/builder.go`, `pkg/template/render.go`, and `pkg/template/module.go` to understand the Go-backed builder pattern.
2. Read `pkg/markdown/module.go`, `pkg/markdown/types.go`, and `pkg/markdown/parser.go` to understand the current Markdown API.
3. Implement service types and renderer without goja imports.
4. Add service tests until `go test ./pkg/markdown -count=1` passes.
5. Add `builder()` export to the markdown module.
6. Add runtime tests using the engine middleware pattern from `pkg/template/module_test.go`.
7. Add TypeScript declarations.
8. Add help pages and examples.
9. Add jsverbs, keeping helpers in an object.
10. Regenerate and build the xgoja binary.
11. Validate docs and update the GOJA-TEXT-005 diary with exact commands and failures.

## File reference map

Start review here:

- `/home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/markdown/module.go` — Current JavaScript module exports and callback traversal pattern.
- `/home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/markdown/types.go` — Current Go-backed Markdown AST node fields.
- `/home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/markdown/parser.go` — Existing parse, HTML render, and validation entrypoints.
- `/home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/markdown/convert.go` — Existing Goldmark-to-project AST conversion.
- `/home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/template/builder.go` — Fluent builder pattern to copy conceptually.
- `/home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/template/render.go` — Go-backed rendered result pattern.
- `/home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/template/module.go` — NativeModule adapter and JS callback handling.
- `/home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/template/typescript.go` — TypeScript declaration pattern.
- `/home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/template/module_test.go` — Runtime integration test pattern.
- `/home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/cmd/goja-text/jsverbs/template.js` — jsverb structure and embedded example pattern.
- `/home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/xgoja/providers/text/text.go` — Provider module registration.
- `/home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/cmd/goja-text/xgoja.yaml` — Generated binary module/jsverb/help/assets selection.

## Open questions

- Should the initial builder include an inline factory, or should phase 1 accept only strings and raw Markdown blocks?
- Should `Table().Rows()` accept arrays of objects and choose columns by key, or only arrays of arrays at first?
- Should `Render()` always run the existing Markdown validator, or should validation remain builder-structural only for speed?
- Should `Raw()` be allowed in tables, or should table cells always be escaped text in phase 1?
- Should jsverbs live under existing `markdown` commands or a new `markdown-builder` command group?

## Final recommendation

Implement `markdown.builder()` as a Go-backed document builder inside the existing `markdown` module. Start with a small typed block model and excellent table support. Keep ordinary strings safe by default, add explicit raw escape hatches, validate aggressively, and reuse the template module's proven patterns for goja exports, TypeScript declarations, runtime tests, help docs, jsverbs, embedded examples, and generated binary validation.
