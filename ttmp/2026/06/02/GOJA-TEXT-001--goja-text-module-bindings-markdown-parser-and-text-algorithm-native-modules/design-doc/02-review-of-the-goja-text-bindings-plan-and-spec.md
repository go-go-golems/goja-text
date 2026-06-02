---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: go-go-goja/cmd/xgoja/internal/generate/gomod.go
      Note: xgoja generated go.mod and package replace behavior
    - Path: go-go-goja/pkg/xgoja/providers/core/core.go
      Note: Reference provider wrapper for NativeModule entries
    - Path: go-go-goja/pkg/xgoja/providers/host/host.go
      Note: Host fs provider config and security boundary
    - Path: goja-text/ttmp/2026/06/02/GOJA-TEXT-001--goja-text-module-bindings-markdown-parser-and-text-algorithm-native-modules/design-doc/01-goja-text-bindings-architecture-design-and-implementation-guide.md
      Note: Reviewed plan/spec under critique
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---





# Review of the Goja Text Bindings Plan and Spec

## Superseding Design Decision: Keep Go-Backed AST Objects

After this review was written, the project decision was clarified: the Markdown AST should remain a Go-backed object graph, not a plain `map[string]any` / lowercase JSON object tree. This is consistent with the broader goja approach in this codebase: Go-side objects preserve runtime type information, support builder-pattern and runtime validation, and allow better error messages when JavaScript passes invalid values back into Go functions.

Therefore, the review's recommendation to prefer JS-safe maps should be read as an implementation caution, not the final design direction. The final design direction is:

- `parse()` returns `*MarkdownNode` and JS accesses exported fields such as `node.Type`, `node.Children`, `node.Level`, and `node.Destination`.
- `walk(ast, visitor)` is the general traversal primitive.
- one-off Go exports for heading/link extraction are intentionally omitted; JavaScript implements those queries with `walk()`.
- if lowercase JSON-style objects are needed later, add an explicit adapter such as `toPlainObject(node)` rather than changing the primary AST representation.

## Executive Summary

This is a technical review of the first goja-text bindings plan, written for the intern who will implement the feature. The plan is strong in its broad architecture mapping: it correctly identifies `modules.NativeModule`, the engine runtime factory, xgoja providers, the host `fs` module, jsverbs, and goldmark as the main pieces of the system. It also makes the right high-level move from a hand-written CLI to an xgoja-generated binary.

The weak parts are mostly at the boundary between "conceptually correct" and "implementation-ready". The plan assumes some goja struct-field behavior that must be verified, does not yet account for xgoja `replace` wiring for the local `goja-text` provider module, duplicates a `goja-repl` section after the xgoja update, and hard-codes convenience extractors (`extractHeadings`, `extractLinks`) before introducing the more general traversal primitive (`walk`). These are normal first-plan issues, but they matter because they are the difference between a document that reads well and a document an engineer can implement without tripping.

The most important review recommendation is: keep the xgoja provider architecture, but tighten the implementation contract around three things:

1. **The JS-facing AST shape** — do not rely on Go struct `json` tags unless the runtime installs a `FieldNameMapper`, or convert nodes to `map[string]any` explicitly.
2. **The xgoja build spec** — add a local `replace: .` for the `goja-text` provider package, otherwise the generated temporary module may not resolve the local package.
3. **The traversal API** — add `walk(ast, callback, options?)` as the primary composability hook, then implement `extractHeadings` and `extractLinks` as JS helpers or thin Go wrappers over the same traversal semantics.

---

## What the Plan Gets Right

### 1. It identifies the correct integration seam: `modules.NativeModule`

The plan correctly starts from `go-go-goja/modules/common.go`, where the central interface is defined:

```go
type NativeModule interface {
    Name() string
    Doc() string
    Loader(*goja.Runtime, *goja.Object)
}
```

That is the right seam. A markdown parser should be a Go-native module with a `Loader` that sets exports on `module.exports`. This means the module can be reused from:

- direct engine tests,
- xgoja generated binaries,
- jsverbs command execution,
- future embedding hosts.

That is a good architectural instinct: build the smallest reusable unit first, then wire it into higher-level tools.

### 2. It uses the existing YAML module as the reference pattern

