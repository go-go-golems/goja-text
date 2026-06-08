package markdown

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	jsonsanitize "github.com/go-go-golems/sanitize/pkg/json"
	yamlsanitize "github.com/go-go-golems/sanitize/pkg/yaml"
	"gopkg.in/yaml.v3"
)

// DocumentBuilder builds a parsed Markdown document through a fluent Go-backed
// API. It intentionally keeps document parsing policy on the Go side so xgoja
// scripts configure behavior through validated methods rather than loose JS maps.
type DocumentBuilder struct {
	source      string
	frontmatter *frontmatterConfig
	blocks      []blockRule
	errors      []string
}

type frontmatterConfig struct {
	enabled  bool
	format   string
	repair   bool
	required bool
}

type blockRule struct {
	name          string
	xmlTags       []string
	fenceInfos    []string
	stripFromBody bool
	required      bool
	json          *jsonBlockConfig
}

type jsonBlockConfig struct {
	enabled bool
	repair  bool
	strict  bool
}

// FrontmatterBuilder configures leading frontmatter parsing for a document.
type FrontmatterBuilder struct {
	parent *DocumentBuilder
	cfg    *frontmatterConfig
}

// BlockSetBuilder configures named structured blocks extracted from a document body.
type BlockSetBuilder struct {
	parent *DocumentBuilder
}

// BlockRuleBuilder configures extraction and parsing behavior for one named block.
type BlockRuleBuilder struct {
	parent *BlockSetBuilder
	rule   *blockRule
}

// JSONBlockBuilder configures JSON parsing behavior for one structured block.
type JSONBlockBuilder struct {
	parent *BlockRuleBuilder
	cfg    *jsonBlockConfig
}

// DocumentValidationResult reports accumulated document-builder validation errors.
type DocumentValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// ParsedDocument is the built, validated document view exposed to JavaScript.
type ParsedDocument struct {
	source      string
	body        string
	ast         *MarkdownNode
	frontmatter *FrontmatterView
	blocks      []*DocumentBlock
}

// FrontmatterView provides typed access to parsed frontmatter values.
type FrontmatterView struct {
	values map[string]any
}

// DocumentBlock is one extracted structured block from a Markdown document.
type DocumentBlock struct {
	name       string
	kind       string
	text       string
	raw        string
	startByte  int
	endByte    int
	jsonValue  any
	jsonParsed bool
	jsonRepair bool
}

type extractedBlock struct {
	block *DocumentBlock
	rule  *blockRule
}

// NewDocumentBuilder returns a fluent builder for a Markdown document source.
func NewDocumentBuilder(source string) *DocumentBuilder {
	return &DocumentBuilder{source: source}
}

// Frontmatter enables and configures leading frontmatter parsing.
func (b *DocumentBuilder) Frontmatter() *FrontmatterBuilder {
	if b.frontmatter == nil {
		b.frontmatter = &frontmatterConfig{enabled: true, format: "yaml"}
	}
	return &FrontmatterBuilder{parent: b, cfg: b.frontmatter}
}

// Blocks configures named structured blocks in the Markdown body.
func (b *DocumentBuilder) Blocks() *BlockSetBuilder {
	return &BlockSetBuilder{parent: b}
}

// Validate checks builder configuration without parsing the document.
func (b *DocumentBuilder) Validate() DocumentValidationResult {
	if b == nil {
		return DocumentValidationResult{Valid: false, Errors: []string{"document builder is nil"}}
	}
	errs := append([]string(nil), b.errors...)
	if b.frontmatter != nil {
		format := normalizeDocumentLabel(b.frontmatter.format)
		if format == "" {
			errs = append(errs, "frontmatter format must not be empty")
		} else if format != "yaml" {
			errs = append(errs, fmt.Sprintf("unsupported frontmatter format %q", b.frontmatter.format))
		}
	}
	seenBlocks := map[string]bool{}
	for _, rule := range b.blocks {
		name := normalizeDocumentLabel(rule.name)
		if !validDocumentLabel(name) {
			errs = append(errs, fmt.Sprintf("invalid block name %q", rule.name))
		}
		if seenBlocks[name] {
			errs = append(errs, fmt.Sprintf("duplicate block name %q", rule.name))
		}
		seenBlocks[name] = true
		if len(rule.xmlTags) == 0 && len(rule.fenceInfos) == 0 {
			errs = append(errs, fmt.Sprintf("block %q requires at least one XML tag or fence source", rule.name))
		}
		for _, tag := range rule.xmlTags {
			if !validDocumentLabel(tag) {
				errs = append(errs, fmt.Sprintf("invalid XML tag %q for block %q", tag, rule.name))
			}
		}
		for _, info := range rule.fenceInfos {
			if !validDocumentLabel(info) {
				errs = append(errs, fmt.Sprintf("invalid fence info %q for block %q", info, rule.name))
			}
		}
	}
	return DocumentValidationResult{Valid: len(errs) == 0, Errors: errs}
}

