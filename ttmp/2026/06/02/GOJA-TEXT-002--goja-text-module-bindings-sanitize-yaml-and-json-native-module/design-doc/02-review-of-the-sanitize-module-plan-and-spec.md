---
Title: Review of the Sanitize Module Plan and Spec
Ticket: GOJA-TEXT-002
Status: active
Topics:
    - goja
    - goja-bindings
    - sanitize
    - yaml
    - json
    - native-modules
    - tree-sitter
DocType: design-doc
Intent: Technical review and mentoring feedback for the intern-facing sanitize module plan
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-goja/modules/exports.go
      Note: SetExport implementation proves dotted names are literal properties
    - Path: ../../../../../../../go-go-goja/modules/yaml/yaml.go
      Note: Reference option decoding and unknown-option rejection pattern
    - Path: ../../../../../../../sanitize/pkg/json/options.go
      Note: |-
        JSON option validation behavior that the JS bridge must preserve
        JSON option validation and rule-name behavior
    - Path: ../../../../../../../sanitize/pkg/json/parse.go
      Note: |-
        Strict JSON parse behavior that affects API design
        StrictParse API and JSON strict-parse behavior
    - Path: ../../../../../../../sanitize/pkg/yaml/options.go
      Note: |-
        YAML option validation behavior that the JS bridge must preserve
        YAML option validation and rule-name behavior
    - Path: design-doc/01-sanitize-native-module-design-and-implementation-guide.md
      Note: Primary plan under review
    - Path: go.mod
      Note: Dependency and local replace wiring needs for sanitize
    - Path: pkg/markdown/module.go
      Note: Existing goja-text NativeModule implementation used as comparison baseline
    - Path: pkg/xgoja/providers/text/text.go
      Note: |-
        Existing xgoja provider that the sanitize module must extend
        Provider blank-import and module registration changes
    - Path: ttmp/2026/06/02/GOJA-TEXT-002--goja-text-module-bindings-sanitize-yaml-and-json-native-module/design-doc/01-sanitize-native-module-design-and-implementation-guide.md
      Note: Primary plan reviewed
    - Path: xgoja.yaml
      Note: Generated binary composition file that must include sanitize
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Review of the Sanitize Module Plan and Spec

This document reviews the GOJA-TEXT-002 sanitize module plan as if it were submitted by an intern for implementation approval. The goal is not to rewrite the plan. The goal is to teach how to evaluate a design document before code is written: which parts are solid, which parts are underspecified, which parts are likely to fail at compile time, and which questions should have been answered by reading the surrounding code more carefully.

The plan is strong in its overall direction. It correctly identifies the sanitize library as a Go library with two parallel domains, YAML and JSON, and it correctly proposes exposing those domains through a single `require("sanitize")` module. It also carries forward the most important design lesson from the Markdown module: domain values should remain Go-backed objects when Go functions will validate or consume them later.

The plan needs tightening before implementation. The largest gaps are not conceptual. They are integration details: dependency wiring, exact goja function signatures, options decoding, nested export mechanics, TypeScript declaration shape, and xgoja build validation. These are the details that usually separate a readable plan from an implementation-ready spec.

---

## 1. Summary Assessment

### What the plan gets right

The plan makes the right high-level architectural choice: expose `sanitize` as one module with two namespaces, `sanitize.yaml` and `sanitize.json`. That matches the user-facing concept. A caller wants one structured-text repair module, not unrelated top-level modules named `yaml` and `json`. It also avoids colliding conceptually with the existing `go-go-goja` core `yaml` module, which already parses and stringifies YAML.

The plan also correctly preserves Go-backed result objects. This is important. The sanitize library's results are not throwaway dictionaries. They are typed reports containing original input, sanitized output, parse errors, lint issues, applied fixes, rule names, and parser status. Keeping these as Go values projected into JavaScript gives the runtime a clear object identity and allows future Go functions to accept these values back without relying on map-shape guesses.