The plan correctly treats `go-go-goja/modules/yaml/yaml.go` as the closest model. The YAML module is small, stateless, and idiomatic:

```go
func (mod m) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)
    modules.SetExport(exports, mod.Name(), "parse", func(input string) (any, error) {
        var out any
        if err := yaml.Unmarshal([]byte(input), &out); err != nil {
            return nil, fmt.Errorf("yaml.parse: %w", err)
        }
        return out, nil
    })
}
```

The markdown module should follow this style: keep the loader simple, return plain Go values, let goja do conversion, and report errors with module-qualified messages such as `markdown.parse: ...`.

### 3. It correctly moves from a custom CLI to xgoja

The initial plan proposed a hand-written `cmd/goja-text/main.go`. The updated plan moves to xgoja. That is the right correction.

xgoja gives the project these things for free:

- a generated binary named `goja-text`,
- `eval` for quick one-liners,
- `run` for JS files,
- `repl` for the TUI REPL,
- `verbs` for jsverbs integration,
- provider-based module selection,
- guarded host modules such as `fs`.

That makes xgoja the right exercise harness. The feature is not "write another CLI"; the feature is "provide text modules that xgoja can assemble into a CLI".

### 4. It correctly identifies the host `fs` provider as the disk access story

The plan correctly points to `go-go-goja/pkg/xgoja/providers/host/host.go` and the explicit config needed to enable filesystem access:

```yaml
- package: go-go-goja-host
  name: fs
  as: fs
  config:
    allow: true
```

That is important. Host filesystem access is not a safe default. xgoja intentionally requires `config.allow=true`. This is the right security boundary to keep in mind: parsing Markdown is data-only, but reading arbitrary files is a host capability.

### 5. It gives the intern enough architectural context

The plan is long, but the breadth is useful. It covers:

- goja value conversion,
- native modules,
- engine runtime ownership,
- xgoja providers,
- jsverbs,
- goldmark,
- testing strategy,
- risks and open questions.

For a new intern, this matters. They need a map before they start editing code.

---

## What Needs Correction

### 1. The current document duplicates the `goja-repl` section

After the xgoja update, the design document contains both:

- the new `Part 5: The goja-repl — Interactive Testing (Alternative to xgoja)`, and
- the older `What is goja-repl?` / `How to Test a New Module in the REPL` / `The replapi.App Layer` section that was not fully removed.

This is not a design flaw, but it is a document hygiene problem. Duplicate sections confuse readers because they cannot tell which one is authoritative.

**Fix:** Collapse the goja-repl material into one short alternative-testing section. xgoja should remain the primary harness.

### 2. The plan assumes `json` struct tags become JS property names

The plan proposes:

```go
type MarkdownNode struct {
    Type     string          `json:"type"`
    Children []*MarkdownNode `json:"children,omitempty"`
    Text     string          `json:"text,omitempty"`
}
```

and then assumes JavaScript can read:

```javascript
node.type
node.children
node.text
```

That may not be true by default.

In goja, exported Go struct fields are exposed through reflection. Lowercase JS property names from `json` tags require a field-name mapper such as:

```go
vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
```

or:

```go
vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
```

No current `go-go-goja` engine code appears to install a field-name mapper globally. That means a `MarkdownNode{Type: "heading"}` may be visible as `node.Type`, not `node.type`, unless the module converts it to a map or configures the runtime.

This is the most important technical gap in the plan.

#### Better options

Use one of these explicit strategies:

**Option A — Return maps for JS-facing AST nodes.**

```go
func nodeToJSMap(n *MarkdownNode) map[string]any {
    out := map[string]any{"type": n.Type}
    if len(n.Children) > 0 {
        children := make([]any, 0, len(n.Children))
        for _, child := range n.Children {
            children = append(children, nodeToJSMap(child))
        }
        out["children"] = children
    }
    if n.Text != "" {
        out["text"] = n.Text
    }
    if n.Level != 0 {
        out["level"] = n.Level
    }
    return out
}
```

This is boring but reliable. JS sees exactly the property names we put in the map.

**Option B — Install `TagFieldNameMapper` for runtimes that use the markdown module.**

This is trickier because xgoja's `RuntimeFactory` creates runtimes centrally. A module loader can call `vm.SetFieldNameMapper(...)`, but that changes runtime-wide behavior and may surprise other modules. This should not be done casually.