// Build parses the configured Markdown document and returns a Go-backed view.
func (b *DocumentBuilder) Build() (*ParsedDocument, error) {
	validation := b.Validate()
	if !validation.Valid {
		return nil, fmt.Errorf("markdown.document.build: invalid config: %s", strings.Join(validation.Errors, "; "))
	}

	body := b.source
	frontmatter := NewFrontmatterView(nil)
	if b.frontmatter != nil && b.frontmatter.enabled {
		values, rest, err := b.parseFrontmatter()
		if err != nil {
			return nil, err
		}
		frontmatter = NewFrontmatterView(values)
		body = rest
	}

	extracted, err := b.extractBlocks(body)
	if err != nil {
		return nil, err
	}
	if len(extracted) > 0 {
		body = stripExtractedBlocks(body, extracted)
	}

	ast, err := Parse(body)
	if err != nil {
		return nil, fmt.Errorf("markdown.document.build: parse body: %w", err)
	}
	blocks := make([]*DocumentBlock, 0, len(extracted))
	for _, item := range extracted {
		blocks = append(blocks, item.block)
	}
	return &ParsedDocument{source: b.source, body: body, ast: ast, frontmatter: frontmatter, blocks: blocks}, nil
}

// YAML configures frontmatter parsing as YAML. YAML is the only supported format
// in the first implementation slice.
func (b *FrontmatterBuilder) YAML() *FrontmatterBuilder {
	b.cfg.format = "yaml"
	return b
}

// Repair enables YAML repair before frontmatter parsing.
func (b *FrontmatterBuilder) Repair() *FrontmatterBuilder {
	b.cfg.repair = true
	return b
}

// Optional allows documents without frontmatter.
func (b *FrontmatterBuilder) Optional() *FrontmatterBuilder {
	b.cfg.required = false
	return b
}

// Required requires a leading frontmatter section.
func (b *FrontmatterBuilder) Required() *FrontmatterBuilder {
	b.cfg.required = true
	return b
}

// End returns to the parent DocumentBuilder.
func (b *FrontmatterBuilder) End() *DocumentBuilder {
	return b.parent
}

// Block adds a named structured block rule.
func (b *BlockSetBuilder) Block(name string) *BlockRuleBuilder {
	rule := blockRule{name: normalizeDocumentLabel(name)}
	b.parent.blocks = append(b.parent.blocks, rule)
	return &BlockRuleBuilder{parent: b, rule: &b.parent.blocks[len(b.parent.blocks)-1]}
}

// End returns to the parent DocumentBuilder.
func (b *BlockSetBuilder) End() *DocumentBuilder {
	return b.parent
}

// FromXMLTag extracts blocks wrapped in the given XML-like tag name.
func (b *BlockRuleBuilder) FromXMLTag(tag string) *BlockRuleBuilder {
	tag = normalizeDocumentLabel(tag)
	if tag != "" {
		b.rule.xmlTags = appendUniqueString(b.rule.xmlTags, tag)
	}
	return b
}

// FromFence extracts fenced code blocks whose first info word matches info.
func (b *BlockRuleBuilder) FromFence(info string) *BlockRuleBuilder {
	info = normalizeDocumentLabel(info)
	if info != "" {
		b.rule.fenceInfos = appendUniqueString(b.rule.fenceInfos, info)
	}
	return b
}

// StripFromBody removes matching blocks before parsing/rendering the Markdown body.
func (b *BlockRuleBuilder) StripFromBody() *BlockRuleBuilder {
	b.rule.stripFromBody = true
	return b
}

// JSON configures matching blocks as JSON payloads.
func (b *BlockRuleBuilder) JSON() *JSONBlockBuilder {
	if b.rule.json == nil {
		b.rule.json = &jsonBlockConfig{enabled: true}
	}
	b.rule.json.enabled = true
	return &JSONBlockBuilder{parent: b, cfg: b.rule.json}
}

// Optional allows this block to be absent.
func (b *BlockRuleBuilder) Optional() *BlockRuleBuilder {
	b.rule.required = false
	return b
}

