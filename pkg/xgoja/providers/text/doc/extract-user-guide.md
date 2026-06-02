---
Title: "goja-text extract user guide"
Slug: goja-text-extract-user-guide
Short: "A guided introduction to finding structured-data candidates inside larger text."
Topics:
- goja-text
- extract
- structured-data
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

The extract module is for text that contains structured data but is not itself just structured data. A Markdown document may start with YAML frontmatter and later contain a JSON code fence. A model response may wrap YAML in `<yaml>...</yaml>`. A pasted snippet may be a raw JSON object with no wrapper at all. `require("extract")` finds those candidates and preserves where they came from.

The module deliberately returns candidates rather than parsed values. A candidate is evidence: it records the raw wrapper, the payload, source positions, format guesses, and confidence. That evidence lets a script decide which candidate to trust, show a reviewer where it came from, or validate it before parsing.

## Start broad with all()

When you do not yet know the document shape, call `extract.all(input)`. It runs the available extractors and sorts candidates by source order.

```js
const extract = require("extract");

const input = `---
title: Demo
---

~~~json
{"ok": true}
~~~

<yaml>name: Alice\n</yaml>
`;

const candidates = extract.all(input);
for (const candidate of candidates) {
  console.log(candidate.Kind, candidate.Format, candidate.StartRow, candidate.Confidence);
}
```

The result is intentionally descriptive. `Kind` tells you whether the candidate came from frontmatter, a Markdown code block, an XML-like tag, or raw text. `Format` records the guessed data format. The row and column fields let a tool point back to the original source.

## Narrow the extractor when the wrapper is known

Once a script knows the source shape, call a specific helper. This makes command behavior easier to explain and tests easier to read.

```js
const frontmatter = extract.frontmatter(markdownText);
const blocks = extract.markdownCodeBlocks(markdownText);
const tagged = extract.xmlTagged(modelResponse);
const raw = extract.rawStructured(possibleJsonOrYaml);
```

The XML helper is XML-like rather than a full XML parser. It recognizes simple same-name wrappers such as `<json>...</json>` and `<yaml>...</yaml>`, which are common in model responses and prompt protocols.

## Validate before parsing

Candidates are not trusted parsed values. Validate a candidate before handing its payload to `JSON.parse()` or to a YAML parser.

```js
const candidate = extract.all(input)[0];
const validation = extract.validate(candidate);

if (!validation.Valid) {
  throw new Error("candidate did not validate");
}

console.log(validation.Sanitized);
```

For JSON and YAML, validation delegates to the same repair semantics exposed by `require("sanitize")`. That means extraction and repair compose: first find the likely payload, then ask the format-specific sanitizer whether it can be made parseable.

## Preserve positions in tools

A command-line tool can return only the payload text, but a review tool should preserve positions. `StartRow`, `StartColumn`, `PayloadStartRow`, and `PayloadStartColumn` are what make it possible to highlight the wrapper and explain the extraction result to a user.

```js
const rows = extract.all(input).map((candidate) => ({
  kind: candidate.Kind,
  format: candidate.Format,
  startsAt: `${candidate.StartRow}:${candidate.StartColumn}`,
  payloadStartsAt: `${candidate.PayloadStartRow}:${candidate.PayloadStartColumn}`,
}));
```

## Decide how to handle overlaps

`all()` currently keeps overlapping candidates. That is a feature of the evidence-first design. A document can have YAML frontmatter and also look like a raw YAML document; a Markdown code block can appear inside a larger wrapper. The module shows the evidence and leaves the policy to the caller.

Common policies are:

- Prefer wrapped candidates over raw whole-document candidates.
- Show every candidate to a human reviewer.
- Validate every candidate and keep the highest-confidence valid result.

## Running the included examples

The demo script reads a sample Markdown document and prints the discovered candidates and validation results.

```bash
./dist/goja-text run examples/js/extract-demo.js
```

The bundled root-mounted JavaScript verbs turn extraction into reusable commands:

```bash
./dist/goja-text extract list examples/text/structured-data-sample.md
./dist/goja-text extract validate examples/text/structured-data-sample.md
```

These verbs are a useful starting point for automation because they return structured rows rather than prose output.

## Key points

- Extraction returns candidates, not trusted values.
- Positions and wrapper metadata are part of the result because they make extraction explainable.
- Use specific helpers when the source shape is known; use `all()` when exploring mixed documents.
- Validate candidates before parsing or applying domain rules.
- Treat overlap handling as an application policy, not a hidden module decision.