**Option C — Keep Go structs internally, but export plain objects.**

This is the best compromise:

- Use typed Go structs internally for conversion and tests.
- Before returning to JS, convert to `map[string]any` or a dedicated `JSNode` map shape.

Recommended direction: **Option C**.

### 3. The xgoja spec is missing a local `replace` for the `goja-text` provider

The proposed spec says:

```yaml
packages:
  - id: goja-text
    import: github.com/go-go-golems/goja-text/pkg/xgoja/providers/text
```

During `xgoja build`, xgoja generates a temporary Go module. That temporary module imports the provider package. If `goja-text` is not published as a module version, `go mod tidy` will not know where to find it unless the spec provides a `replace`.

The buildspec supports package-level `replace` fields:

```yaml
packages:
  - id: goja-text
    import: github.com/go-go-golems/goja-text/pkg/xgoja/providers/text
    replace: .
```

Because `replace` is resolved relative to `spec.BaseDir`, `replace: .` points to the local `goja-text/` module when `xgoja.yaml` lives at `goja-text/xgoja.yaml`.

A development spec should probably be:

```yaml
name: goja-text
target:
  kind: xgoja
  output: dist/goja-text
packages:
  - id: goja-text
    import: github.com/go-go-golems/goja-text/pkg/xgoja/providers/text
    replace: .
  - id: go-go-goja-core
    import: github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/core
  - id: go-go-goja-host
    import: github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/host
```

The build command still needs the go-go-goja replacement:

```bash
go run ../go-go-goja/cmd/xgoja build \
  -f xgoja.yaml \
  --xgoja-replace ../go-go-goja
```

**Lesson:** whenever xgoja imports a provider from a local, unpublished module, the spec needs a package-level `replace`.

### 4. The plan overcommits to `extractHeadings` and `extractLinks` as Go exports

The plan lists:

```text
extractHeadings(input)
extractLinks(input)
```

as first-class exports. These are useful, but they are not the best primitive. They are specific queries. The general primitive is tree traversal.

A better module API should include:

```javascript
markdown.walk(ast, callback, options?)
```

Then `extractHeadings` and `extractLinks` can be implemented in JavaScript:

```javascript
function extractHeadings(input) {
  const ast = markdown.parse(input);
  const headings = [];
  markdown.walk(ast, node => {
    if (node.type === "heading") headings.push(node);
  });
  return headings;
}
```

This makes the module more composable. Users can extract:

- all links,
- headings by level,
- fenced code blocks by language,
- all paragraphs containing a pattern,
- all images without alt text,
- outline trees,
- internal links only,
- broken reference links.

You do not want to add a new Go export for every query.

#### Recommended traversal contract

```typescript
type WalkResult = void | boolean | "skip" | "stop";

declare function walk(
  root: MarkdownNode,
  visitor: (node: MarkdownNode, ctx: WalkContext) => WalkResult,
  options?: WalkOptions,
): void;

interface WalkContext {
  parent?: MarkdownNode;
  depth: number;
  index: number;
  path: number[];
  entering: boolean;
}

interface WalkOptions {
  order?: "pre" | "post" | "both";
  maxDepth?: number;
}
```

Return semantics:

- `undefined` / `true`: continue normally
- `false` / `"skip"`: skip children of this node
- `"stop"`: stop traversal entirely

This maps well to goldmark's own `ast.Walk` idea, but gives JS users a familiar visitor API.

### 5. The plan should distinguish internal AST nodes from JS API nodes

The plan mixes three concepts:

1. goldmark's `ast.Node`,
2. the Go-side `MarkdownNode` struct,
3. the JS-visible node object.

These should be separated in the design.

Recommended terminology:

```text
goldmark AST node
  ↓ convertGoldmarkNode(...)
internal MarkdownNode struct
  ↓ exportMarkdownNode(...)
JS MarkdownNode object
```

This gives us freedom to change one layer without breaking the others.

For example:

- goldmark stores source text in `text.Segment` references to `source []byte`;
- internal Go nodes may store `StartLine`, `EndLine`, and `Segments`;
- JS nodes should probably expose simple serializable fields like `type`, `text`, `children`, `position`.

