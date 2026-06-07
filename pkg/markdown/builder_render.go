package markdown

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var markdownSpecials = strings.NewReplacer(
	"\\", "\\\\",
	"*", "\\*",
	"_", "\\_",
	"`", "\\`",
	"[", "\\[",
	"]", "\\]",
	"<", "\\<",
	">", "\\>",
)

var codeLanguagePattern = regexp.MustCompile(`^[A-Za-z0-9_+.#-]*$`)

func renderMarkdownDocument(doc *markdownDocument) (string, error) {
	if doc == nil || len(doc.Blocks) == 0 {
		return "", nil
	}
	parts := make([]string, 0, len(doc.Blocks))
	for _, block := range doc.Blocks {
		rendered, err := renderMarkdownBlock(block)
		if err != nil {
			return "", err
		}
		rendered = strings.Trim(rendered, "\n")
		if rendered != "" {
			parts = append(parts, rendered)
		}
	}
	if len(parts) == 0 {
		return "", nil
	}
	return strings.Join(parts, "\n\n") + "\n", nil
}

func renderMarkdownBlock(block markdownBlock) (string, error) {
	switch b := block.(type) {
	case headingBlock:
		return strings.Repeat("#", b.Level) + " " + renderInlineNormal(b.Text), nil
	case paragraphBlock:
		return normalizeParagraph(renderInlineNormal(b.Inlines)), nil
	case rawBlock:
		return b.Markdown, nil
	case thematicBreakBlock:
		return "---", nil
	case blockquoteBlock:
		return prefixLines(b.Lines, "> "), nil
	case calloutBlock:
		title := strings.TrimSpace(b.Title)
		first := "> [!" + strings.ToUpper(strings.TrimSpace(b.Kind)) + "]"
		if title != "" {
			first += " " + title
		}
		lines := append([]string{first}, prefixedBodyLines(b.Lines, "> ")...)
		return strings.Join(lines, "\n"), nil
	case listBlock:
		return renderList(b), nil
	case checklistBlock:
		return renderChecklist(b), nil
	case codeBlock:
		fence := fenceFor(b.Code)
		return fence + b.Language + "\n" + strings.TrimRight(b.Code, "\n") + "\n" + fence, nil
	case tableBlock:
		return renderTable(b)
	default:
		return "", fmt.Errorf("markdown.builder.render: unsupported block %T", block)
	}
}

func renderInlineNormal(inlines []markdownInline) string {
	var b strings.Builder
	for _, in := range inlines {
		b.WriteString(renderOneInline(in, false))
	}
	return b.String()
}

func renderInlineTable(inlines []markdownInline) string {
	var b strings.Builder
	for _, in := range inlines {
		b.WriteString(renderOneInline(in, true))
	}
	return b.String()
}

func renderOneInline(in markdownInline, tableCell bool) string {
	switch v := in.(type) {
	case TextInline:
		return escapeMarkdownText(v.Text, tableCell)
	case RawInline:
		return v.Markdown
	case CodeInline:
		return renderCodeSpan(v.Code, tableCell)
	case EmphasisInline:
		return "*" + renderInlineNormal(v.Children) + "*"
	case StrongInline:
		return "**" + renderInlineNormal(v.Children) + "**"
	case LinkInline:
		text := renderInlineNormal(v.Text)
		url := strings.ReplaceAll(v.URL, ")", "%29")
		if v.Title != "" {
			return "[" + text + "](" + url + " " + strconv.Quote(v.Title) + ")"
		}
		return "[" + text + "](" + url + ")"
	default:
		return ""
	}
}

func escapeMarkdownText(s string, tableCell bool) string {
	s = markdownSpecials.Replace(s)
	if tableCell {
		s = strings.ReplaceAll(s, "|", `\|`)
		s = strings.ReplaceAll(s, "\r\n", "<br>")
		s = strings.ReplaceAll(s, "\n", "<br>")
		s = strings.ReplaceAll(s, "\r", "<br>")
	}
	return s
}

func normalizeParagraph(s string) string {
	fields := strings.Fields(strings.ReplaceAll(s, "\u00a0", " "))
	return strings.Join(fields, " ")
}