The plan includes useful decision records. This is a good habit. The most valuable records are the ones for module shape, Go-backed objects, options mapping, and full result return. These choices affect public API compatibility and should not be hidden inside prose.

The plan reads enough of the sanitize library to explain the iterative repair loop accurately. The core model is correct: parse, lint, apply fixes, reparse, stop when clean or when no progress is possible. That is the most important behavioral fact for anyone exposing this library to JavaScript.

### What the plan gets partially right

The JavaScript API is directionally good, but the implementation sketch is not precise enough. The plan says to use an options object, which is correct. But its pseudocode mixes `map[string]any` signatures with `goja.Value` decoding. You need to choose one bridge style. If you want robust validation and optional second arguments, use `goja.Value` or `goja.FunctionCall` explicitly. If you use `map[string]any`, you get simpler conversion but less control over `undefined`, arrays, number types, unknown keys, and error messages.

The xgoja provider update is also directionally correct, but incomplete. Adding `"sanitize"` to `textModuleNames` is necessary. It is not sufficient. The provider file must blank-import the sanitize module package so `init()` registers it before `modules.GetModule("sanitize")` runs. The plan mentions this, but it should treat the blank import as a hard dependency, not a side note.

The plan notes tree-sitter dependency risk, which is good, but does not make dependency wiring explicit enough. `goja-text/go.mod` currently does not require `github.com/go-go-golems/sanitize`, and it does not have a local `replace` for `../sanitize`. Without that, the implementation cannot compile against the local sanitize checkout in this workspace.

### What needs correction before implementation

The plan has a likely compile-time issue in the nested export pseudocode. It calls:

```go
modules.SetExport(yamlObj, "sanitize", "yaml.sanitize", func(...) { ... })
```

`modules.SetExport` simply calls `exports.Set(name, fn)`. If the target object is `yamlObj`, the export name should normally be `"sanitize"`, not `"yaml.sanitize"`. Otherwise JavaScript may see a property literally named `yaml.sanitize` on the `yaml` object instead of `sanitize.yaml.sanitize`.

The correct shape is:

```go
yamlObj := vm.NewObject()
modules.SetExport(yamlObj, mod.Name(), "sanitize", yamlSanitizeFn)
modules.SetExport(yamlObj, mod.Name(), "lint", yamlLintFn)
modules.SetExport(yamlObj, mod.Name(), "parseTree", yamlParseTreeFn)
modules.SetExport(yamlObj, mod.Name(), "rules", yamlRulesFn)
modules.SetExport(yamlObj, mod.Name(), "examples", yamlExamplesFn)

_ = exports.Set("yaml", yamlObj)
```

The same applies to the JSON namespace.

The plan also needs a concrete dependency step:

```go
require github.com/go-go-golems/sanitize v0.0.0

replace github.com/go-go-golems/sanitize => ../sanitize
```

The exact version may be resolved by `go mod tidy`, but the local replace is non-negotiable for this workspace unless the module is published and intentionally pinned.

---

## 2. What Is Good

### The plan starts from the actual library instead of inventing a wrapper first

A common mistake in binding work is to design the JavaScript API before understanding the Go library. This plan does better. It identifies the two package roots, `sanitize/pkg/yaml` and `sanitize/pkg/json`, and reads their common structure: `types.go`, `sanitize.go`, `lint.go`, `parse.go`, `rules.go`, `options.go`, and `examples.go`.

That matters because bindings should be thin when the underlying library already has a coherent API. The sanitize packages already expose the right operations. The goja module should adapt them to JavaScript, not reinterpret the library.

### The namespace design is the right public shape

The proposed API:

```js
const sanitize = require("sanitize");

sanitize.yaml.sanitize(input, options);
sanitize.yaml.lint(input, options);
sanitize.yaml.parseTree(input);
sanitize.yaml.rules();
sanitize.yaml.examples();

sanitize.json.sanitize(input, options);
sanitize.json.lint(input, options);
sanitize.json.parseTree(input);
sanitize.json.rules();
sanitize.json.examples();
```