### 6. The plan under-specifies goldmark extensions

The plan mentions CommonMark and goldmark extensions but does not decide what the first module supports.

This matters because users will immediately ask for:

- tables,
- task lists,
- strikethrough,
- autolinks,
- frontmatter,
- footnotes,
- heading IDs.

Goldmark supports many of these through extensions, but the AST shape changes when extensions are enabled.

Recommended MVP decision:

```text
Phase 1 supports CommonMark only.
Phase 2 adds a parse option object:
  markdown.parse(input, { gfm: true, frontmatter: true })
```

Do not silently enable every extension in v1. Keep the initial AST contract small and tested.

### 7. The TypeScript declaration plan is too loose

The current plan uses `spec.Any()` heavily:

```go
{ Name: "parse", Returns: spec.Any() }
```

That is fine for the first smoke test, but it misses one of xgoja's advantages: generated tools can ship useful API docs and typings.

At minimum, the module should use `RawDTS` to define the AST contract:

```go
RawDTS: []string{
  "export interface MarkdownNode {",
  "  type: string;",
  "  children?: MarkdownNode[];",
  "  text?: string;",
  "  level?: number;",
  "  destination?: string;",
  "  title?: string;",
  "  language?: string;",
  "  position?: SourcePosition;",
  "}",
  "export interface SourcePosition {",
  "  startLine: number;",
  "  startColumn: number;",
  "  endLine?: number;",
  "  endColumn?: number;",
  "}",
}
```

Then `parse` can return `spec.Named("MarkdownNode")` instead of `spec.Any()`.

### 8. The validation story is weak

The plan includes `validate(input)` but Markdown parsers often accept malformed-looking input by design. CommonMark is permissive. Many things that humans call "invalid" are valid Markdown.

So `validate()` needs a sharper definition. It should probably mean one of:

1. **Parser diagnostics** — only if goldmark exposes parse errors for enabled extensions.
2. **Lint diagnostics** — style or semantic rules (missing alt text, heading jumps, duplicate IDs).
3. **Structural sanity checks** — internal invariants after conversion.

Do not ship a vague `validate()` that always returns `{valid: true}`. That creates false confidence.

Better split:

```javascript
markdown.parse(input)              // parse only
markdown.lint(input, rules?)       // semantic/style diagnostics
markdown.checkAST(ast)             // internal invariant check, mostly for tests
```

For MVP, it is okay to omit `validate()` entirely until it has a meaningful contract.

---

## What Could Be Better in the Document Structure

### 1. Start with a smaller implementation target

The document is broad. That is good for orientation, but it risks making Phase 1 feel bigger than it is.

Phase 1 should fit on one screen:

```text
Phase 1 MVP:
1. Fix goja-text module path.
2. Add markdown NativeModule with parse(input) and renderHTML(input).
3. Return JS-safe maps, not reflected structs.
4. Add xgoja text provider.
5. Add xgoja.yaml with goja-text provider replace and host fs allow.
6. Build dist/goja-text.
7. Run one JS script that reads a .md file from disk and prints heading count.
```

Everything else should be Phase 2+.

### 2. Use "decision records" for choices that matter

The plan makes good decisions, but some are buried in prose. For interns, important decisions should be called out as explicit records:

```text
Decision: Use xgoja instead of custom CLI.
Reason: xgoja gives eval/run/repl/jsverbs and provider composition.
Consequence: Need xgoja provider package and xgoja.yaml replace entries.
```

Good decisions deserve traceability.

### 3. Separate tested facts from assumptions

Some statements are observed facts from files. Others are assumptions.

Examples:

- Fact: xgoja host provider requires `config.allow=true` for host fs.
- Fact: xgoja generated binary imports provider packages and calls their `Register()` functions.
- Assumption: `json` tags will define JS property names for reflected structs.
- Assumption: `validate()` can produce meaningful Markdown parser diagnostics.

Interns should learn to mark assumptions explicitly. A plan that says "verify this" is stronger than a plan that accidentally treats an assumption as a fact.

---

## What the Intern Should Have Looked At

### 1. goja field-name mapping documentation

They should have looked at goja's `FieldNameMapper` support before committing to lowercase JS properties from Go structs.

Relevant API concepts:

```go
vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
```

