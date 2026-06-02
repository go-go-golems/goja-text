package sanitize

import "github.com/go-go-golems/go-go-goja/pkg/tsgen/spec"

func (module) TypeScriptModule() *spec.Module {
	return &spec.Module{
		Name: "sanitize",
		RawDTS: []string{
			"export type UnknownOptionPolicy = 'reject' | 'allow' | 'collect';",
			"export interface ValidationResult { Valid: boolean; Errors?: string[]; Unknown?: string[]; }",
			"export interface ErrorNode { Type: string; StartByte: number; EndByte: number; StartRow: number; StartCol: number; EndRow: number; EndCol: number; Text: string; }",
			"export interface LintIssue { Rule: string; Source: string; Description: string; StartByte: number; EndByte: number; StartRow: number; StartCol: number; EndRow: number; EndCol: number; Row: number; }",
			"export interface Fix { Rule: string; Description: string; Before: string; After: string; }",
			"export interface YamlRuleSpec { Name: string; Summary: string; Lints: boolean; Fixes: boolean; DefaultEnabled: boolean; }",
			"export interface JsonRuleSpec extends YamlRuleSpec { ParseAware: boolean; }",
			"export interface YamlExample { Name: string; Description: string; YAML: string; Category?: string; Source?: string; Filename?: string; }",
			"export interface JsonExample { Name: string; Description: string; JSON: string; Category?: string; Source?: string; Filename?: string; }",
			"export interface YamlResult { Original: string; Sanitized: string; TreeText: string; OriginalTreeText: string; Errors: ErrorNode[]; OriginalErrors: ErrorNode[]; LintIssues: LintIssue[]; OriginalLintIssues: LintIssue[]; Fixes: Fix[]; ParseClean: boolean; LintClean: boolean; }",
			"export interface JsonResult extends YamlResult { StrictParseClean: boolean; OriginalStrictParseClean: boolean; }",
			"export interface YamlParseTreeResult { TreeText: string; Errors: ErrorNode[]; }",
			"export interface JsonParseTreeResult { TreeText: string; Errors: ErrorNode[]; }",
			"export interface StrictParseResult { Valid: boolean; Error?: string; }",
			"export interface YamlConfig { MaxIterations: number; TabWidth: number; OnlyRules?: string[]; DisabledRules?: string[]; UnknownPolicy: UnknownOptionPolicy; Unknown?: string[]; }",
			"export interface JsonConfig { MaxIterations: number; OnlyRules?: string[]; DisabledRules?: string[]; UnknownPolicy: UnknownOptionPolicy; Unknown?: string[]; }",
			"export interface YamlOptionsBuilder { MaxIterations(n: number): YamlOptionsBuilder; TabWidth(n: number): YamlOptionsBuilder; OnlyRules(...rules: string[]): YamlOptionsBuilder; DisabledRules(...rules: string[]): YamlOptionsBuilder; RejectUnknownOptions(): YamlOptionsBuilder; AllowUnknownOptions(): YamlOptionsBuilder; CollectUnknownOptions(): YamlOptionsBuilder; FromObject(options: Record<string, unknown>): YamlOptionsBuilder; Validate(): ValidationResult; Build(): YamlConfig; }",
			"export interface JsonOptionsBuilder { MaxIterations(n: number): JsonOptionsBuilder; OnlyRules(...rules: string[]): JsonOptionsBuilder; DisabledRules(...rules: string[]): JsonOptionsBuilder; RejectUnknownOptions(): JsonOptionsBuilder; AllowUnknownOptions(): JsonOptionsBuilder; CollectUnknownOptions(): JsonOptionsBuilder; FromObject(options: Record<string, unknown>): JsonOptionsBuilder; Validate(): ValidationResult; Build(): JsonConfig; }",
			"export const yaml: { options(): YamlOptionsBuilder; sanitize(input: string, config?: YamlConfig): YamlResult; lint(input: string, config?: YamlConfig): LintIssue[]; parseTree(input: string): YamlParseTreeResult; rules(): YamlRuleSpec[]; examples(): YamlExample[]; };",
			"export const json: { options(): JsonOptionsBuilder; sanitize(input: string, config?: JsonConfig): JsonResult; lint(input: string, config?: JsonConfig): LintIssue[]; parseTree(input: string): JsonParseTreeResult; strictParse(input: string): StrictParseResult; rules(): JsonRuleSpec[]; examples(): JsonExample[]; };",
		},
	}
}