is a clear API. It makes format-specific behavior visible. YAML has `tabWidth`; JSON has strict parse state. Keeping the namespaces separate prevents the API from pretending the two formats are identical.

This shape is also better than two top-level modules because `require("yaml")` already means something different in the current go-go-goja ecosystem. The existing core `yaml` module parses and stringifies YAML. A top-level `require("yaml")` sanitizer would confuse two separate roles: data serialization and malformed-document repair.

### The plan preserves the sanitize library's conservative behavior

The sanitize library is not a free-form repair engine. It applies known rules, validates rule names, iterates up to a maximum, and stops when no progress is possible. The plan preserves that behavior by returning the full `Result`, including original state, final state, and applied fixes.

This is important because callers need to know whether a repair succeeded. A string-only function such as `sanitize.yaml.fix(input) -> string` would be too weak. It would hide whether parse errors remain, whether lint issues remain, and which fixers changed the input.

### The testing plan understands public API contracts

The proposed tests are not limited to "does the function return something." They check field access from JavaScript, rule catalog fields, options behavior, and xgoja integration. That is the right level of testing for a native module. The public contract is not only the Go function result. It is the shape seen by JavaScript through goja reflection.

---

## 3. What Is Bad or Risky

### The plan under-specifies go.mod wiring

The most immediate implementation blocker is dependency wiring. `goja-text/go.mod` currently includes `go-go-goja`, `goja`, `goja_nodejs`, `logcopter`, and `goldmark`, but not `github.com/go-go-golems/sanitize`. The workspace contains `../sanitize`, and `go.work` includes it, but the module still needs a direct dependency if `pkg/sanitize` imports it.

The implementation plan should include a step like:

```bash
cd /home/manuel/workspaces/2026-06-02/goja-text/goja-text
go get github.com/go-go-golems/sanitize@v0.0.0
# if that fails because no such revision exists, add local replace manually
```

A likely final `go.mod` snippet is:

```go
require github.com/go-go-golems/sanitize v0.0.0-00010101000000-000000000000

replace github.com/go-go-golems/sanitize => ../sanitize
```

The exact pseudo-version should be whatever `go mod tidy` produces after local replacement, but the plan should explicitly mention the local replace. This was a known lesson from GOJA-TEXT-001, where generated xgoja builds also needed explicit local replacement.

### The nested export pseudocode is wrong enough to mislead implementation

This is the most concrete code-level issue in the spec. The plan says:

```go
modules.SetExport(yamlObj, "sanitize", "yaml.sanitize", func(...) { ... })
```

But `modules.SetExport` does not parse dotted paths. It simply calls `Set(name, value)`. Passing `"yaml.sanitize"` sets one property with a dot in its name. JavaScript would need bracket syntax:

```js
sanitize.yaml["yaml.sanitize"](...)
```

That is not the intended API.

Correct pseudocode:

```go
func (mod module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)

    yamlObj := vm.NewObject()
    modules.SetExport(yamlObj, mod.Name(), "sanitize", func(input string, options goja.Value) (*yamlsanitize.Result, error) {
        opts, err := decodeYamlOptions(vm, options)
        if err != nil {
            return nil, err
        }
        result, err := yamlsanitize.SanitizeWithOptions(input, opts...)
        if err != nil {
            return nil, fmt.Errorf("sanitize.yaml.sanitize: %w", err)
        }
        return &result, nil
    })
    modules.SetExport(yamlObj, mod.Name(), "lint", yamlLintFn)
    modules.SetExport(yamlObj, mod.Name(), "parseTree", yamlParseTreeFn)
    modules.SetExport(yamlObj, mod.Name(), "rules", func() []yamlsanitize.RuleSpec {
        return yamlsanitize.RuleCatalog()
    })
    modules.SetExport(yamlObj, mod.Name(), "examples", func() []yamlsanitize.Example {
        return yamlsanitize.Examples
    })

    if err := exports.Set("yaml", yamlObj); err != nil {
        // log or panic consistently with module style
    }
}
```

