package tui

import "testing"

func TestComputeDiff_BothEmpty(t *testing.T) {
	result := ComputeDiff("", "")
	if len(result.Lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(result.Lines))
	}
	if result.Stats.Additions != 0 || result.Stats.Deletions != 0 || result.Stats.Equal != 0 {
		t.Errorf("expected zero stats, got %+v", result.Stats)
	}
}

func TestComputeDiff_IdenticalTexts(t *testing.T) {
	text := "line1\nline2\nline3\n"
	result := ComputeDiff(text, text)

	if result.Stats.Additions != 0 {
		t.Errorf("additions = %d, want 0", result.Stats.Additions)
	}
	if result.Stats.Deletions != 0 {
		t.Errorf("deletions = %d, want 0", result.Stats.Deletions)
	}
	if result.Stats.Equal != 3 {
		t.Errorf("equal = %d, want 3", result.Stats.Equal)
	}

	for i, dl := range result.Lines {
		if dl.Type != DiffEqual {
			t.Errorf("line %d: type = %v, want DiffEqual", i, dl.Type)
		}
		if dl.OldLineNo == 0 || dl.NewLineNo == 0 {
			t.Errorf("line %d: line numbers should be non-zero for equal lines", i)
		}
	}
}

func TestComputeDiff_OldEmpty(t *testing.T) {
	result := ComputeDiff("", "alpha\nbeta\n")

	if result.Stats.Additions != 2 {
		t.Errorf("additions = %d, want 2", result.Stats.Additions)
	}
	if result.Stats.Deletions != 0 {
		t.Errorf("deletions = %d, want 0", result.Stats.Deletions)
	}
	for _, dl := range result.Lines {
		if dl.Type != DiffInsert {
			t.Errorf("expected DiffInsert, got %v", dl.Type)
		}
		if dl.OldLineNo != 0 {
			t.Errorf("insert line should have OldLineNo=0, got %d", dl.OldLineNo)
		}
	}
}

func TestComputeDiff_NewEmpty(t *testing.T) {
	result := ComputeDiff("alpha\nbeta\n", "")

	if result.Stats.Deletions != 2 {
		t.Errorf("deletions = %d, want 2", result.Stats.Deletions)
	}
	if result.Stats.Additions != 0 {
		t.Errorf("additions = %d, want 0", result.Stats.Additions)
	}
	for _, dl := range result.Lines {
		if dl.Type != DiffDelete {
			t.Errorf("expected DiffDelete, got %v", dl.Type)
		}
		if dl.NewLineNo != 0 {
			t.Errorf("delete line should have NewLineNo=0, got %d", dl.NewLineNo)
		}
	}
}

func TestComputeDiff_SingleLineChange(t *testing.T) {
	result := ComputeDiff("hello world\n", "hello earth\n")

	if result.Stats.Additions != 1 {
		t.Errorf("additions = %d, want 1", result.Stats.Additions)
	}
	if result.Stats.Deletions != 1 {
		t.Errorf("deletions = %d, want 1", result.Stats.Deletions)
	}

	// Should have word-level segments for the paired change
	foundDelete := false
	foundInsert := false
	for _, dl := range result.Lines {
		switch dl.Type {
		case DiffDelete:
			foundDelete = true
			if len(dl.Segments) == 0 {
				t.Error("delete line should have word-level segments")
			}
		case DiffInsert:
			foundInsert = true
			if len(dl.Segments) == 0 {
				t.Error("insert line should have word-level segments")
			}
		}
	}
	if !foundDelete {
		t.Error("expected a DiffDelete line")
	}
	if !foundInsert {
		t.Error("expected a DiffInsert line")
	}
}

func TestComputeDiff_MultiLineWithContext(t *testing.T) {
	oldText := "line1\nline2\nline3\nline4\nline5\n"
	newText := "line1\nLINE2\nline3\nline4\nline5\nextra\n"

	result := ComputeDiff(oldText, newText)

	if result.Stats.Equal != 4 {
		t.Errorf("equal = %d, want 4", result.Stats.Equal)
	}
	// line2 -> LINE2 = 1 deletion + 1 addition; extra = 1 addition
	if result.Stats.Additions != 2 {
		t.Errorf("additions = %d, want 2", result.Stats.Additions)
	}
	if result.Stats.Deletions != 1 {
		t.Errorf("deletions = %d, want 1", result.Stats.Deletions)
	}
}

