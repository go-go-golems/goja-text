package extract

import "testing"

func TestLineIndexPositions(t *testing.T) {
	li := newLineIndex("alpha\nbeta\ngamma")
	tests := []struct {
		offset int
		row    int
		col    int
	}{
		{0, 0, 0},
		{3, 0, 3},
		{5, 0, 5},
		{6, 1, 0},
		{10, 1, 4},
		{11, 2, 0},
		{12, 2, 1},
		{999, 2, 5},
	}
	for _, tt := range tests {
		got := li.position(tt.offset)
		if got.Row != tt.row || got.Col != tt.col {
			t.Fatalf("position(%d) = (%d,%d), want (%d,%d)", tt.offset, got.Row, got.Col, tt.row, tt.col)
		}
	}
}

func TestSplitSourceLines(t *testing.T) {
	lines := splitSourceLines("a\nb\nccc")
	if len(lines) != 3 {
		t.Fatalf("len(lines) = %d, want 3", len(lines))
	}
	if lines[0].Text != "a" || lines[0].StartByte != 0 || lines[0].EndByte != 1 || !lines[0].HasNewline {
		t.Fatalf("line0 = %#v", lines[0])
	}
	if lines[2].Text != "ccc" || lines[2].StartByte != 4 || lines[2].EndByte != 7 || lines[2].HasNewline {
		t.Fatalf("line2 = %#v", lines[2])
	}
}