This is a small correction, but it prevents a real API bug.

### The options-decoding sketch is too casual about goja number and array conversion

The plan's options-decoding sketch checks only `int64` for numeric options:

```go
if n, ok := v.Export().(int64); ok && n > 0 { ... }
```

Existing module code in `go-go-goja/modules/yaml/yaml.go` handles `int64`, `int`, and `float64` for numeric options. The sanitize module should follow that pattern. JavaScript numbers are not Go integers. Depending on goja export behavior and call path, you may see `int64`, `int`, or `float64`.

A better helper:

```go
func numberOption(obj *goja.Object, name string) (int, bool, error) {
    v := obj.Get(name)
    if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
        return 0, false, nil
    }
    switch n := v.Export().(type) {
    case int:
        return n, true, nil
    case int64:
        return int(n), true, nil
    case float64:
        if n != float64(int(n)) {
            return 0, true, fmt.Errorf("%s must be an integer", name)
        }
        return int(n), true, nil
    default:
        return 0, true, fmt.Errorf("%s must be a number, got %T", name, n)
    }
}
```

The same issue applies to arrays. `v.Export()` may not always give `[]any` in the way the pseudocode assumes. It is safer to either export and handle `[]any`, or explicitly inspect the JS object as an array-like value. This is important enough to test from JavaScript.

### Unknown option behavior is undecided

Decision Record 3 says unknown keys should be "ignored or rejected based on a validation flag." That is not a decision. It is a placeholder. The implementation should choose one behavior before code is written.

For this module, unknown options should be rejected by default. The existing core YAML module rejects unknown `stringify` options. The sanitize library itself validates unknown rule names. Silent option ignoring makes typos hard to detect:

```js
sanitize.yaml.sanitize(input, { maxIteratons: 1 }); // typo
```

If unknown options are ignored, this silently uses the default max iterations. A caller may believe a configuration was applied when it was not. The review recommendation is:

- Reject unknown options in Phase 1.
- Error messages should include the namespace and function, e.g. `sanitize.yaml.sanitize: unknown option "maxIteratons"`.
- If permissive behavior is ever needed, add an explicit `allowUnknownOptions` escape hatch later.

### The plan should not promise Go-backed validation benefits without specifying value-consuming functions

For Markdown, Go-backed AST objects clearly matter because JavaScript passes nodes back into Go functions: `walk(root, visitor)`, `textContent(node)`, and `validate(node)`. For sanitize, the initial API mostly returns result objects and does not consume them again.

Keeping results Go-backed is still acceptable and consistent. But the plan should be more precise: the immediate benefits are consistent field projection and avoiding manual conversion, not necessarily runtime validation on result objects. If future helper functions consume `Result`, then type identity becomes more valuable.

A stronger statement would be:

> Return structs directly in Phase 1 for consistency and minimal conversion. Add runtime tests that pin PascalCase field access. If a future API consumes `Result` or `LintIssue` values, that API should accept the Go-backed types directly. If JSON-style serialization is needed, add an explicit `toPlainObject` or `toJSONResult` helper.

### TypeScript declarations are more complex than the plan admits

The plan says to implement TypeScript declarations with `RawDTS`, but does not show how nested namespaces should be represented in the `spec.Module` structure. The Markdown module has flat functions, so it is not a sufficient reference for nested exports.

For Phase 1, the safest path is to put the full declaration in `RawDTS` and keep `Functions` minimal or empty if the generator cannot represent nested namespaces cleanly.

Example shape:

```ts
export interface ErrorNode { Type: string; StartByte: number; /* ... */ }
export interface LintIssue { Rule: string; Source: string; Description: string; /* ... */ }
export interface Fix { Rule: string; Description: string; Before: string; After: string; }
export interface YamlResult { Original: string; Sanitized: string; /* ... */ }
export interface JsonResult extends YamlResult { StrictParseClean: boolean; OriginalStrictParseClean: boolean; }
export interface YamlOptions { maxIterations?: number; tabWidth?: number; onlyRules?: string[]; disabledRules?: string[]; }
export interface JsonOptions { maxIterations?: number; onlyRules?: string[]; disabledRules?: string[]; }
export interface ParseTreeResult { TreeText: string; Errors: ErrorNode[]; }

export const yaml: {
  sanitize(input: string, options?: YamlOptions): YamlResult;
  lint(input: string, options?: YamlOptions): LintIssue[];
  parseTree(input: string): ParseTreeResult;
  rules(): YamlRuleSpec[];
  examples(): YamlExample[];
};

export const json: {
  sanitize(input: string, options?: JsonOptions): JsonResult;
  lint(input: string, options?: JsonOptions): LintIssue[];
  parseTree(input: string): ParseTreeResult;
  rules(): JsonRuleSpec[];
  examples(): JsonExample[];
};
```

This avoids forcing the `spec.Function` model to represent nested properties before verifying that it supports them.

### The plan should distinguish parse diagnostics from repair results more clearly

`parseTree(input)` returns a tree text and structural parser errors. `lint(input)` returns rule-based issues, including parse-derived lint issues. `sanitize(input)` returns original and final parse/lint state plus fixes.

These three operations overlap but are not equivalent. The plan should teach the intern to preserve those distinctions:

- `parseTree` is for debugging parser structure.
- `lint` is for actionable diagnostics.
- `sanitize` is for attempting repair and reporting before/after state.

If these functions are conflated, the JavaScript API will become confusing.

---

## 4. What Could Be Better

### Add a concrete API example with expected output

The design has many interface shapes, but it should include at least one exact run example with expected output. This helps the implementer understand the target behavior.

Example:

```js
const sanitize = require("sanitize");

const result = sanitize.yaml.sanitize("name:Alice\n");

console.log(result.Sanitized);              // "name: Alice\n"
console.log(result.OriginalLintIssues[0].Rule); // "missing_space_after_colon"
console.log(result.Fixes[0].Rule);          // "missing_space_after_colon"
console.log(result.LintClean);              // true
```

For JSON:

```js
const result = sanitize.json.sanitize("```json\n{'ok': True,}\n```\n");

console.log(result.Sanitized);        // should be strict JSON if all applied fixes succeed
console.log(result.StrictParseClean); // true or false depending on remaining issues
console.log(result.Fixes.map(f => f.Rule));
```

Expected-output examples make tests easier to write.

### Add a dependency and build-validation phase before module implementation

The plan's Phase 1 starts with `pkg/sanitize/types.go`, but the first implementation step should be dependency wiring:

1. Add `github.com/go-go-golems/sanitize` to `goja-text/go.mod`.
2. Add `replace github.com/go-go-golems/sanitize => ../sanitize`.
3. Run `go mod tidy`.
4. Run a tiny compile probe importing both sanitize packages.

This reduces risk. If tree-sitter or module versions create build failures, you want to discover that before writing the goja wrapper.

### Add a small adapter layer instead of putting all logic in `module.go`

The plan puts everything in `module.go`. That may become too large because YAML and JSON each need sanitize, lint, parse tree, rules, examples, and option decoding.

A cleaner layout:

```text
pkg/sanitize/
  module.go          # NativeModule, Loader, namespace wiring
  types.go           # ParseTreeResult types and shared helpers
  options.go         # JS option decoding helpers
  yaml.go            # YAML wrapper functions
  json.go            # JSON wrapper functions
  typescript.go      # RawDTS declaration builder
  module_test.go     # runtime tests
  options_test.go    # option decoding tests
```

This keeps `module.go` focused on registration and export wiring.