This is directly relevant to `node.type` vs `node.Type`.

### 2. xgoja `RenderGoMod` and package-level `replace`

They should have looked at:

- `go-go-goja/cmd/xgoja/internal/generate/gomod.go`
- `go-go-goja/cmd/xgoja/internal/buildspec/spec.go`
- `go-go-goja/cmd/xgoja/internal/buildspec/load.go`

This reveals how generated builds resolve provider imports and why local provider packages need `replace`.

### 3. xgoja examples with host provider

The plan did look at these, but it should have pulled the `replace` and config lessons into the final spec:

- `go-go-goja/examples/xgoja/01-core-provider/xgoja.yaml`
- `go-go-goja/examples/xgoja/02-host-provider/xgoja.yaml`
- `go-go-goja/examples/xgoja/06-runtime-filesystem/xgoja.yaml`

### 4. Goldmark extension and AST APIs

They should inspect goldmark's extension packages before claiming future support. In particular:

- `github.com/yuin/goldmark/extension`
- `github.com/yuin/goldmark/parser`
- `github.com/yuin/goldmark/ast`
- `github.com/yuin/goldmark/text`

This will clarify how to support GFM features without inventing an AST shape that won't fit extension nodes.

### 5. Existing module tests

They should look at test patterns in:

- `go-go-goja/modules/yaml/yaml_test.go`
- `go-go-goja/modules/fs/fs_test.go`
- `go-go-goja/pkg/xgoja/app/*_test.go`
- `go-go-goja/cmd/xgoja/internal/buildspec/*_test.go`

The implementation plan should mirror these tests instead of inventing a new testing style.

---

## Recommended Revised API

### Minimal MVP API

```javascript
const markdown = require("markdown");

const ast = markdown.parse(source);
const html = markdown.renderHTML(source);
markdown.walk(ast, (node, ctx) => {
  if (node.type === "heading") console.log(node.level, markdown.textContent(node));
});
```

### Exports

| Export | Phase | Purpose |
|---|---:|---|
| `parse(input, options?)` | 1 | Parse Markdown into JS-safe AST object |
| `renderHTML(input, options?)` | 1 | Render Markdown to HTML |
| `walk(ast, visitor, options?)` | 1 | General AST traversal primitive |
| `textContent(node)` | 1 | Extract plain text from a node subtree |
| `extractHeadings(inputOrAst)` | 2 | Convenience helper built on `walk` |
| `extractLinks(inputOrAst)` | 2 | Convenience helper built on `walk` |
| `lint(inputOrAst, rules?)` | 3 | Diagnostics / semantic checks |

### Example implementation of JS helpers using `walk`

```javascript
function extractHeadings(inputOrAst) {
  const ast = typeof inputOrAst === "string" ? markdown.parse(inputOrAst) : inputOrAst;
  const headings = [];
  markdown.walk(ast, node => {
    if (node.type === "heading") {
      headings.push({
        level: node.level,
        text: markdown.textContent(node),
        node,
      });
    }
  });
  return headings;
}

function extractLinks(inputOrAst) {
  const ast = typeof inputOrAst === "string" ? markdown.parse(inputOrAst) : inputOrAst;
  const links = [];
  markdown.walk(ast, node => {
    if (node.type === "link") {
      links.push({
        destination: node.destination,
        title: node.title || "",
        text: markdown.textContent(node),
        node,
      });
    }
  });
  return links;
}
```

---

## Recommended Revised xgoja Spec

Use this as the implementation baseline:

```yaml
name: goja-text
target:
  kind: xgoja
  output: dist/goja-text
packages:
  - id: goja-text
    import: github.com/go-go-golems/goja-text/pkg/xgoja/providers/text
    replace: .
  - id: go-go-goja-core
    import: github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/core
  - id: go-go-goja-host
    import: github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/host
runtimes:
  main:
    modules:
      - package: goja-text
        name: markdown
        as: markdown
      - package: go-go-goja-core
        name: path
        as: path
      - package: go-go-goja-core
        name: yaml
        as: yaml
      - package: go-go-goja-host
        name: fs
        as: fs
        config:
          allow: true
commands:
  eval:
    enabled: true
    runtime: main
  run:
    enabled: true
    runtime: main
  repl:
    enabled: true
    runtime: main
  jsverbs:
    enabled: false
    runtime: main
```

