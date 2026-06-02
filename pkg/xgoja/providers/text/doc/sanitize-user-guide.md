---
Title: "goja-text sanitize user guide"
Slug: goja-text-sanitize-user-guide
Short: "A guided introduction to repairing YAML and JSON safely from JavaScript."
Topics:
- goja-text
- sanitize
- yaml
- json
- guide
Commands:
- goja-text eval
- goja-text run
- goja-text markdown
- goja-text sanitize
- goja-text extract
Flags: []
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

Structured text often arrives almost correct. A model wraps JSON in a Markdown fence. A human writes `name:Alice` instead of `name: Alice`. A log captures JSON with a trailing comma. The sanitize module handles that boundary layer: it repairs common YAML and JSON syntax problems, records what it changed, and returns text that downstream code can parse or validate more deliberately.

The important boundary is between repair and trust. `sanitize` can make malformed structured text parseable, but it does not prove that the resulting value satisfies an application schema. A good script repairs first, inspects the fixes, then performs domain validation.

## Choose the format namespace

The module has two namespaces, `sanitize.yaml` and `sanitize.json`. Choose the namespace before configuring or repairing input.

```js
const sanitize = require("sanitize");

const yamlResult = sanitize.yaml.sanitize("name:Alice\nitems:\n- one");
const jsonResult = sanitize.json.sanitize("~~~json\n{'ok': True,}\n~~~");

console.log(yamlResult.Sanitized);
console.log(jsonResult.Sanitized);
```

Results are Go-backed objects. Read fields with PascalCase names such as `Sanitized`, `Fixes`, `Issues`, and `StrictParseClean`.

## Inspect what changed

Repair is most useful when it is explainable. Each result includes fix metadata, so command-line tools can show the cleaned value and the evidence for how it was obtained.

```js
const result = sanitize.json.sanitize("{ok: true,}");

for (const fix of result.Fixes) {
  console.log(`${fix.Rule}: ${fix.Message || "applied"}`);
}
```

This is especially important for generated content. If a repair rule changes, or if an upstream model starts emitting a new malformed shape, the fix list gives you a place to debug the behavior.

## Configure repair with builders

Options use Go-backed builders rather than raw JavaScript option objects as the primary API. Builders let Go reject unknown option names, validate values, and return clearer runtime errors.

```js
const options = sanitize.yaml.options()
  .MaxIterations(8)
  .TabWidth(2)
  .OnlyRules("normalize-space", "fix-colons")
  .Build();

const result = sanitize.yaml.sanitize(input, options);
```

When options come from a user-provided object, import them explicitly. Unknown options are rejected by default; use `CollectUnknownOptions()` when you want diagnostics without failing immediately.

```js
const options = sanitize.json.options()
  .CollectUnknownOptions()
  .FromObject(userConfig)
  .Validate()
  .Build();
```

## Repair before parsing

For JSON, a common workflow is to repair, check that strict parsing is clean, and then parse with JavaScript.

```js
const repaired = sanitize.json.sanitize(input);
if (!repaired.StrictParseClean) {
  throw new Error("JSON still does not parse after repair");
}

const value = JSON.parse(repaired.Sanitized);
```

For YAML, pair the sanitized text with the YAML parser available in your runtime. The key is the same: keep repair separate from the application decision about whether the value is acceptable.

## Running the included examples

The demo script reads intentionally broken files and prints the repaired text and fix lists.

```bash
./dist/goja-text run examples/js/sanitize-demo.js
```

The bundled root-mounted JavaScript verbs expose the same behavior as structured commands:

```bash
./dist/goja-text sanitize yaml examples/yaml/broken.yaml
./dist/goja-text sanitize json examples/json/broken.json
```

Because JavaScript verbs return structured values, the same command can be rendered as JSON, YAML, or a table by Glazed output flags.

## Key points

- Repair is not schema validation. Sanitize syntax first, then validate the resulting value for your domain.
- Inspect `Fixes` when input is untrusted or generated.
- Prefer Go-backed option builders so unknown options and invalid values are caught at the module boundary.
- Keep raw input and sanitized output together in logs when auditability matters.
