# {{ .module }} API Reference

{{ .summary }}

| Function | Purpose | Example |
| --- | --- | --- |
{{ range .functions -}}
| `{{ .name }}` | {{ .purpose }} | `{{ .example }}` |
{{ end }}

## Notes

{{ range .notes -}}
- {{ . }}
{{ end }}
