---
Title: "goja-text sanitize JavaScript API reference"
Slug: goja-text-sanitize-api-reference
Short: "Reference for require(\"sanitize\") YAML and JSON repair helpers."
Topics:
- goja-text
- sanitize
- yaml
- json
- javascript
Commands:
- goja-text
- goja-text eval
- goja-text run
Flags: []
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

Use `require("sanitize")` to lint, repair, inspect, and strictly parse YAML and JSON strings from JavaScript.

The module exposes two namespaces: `sanitize.yaml` and `sanitize.json`. Both namespaces return Go-backed result objects with PascalCase fields such as `Sanitized`, `Fixes`, `Issues`, and `StrictParseClean`.

## Loading

```js
const sanitize = require("sanitize");
```

## YAML namespace

### sanitize.yaml.options()

Returns a Go-backed YAML options builder.

```js
const config = sanitize.yaml.options()
  .MaxIterations(5)
  .TabWidth(2)
  .Build();
```

Builder methods include `MaxIterations(n)`, `TabWidth(n)`, `OnlyRules(...rules)`, `DisabledRules(...rules)`, `RejectUnknownOptions()`, `AllowUnknownOptions()`, `CollectUnknownOptions()`, `FromObject(obj)`, `Validate()`, and `Build()`.

### sanitize.yaml.sanitize(input, options?)

Repairs YAML-like input and returns a sanitize result.

```js
const result = sanitize.yaml.sanitize("name:Alice\n");
console.log(result.Sanitized);
console.log(result.Fixes.map((fix) => fix.Rule));
```

### sanitize.yaml.lint(input, options?)

Returns diagnostics without using the sanitized output as the main result.

### sanitize.yaml.parseTree(input, options?)

Returns a parse-tree view for debugging and explanation.

### sanitize.yaml.rules()

Returns the YAML rule catalog.

### sanitize.yaml.examples()

Returns example YAML inputs from the underlying sanitize package.

## JSON namespace

### sanitize.json.options()

Returns a Go-backed JSON options builder. JSON has no YAML tab-width option.

### sanitize.json.sanitize(input, options?)

Repairs JSON-like input, including common LLM wrapper and syntax issues.

```js
const result = sanitize.json.sanitize("~~~json\n{'ok': True,}\n~~~\n");
console.log(result.Sanitized);
console.log(result.StrictParseClean);
```

### sanitize.json.lint(input, options?)

Returns JSON diagnostics.

### sanitize.json.parseTree(input, options?)

Returns a parse-tree view for debugging.

### sanitize.json.strictParse(input)

Parses strict JSON and fails if repair would be required.

### sanitize.json.rules()

Returns the JSON rule catalog.

### sanitize.json.examples()

Returns example JSON inputs from the underlying sanitize package.

## Unknown option policy

Builders reject unknown keys imported through `FromObject()` by default. Use `CollectUnknownOptions()` to keep diagnostics instead of failing immediately, or `AllowUnknownOptions()` only when the script explicitly tolerates ignored fields.
