package markdown

import (
	"fmt"
	"reflect"
	"strings"
)

// MarkdownBuilder builds Markdown documents through a fluent Go-backed API.
type MarkdownBuilder struct {
	doc    markdownDocument
	errors []string
}

// NewMarkdownBuilder creates an empty Markdown document builder.
func NewMarkdownBuilder() *MarkdownBuilder {
	return &MarkdownBuilder{}
}

func (b *MarkdownBuilder) Title(text any) *MarkdownBuilder {
	return b.Heading(1, text)
}

func (b *MarkdownBuilder) Heading(level int, text any) *MarkdownBuilder {
	if level < 1 || level > 6 {
		b.addError("heading level must be 1..6, got %d", level)
		return b
	}
	inlines, err := normalizeInlineInputs([]any{text})
	if err != nil {
		b.addError("heading: %v", err)
		return b
	}
	b.doc.Blocks = append(b.doc.Blocks, headingBlock{Level: level, Text: inlines})
	return b
}

func (b *MarkdownBuilder) Paragraph(parts ...any) *MarkdownBuilder {
	inlines, err := normalizeInlineInputs(parts)
	if err != nil {
		b.addError("paragraph: %v", err)
		return b
	}
	b.doc.Blocks = append(b.doc.Blocks, paragraphBlock{Inlines: inlines})
	return b
}

func (b *MarkdownBuilder) Text(text string) *MarkdownBuilder {
	return b.Paragraph(text)
}

func (b *MarkdownBuilder) Raw(markdown string) *MarkdownBuilder {
	b.doc.Blocks = append(b.doc.Blocks, rawBlock{Markdown: markdown})
	return b
}

func (b *MarkdownBuilder) ThematicBreak() *MarkdownBuilder {
	b.doc.Blocks = append(b.doc.Blocks, thematicBreakBlock{})
	return b
}

func (b *MarkdownBuilder) Blockquote(body any) *MarkdownBuilder {
	b.doc.Blocks = append(b.doc.Blocks, blockquoteBlock{Lines: normalizeBodyLines(body)})
	return b
}

func (b *MarkdownBuilder) Callout(kind, title string, body any) *MarkdownBuilder {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		b.addError("callout kind must not be empty")
		return b
	}
	b.doc.Blocks = append(b.doc.Blocks, calloutBlock{Kind: kind, Title: title, Lines: normalizeBodyLines(body)})
	return b
}

func (b *MarkdownBuilder) BulletList(items any) *MarkdownBuilder {
	listItems, err := normalizeListItems(items)
	if err != nil {
		b.addError("bullet list: %v", err)
		return b
	}
	b.doc.Blocks = append(b.doc.Blocks, listBlock{Items: listItems})
	return b
}

func (b *MarkdownBuilder) OrderedList(items any, start ...int) *MarkdownBuilder {
	listItems, err := normalizeListItems(items)
	if err != nil {
		b.addError("ordered list: %v", err)
		return b
	}
	startAt := 1
	if len(start) > 0 {
		startAt = start[0]
	}
	if startAt <= 0 {
		b.addError("ordered list start must be positive, got %d", startAt)
		return b
	}
	b.doc.Blocks = append(b.doc.Blocks, listBlock{Ordered: true, Start: startAt, Items: listItems})
	return b
}

func (b *MarkdownBuilder) Checklist(items any) *MarkdownBuilder {
	checkItems, err := normalizeChecklistItems(items)
	if err != nil {
		b.addError("checklist: %v", err)
		return b
	}
	b.doc.Blocks = append(b.doc.Blocks, checklistBlock{Items: checkItems})
	return b
}

func (b *MarkdownBuilder) CodeBlock(language, code string) *MarkdownBuilder {
	language = strings.TrimSpace(language)
	if !codeLanguagePattern.MatchString(language) {
		b.addError("code block language %q must not contain whitespace", language)
		return b
	}
	b.doc.Blocks = append(b.doc.Blocks, codeBlock{Language: language, Code: code})
	return b
}

func (b *MarkdownBuilder) Table() *TableBuilder {
	return &TableBuilder{parent: b, table: &tableBlock{}}
}

func (b *MarkdownBuilder) Validate() ValidationResult {
	errs := append([]string(nil), b.errors...)
	for i, block := range b.doc.Blocks {
		switch v := block.(type) {
		case headingBlock:
			if v.Level < 1 || v.Level > 6 {
				errs = append(errs, fmt.Sprintf("block %d: heading level must be 1..6", i+1))
			}
		case tableBlock:
			errs = append(errs, validateTableBlock(i+1, v)...)
		}
	}
	return ValidationResult{Valid: len(errs) == 0, Errors: errs}
}

