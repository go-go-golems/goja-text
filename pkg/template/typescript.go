package template

import "github.com/go-go-golems/go-go-goja/pkg/tsgen/spec"

func (module) TypeScriptModule() *spec.Module {
	return &spec.Module{
		Name: "template",
		RawDTS: []string{
			"export type TemplateMode = 'text' | 'html';",
			"export type MissingKeyPolicy = 'default' | 'invalid' | 'zero' | 'error';",
			"export type FuncSetName = 'none' | 'sprig' | 'glazed';",
			"export interface ValidationResult { Valid: boolean; Errors?: string[]; }",
			"export interface TemplateConfig { Mode: TemplateMode; Name: string; FuncSets?: string[]; MissingKey: MissingKeyPolicy; LeftDelim?: string; RightDelim?: string; }",
			"export interface RenderResult { Text: string; TemplateName: string; Mode: TemplateMode; Bytes: number; }",
			"export interface TemplateInfo { Name: string; Defined: boolean; Mode: TemplateMode; }",
			"export interface TemplateBuilder {",
			"  Name(name: string): TemplateBuilder;",
			"  Funcs(...names: FuncSetName[]): TemplateBuilder;",
			"  MissingKey(policy: MissingKeyPolicy): TemplateBuilder;",
			"  Delims(left: string, right: string): TemplateBuilder;",
			"  Validate(): ValidationResult;",
			"  BuildConfig(): TemplateConfig;",
			"  Parse(source: string): TemplateSet;",
			"  ParseNamed(name: string, source: string): TemplateSet;",
			"}",
			"export interface TemplateSet {",
			"  Mode: TemplateMode;",
			"  Name: string;",
			"  Render(data?: unknown): RenderResult;",
			"  RenderString(data?: unknown): string;",
			"  RenderTemplate(name: string, data?: unknown): RenderResult;",
			"  Templates(): TemplateInfo[];",
			"  Lookup(name: string): TemplateInfo | undefined;",
			"}",
			"export function text(): TemplateBuilder;",
			"export function html(): TemplateBuilder;",
			"export function renderText(source: string, data?: unknown): RenderResult;",
			"export function renderHTML(source: string, data?: unknown): RenderResult;",
		},
	}
}