// Required requires at least one matching block.
func (b *BlockRuleBuilder) Required() *BlockRuleBuilder {
	b.rule.required = true
	return b
}

// End returns to the parent BlockSetBuilder.
func (b *BlockRuleBuilder) End() *BlockSetBuilder {
	return b.parent
}

// Repair enables JSON repair before parsing.
func (b *JSONBlockBuilder) Repair() *JSONBlockBuilder {
	b.cfg.repair = true
	b.cfg.strict = false
	return b
}

// Strict disables repair and requires strict JSON.
func (b *JSONBlockBuilder) Strict() *JSONBlockBuilder {
	b.cfg.strict = true
	b.cfg.repair = false
	return b
}

// Optional allows JSON parsing to be skipped for empty or absent optional blocks.
func (b *JSONBlockBuilder) Optional() *JSONBlockBuilder {
	b.parent.rule.required = false
	return b
}

// Required requires a matching block with parseable JSON.
func (b *JSONBlockBuilder) Required() *JSONBlockBuilder {
	b.parent.rule.required = true
	return b
}

// End returns to the parent BlockRuleBuilder.
func (b *JSONBlockBuilder) End() *BlockRuleBuilder {
	return b.parent
}

// Source returns the original source passed to markdown.document().
func (d *ParsedDocument) Source() string {
	if d == nil {
		return ""
	}
	return d.source
}

// Body returns the parsed document body after frontmatter and stripped blocks.
func (d *ParsedDocument) Body() string {
	if d == nil {
		return ""
	}
	return d.body
}

// AST returns the Markdown AST for Body().
func (d *ParsedDocument) AST() *MarkdownNode {
	if d == nil {
		return nil
	}
	return d.ast
}

// Frontmatter returns typed frontmatter accessors.
func (d *ParsedDocument) Frontmatter() *FrontmatterView {
	if d == nil || d.frontmatter == nil {
		return NewFrontmatterView(nil)
	}
	return d.frontmatter
}

// FirstHeading returns the first heading text in the parsed body, or fallback.
func (d *ParsedDocument) FirstHeading(fallback ...string) string {
	if d == nil || d.ast == nil {
		return firstStringFallback(fallback)
	}
	var title string
	var visit func(*MarkdownNode)
	visit = func(node *MarkdownNode) {
		if node == nil || title != "" {
			return
		}
		if node.Type == "heading" {
			if text, err := TextContent(node); err == nil {
				title = strings.TrimSpace(text)
			}
			return
		}
		for _, child := range node.Children {
			visit(child)
		}
	}
	visit(d.ast)
	if title == "" {
		return firstStringFallback(fallback)
	}
	return title
}

// RenderHTML renders the parsed body to HTML.
func (d *ParsedDocument) RenderHTML() (string, error) {
	if d == nil {
		return "", fmt.Errorf("markdown.document.renderHTML: document is nil")
	}
	return RenderHTML(d.body)
}

// Blocks returns all extracted blocks.
func (d *ParsedDocument) Blocks() []*DocumentBlock {
	if d == nil {
		return nil
	}
	return append([]*DocumentBlock(nil), d.blocks...)
}

// Block returns the first extracted block by name, or nil.
func (d *ParsedDocument) Block(name string) *DocumentBlock {
	if d == nil {
		return nil
	}
	name = normalizeDocumentLabel(name)
	for _, block := range d.blocks {
		if block.name == name {
			return block
		}
	}
	return nil
}

// NewFrontmatterView creates a typed view over frontmatter values.
func NewFrontmatterView(values map[string]any) *FrontmatterView {
	out := map[string]any{}
	for k, v := range values {
		out[k] = normalizeDocumentValue(v)
	}
	return &FrontmatterView{values: out}
}

// Has reports whether a frontmatter key exists.
func (v *FrontmatterView) Has(name string) bool {
	if v == nil {
		return false
	}
	_, ok := v.values[name]
	return ok
}

// Value returns the raw normalized frontmatter value.
func (v *FrontmatterView) Value(name string) any {
	if v == nil {
		return nil
	}
	return v.values[name]
}

// String returns a frontmatter value coerced to string, or fallback/empty string.
func (v *FrontmatterView) String(name string, fallback ...string) string {
	if v == nil {
		return firstStringFallback(fallback)
	}
	value, ok := v.values[name]
	if !ok || value == nil {
		return firstStringFallback(fallback)
	}
	switch x := value.(type) {
	case string:
		if x == "" {
			return firstStringFallback(fallback)
		}
		return x
	case fmt.Stringer:
		return x.String()
	case bool:
		return strconv.FormatBool(x)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return fmt.Sprint(x)
	}
}

