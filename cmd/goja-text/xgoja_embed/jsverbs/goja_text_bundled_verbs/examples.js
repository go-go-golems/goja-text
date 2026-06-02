function fixtures() {
  return [
    {
      kind: "markdown",
      path: "examples/markdown/sample.md",
      tryCommand: "goja-text markdown toc examples/markdown/sample.md"
    },
    {
      kind: "yaml",
      path: "examples/yaml/broken.yaml",
      tryCommand: "goja-text sanitize yaml examples/yaml/broken.yaml"
    },
    {
      kind: "json",
      path: "examples/json/broken.json",
      tryCommand: "goja-text sanitize json examples/json/broken.json"
    },
    {
      kind: "structured-text",
      path: "examples/text/structured-data-sample.md",
      tryCommand: "goja-text extract validate examples/text/structured-data-sample.md"
    }
  ];
}

__verb__("fixtures", {
  short: "List bundled repository fixtures and useful commands"
});

function tour() {
  return [
    { step: 1, command: "goja-text help goja-text-markdown-user-guide", purpose: "Learn the Markdown AST and walk() model" },
    { step: 2, command: "goja-text markdown toc examples/markdown/sample.md", purpose: "Build a table of contents" },
    { step: 3, command: "goja-text help goja-text-sanitize-user-guide", purpose: "Learn repair vs validation" },
    { step: 4, command: "goja-text sanitize json examples/json/broken.json", purpose: "Repair a JSON-like file" },
    { step: 5, command: "goja-text help goja-text-extract-user-guide", purpose: "Learn candidate extraction" },
    { step: 6, command: "goja-text extract validate examples/text/structured-data-sample.md", purpose: "Find and validate structured payloads" }
  ];
}

__verb__("tour", {
  short: "Show a short goja-text command tour"
});