func (b *MarkdownBuilder) Render() (*MarkdownRenderResult, error) {
	validation := b.Validate()
	if !validation.Valid {
		return nil, fmt.Errorf("markdown.builder.validate: %s", strings.Join(validation.Errors, "; "))
	}
	text, err := renderMarkdownDocument(&b.doc)
	if err != nil {
		return nil, err
	}
	return &MarkdownRenderResult{Text: text, Bytes: len([]byte(text)), Blocks: len(b.doc.Blocks)}, nil
}

func (b *MarkdownBuilder) RenderString() (string, error) {
	result, err := b.Render()
	if err != nil {
		return "", err
	}
	return result.Text, nil
}

func (b *MarkdownBuilder) RenderHTML() (string, error) {
	text, err := b.RenderString()
	if err != nil {
		return "", err
	}
	return RenderHTML(text)
}

func (b *MarkdownBuilder) addError(format string, args ...any) {
	b.errors = append(b.errors, fmt.Sprintf(format, args...))
}

func normalizeInlineInputs(parts []any) ([]markdownInline, error) {
	out := make([]markdownInline, 0, len(parts))
	for _, part := range parts {
		inlines, err := normalizeInlineInput(part)
		if err != nil {
			return nil, err
		}
		out = append(out, inlines...)
	}
	return out, nil
}

func normalizeInlineInput(part any) ([]markdownInline, error) {
	switch v := part.(type) {
	case nil:
		return nil, nil
	case string:
		return []markdownInline{TextInline{Text: v}}, nil
	case fmt.Stringer:
		return []markdownInline{TextInline{Text: v.String()}}, nil
	case TextInline:
		return []markdownInline{v}, nil
	case RawInline:
		return []markdownInline{v}, nil
	case CodeInline:
		return []markdownInline{v}, nil
	case EmphasisInline:
		return []markdownInline{v}, nil
	case StrongInline:
		return []markdownInline{v}, nil
	case LinkInline:
		return []markdownInline{v}, nil
	case []markdownInline:
		return v, nil
	case []any:
		return normalizeInlineInputs(v)
	default:
		rv := reflect.ValueOf(part)
		if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
			parts := make([]any, 0, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				parts = append(parts, rv.Index(i).Interface())
			}
			return normalizeInlineInputs(parts)
		}
		return []markdownInline{TextInline{Text: fmt.Sprint(part)}}, nil
	}
}

func normalizeListItems(items any) ([]listItem, error) {
	values, err := normalizeAnySlice(items)
	if err != nil {
		return nil, err
	}
	out := make([]listItem, 0, len(values))
	for _, value := range values {
		inlines, err := normalizeInlineInput(value)
		if err != nil {
			return nil, err
		}
		out = append(out, listItem{Inlines: inlines})
	}
	return out, nil
}

func normalizeChecklistItems(items any) ([]checklistItem, error) {
	values, err := normalizeAnySlice(items)
	if err != nil {
		return nil, err
	}
	out := make([]checklistItem, 0, len(values))
	for _, value := range values {
		checked, text := false, value
		if m, ok := value.(map[string]any); ok {
			if v, ok := m["checked"].(bool); ok {
				checked = v
			}
			if v, ok := m["Checked"].(bool); ok {
				checked = v
			}
			if v, ok := m["text"]; ok {
				text = v
			} else if v, ok := m["Text"]; ok {
				text = v
			}
		}
		inlines, err := normalizeInlineInput(text)
		if err != nil {
			return nil, err
		}
		out = append(out, checklistItem{Checked: checked, Inlines: inlines})
	}
	return out, nil
}

func normalizeAnySlice(value any) ([]any, error) {
	if value == nil {
		return nil, nil
	}
	if values, ok := value.([]any); ok {
		return values, nil
	}
	rv := reflect.ValueOf(value)
	if !rv.IsValid() || (rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array) {
		return nil, fmt.Errorf("expected array or slice, got %T", value)
	}
	out := make([]any, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		out = append(out, rv.Index(i).Interface())
	}
	return out, nil
}

func normalizeBodyLines(body any) []string {
	if body == nil {
		return nil
	}
	switch v := body.(type) {
	case string:
		return strings.Split(strings.TrimRight(v, "\n"), "\n")
	case *MarkdownBuilder:
		text, err := v.RenderString()
		if err != nil {
			return []string{err.Error()}
		}
		return strings.Split(strings.TrimRight(text, "\n"), "\n")
	default:
		return strings.Split(fmt.Sprint(v), "\n")
	}
}