// Number returns a frontmatter value coerced to float64, or fallback/zero.
func (v *FrontmatterView) Number(name string, fallback ...float64) float64 {
	if v == nil {
		return firstNumberFallback(fallback)
	}
	value, ok := v.values[name]
	if !ok || value == nil {
		return firstNumberFallback(fallback)
	}
	switch x := value.(type) {
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case float64:
		return x
	case float32:
		return float64(x)
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
		if err == nil {
			return parsed
		}
	}
	return firstNumberFallback(fallback)
}

// Bool returns a frontmatter value coerced to bool, or fallback/false.
func (v *FrontmatterView) Bool(name string, fallback ...bool) bool {
	if v == nil {
		return firstBoolFallback(fallback)
	}
	value, ok := v.values[name]
	if !ok || value == nil {
		return firstBoolFallback(fallback)
	}
	switch x := value.(type) {
	case bool:
		return x
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(x))
		if err == nil {
			return parsed
		}
	}
	return firstBoolFallback(fallback)
}

// Keys returns frontmatter keys in stable order.
func (v *FrontmatterView) Keys() []string {
	if v == nil {
		return nil
	}
	keys := make([]string, 0, len(v.values))
	for k := range v.values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ToObject returns a shallow copy of the normalized frontmatter values.
func (v *FrontmatterView) ToObject() map[string]any {
	if v == nil {
		return map[string]any{}
	}
	out := map[string]any{}
	for k, value := range v.values {
		out[k] = value
	}
	return out
}

func (b *DocumentBlock) Name() string {
	if b == nil {
		return ""
	}
	return b.name
}

func (b *DocumentBlock) Kind() string {
	if b == nil {
		return ""
	}
	return b.kind
}

func (b *DocumentBlock) Text() string {
	if b == nil {
		return ""
	}
	return b.text
}

func (b *DocumentBlock) Raw() string {
	if b == nil {
		return ""
	}
	return b.raw
}

func (b *DocumentBlock) StartByte() int {
	if b == nil {
		return 0
	}
	return b.startByte
}

func (b *DocumentBlock) EndByte() int {
	if b == nil {
		return 0
	}
	return b.endByte
}

// JSONValue returns the parsed JSON value for this block. If the block was not
// configured as JSON, it attempts strict JSON parsing on demand.
func (b *DocumentBlock) JSONValue() (any, error) {
	if b == nil {
		return nil, fmt.Errorf("markdown.document.block.jsonValue: block is nil")
	}
	if b.jsonParsed {
		return b.jsonValue, nil
	}
	value, err := parseDocumentJSON(b.text, b.jsonRepair)
	if err != nil {
		return nil, fmt.Errorf("markdown.document.block %q: json: %w", b.name, err)
	}
	b.jsonValue = value
	b.jsonParsed = true
	return value, nil
}

func (b *DocumentBuilder) parseFrontmatter() (map[string]any, string, error) {
	match := frontmatterPattern.FindStringSubmatchIndex(b.source)
	if match == nil {
		if b.frontmatter.required {
			return nil, "", fmt.Errorf("markdown.document.build: required frontmatter not found")
		}
		return map[string]any{}, b.source, nil
	}
	text := b.source[match[2]:match[3]]
	if b.frontmatter.repair {
		result := yamlsanitize.Sanitize(text)
		text = result.Sanitized
	}
	values, err := parseDocumentYAMLMap(text)
	if err != nil {
		return nil, "", fmt.Errorf("markdown.document.build: frontmatter: parse yaml: %w", err)
	}
	return values, b.source[match[1]:], nil
}

func (b *DocumentBuilder) extractBlocks(body string) ([]extractedBlock, error) {
	var extracted []extractedBlock
	for i := range b.blocks {
		rule := &b.blocks[i]
		matches := extractBlocksForRule(body, rule)
		if rule.required && len(matches) == 0 {
			return nil, fmt.Errorf("markdown.document.build: required block %q not found", rule.name)
		}
		for _, block := range matches {
			if rule.json != nil && rule.json.enabled {
				value, err := parseDocumentJSON(block.text, rule.json.repair)
				if err != nil {
					return nil, fmt.Errorf("markdown.document.build: block %q: json: %w", rule.name, err)
				}
				block.jsonValue = value
				block.jsonParsed = true
				block.jsonRepair = rule.json.repair
			}
			extracted = append(extracted, extractedBlock{block: block, rule: rule})
		}
	}
	sort.SliceStable(extracted, func(i, j int) bool {
		return extracted[i].block.startByte < extracted[j].block.startByte
	})
	return extracted, nil
}

func extractBlocksForRule(body string, rule *blockRule) []*DocumentBlock {
	var blocks []*DocumentBlock
	wantedTags := stringSet(rule.xmlTags)
	for _, match := range xmlBlockPattern.FindAllStringSubmatchIndex(body, -1) {
		tag := normalizeDocumentLabel(body[match[2]:match[3]])
		closingTag := normalizeDocumentLabel(body[match[6]:match[7]])
		if tag != closingTag || !wantedTags[tag] {
			continue
		}
		blocks = append(blocks, &DocumentBlock{name: rule.name, kind: "xml", text: body[match[4]:match[5]], raw: body[match[0]:match[1]], startByte: match[0], endByte: match[1]})
	}
	wantedFences := stringSet(rule.fenceInfos)
	for _, match := range fencedBlockPattern.FindAllStringSubmatchIndex(body, -1) {
		info := strings.TrimSpace(body[match[2]:match[3]])
		first := ""
		if fields := strings.Fields(info); len(fields) > 0 {
			first = normalizeDocumentLabel(fields[0])
		}
		if !wantedFences[first] {
			continue
		}
		blocks = append(blocks, &DocumentBlock{name: rule.name, kind: "fence", text: body[match[4]:match[5]], raw: body[match[0]:match[1]], startByte: match[0], endByte: match[1]})
	}
	sort.SliceStable(blocks, func(i, j int) bool { return blocks[i].startByte < blocks[j].startByte })
	return blocks
}

func stripExtractedBlocks(body string, extracted []extractedBlock) string {
	type span struct{ start, end int }
	spans := []span{}
	for _, item := range extracted {
		if item.rule.stripFromBody {
			spans = append(spans, span{start: item.block.startByte, end: item.block.endByte})
		}
	}
	if len(spans) == 0 {
		return body
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].start > spans[j].start })
	for _, s := range spans {
		if s.start >= 0 && s.end <= len(body) && s.start <= s.end {
			body = body[:s.start] + body[s.end:]
		}
	}
	return strings.TrimSpace(body)
}