### Add tests that pin lowercase absence if the API intentionally uses PascalCase

The Markdown module explicitly tests `ast.type === undefined`. The sanitize module should do the same for at least one result object:

```js
const result = sanitize.yaml.sanitize("name:Alice\n");
({
  sanitized: result.Sanitized,
  lowercaseMissing: result.sanitized === undefined,
  issueRule: result.OriginalLintIssues[0].Rule,
  issueLowercaseMissing: result.OriginalLintIssues[0].rule === undefined,
});
```

This prevents accidental field-name mapper changes from silently altering the public API.

### Add tests for unknown-rule and conflicting-rule errors

The sanitize library validates rule names and rejects overlap between `onlyRules` and `disabledRules`. The goja module should preserve those errors.

Test cases:

```js
sanitize.yaml.sanitize("name:Alice\n", { onlyRules: ["not_a_rule"] });
sanitize.yaml.sanitize("name:Alice\n", { onlyRules: ["tab_indent"], disabledRules: ["tab_indent"] });
sanitize.json.lint("{'x': 1}", { disabledRules: ["not_a_rule"] });
```

Each should throw a useful Go-backed error into JavaScript.

### Add strict JSON validation as either an explicit non-goal or a Phase 1 function

The plan leaves `sanitize.json.strictParse(input)` as an open question. That is acceptable, but the review recommendation is to decide before implementation. The underlying Go function already exists and is small:

```go
func StrictParse(src string) error
```

There are two reasonable choices:

1. Do not expose it in Phase 1. State clearly that strict status is available through `json.sanitize(input).StrictParseClean` and `json.lint(input)`.
2. Expose `json.strictParse(input)` returning `{ Valid: boolean, Error?: string }`.

The first option keeps the module smaller. The second option is useful because callers often want validation without repair. The important thing is to choose intentionally.

---

## 5. What the Intern Should Have Known

### `modules.SetExport` does not understand dotted names

This is the kind of detail that comes from reading the helper implementation, not only using examples. `SetExport` is tiny:

```go
func SetExport(exports settableObject, moduleName, name string, fn interface{}) {
    if err := exports.Set(name, fn); err != nil { ... }
}
```

Because it only calls `Set`, a dotted name is just a string key. If you want nested exports, create nested objects yourself and set functions on those objects.

### goja argument conversion is convenient but not magic

goja can convert JavaScript values into Go function parameters, but optional arguments, arrays, integer validation, and unknown keys still need careful handling. The existing YAML module's `stringify` function is a good local reference because it handles numeric option conversion and rejects unknown keys.

The intern should have looked at `go-go-goja/modules/yaml/yaml.go` before writing the options-decoding pseudocode.

### A workspace is not a substitute for module dependency wiring

The Go workspace helps local development, but each module still needs correct `require` and `replace` entries when it imports another local module. GOJA-TEXT-001 already ran into this with local `go-go-goja` replacement. GOJA-TEXT-002 will run into the same issue with `sanitize`.

The intern should have checked `goja-text/go.mod` and asked: "If I add `import yamlsanitize \"github.com/go-go-golems/sanitize/pkg/yaml\"`, how will this module resolve it?"

### Existing module names matter

The go-go-goja ecosystem already has a `yaml` module. A plan that proposes top-level `require("yaml")` for sanitization would collide conceptually with existing behavior. The intern avoided that, which is good, but the review should make this explicit because module naming is a public API decision.

### Tree-sitter bindings may affect generated binary builds

The sanitize library uses tree-sitter grammar packages. That is not a reason to avoid the module, but it is a reason to validate both normal tests and xgoja generated builds early. Unit tests can pass while generated builds fail due to replacement paths, CGo settings, or temporary-module dependency resolution.

---

## 6. What They Should Have Looked At

### `go-go-goja/modules/exports.go`

This file is only a few lines, but it determines how nested exports must be implemented. Reading it would have prevented the `"yaml.sanitize"` pseudocode problem.

