function fixtures() {
  return [
    {
      kind: "markdown",
      path: "examples/markdown/sample.md",
      tryCommand: "goja-text markdown toc examples/markdown/sample.md"
    },
    {
      kind: "chunking",
      path: "examples/markdown/chunking-sample.md",
      tryCommand: "goja-text chunking pack examples/markdown/chunking-sample.md --max-units 180"
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
    },
    {
      kind: "template-demo",
      path: "examples/js/template-demo.js",
      tryCommand: "goja-text run examples/js/template-demo.js"
    },
    {
      kind: "embedded-template-assets",
      path: "/templates inside fs:assets",
      tryCommand: "goja-text template examples"
    }
  ];
}

__verb__("fixtures", {
  short: "List bundled repository fixtures and useful commands"
});

function tour() {
  return [
    { step: 1, command: "goja-text help goja-text-chunking-user-guide", purpose: "Learn lossless segmentation and budgeted packing" },
    { step: 2, command: "goja-text chunking pack examples/markdown/chunking-sample.md --max-units 180", purpose: "Build source-addressable chunks" },
    { step: 3, command: "goja-text help goja-text-markdown-user-guide", purpose: "Learn the Markdown AST and walk() model" },
    { step: 4, command: "goja-text markdown toc examples/markdown/sample.md", purpose: "Build a table of contents" },
    { step: 5, command: "goja-text help goja-text-sanitize-user-guide", purpose: "Learn repair vs validation" },
    { step: 6, command: "goja-text sanitize json examples/json/broken.json", purpose: "Repair a JSON-like file" },
    { step: 7, command: "goja-text help goja-text-extract-user-guide", purpose: "Learn candidate extraction" },
    { step: 8, command: "goja-text extract validate examples/text/structured-data-sample.md", purpose: "Find and validate structured payloads" },
    { step: 9, command: "goja-text help goja-text-template-writing-documentation", purpose: "Learn documentation rendering with Go templates" },
    { step: 10, command: "goja-text template helper-demo --name goja-text", purpose: "Try the template JSFunc helper command" },
    { step: 11, command: "goja-text template examples", purpose: "List embedded reusable template assets" },
    { step: 12, command: "goja-text template example report", purpose: "Render a bundled Markdown report template" }
  ];
}

__verb__("tour", {
  short: "Show a short goja-text command tour"
});