func parseDocumentYAMLMap(text string) (map[string]any, error) {
	if strings.TrimSpace(text) == "" {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := yaml.Unmarshal([]byte(text), &out); err != nil {
		return nil, err
	}
	if out == nil {
		return map[string]any{}, nil
	}
	return normalizeDocumentValue(out).(map[string]any), nil
}

func parseDocumentJSON(text string, repair bool) (any, error) {
	text = strings.TrimSpace(text)
	if repair {
		result := jsonsanitize.Sanitize(text)
		text = result.Sanitized
	}
	var out any
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return nil, err
	}
	return normalizeDocumentValue(out), nil
}

func normalizeDocumentValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := map[string]any{}
		for k, item := range v {
			out[k] = normalizeDocumentValue(item)
		}
		return out
	case map[any]any:
		out := map[string]any{}
		for k, item := range v {
			out[fmt.Sprint(k)] = normalizeDocumentValue(item)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = normalizeDocumentValue(item)
		}
		return out
	default:
		return value
	}
}

func normalizeDocumentLabel(label string) string {
	return strings.ToLower(strings.TrimSpace(label))
}

func validDocumentLabel(label string) bool {
	return documentLabelPattern.MatchString(label)
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}

func firstStringFallback(fallback []string) string {
	if len(fallback) > 0 {
		return fallback[0]
	}
	return ""
}

func firstNumberFallback(fallback []float64) float64 {
	if len(fallback) > 0 {
		return fallback[0]
	}
	return 0
}

func firstBoolFallback(fallback []bool) bool {
	if len(fallback) > 0 {
		return fallback[0]
	}
	return false
}

var (
	frontmatterPattern   = regexp.MustCompile(`(?s)^---\s*\r?\n(.*?)\r?\n---\s*\r?\n?`)
	xmlBlockPattern      = regexp.MustCompile(`(?is)<([a-z][a-z0-9_-]*)\b[^>]*>(.*?)</\s*([a-z][a-z0-9_-]*)\s*>`)
	fencedBlockPattern   = regexp.MustCompile("(?is)```([^\n\r`]*)\r?\n(.*?)\r?\n```")
	documentLabelPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
)