### `go-go-goja/modules/yaml/yaml.go`

This is the best local reference for option decoding in a native module. It shows:

- optional `map[string]any` options
- number conversion from several Go numeric types
- unknown option rejection
- error wrapping with function names

### `goja-text/go.mod`

This file shows the current dependency story. It already has `go-go-goja` replaced locally, but it does not yet depend on `sanitize`. This should have been called out in the implementation checklist.

### `goja-text/pkg/markdown/module_test.go`

This file shows how to write runtime tests that execute JavaScript under the goja engine. The sanitize module needs the same kind of tests, especially for PascalCase field access and lowercase absence.

### `sanitize/pkg/json/parse.go`

This file includes `StrictParse`, which creates a real API design question. The intern noticed it in the open questions, but should have given a stronger recommendation.

### `sanitize/pkg/yaml/options.go` and `sanitize/pkg/json/options.go`

These files show validation behavior for rule names and overlapping option sets. The JavaScript bridge must not bypass that validation.

---

## 7. Recommended Corrections to the Plan

### Correction 1: Add dependency wiring as Phase 0

Before implementing the module, add a Phase 0:

```md
### Phase 0: Dependency wiring and compile probe

- Add `github.com/go-go-golems/sanitize` to `goja-text/go.mod`.
- Add local replace: `replace github.com/go-go-golems/sanitize => ../sanitize`.
- Run `go mod tidy`.
- Add or run a tiny import probe that imports both `pkg/yaml` and `pkg/json`.
- Run `go test ./... -count=1` before module code is added.
```

### Correction 2: Fix nested export pseudocode

Use nested objects and simple export names:

```go
yamlObj := vm.NewObject()
modules.SetExport(yamlObj, mod.Name(), "sanitize", yamlSanitize)
modules.SetExport(yamlObj, mod.Name(), "lint", yamlLint)
modules.SetExport(yamlObj, mod.Name(), "parseTree", yamlParseTree)
modules.SetExport(yamlObj, mod.Name(), "rules", yamlRules)
modules.SetExport(yamlObj, mod.Name(), "examples", yamlExamples)
_ = exports.Set("yaml", yamlObj)
```

Do not use `"yaml.sanitize"` as the property name.

### Correction 3: Reject unknown options

The options decoder should reject any key outside the allowed set:

YAML allowed keys:

- `maxIterations`
- `tabWidth`
- `onlyRules`
- `disabledRules`

JSON allowed keys:

- `maxIterations`
- `onlyRules`
- `disabledRules`

Unknown keys should return errors. This matches the existing `yaml.stringify` module behavior and prevents typo-driven misconfiguration.

### Correction 4: Split implementation files

Prefer this layout over one large `module.go`:

```text
pkg/sanitize/module.go
pkg/sanitize/options.go
pkg/sanitize/types.go
pkg/sanitize/yaml.go
pkg/sanitize/json.go
pkg/sanitize/typescript.go
pkg/sanitize/module_test.go
pkg/sanitize/options_test.go
```

This is easier to review and maintain.

### Correction 5: Make TypeScript declarations namespace-aware

Use `RawDTS` for nested `yaml` and `json` exports. Do not assume `spec.Function` can model nested module objects unless verified.

### Correction 6: Decide `json.strictParse` before implementation

Recommended Phase 1 decision: expose `json.strictParse(input)` because the underlying function is simple and useful.

Suggested return shape:

```ts
export interface StrictParseResult {
  Valid: boolean;
  Error?: string;
}
```

Suggested behavior:

```js
const check = sanitize.json.strictParse("{'x': 1}");
console.log(check.Valid); // false
console.log(check.Error); // strict parser message
```

If this is rejected for Phase 1, the plan should explicitly state that strict parse status is only available through `json.sanitize` and `json.lint`.

---

## 8. A Better Implementation Sketch

The following sketch is not complete code. It shows the structure the intern should aim for.

