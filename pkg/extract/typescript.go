package extract

import "github.com/go-go-golems/go-go-goja/pkg/tsgen/spec"

func (module) TypeScriptModule() *spec.Module {
	return &spec.Module{
		Name: "extract",
		RawDTS: []string{
			"export interface ExtractionCandidate {",
			"  Kind: string;",
			"  Format: string;",
			"  Text: string;",
			"  Raw: string;",
			"  Wrapper: string;",
			"  Label?: string;",
			"  Info?: string;",
			"  StartByte: number; EndByte: number; StartRow: number; StartColumn: number; EndRow: number; EndColumn: number;",
			"  PayloadStartByte: number; PayloadEndByte: number; PayloadStartRow: number; PayloadStartColumn: number; PayloadEndRow: number; PayloadEndColumn: number;",
			"  Confidence: number;",
			"  Diagnostics?: string[];",
			"}",
			"export interface ExtractOptions { Formats?: string[]; Tags?: string[]; Extractors?: string[]; IncludeDiagnostics: boolean; InferFormat: boolean; MinConfidence: number; MaxCandidates: number; }",
			"export interface ExtractionValidationResult { Valid: boolean; Errors?: string[]; }",
			"export interface CandidateValidationResult { Candidate: ExtractionCandidate; Valid: boolean; Format: string; Sanitized?: string; Errors?: string[]; Fixes?: unknown; Issues?: unknown; }",
			"export interface ExtractOptionsBuilder {",
			"  Formats(...formats: string[]): ExtractOptionsBuilder;",
			"  Tags(...tags: string[]): ExtractOptionsBuilder;",
			"  Extractors(...extractors: string[]): ExtractOptionsBuilder;",
			"  IncludeDiagnostics(enabled: boolean): ExtractOptionsBuilder;",
			"  InferFormat(enabled: boolean): ExtractOptionsBuilder;",
			"  MinConfidence(n: number): ExtractOptionsBuilder;",
			"  MaxCandidates(n: number): ExtractOptionsBuilder;",
			"  Validate(): ExtractionValidationResult;",
			"  Build(): ExtractOptions;",
			"}",
			"export function options(): ExtractOptionsBuilder;",
			"export function markdownCodeBlocks(input: string, options?: ExtractOptions): ExtractionCandidate[];",
			"export function xmlTagged(input: string, options?: ExtractOptions): ExtractionCandidate[];",
			"export function rawStructured(input: string, options?: ExtractOptions): ExtractionCandidate[];",
			"export function frontmatter(input: string, options?: ExtractOptions): ExtractionCandidate[];",
			"export function all(input: string, options?: ExtractOptions): ExtractionCandidate[];",
			"export function validate(candidate: ExtractionCandidate, options?: ExtractOptions): CandidateValidationResult;",
		},
	}
}