Start with `jsverbs.enabled: false` unless you also add a `jsverbs:` source. Enabling the command without sources is not necessarily wrong, but it may confuse a first smoke test.

Build command:

```bash
cd goja-text
go run ../go-go-goja/cmd/xgoja build \
  -f xgoja.yaml \
  --xgoja-replace ../go-go-goja
```

Smoke script:

```javascript
const fs = require("fs");
const markdown = require("markdown");

const source = fs.readFileSync("README.md", "utf-8");
const ast = markdown.parse(source);
let headings = [];
markdown.walk(ast, node => {
  if (node.type === "heading") headings.push(markdown.textContent(node));
});
console.log(JSON.stringify({ blocks: ast.children.length, headings }, null, 2));
```

Run:

```bash
dist/goja-text run examples/js/readme-headings.js
```

---

## Review Scorecard

| Area | Rating | Notes |
|---|---:|---|
| Architecture discovery | Strong | Correctly mapped modules, engine, xgoja, jsverbs, goldmark |
| Choice of xgoja | Strong | Right harness; avoids custom CLI drift |
| Native module design | Good | Correct seam, but JS AST shape needs correction |
| xgoja provider spec | Good idea, incomplete details | Needs package-level `replace: .`; maybe disable jsverbs until sources exist |
| JS API design | Mixed | `parse` and `renderHTML` good; add `walk`; delay vague `validate` |
| TypeScript declaration plan | Weak | Too much `any`; needs explicit `MarkdownNode` interfaces |
| Testing plan | Good but broad | Needs exact first smoke test and xgoja build validation |
| Document hygiene | Needs cleanup | Duplicate goja-repl section; some sections too long for Phase 1 implementation |

---

## Advice for Next Time

### 1. Always validate the build path, not just the runtime path

It is not enough to say "this package imports that package." With generated binaries, ask:

- What temporary module is generated?
- What does its `go.mod` contain?
- How will local, unpublished provider packages resolve?
- Which `replace` directives are needed?

xgoja is a build system, not just a runtime framework. Build resolution is part of the design.

### 2. Verify reflection assumptions with a tiny probe

When designing JS APIs backed by Go structs, write a tiny proof:

```go
vm := goja.New()
vm.Set("node", &MarkdownNode{Type: "heading"})
v, err := vm.RunString(`node.type || node.Type`)
```

This would immediately reveal whether `node.type` works. Never assume tags affect a JS runtime unless you verified the mapper.

### 3. Prefer primitives over convenience APIs

`extractHeadings` and `extractLinks` are useful, but `walk` is the real primitive. Design the primitive first, then build convenience helpers on top.

A good API lets users invent things you did not predict.

### 4. Do not ship vague validation APIs

If you cannot explain what `validate()` catches, do not add it yet. Markdown is permissive. A vague validator becomes a source of confusion.

Use precise names:

- `parse`
- `lint`
- `checkAST`
- `diagnoseLinks`

### 5. Keep Phase 1 brutally small

A good Phase 1 should answer one question: "Can we parse a Markdown file from disk inside a generated xgoja binary?"

Everything else is Phase 2.

### 6. Mark assumptions explicitly

A high-quality plan says:

```text
Assumption: goja struct tags can expose lower-case properties.
Verification: write TestMarkdownNodeJSPropertyNames before implementation.
```

That is better than presenting the assumption as fact.

---

## Final Recommendation

The intern's plan is directionally good and should be used as the starting point, not thrown away. The core architecture is right: `NativeModule` + xgoja provider + host `fs` + generated binary. The implementation should begin only after tightening the JS AST export strategy and xgoja build spec.

Before coding, make these document corrections:

1. Remove duplicate `goja-repl` section.
2. Add `walk(ast, visitor, options?)` to the MVP API.
3. Change AST export guidance from reflected structs to JS-safe maps, or explicitly install/test a field-name mapper.
4. Add `replace: .` to the local `goja-text` provider package in `xgoja.yaml`.
5. Delay or rename `validate()` until its semantics are clear.
6. Add concrete TypeScript declarations for `MarkdownNode`, `WalkContext`, and `WalkOptions`.

If those changes are made, the plan becomes implementation-ready.
