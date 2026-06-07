# {{ .title }}

{{ .summary }}

## Highlights

{{ range .highlights -}}
- **{{ .name }}** — {{ .description }}
{{ end }}

## Metadata

{{ toYaml .metadata }}
