const template = require("template");

const textResult = template.text()
  .Name("summary")
  .Funcs("sprig", "glazed")
  .Parse([
    "Project: {{ .Project | upper }}",
    "Items:",
    "{{ range .Items }}- {{ . | trim }}",
    "{{ end }}",
  ].join("\n"))
  .Render({
    Project: "goja-text",
    Items: [" markdown ", " sanitize ", " templates "],
  });

const named = template.text().Name("report").Parse(`
{{ define "title" }}Report for {{ .Project }}{{ end }}
{{ define "body" -}}
# {{ template "title" . }}
{{ range .Items }}- {{ . }}
{{ end }}
{{- end }}
`);

const namedResult = named.RenderTemplate("body", {
  Project: "goja-text",
  Items: ["Markdown", "Sanitize", "Templates"],
});

const helperResult = template.text()
  .JSFunc("surround", (left, value, right) => `${left}${String(value).toUpperCase()}${right}`)
  .Parse('{{ surround "[" .Name "]" }}')
  .Render({ Name: "ada" });

const htmlResult = template.html()
  .JSFunc("unsafe", () => "<script>alert(1)</script>")
  .Parse('<p>{{ .Name }}</p><a href="{{ .URL }}">open</a><div>{{ unsafe }}</div>')
  .Render({ Name: "<Ada>", URL: "javascript:alert(1)" });

console.log(JSON.stringify({
  Text: textResult.Text,
  Named: namedResult.Text,
  Templates: named.Templates().map((t) => t.Name),
  Helper: helperResult.Text,
  HTML: htmlResult.Text,
  Convenience: template.renderText("Hello {{ .Name }}", { Name: "intern" }).Text,
}, null, 2));