func renderList(block listBlock) string {
	lines := make([]string, 0, len(block.Items))
	start := block.Start
	if start <= 0 {
		start = 1
	}
	for i, item := range block.Items {
		marker := "-"
		if block.Ordered {
			marker = fmt.Sprintf("%d.", start+i)
		}
		lines = append(lines, marker+" "+renderInlineNormal(item.Inlines))
	}
	return strings.Join(lines, "\n")
}

func renderChecklist(block checklistBlock) string {
	lines := make([]string, 0, len(block.Items))
	for _, item := range block.Items {
		mark := " "
		if item.Checked {
			mark = "x"
		}
		lines = append(lines, "- ["+mark+"] "+renderInlineNormal(item.Inlines))
	}
	return strings.Join(lines, "\n")
}

func renderTable(block tableBlock) (string, error) {
	columns := len(block.Columns)
	if columns == 0 {
		return "", fmt.Errorf("markdown.builder.table: table requires at least one column")
	}
	widths := make([]int, columns)
	headers := make([]string, columns)
	for i, column := range block.Columns {
		headers[i] = strings.TrimSpace(renderInlineTable(column.Label))
		widths[i] = max(widths[i], displayWidth(headers[i]))
	}
	rows := make([][]string, len(block.Rows))
	for r, row := range block.Rows {
		if len(row) != columns {
			return "", fmt.Errorf("markdown.builder.table: row %d has %d cells, want %d", r+1, len(row), columns)
		}
		rows[r] = make([]string, columns)
		for c, cell := range row {
			text := strings.TrimSpace(renderInlineTable([]markdownInline{cell}))
			rows[r][c] = text
			widths[c] = max(widths[c], displayWidth(text))
		}
	}
	lines := []string{renderTableRow(headers, widths), renderTableAlignmentRow(block.Columns, widths)}
	for _, row := range rows {
		lines = append(lines, renderTableRow(row, widths))
	}
	return strings.Join(lines, "\n"), nil
}

func renderTableRow(cells []string, widths []int) string {
	parts := make([]string, len(cells))
	for i, cell := range cells {
		parts[i] = " " + padRight(cell, widths[i]) + " "
	}
	return "|" + strings.Join(parts, "|") + "|"
}

func renderTableAlignmentRow(columns []tableColumn, widths []int) string {
	parts := make([]string, len(columns))
	for i, column := range columns {
		width := max(widths[i], 3)
		marker := strings.Repeat("-", width)
		switch column.Align {
		case AlignDefault:
		case AlignLeft:
			marker = ":" + strings.Repeat("-", max(width-1, 2))
		case AlignCenter:
			marker = ":" + strings.Repeat("-", max(width-2, 1)) + ":"
		case AlignRight:
			marker = strings.Repeat("-", max(width-1, 2)) + ":"
		}
		parts[i] = " " + marker + " "
	}
	return "|" + strings.Join(parts, "|") + "|"
}

func renderCodeSpan(code string, tableCell bool) string {
	fence := strings.Repeat("`", max(1, longestRun(code, '`')+1))
	text := fence + code + fence
	if tableCell {
		text = strings.ReplaceAll(text, "|", `\|`)
	}
	return text
}

func fenceFor(code string) string {
	return strings.Repeat("`", max(3, longestRun(code, '`')+1))
}

func longestRun(s string, needle rune) int {
	best, cur := 0, 0
	for _, r := range s {
		if r == needle {
			cur++
			best = max(best, cur)
		} else {
			cur = 0
		}
	}
	return best
}

func prefixLines(lines []string, prefix string) string {
	return strings.Join(prefixedBodyLines(lines, prefix), "\n")
}

func prefixedBodyLines(lines []string, prefix string) []string {
	if len(lines) == 0 {
		return []string{prefix}
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		for _, split := range strings.Split(line, "\n") {
			out = append(out, prefix+split)
		}
	}
	return out
}

func displayWidth(s string) int {
	return utf8.RuneCountInString(s)
}

func padRight(s string, width int) string {
	padding := width - displayWidth(s)
	if padding <= 0 {
		return s
	}
	return s + strings.Repeat(" ", padding)
}