```go
func (mod module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)

    yamlObj := vm.NewObject()
    setYamlExports(vm, yamlObj)
    mustSet(exports, "yaml", yamlObj)

    jsonObj := vm.NewObject()
    setJsonExports(vm, jsonObj)
    mustSet(exports, "json", jsonObj)
}

func setYamlExports(vm *goja.Runtime, obj *goja.Object) {
    modules.SetExport(obj, "sanitize", "sanitize", func(input string, options goja.Value) (*yamlsanitize.Result, error) {
        opts, err := decodeYamlOptions(vm, options)
        if err != nil {
            return nil, fmt.Errorf("sanitize.yaml.sanitize: %w", err)
        }
        result, err := yamlsanitize.SanitizeWithOptions(input, opts...)
        if err != nil {
            return nil, fmt.Errorf("sanitize.yaml.sanitize: %w", err)
        }
        return &result, nil
    })

    modules.SetExport(obj, "sanitize", "lint", func(input string, options goja.Value) ([]yamlsanitize.LintIssue, error) {
        opts, err := decodeYamlOptions(vm, options)
        if err != nil {
            return nil, fmt.Errorf("sanitize.yaml.lint: %w", err)
        }
        return yamlsanitize.LintWithOptions(input, opts...)
    })

    modules.SetExport(obj, "sanitize", "parseTree", func(input string) (*YamlParseTreeResult, error) {
        tree, errors, err := yamlsanitize.ParseTree(input)
        if err != nil {
            return nil, fmt.Errorf("sanitize.yaml.parseTree: %w", err)
        }
        return &YamlParseTreeResult{TreeText: tree, Errors: errors}, nil
    })

    modules.SetExport(obj, "sanitize", "rules", func() []yamlsanitize.RuleSpec {
        return yamlsanitize.RuleCatalog()
    })

    modules.SetExport(obj, "sanitize", "examples", func() []yamlsanitize.Example {
        return yamlsanitize.Examples
    })
}
```

The JSON side is the same shape, with `jsonsanitize` types and no `tabWidth` option.

---

## 9. Suggested Review Checklist Before Implementation Starts

Before code is written, the intern should answer these questions:

1. Does `goja-text/go.mod` require and replace `github.com/go-go-golems/sanitize` locally?
2. Does the implementation set `sanitize.yaml.sanitize`, not `sanitize.yaml["yaml.sanitize"]`?
3. Are unknown options rejected?
4. Are `maxIterations` and `tabWidth` validated as positive integers?
5. Are `onlyRules` and `disabledRules` validated as arrays of strings?
6. Are rule-name errors from the sanitize library preserved and wrapped with function context?
7. Does the xgoja provider blank-import `pkg/sanitize`?
8. Does `xgoja.yaml` include the `sanitize` module in the runtime?
9. Do runtime tests prove PascalCase field access and lowercase absence?
10. Does the generated xgoja binary build successfully with the local sanitize dependency?

---

## 10. Final Advice to the Intern

Your plan is strong because it starts from the actual library and proposes a small binding layer rather than a rewrite. That is the right instinct. A binding should preserve the semantics of the Go library unless there is a concrete reason to adapt them.

The next improvement is precision. When a plan includes pseudocode, the pseudocode should be close enough to compile in spirit. It does not need imports and every helper, but it should not rely on behavior that the helper functions do not have. In this plan, the dotted `SetExport` names and loose option decoding are the main examples. These are easy to fix once noticed, but they are exactly the kind of errors that slow down implementation.

For future plans, always do four small checks before finalizing the spec:

1. Read the helper functions you are using, not only the modules that call them.
2. Check `go.mod` and local `replace` requirements before assuming imports will compile.
3. Write one exact JavaScript example with expected output.
4. Convert every "maybe" in a decision record into either a decision or an explicit deferred question.

If you apply those checks, your next plan will be much closer to implementation-ready.
