---
Title: 'Goja Text Module Bindings: Sanitize YAML and JSON Native Module — Design and Implementation Guide'
Ticket: GOJA-TEXT-002
Status: active
Topics:
    - goja
    - goja-bindings
    - sanitize
    - yaml
    - json
    - native-modules
    - text-algorithms
    - tree-sitter
DocType: design-doc
Intent: ""
Owners: []
RelatedFiles:
    - Path: ../../../../../../../sanitize/pkg/json/options.go
      Note: JSON functional options (no WithTabWidth)
    - Path: ../../../../../../../sanitize/pkg/json/sanitize.go
      Note: JSON sanitize algorithm
    - Path: ../../../../../../../sanitize/pkg/json/types.go
      Note: JSON result types (adds StrictParseClean)
    - Path: ../../../../../../../sanitize/pkg/yaml/options.go
      Note: YAML functional options (WithMaxIterations
    - Path: ../../../../../../../sanitize/pkg/yaml/sanitize.go
      Note: YAML sanitize algorithm and iterative fix loop
    - Path: ../../../../../../../sanitize/pkg/yaml/types.go
      Note: YAML result types (Result
    - Path: pkg/markdown/module.go
      Note: Reference native module implementation pattern
    - Path: pkg/xgoja/providers/text/text.go
      Note: Reference xgoja provider wrapping pattern
    - Path: xgoja.yaml
      Note: Reference xgoja build spec
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Goja Text Module Bindings: Sanitize YAML and JSON Native Module

## Design and Implementation Guide

This document is an intern-ready guide for implementing goja bindings for the `sanitize` library. The goal is to expose YAML and JSON sanitization, linting, parse-tree inspection, rule catalog enumeration, and example retrieval to JavaScript through `go-go-goja`'s native module system. The document covers the sanitize library's architecture, the goja native module pattern, the xgoja provider model, the proposed API surface, and a phased implementation plan with exact file references.

---

## Part 0: Executive Summary

The `sanitize` library at `/home/manuel/workspaces/2026-06-02/goja-text/sanitize` provides structured-text linting and heuristic fixing for YAML and JSON, powered by tree-sitter. It currently operates as a standalone Go CLI (`sanitize fix`, `sanitize lint`, `sanitize serve`) and as a Go library (`pkg/yaml`, `pkg/json`). This project adds a `require("sanitize")` native module to `goja-text` so JavaScript scripts running inside `go-go-goja` can lint and fix YAML and JSON without leaving the JavaScript runtime.

The module follows the same pattern as the existing `markdown` module in `goja-text`: a `modules.NativeModule` implementation registered through `init()`, wrapped by an xgoja provider, and composed into a generated binary via `xgoja.yaml`. The module exposes Go-backed result objects (`Result`, `LintIssue`, `ErrorNode`, `Fix`, `RuleSpec`, `Example`) so Go-side validation and error reporting remain possible when JavaScript passes values back into module functions.

---

## Part 1: The Problem Statement

### What we want

JavaScript scripts running inside a `go-go-goja` runtime should be able to:

1. Lint a YAML or JSON string and receive a list of issues with rule names, descriptions, and source positions.
2. Sanitize (lint + fix) a YAML or JSON string and receive the fixed output, the list of applied fixes, and the before/after parse state.
3. Inspect the tree-sitter parse tree of a YAML or JSON string for debugging.
4. Enumerate available lint/fix rules and their metadata (name, summary, whether it lints, whether it fixes, whether it is enabled by default).
5. Retrieve built-in examples of broken and valid YAML and JSON snippets.
6. Configure sanitizer behavior (max iterations, tab width for YAML, rule enablement/disablement) from JavaScript.

### What we have today

The `sanitize` library already provides all of the above at the Go level:

- `pkg/yaml.Sanitize(src, opts...)` returns a `yamlsanitize.Result`
- `pkg/yaml.Lint(src)` returns `[]yamlsanitize.LintIssue`
- `pkg/yaml.ParseTree(src)` returns `(treeText, errors, err)`
- `pkg/yaml.RuleCatalog()` returns `[]yamlsanitize.RuleSpec`
- `pkg/yaml.Examples` is a `[]yamlsanitize.Example` var
- Same for `pkg/json`

The library uses tree-sitter parsers for both YAML and JSON. It applies heuristics iteratively: parse, lint, fix, re-parse, re-lint, until the document is clean or no more fixes can be applied. The Go API is functional-options-based.

### What we do NOT want

- We do not want to rewrite sanitization logic in JavaScript.
- We do not want to convert Go result structs into plain `map[string]any` as the primary representation. Go-backed objects preserve type safety for functions that receive results back from JavaScript.
- We do not want to add query-specific helpers such as `extractDuplicateKeys` or `extractTrailingCommas`. The result objects are inspectable directly; JavaScript can filter arrays of `LintIssue` or `Fix` values using standard array methods.
- We do not want to expose tree-sitter's raw C API or node pointers to JavaScript. The parse tree is returned as a pre-rendered string.

---

## Part 2: The Sanitize Library — Architecture Overview

Before designing the goja module, understand the sanitize library's structure. It has two parallel packages with nearly identical shapes.

### Package structure

```text
sanitize/
  pkg/yaml/
    types.go        — ErrorNode, LintIssue, Fix, Result, Example, RuleSpec
    sanitize.go     — Sanitize(), SanitizeWithOptions(), sanitizeWithConfig()
    lint.go         — Lint(), LintWithOptions(), lintWithConfig(), lintIssuesFromAnalysis()
    parse.go        — ParseTree(), newParser(), collectErrors()
    rules.go        — ruleCatalog, RuleCatalog(), KnownRule(), LookupRule(), ValidateRuleNames()
    options.go      — Option, config, WithMaxIterations(), WithTabWidth(), WithOnlyRules(), WithDisabledRules()
    analysis.go     — documentAnalysis, analyzeDocument()
    fix.go          — applyFixes(), individual fixers
    examples.go     — Examples var
  pkg/json/
    types.go        — Same shape as yaml/types.go (no TabWidth)
    sanitize.go     — Same shape as yaml/sanitize.go
    lint.go         — Same shape as yaml/lint.go (adds strict_parse_error, multiple_top_level_values)
    parse.go        — Same shape as yaml/parse.go (adds StrictParse())
    rules.go        — Same shape as yaml/rules.go (different catalog)
    options.go      — Same shape as yaml/options.go (no WithTabWidth)
    analysis.go     — documentAnalysis, analyzeDocument()
    fix.go          — applyFixes(), individual fixers
    examples.go     — Examples var
```

### Core data types

Both packages share the same result shape:

```go
type Result struct {
    Original           string
    Sanitized          string
    TreeText           string
    OriginalTreeText   string
    Errors             []ErrorNode
    OriginalErrors     []ErrorNode
    LintIssues         []LintIssue
    OriginalLintIssues []LintIssue
    Fixes              []Fix
    ParseClean         bool
    LintClean          bool
}
```

JSON adds `StrictParseClean` and `OriginalStrictParseClean` because JSON has a strict `encoding/json` parse check that YAML does not.

```go
type ErrorNode struct {
    Type      string // "ERROR" or "MISSING"
    StartByte uint
    EndByte   uint
    StartRow  uint
    StartCol  uint
    EndRow    uint
    EndCol    uint
    Text      string
}

type LintIssue struct {
    Rule        string
    Source      string // "parse", "heuristic", "strict-parser"
    Description string
    StartByte   uint
    EndByte     uint
    StartRow    uint
    StartCol    uint
    EndRow      uint
    EndCol      uint
    Row         int    // 0-indexed alias for StartRow
}

type Fix struct {
    Rule        string
    Description string
    Before      string
    After       string
}

type RuleSpec struct {
    Name           string
    Summary        string
    Lints          bool
    Fixes          bool
    DefaultEnabled bool
    // JSON only:
    ParseAware     bool
}

type Example struct {
    Name        string
    Description string
    YAML        string // or JSON in json package
    Category    string
    Source      string
    Filename    string
}
```

### The sanitize algorithm

Both packages follow the same iterative pattern:

```text
sanitize(src, config):
    original = src
    allFixes = []

    capture original parse state (tree, errors, lint issues)

    for iter = 0 to maxIterations:
        doc = analyzeDocument(src)
        errors = doc.ParseErrors
        lintIssues = lint(doc)

        if len(errors) == 0 and len(lintIssues) == 0:
            return Result with ParseClean=true, LintClean=true

        fixed, fixes = applyFixes(src, doc, config)
        if len(fixes) == 0:
            return Result with current state (no more progress)

        allFixes.append(fixes)
        src = fixed

    return Result with final state
```

The algorithm is conservative: it only applies fixes that make measurable progress. If a fix round produces no changes, the loop stops.

### Options system

Both packages use a functional options pattern:

```go
type Option func(*config)

// YAML and JSON both support:
func WithMaxIterations(n int) Option
func WithOnlyRules(rules ...string) Option
func WithDisabledRules(rules ...string) Option

// YAML only:
func WithTabWidth(w int) Option
```

Options are validated at build time: unknown rule names are rejected, and a rule cannot be both enabled and disabled.

---

## Part 3: The goja Native Module Pattern — How Sanitize Fits In

The module pattern is identical to the `markdown` module already implemented in `goja-text`. This section summarizes the pattern for an intern who has not yet read the markdown design doc in detail.

### The NativeModule interface

```go
type NativeModule interface {
    Name() string
    Doc() string
    Loader(*goja.Runtime, *goja.Object)
}
```

A native module:

1. Implements `NativeModule`.
2. Calls `modules.Register(&module{})` in `init()`.
3. In `Loader`, reads `exports := moduleObj.Get("exports").(*goja.Object)` and sets properties on it.
4. Optionally implements `modules.TypeScriptDeclarer` for TypeScript declarations.

When `require("sanitize")` is called from JavaScript, goja finds the registered module, creates a new module object, and calls `Loader`. The exports set in `Loader` become the value returned by `require("sanitize")`.

### Go struct projection

goja automatically projects Go struct exported fields as JavaScript object properties. The `sanitize` module's `Result`, `LintIssue`, `ErrorNode`, `Fix`, `RuleSpec`, and `Example` structs will all be visible in JavaScript with PascalCase field names:

```js
const result = sanitize.yaml.sanitize("broken:yaml\n");
console.log(result.Sanitized);        // fixed YAML string
console.log(result.ParseClean);       // boolean
console.log(result.LintIssues[0].Rule);      // "missing_space_after_colon"
console.log(result.LintIssues[0].Description);
console.log(result.Fixes[0].Before);
console.log(result.Fixes[0].After);
```

This is the same design decision as the markdown module: Go-backed objects preserve runtime type information and enable Go-side validation.

### Registration paths

There are three ways a module becomes available in a goja runtime:

1. **Blank import + global registry**: The module package calls `modules.Register()` in `init()`. A program that blank-imports the package gets the module in `modules.DefaultRegistry`. The go-go-goja engine builder can then load it via `MiddlewareOnly("sanitize")` or similar.
2. **xgoja provider**: The module is wrapped in a `providerapi.Module` and registered through an xgoja provider package. This is the path used for generated binaries.
3. **RuntimeModuleSpec**: A runtime factory can be configured with explicit module specs at construction time.

For `goja-text`, we use Path 1 for unit tests and Path 2 for the generated xgoja binary.

---

## Part 4: The xgoja Build System

xgoja builds a custom binary from a declarative `xgoja.yaml` spec. The spec declares provider packages, enabled modules, runtime configuration, and available commands.

The existing `goja-text/xgoja.yaml` already includes:

- `goja-text` provider (with `replace: .` for local resolution)
- `go-go-goja-core` provider (path, yaml modules)
- `go-go-goja-host` provider (fs module with `allow: true`)
- Commands: eval, run, repl

The new sanitize module will be added to the `goja-text` provider. No new provider package is needed unless the module is split into a separate package.

---

## Part 5: Decision Records

### Decision 1: Expose YAML and JSON through a single `sanitize` module

- **Context:** The sanitize library has two distinct Go packages (`pkg/yaml` and `pkg/json`). From JavaScript, we could expose `require("yaml")` and `require("json")` as separate modules, or expose a single `require("sanitize")` module with nested namespaces.
- **Options considered:**
  - Two separate modules: `require("yaml")` and `require("json")`. This maps 1:1 with Go packages but pollutes the global module namespace.
  - Single module with `sanitize.yaml.*` and `sanitize.json.*` namespaces. This keeps the module namespace clean and groups related functionality.
  - Single module with format-dispatched functions like `sanitize.sanitize(input, {format: "yaml"})`. This is more compact but loses the natural grouping of format-specific options.
- **Decision:** Single `sanitize` module with `sanitize.yaml` and `sanitize.json` namespaces.
- **Rationale:** The sanitize library is conceptually one tool that operates on two formats. Grouping under one module keeps the global namespace clean. The nested namespace pattern is familiar from Node.js built-ins and makes the API self-documenting.
- **Consequences:** The module loader must create two sub-objects (`yaml` and `json`) on the exports object. Each sub-object gets its own set of functions.
- **Status:** accepted

### Decision 2: Keep result objects as Go-backed structs

- **Context:** The sanitize library returns rich structs (`Result`, `LintIssue`, etc.). JavaScript will inspect these objects and may pass them back into Go functions.
- **Options considered:**
  - Plain `map[string]any` with lowercase keys. This is more idiomatic for JavaScript but loses type safety.
  - JSON-serialized strings. This would require JavaScript to `JSON.parse` everything, adding overhead and losing object identity.
  - Go-backed structs with PascalCase field names. This preserves type identity and enables Go-side validation.
- **Decision:** Go-backed structs projected by goja reflection.
- **Rationale:** Same reasoning as the markdown module. The result objects may be passed back into Go functions. Go-backed objects let those functions validate inputs and report precise errors. The `Result` struct is a domain object, not merely a data transfer object.
- **Consequences:** JavaScript uses `result.Sanitized`, `issue.Rule`, `fix.Before`, etc. If lowercase JSON-style objects are needed later, an explicit adapter can be added.
- **Status:** accepted

### Decision 3: Use Go-backed builder/config objects instead of raw options objects

- **Context:** The sanitize library uses Go functional options (`WithMaxIterations`, `WithOnlyRules`, etc.). A plain JavaScript options object is convenient, but it makes unknown-option handling, rule-name validation, cross-field validation, and future runtime validation policies harder to control consistently. The user explicitly wants unknown-option behavior to be controllable and wants the Go side to provide more complex validation rules at runtime.
- **Options considered:**
  - Positional parameters such as `sanitize.yaml.sanitize(input, maxIterations, tabWidth, onlyRules, disabledRules)`. This is brittle and hard to extend.
  - Plain options objects such as `sanitize.yaml.sanitize(input, {maxIterations: 5})`. This is familiar JavaScript, but every function must re-decode and re-validate the object. Unknown-key policy becomes an ad hoc per-call concern.
  - Go-backed builder/config objects such as `sanitize.yaml.options().MaxIterations(5).TabWidth(4).RejectUnknownOptions().Build()`. This gives Go a durable place to enforce validation, preserve policy, and add richer runtime checks.
- **Decision:** Expose Go-backed builder/config objects as the primary configuration API. Keep direct calls with no options for defaults. Do not make raw options objects the primary Phase 1 API.
- **Rationale:** A builder object lets Go own validation semantics. It can reject unknown options, allow unknown options, collect unknown options for diagnostics, validate rule-name overlap, normalize rule arrays, and later support more complex policies without changing every sanitize/lint function signature. This follows the goja project preference for Go-backed domain objects when runtime validation matters.
- **Consequences:** JavaScript callers use PascalCase builder methods because the builder is Go-backed: `MaxIterations`, `TabWidth`, `OnlyRules`, `DisabledRules`, `RejectUnknownOptions`, `AllowUnknownOptions`, `CollectUnknownOptions`, `Build`, and `Validate`. Tests must pin those method names. If a lowerCamel convenience wrapper is desired later, add it deliberately as a JS adapter.
- **Status:** accepted

### Decision 5: Pin sanitize to `v0.0.2` without a local replace

- **Context:** The sanitize repository is present locally for source inspection, and the module tag `github.com/go-go-golems/sanitize@v0.0.2` resolves as a published dependency. The local checkout is reference material for this task, not a library we plan to modify.
- **Options considered:** Use an unversioned local-only import, use a pseudo-version from the local checkout, require `v0.0.2` with a local replace, or require only the published pinned `v0.0.2` version.
- **Decision:** Require `github.com/go-go-golems/sanitize v0.0.2` without a local replace.
- **Rationale:** The pinned module version records the dependency boundary and keeps xgoja generated builds reproducible outside this workspace. `go.work` is enough for local workspace awareness when needed, but the goja-text module should not depend on a local sanitize checkout if we are not editing sanitize.
- **Consequences:** `go mod tidy` should add `github.com/go-go-golems/sanitize v0.0.2` and its transitive dependencies. xgoja generated builds should resolve sanitize from the module proxy/GitHub. The only local replacement still required during development is the existing xgoja replacement for the local `go-go-goja` checkout, because this workspace is actively using that local basis.
- **Status:** accepted

### Decision 4: Return both original and sanitized state in Result

- **Context:** The sanitize library's `Result` struct includes both `OriginalErrors`/`OriginalLintIssues` and post-fix `Errors`/`LintIssues`, plus `OriginalTreeText` and `TreeText`.
- **Options considered:**
  - Return only the final state. This is simpler but loses the ability to compare before/after in JavaScript.
  - Return the full `Result` struct as-is. This is what the library already does.
- **Decision:** Return the full `Result` struct, including original and final state.
- **Rationale:** The comparison between original and final state is a core use case for the sanitize library. JavaScript callers will want to show before/after diffs. The library already captures this information; there is no benefit to stripping it.
- **Consequences:** The `Result` object is larger. For very large inputs with many errors, this could use more memory. This is acceptable for the intended use case.
- **Status:** accepted

---

## Part 6: Proposed Module API

### Module structure

```js
const sanitize = require("sanitize");

// Namespaces
sanitize.yaml   // YAML operations
sanitize.json   // JSON operations
```

### YAML API

```js
// Sanitize YAML (lint + fix)
const result = sanitize.yaml.sanitize(input, options);

// Lint YAML only
const issues = sanitize.yaml.lint(input, options);

// Parse tree inspection
const tree = sanitize.yaml.parseTree(input);

// Rule catalog
const rules = sanitize.yaml.rules();

// Built-in examples
const examples = sanitize.yaml.examples();
```

### JSON API

```js
// Sanitize JSON (lint + fix)
const result = sanitize.json.sanitize(input, options);

// Lint JSON only
const issues = sanitize.json.lint(input, options);

// Parse tree inspection
const tree = sanitize.json.parseTree(input);

// Rule catalog
const rules = sanitize.json.rules();

// Built-in examples
const examples = sanitize.json.examples();
```

### Builder/config API shape

Both YAML and JSON expose a Go-backed options builder. Direct calls without a config use the sanitize library defaults. Calls with a config require a value produced by `Build()` so the Go side can validate and normalize options before they are used.

```js
const yamlConfig = sanitize.yaml.options()
  .MaxIterations(5)
  .TabWidth(4)
  .OnlyRules("tab_indent", "missing_space_after_colon")
  .RejectUnknownOptions()
  .Build();

const result = sanitize.yaml.sanitize(input, yamlConfig);
```

JSON uses the same pattern without `TabWidth`:

```js
const jsonConfig = sanitize.json.options()
  .MaxIterations(3)
  .DisabledRules("duplicate_key")
  .CollectUnknownOptions()
  .Build();

const result = sanitize.json.sanitize(input, jsonConfig);
```

The builder methods are intentionally PascalCase because the builder is a Go-backed object projected through goja:

- `MaxIterations(n)` — positive integer, default 10
- `TabWidth(n)` — YAML only; positive integer, default 2
- `OnlyRules(...rules)` — restrict lint/fix to known rule names
- `DisabledRules(...rules)` — disable known rule names
- `RejectUnknownOptions()` — reject unknown keys when importing raw option objects; recommended default
- `AllowUnknownOptions()` — ignore unknown keys when importing raw option objects
- `CollectUnknownOptions()` — record unknown keys for diagnostics without failing immediately
- `FromObject(obj)` — optional bridge for callers that receive plain JS options dynamically
- `Validate()` — return a Go-backed validation result without building
- `Build()` — return an immutable Go-backed config object or throw a validation error

The initial implementation should support `Build()`-produced config objects as the primary path. `FromObject` can be implemented in Phase 1 if dynamic option import is needed for tests; otherwise it can be added in Phase 2. The important design rule is that all raw object decoding goes through the builder so unknown-option policy and cross-field validation are centralized.

### Result object shape

The `sanitize` function returns a `Result` object:

```js
{
  Original: string,
  Sanitized: string,
  TreeText: string,
  OriginalTreeText: string,
  Errors: ErrorNode[],
  OriginalErrors: ErrorNode[],
  LintIssues: LintIssue[],
  OriginalLintIssues: LintIssue[],
  Fixes: Fix[],
  ParseClean: boolean,
  LintClean: boolean,
  // JSON only:
  StrictParseClean: boolean,
  OriginalStrictParseClean: boolean
}
```

### ErrorNode shape

```js
{
  Type: string,        // "ERROR" or "MISSING"
  StartByte: number,
  EndByte: number,
  StartRow: number,
  StartCol: number,
  EndRow: number,
  EndCol: number,
  Text: string
}
```

### LintIssue shape

```js
{
  Rule: string,
  Source: string,      // "parse", "heuristic", "strict-parser"
  Description: string,
  StartByte: number,
  EndByte: number,
  StartRow: number,
  StartCol: number,
  EndRow: number,
  EndCol: number,
  Row: number
}
```

### Fix shape

```js
{
  Rule: string,
  Description: string,
  Before: string,
  After: string
}
```

### RuleSpec shape

```js
{
  Name: string,
  Summary: string,
  Lints: boolean,
  Fixes: boolean,
  DefaultEnabled: boolean,
  // JSON only:
  ParseAware: boolean
}
```

### Example shape

```js
{
  Name: string,
  Description: string,
  YAML: string,   // or JSON in json examples
  Category: string,
  Source: string,
  Filename: string
}
```

### parseTree return shape

`parseTree` returns an object:

```js
{
  treeText: string,
  errors: ErrorNode[]
}
```

---

## Part 7: Implementation Plan

### Phase 1: Core sanitize module

Create `pkg/sanitize/` with the native module implementation.

**Step 1.1: Create `pkg/sanitize/types.go`**

Re-export or alias the types from `sanitize/pkg/yaml` and `sanitize/pkg/json`. The module will work with:

- `yamlsanitize.Result`, `yamlsanitize.LintIssue`, etc.
- `jsonsanitize.Result`, `jsonsanitize.LintIssue`, etc.

No new types are needed unless we want to unify the two result shapes. For the first implementation, keep them separate: JavaScript sees `sanitize.yaml.sanitize()` returning a YAML `Result` and `sanitize.json.sanitize()` returning a JSON `Result`.

**Step 1.2: Create `pkg/sanitize/module.go`**

Implement the `modules.NativeModule` interface:

```go
package sanitize

import (
    "fmt"

    "github.com/dop251/goja"
    "github.com/go-go-golems/go-go-goja/modules"
    "github.com/go-go-golems/go-go-goja/pkg/tsgen/spec"
    yamlsanitize "github.com/go-go-golems/sanitize/pkg/yaml"
    jsonsanitize "github.com/go-go-golems/sanitize/pkg/json"
)

type module struct{}

func (module) Name() string { return "sanitize" }

func (module) Doc() string { /* ... */ }

func (module) TypeScriptModule() *spec.Module { /* ... */ }

func (mod module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)

    // Create yaml sub-namespace
    yamlObj := vm.NewObject()
    modules.SetExport(yamlObj, "sanitize", "yaml.sanitize", func(input string, opts map[string]any) (*yamlsanitize.Result, error) {
        // decode opts into yamlsanitize.Option values
        // call yamlsanitize.SanitizeWithOptions(input, opts...)
    })
    // ... yaml.lint, yaml.parseTree, yaml.rules, yaml.examples

    // Create json sub-namespace
    jsonObj := vm.NewObject()
    // ... same pattern for jsonsanitize

    exports.Set("yaml", yamlObj)
    exports.Set("json", jsonObj)
}

func init() {
    modules.Register(&module{})
}
```

**Step 1.3: Implement builder/config objects**

Create Go-backed builders and immutable config values. The builder owns all JavaScript-facing validation policy. Sanitizer functions accept either no config or a built config value. They should not decode arbitrary options objects independently.

Core types:

```go
type UnknownOptionPolicy string

const (
    UnknownOptionReject  UnknownOptionPolicy = "reject"
    UnknownOptionAllow   UnknownOptionPolicy = "allow"
    UnknownOptionCollect UnknownOptionPolicy = "collect"
)

type ValidationResult struct {
    Valid   bool
    Errors  []string
    Unknown []string
}

type YamlConfig struct {
    MaxIterations int
    TabWidth      int
    OnlyRules     []string
    DisabledRules []string
    UnknownPolicy UnknownOptionPolicy
    Unknown       []string
}

type JsonConfig struct {
    MaxIterations int
    OnlyRules     []string
    DisabledRules []string
    UnknownPolicy UnknownOptionPolicy
    Unknown       []string
}
```

Builder methods should be chainable and Go-backed:

```go
func (b *YamlOptionsBuilder) MaxIterations(n int) *YamlOptionsBuilder
func (b *YamlOptionsBuilder) TabWidth(n int) *YamlOptionsBuilder
func (b *YamlOptionsBuilder) OnlyRules(rules ...string) *YamlOptionsBuilder
func (b *YamlOptionsBuilder) DisabledRules(rules ...string) *YamlOptionsBuilder
func (b *YamlOptionsBuilder) RejectUnknownOptions() *YamlOptionsBuilder
func (b *YamlOptionsBuilder) AllowUnknownOptions() *YamlOptionsBuilder
func (b *YamlOptionsBuilder) CollectUnknownOptions() *YamlOptionsBuilder
func (b *YamlOptionsBuilder) Validate() ValidationResult
func (b *YamlOptionsBuilder) Build() (*YamlConfig, error)
func (c *YamlConfig) Options() []yamlsanitize.Option
```

The JSON builder mirrors this shape without `TabWidth`. `Build()` should validate positive integer settings, unknown option policy, rule names, and overlap between `OnlyRules` and `DisabledRules`. It should call the sanitize library's `ValidateRuleNames` by constructing the underlying options through `SanitizeWithOptions`/`LintWithOptions` or by directly using the library's validation helpers where available.

**Step 1.4: Export functions**

For each namespace, implement:

- `sanitize(input string, config *YamlConfig|*JsonConfig)` — calls `SanitizeWithOptions` with options produced by the built config
- `lint(input string, config *YamlConfig|*JsonConfig)` — calls `LintWithOptions` with options produced by the built config
- `options()` — returns a Go-backed options builder for the format
- `parseTree(input string) (*ParseTreeResult, error)` — calls `ParseTree`
- `rules() ([]RuleSpec, error)` — calls `RuleCatalog`
- `examples() ([]Example, error)` — returns the built-in `Examples` var

The `parseTree` function wraps the two return values plus error into a single object:

```go
type ParseTreeResult struct {
    TreeText string
    Errors   []ErrorNode
}
```

**Step 1.5: TypeScript declarations**

Implement `TypeScriptModule()` with `RawDTS` entries for all interfaces and function signatures. The declarations should cover:

- `SanitizeOptions`
- `YamlSanitizeOptions` (extends SanitizeOptions with `tabWidth`)
- `Result`
- `ErrorNode`
- `LintIssue`
- `Fix`
- `RuleSpec`
- `Example`
- `ParseTreeResult`
- Functions for `yaml.*` and `json.*`

### Phase 2: Tests

**Step 2.1: Go-side tests in `pkg/sanitize/sanitize_test.go`**

Test that `Sanitize`, `Lint`, `ParseTree`, `Rules`, and `Examples` work correctly when called from Go through the module's internal functions.

**Step 2.2: JavaScript runtime tests in `pkg/sanitize/module_test.go`**

Test that:

- `require("sanitize")` loads successfully
- `sanitize.yaml.sanitize("broken:yaml\n")` returns a `Result` with `Sanitized`, `ParseClean`, `LintClean`
- `sanitize.yaml.lint("broken:yaml\n")` returns `[]LintIssue`
- `sanitize.yaml.rules()` returns `[]RuleSpec`
- `sanitize.json.sanitize("{'bad': 'json'}\n")` returns a `Result`
- Options objects are decoded correctly (`maxIterations`, `tabWidth`, `onlyRules`, `disabledRules`)
- Invalid options produce errors

**Step 2.3: Edge-field regression probes**

Test that the important fields on `Result`, `LintIssue`, `ErrorNode`, and `Fix` are accessible from JavaScript:

- `result.Sanitized`, `result.ParseClean`, `result.LintClean`
- `result.Errors[0].Type`, `result.Errors[0].StartRow`, `result.Errors[0].Text`
- `result.LintIssues[0].Rule`, `result.LintIssues[0].Description`
- `result.Fixes[0].Rule`, `result.Fixes[0].Before`, `result.Fixes[0].After`
- `rules[0].Name`, `rules[0].Summary`, `rules[0].Lints`, `rules[0].Fixes`

### Phase 3: xgoja provider integration

**Step 3.1: Update `pkg/xgoja/providers/text/text.go`**

Add `"sanitize"` to `textModuleNames`:

```go
var textModuleNames = []string{
    "markdown",
    "sanitize",
}
```

The provider already wraps modules from `modules.GetModule(name)`. No other changes are needed.

**Step 3.2: Blank-import in provider package**

Add a blank import of `pkg/sanitize` in `pkg/xgoja/providers/text/text.go` to trigger `init()` registration:

```go
import (
    _ "github.com/go-go-golems/goja-text/pkg/sanitize"
    // existing imports...
)
```

**Step 3.3: Update `xgoja.yaml`**

Add the `sanitize` module to the runtime modules list:

```yaml
runtimes:
  main:
    modules:
      - package: goja-text
        name: markdown
        as: markdown
      - package: goja-text
        name: sanitize
        as: sanitize
      # ... existing modules
```

**Step 3.4: Add demo script**

Create `examples/js/sanitize-demo.js`:

```js
const fs = require("fs");
const sanitize = require("sanitize");

// Read a broken YAML file
const yamlSource = fs.readFileSync("examples/yaml/broken.yaml", "utf-8");
const yamlResult = sanitize.yaml.sanitize(yamlSource);
console.log("YAML sanitized:", yamlResult.Sanitized);
console.log("YAML parse clean:", yamlResult.ParseClean);
console.log("YAML lint clean:", yamlResult.LintClean);
console.log("YAML fixes:", yamlResult.Fixes.length);

// Read a broken JSON file
const jsonSource = fs.readFileSync("examples/json/broken.json", "utf-8");
const jsonResult = sanitize.json.sanitize(jsonSource);
console.log("JSON sanitized:", jsonResult.Sanitized);
console.log("JSON parse clean:", jsonResult.ParseClean);
console.log("JSON strict parse clean:", jsonResult.StrictParseClean);

// List rules
const yamlRules = sanitize.yaml.rules();
console.log("YAML rules:", yamlRules.map(r => r.Name));

const jsonRules = sanitize.json.rules();
console.log("JSON rules:", jsonRules.map(r => r.Name));
```

**Step 3.5: Build and validate**

Build the xgoja binary:

```bash
go run ../go-go-goja/cmd/xgoja build \
  -f xgoja.yaml \
  --xgoja-replace /home/manuel/workspaces/2026-06-02/goja-text/go-go-goja
```

Run smoke tests:

```bash
./dist/goja-text eval 'const s = require("sanitize"); JSON.stringify(s.yaml.rules().map(r => r.Name))'
./dist/goja-text run examples/js/sanitize-demo.js
```

### Phase 4: Documentation

**Step 4.1: Update `goja-text/README.md`**

Add a section for the sanitize module alongside the markdown module.

**Step 4.2: Update ticket docs**

- Add diary entries for each phase
- Update changelog
- Relate modified files

---

## Part 8: File Layout

```text
goja-text/
  pkg/
    markdown/              # existing
    sanitize/
      types.go             # ParseTreeResult (if needed)
      module.go            # NativeModule implementation
      module_test.go       # JavaScript runtime tests
      sanitize_test.go     # Go-side tests
    xgoja/providers/text/
      text.go              # updated to include "sanitize"
  examples/
    js/
      markdown-demo.js     # existing
      sanitize-demo.js     # new
    yaml/
      broken.yaml          # new test fixture
    json/
      broken.json          # new test fixture
  xgoja.yaml               # updated
  README.md                # updated
```

---

## Part 9: Key Source References

### Sanitize library

- `sanitize/pkg/yaml/types.go` — `Result`, `LintIssue`, `ErrorNode`, `Fix`, `RuleSpec`, `Example`
- `sanitize/pkg/yaml/sanitize.go` — `Sanitize()`, `SanitizeWithOptions()`, iterative fix loop
- `sanitize/pkg/yaml/lint.go` — `Lint()`, `LintWithOptions()`, regex-based line linting
- `sanitize/pkg/yaml/parse.go` — `ParseTree()`, tree-sitter parser setup
- `sanitize/pkg/yaml/rules.go` — `RuleCatalog()`, `KnownRule()`, `LookupRule()`
- `sanitize/pkg/yaml/options.go` — `Option`, `config`, `WithMaxIterations()`, `WithTabWidth()`, `WithOnlyRules()`, `WithDisabledRules()`
- `sanitize/pkg/yaml/examples.go` — `Examples` var
- `sanitize/pkg/json/types.go` — Same shape as yaml (adds `StrictParseClean`)
- `sanitize/pkg/json/sanitize.go` — Same shape as yaml
- `sanitize/pkg/json/lint.go` — Same shape as yaml (adds `strict_parse_error`, `multiple_top_level_values`)
- `sanitize/pkg/json/parse.go` — Same shape as yaml (adds `StrictParse()`)
- `sanitize/pkg/json/rules.go` — JSON rule catalog
- `sanitize/pkg/json/options.go` — Same shape as yaml (no `WithTabWidth`)
- `sanitize/pkg/json/examples.go` — `Examples` var

### goja-text existing infrastructure

- `goja-text/pkg/markdown/module.go` — Reference native module implementation
- `goja-text/pkg/xgoja/providers/text/text.go` — Reference xgoja provider wrapping
- `goja-text/xgoja.yaml` — Reference build spec
- `goja-text/go.mod` — Module dependencies

### go-go-goja framework

- `go-go-goja/modules/common.go` — `NativeModule`, `Registry`, `Register()`
- `go-go-goja/modules/exports.go` — `SetExport()`
- `go-go-goja/modules/typing.go` — `TypeScriptDeclarer`
- `go-go-goja/pkg/xgoja/providerapi/registry.go` — `Registry`, `Package()`
- `go-go-goja/pkg/xgoja/providerapi/module.go` — `Module`, `Entry`, `New()`
- `go-go-goja/pkg/xgoja/providers/core/core.go` — Reference provider implementation

---

## Part 10: Testing Strategy

### Unit tests (Go side)

1. **Sanitize function tests**: Call the module's internal sanitize functions directly with known broken inputs and verify the `Result` struct fields.
2. **Options decoding tests**: Pass various JS options objects and verify the correct Go `Option` values are produced.
3. **Error handling tests**: Pass invalid options (unknown rules, conflicting only/disabled) and verify errors are returned.

### Runtime tests (JavaScript side)

1. **Module loading**: `require("sanitize")` loads without error.
2. **YAML sanitize**: Broken YAML input produces a `Result` with `ParseClean == false`, `LintClean == false`, and non-empty `Fixes`.
3. **JSON sanitize**: Broken JSON input produces a `Result` with appropriate error state.
4. **Lint only**: `lint()` returns `[]LintIssue` without applying fixes.
5. **Parse tree**: `parseTree()` returns a string and error array.
6. **Rules**: `rules()` returns the full catalog with correct `Name`, `Summary`, `Lints`, `Fixes`, `DefaultEnabled`.
7. **Examples**: `examples()` returns non-empty arrays.
8. **Options**: `sanitize()` with `maxIterations: 1` stops after one iteration. `sanitize()` with `onlyRules: ["tab_indent"]` only reports tab issues.
9. **Field access**: All important fields are accessible from JavaScript with PascalCase names.

### Integration tests (xgoja binary)

1. Build the binary with `xgoja build`.
2. Run `eval` with a one-liner that loads `sanitize` and calls `yaml.sanitize()`.
3. Run the demo script that reads files from disk.
4. Verify both YAML and JSON operations work end-to-end.

---

## Part 11: Risks, Alternatives, and Open Questions

### Risk 1: Tree-sitter parser dependencies

The sanitize library depends on tree-sitter grammars for YAML and JSON. These are C libraries with Go bindings. When the module is loaded in goja, the tree-sitter parsers are initialized. This should work in the same process, but the intern should verify that the generated xgoja binary can load and use tree-sitter without issues.

### Risk 2: Builder/config API evolution

The sanitize library may add new options in the future. A Go-backed builder gives us a controlled extension point: add a method, add validation, and keep existing built config objects stable. Raw JavaScript object import, if supported via `FromObject`, must route through the builder's unknown-option policy (`reject`, `allow`, or `collect`) so typos and future option additions are handled deliberately rather than silently.

### Risk 3: Result object size

For very large inputs with many errors, the `Result` object includes both original and final state, plus tree text. This could be large. The current design returns the full result. If memory becomes an issue, a future version could add an option to omit `TreeText` and `OriginalTreeText`.

### Risk 4: Rule name stability

JavaScript code may reference rule names in `onlyRules` and `disabledRules`. If the sanitize library renames a rule, JavaScript code using the old name will get a validation error. This is a coupling between the Go library and JavaScript callers. The module should pass through the Go library's `ValidateRuleNames` error so JavaScript callers see a clear message.

### Alternative 1: Separate modules for YAML and JSON

Instead of `require("sanitize")` with `yaml` and `json` sub-objects, we could have `require("yaml")` and `require("json")` as separate modules. This was rejected because it pollutes the global module namespace and the two formats share the same conceptual tool.

### Alternative 2: Unified result type

We could define a single `Result` type that covers both YAML and JSON, with optional fields for JSON-specific values. This was rejected because it adds complexity and the two formats already have distinct Go types in the sanitize library. Keeping them separate makes the mapping straightforward.

### Open question 1: Should `parseTree` return a structured tree?

Currently, `parseTree` returns a pre-rendered string (`treeText`). The sanitize library does not expose the raw tree-sitter node tree as a traversable structure. If JavaScript callers need programmatic tree access, a future enhancement could walk the tree-sitter AST and convert it into a JavaScript-accessible node tree. This is out of scope for Phase 1.

### Open question 2: Should the module expose `StrictParse` for JSON?

The sanitize library has `jsonsanitize.StrictParse(src)` which validates with `encoding/json`. The module exposes this indirectly through `result.StrictParseClean`. A dedicated `sanitize.json.strictParse(input)` function could be useful for callers who only want validation, not fixing. This is a candidate for Phase 2.

### Open question 3: Streaming or chunked processing

The current API processes the entire input as a single string. For very large YAML or JSON files, this could be memory-intensive. Streaming sanitization is not supported by the underlying sanitize library and is out of scope.

---

## Part 12: Decision Record Summary

| Decision | Status | Key consequence |
| --- | --- | --- |
| Single `sanitize` module with `yaml`/`json` namespaces | accepted | Cleaner global namespace, mirrors conceptual tool |
| Go-backed result structs with PascalCase fields | accepted | Type safety for Go validation, consistent with markdown module |
| Go-backed builder/config objects mirror Go functional options | accepted | Go owns unknown-option policy and complex runtime validation |
| Full Result with original + final state | accepted | Enables before/after comparison in JavaScript |
| `parseTree` returns string + errors, not structured tree | accepted | Simplest mapping of existing library API |

---

## Part 13: Implementation Checklist

### Phase 1: Core module

- [ ] Phase 0: Add pinned sanitize dependency (`v0.0.2`) to `go.mod` without a local replace
- [ ] Create `pkg/sanitize/types.go` with `ParseTreeResult`, `ValidationResult`, config structs, and unknown-option policy types
- [ ] Create `pkg/sanitize/options.go` with YAML/JSON builder implementations and validation
- [ ] Create `pkg/sanitize/module.go` with `NativeModule` implementation and namespace wiring
- [ ] Implement `yaml.sanitize`, `yaml.lint`, `yaml.parseTree`, `yaml.rules`, `yaml.examples`
- [ ] Implement `json.sanitize`, `json.lint`, `json.parseTree`, `json.rules`, `json.examples`
- [ ] Implement `TypeScriptModule()` with declarations
- [ ] Add blank import in `pkg/xgoja/providers/text/text.go`
- [ ] Update `textModuleNames` to include `"sanitize"`

### Phase 2: Tests

- [ ] Go-side tests in `pkg/sanitize/sanitize_test.go`
- [ ] JavaScript runtime tests in `pkg/sanitize/module_test.go`
- [ ] Edge-field regression probes (Result, LintIssue, ErrorNode, Fix, RuleSpec, Example)
- [ ] Builder/config validation tests, including unknown-option policy tests
- [ ] Error handling tests

### Phase 3: Integration

- [ ] Update `xgoja.yaml` to include `sanitize` module
- [ ] Create `examples/js/sanitize-demo.js`
- [ ] Create test fixtures (`examples/yaml/broken.yaml`, `examples/json/broken.json`)
- [ ] Build xgoja binary
- [ ] Run smoke tests (`eval`, `run`)

### Phase 4: Documentation

- [ ] Update `goja-text/README.md`
- [ ] Update ticket diary
- [ ] Update ticket changelog
- [ ] Relate modified files
- [ ] Upload to reMarkable