func TestComputeDiff_PureAdditions(t *testing.T) {
	oldText := "aaa\nccc\n"
	newText := "aaa\nbbb\nccc\n"
	result := ComputeDiff(oldText, newText)

	if result.Stats.Additions != 1 {
		t.Errorf("additions = %d, want 1", result.Stats.Additions)
	}
	if result.Stats.Deletions != 0 {
		t.Errorf("deletions = %d, want 0", result.Stats.Deletions)
	}
	if result.Stats.Equal != 2 {
		t.Errorf("equal = %d, want 2", result.Stats.Equal)
	}
}

func TestComputeDiff_PureDeletions(t *testing.T) {
	oldText := "aaa\nbbb\nccc\n"
	newText := "aaa\nccc\n"
	result := ComputeDiff(oldText, newText)

	if result.Stats.Deletions != 1 {
		t.Errorf("deletions = %d, want 1", result.Stats.Deletions)
	}
	if result.Stats.Additions != 0 {
		t.Errorf("additions = %d, want 0", result.Stats.Additions)
	}
}

func TestComputeDiff_TrailingNewline(t *testing.T) {
	// With and without trailing newline should produce same diff for content
	withNewline := ComputeDiff("a\nb\n", "a\nc\n")
	withoutNewline := ComputeDiff("a\nb", "a\nc")

	if withNewline.Stats != withoutNewline.Stats {
		t.Errorf("trailing newline should not affect diff stats: with=%+v without=%+v",
			withNewline.Stats, withoutNewline.Stats)
	}
}

func TestComputeDiff_LineNumbersMonotonic(t *testing.T) {
	oldText := "a\nb\nc\nd\n"
	newText := "a\nX\nY\nc\nd\nZ\n"
	result := ComputeDiff(oldText, newText)

	lastOld, lastNew := 0, 0
	for _, dl := range result.Lines {
		if dl.OldLineNo > 0 {
			if dl.OldLineNo < lastOld {
				t.Errorf("OldLineNo went backwards: %d after %d", dl.OldLineNo, lastOld)
			}
			lastOld = dl.OldLineNo
		}
		if dl.NewLineNo > 0 {
			if dl.NewLineNo < lastNew {
				t.Errorf("NewLineNo went backwards: %d after %d", dl.NewLineNo, lastNew)
			}
			lastNew = dl.NewLineNo
		}
	}
}

func TestSplitLines_Empty(t *testing.T) {
	lines := splitLines("")
	if lines != nil {
		t.Errorf("expected nil for empty string, got %v", lines)
	}
}

func TestSplitLines_TrailingNewline(t *testing.T) {
	lines := splitLines("a\nb\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %v", len(lines), lines)
	}
}

func TestSplitLines_NoTrailingNewline(t *testing.T) {
	lines := splitLines("a\nb")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %v", len(lines), lines)
	}
}

func TestComputeWordSegments_PartialChange(t *testing.T) {
	oldSegs, newSegs := computeWordSegments("hello world", "hello earth")

	if len(oldSegs) == 0 || len(newSegs) == 0 {
		t.Fatal("expected non-empty segments")
	}

	// Should have at least one unchanged segment ("hello ") and one changed segment
	hasUnchanged := false
	hasChanged := false
	for _, s := range oldSegs {
		if s.Changed {
			hasChanged = true
		} else {
			hasUnchanged = true
		}
	}
	if !hasUnchanged || !hasChanged {
		t.Errorf("old segments should have both changed and unchanged parts, got %+v", oldSegs)
	}
}

func TestComputeWordSegments_IdenticalLines(t *testing.T) {
	oldSegs, newSegs := computeWordSegments("same text", "same text")

	// All segments should be unchanged
	for _, s := range oldSegs {
		if s.Changed {
			t.Errorf("old segment should not be changed: %+v", s)
		}
	}
	for _, s := range newSegs {
		if s.Changed {
			t.Errorf("new segment should not be changed: %+v", s)
		}
	}
}
