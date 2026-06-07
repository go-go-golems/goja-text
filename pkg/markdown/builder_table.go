package markdown

import (
	"fmt"
	"strings"
)

// TableBuilder incrementally builds a Markdown table and returns to its parent
// MarkdownBuilder with End().
type TableBuilder struct {
	parent *MarkdownBuilder
	table  *tableBlock
	closed bool
}

func (t *TableBuilder) Columns(columns ...any) *TableBuilder {
	if !t.ensureOpen("columns") {
		return t
	}
	t.table.Columns = t.table.Columns[:0]
	for _, column := range columns {
		parsed, err := normalizeTableColumn(column)
		if err != nil {
			t.parent.addError("table columns: %v", err)
			continue
		}
		t.table.Columns = append(t.table.Columns, parsed)
	}
	return t
}

func (t *TableBuilder) Align(alignments ...string) *TableBuilder {
	if !t.ensureOpen("align") {
		return t
	}
	for i, alignment := range alignments {
		if i >= len(t.table.Columns) {
			t.parent.addError("table align: alignment %d has no matching column", i+1)
			continue
		}
		parsed, err := parseTableAlignment(alignment)
		if err != nil {
			t.parent.addError("table align: %v", err)
			continue
		}
		t.table.Columns[i].Align = parsed
	}
	return t
}

func (t *TableBuilder) Row(cells ...any) *TableBuilder {
	if !t.ensureOpen("row") {
		return t
	}
	row := make([]markdownInline, 0, len(cells))
	for _, cell := range cells {
		inlines, err := normalizeInlineInput(cell)
		if err != nil {
			t.parent.addError("table row: %v", err)
			continue
		}
		row = append(row, cellInline(inlines))
	}
	t.table.Rows = append(t.table.Rows, row)
	return t
}

func (t *TableBuilder) Rows(rows any) *TableBuilder {
	if !t.ensureOpen("rows") {
		return t
	}
	values, err := normalizeAnySlice(rows)
	if err != nil {
		t.parent.addError("table rows: %v", err)
		return t
	}
	for _, value := range values {
		cells, err := normalizeAnySlice(value)
		if err != nil {
			t.parent.addError("table rows: %v", err)
			continue
		}
		t.Row(cells...)
	}
	return t
}

func (t *TableBuilder) End() *MarkdownBuilder {
	if t == nil || t.parent == nil {
		return nil
	}
	if t.closed {
		t.parent.addError("table end: table already ended")
		return t.parent
	}
	t.closed = true
	t.parent.doc.Blocks = append(t.parent.doc.Blocks, *t.table)
	return t.parent
}

func (t *TableBuilder) ensureOpen(method string) bool {
	if t == nil || t.parent == nil || t.table == nil {
		return false
	}
	if t.closed {
		t.parent.addError("table %s: table already ended", method)
		return false
	}
	return true
}

func validateTableBlock(blockIndex int, table tableBlock) []string {
	var errs []string
	if len(table.Columns) == 0 {
		errs = append(errs, fmt.Sprintf("block %d: table requires at least one column", blockIndex))
	}
	for i, column := range table.Columns {
		if strings.TrimSpace(renderInlineTable(column.Label)) == "" {
			errs = append(errs, fmt.Sprintf("block %d: table column %d label must not be empty", blockIndex, i+1))
		}
		if _, err := parseTableAlignment(string(column.Align)); err != nil {
			errs = append(errs, fmt.Sprintf("block %d: table column %d: %v", blockIndex, i+1, err))
		}
	}
	for i, row := range table.Rows {
		if len(row) != len(table.Columns) {
			errs = append(errs, fmt.Sprintf("block %d: table row %d has %d cells, want %d", blockIndex, i+1, len(row), len(table.Columns)))
		}
	}
	return errs
}

func normalizeTableColumn(column any) (tableColumn, error) {
	switch v := column.(type) {
	case tableColumn:
		return v, nil
	case string:
		return tableColumn{Label: []markdownInline{TextInline{Text: v}}, Align: AlignDefault}, nil
	case map[string]any:
		label, ok := v["label"]
		if !ok {
			label = v["Label"]
		}
		inlines, err := normalizeInlineInput(label)
		if err != nil {
			return tableColumn{}, err
		}
		alignment := AlignDefault
		if raw, ok := v["align"]; ok {
			alignment, err = parseTableAlignment(fmt.Sprint(raw))
			if err != nil {
				return tableColumn{}, err
			}
		} else if raw, ok := v["Align"]; ok {
			alignment, err = parseTableAlignment(fmt.Sprint(raw))
			if err != nil {
				return tableColumn{}, err
			}
		}
		return tableColumn{Label: inlines, Align: alignment}, nil
	default:
		inlines, err := normalizeInlineInput(column)
		if err != nil {
			return tableColumn{}, err
		}
		return tableColumn{Label: inlines, Align: AlignDefault}, nil
	}
}

func parseTableAlignment(value string) (TableAlignment, error) {
	switch TableAlignment(strings.ToLower(strings.TrimSpace(value))) {
	case "", AlignDefault:
		return AlignDefault, nil
	case AlignLeft:
		return AlignLeft, nil
	case AlignCenter:
		return AlignCenter, nil
	case AlignRight:
		return AlignRight, nil
	default:
		return AlignDefault, fmt.Errorf("unknown table alignment %q", value)
	}
}

func cellInline(inlines []markdownInline) markdownInline {
	if len(inlines) == 1 {
		return inlines[0]
	}
	return RawInline{Markdown: renderInlineTable(inlines)}
}

// InlineFactory creates explicit inline nodes for MarkdownBuilder methods.
type InlineFactory struct{}

func NewInlineFactory() InlineFactory { return InlineFactory{} }

func (InlineFactory) Text(text string) TextInline   { return TextInline{Text: text} }
func (InlineFactory) Raw(markdown string) RawInline { return RawInline{Markdown: markdown} }
func (InlineFactory) Code(code string) CodeInline   { return CodeInline{Code: code} }

func (InlineFactory) Em(parts ...any) EmphasisInline {
	inlines, _ := normalizeInlineInputs(parts)
	return EmphasisInline{Children: inlines}
}

func (InlineFactory) Strong(parts ...any) StrongInline {
	inlines, _ := normalizeInlineInputs(parts)
	return StrongInline{Children: inlines}
}

func (InlineFactory) Link(text any, url string, title ...string) LinkInline {
	inlines, _ := normalizeInlineInput(text)
	ret := LinkInline{Text: inlines, URL: url}
	if len(title) > 0 {
		ret.Title = title[0]
	}
	return ret
}
