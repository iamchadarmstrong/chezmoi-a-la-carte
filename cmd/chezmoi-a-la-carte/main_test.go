package main

import (
	"strings"
	"testing"

	"a-la-carte/internal/app"
)

// Helper: create a minimal manifest for testing
func testManifest() app.Manifest {
	return app.Manifest{
		"foo": {Name: "Foo", Desc: "Foo desc", Bin: []string{"foo"}},
		"bar": {Name: "Bar", Desc: "Bar desc", Bin: []string{"bar"}},
		"baz": {Name: "Baz", Desc: "Baz desc", Bin: []string{"baz"}},
	}
}

func newTestModel() model {
	manifest := testManifest()
	var keys []string
	for k := range manifest {
		keys = append(keys, k)
	}
	return model{
		manifest: manifest,
		entries:  keys,
		visible:  keys,
		selected: 0,
	}
}

func SkipTuiTestListAlwaysFixedHeight(t *testing.T) {
	m := newTestModel()
	view := m.View()
	lines := strings.Split(view, "\n")
	// The list area is always listHeight lines (search bar + list)
	listLines := 2 + listHeight // header + search + list
	if len(lines) < listLines {
		t.Fatalf("expected at least %d lines, got %d", listLines, len(lines))
	}
}

func TestListPaddingWhenFiltered(t *testing.T) {
	m := newTestModel()
	m.search = "foo"
	m.filter()
	view := m.View()
	lines := strings.Split(view, "\n")
	// List area should still be padded to listHeight
	listStart := 2 // header + search
	listEnd := listStart + listHeight
	listBlock := lines[listStart:listEnd]
	found := 0
	for _, l := range listBlock {
		if strings.Contains(l, "Foo") {
			found++
		}
	}
	if found != 1 {
		t.Errorf("expected 1 visible entry with 'Foo', got %d", found)
	}
}

func TestNoResultsMessageAndDetailsPlaceholder(t *testing.T) {
	m := newTestModel()
	m.search = "zzzz"
	m.filter()
	view := m.View()
	if !strings.Contains(view, "No results found") {
		t.Error("no results message not shown")
	}
	if !strings.Contains(view, "No details available.") {
		t.Error("details placeholder not shown")
	}
}

func SkipTuiTestDetailsPanelFixedHeight(t *testing.T) {
	m := newTestModel()
	view := m.View()
	lines := strings.Split(view, "\n")
	// Find details panel start (look for "Details")
	detailIdx := -1
	for i, l := range lines {
		if strings.Contains(l, "Details") {
			detailIdx = i
			break
		}
	}
	if detailIdx == -1 {
		t.Fatal("details panel not found")
	}
	panelLines := 0
	for i := detailIdx; i < len(lines) && panelLines < detailHeight; i++ {
		panelLines++
	}
	if panelLines != detailHeight {
		t.Errorf("details panel not fixed height: got %d, want %d", panelLines, detailHeight)
	}
}

func SkipTuiTestNoPanicOnEmptyList() {
	m := newTestModel()
	m.visible = []string{}
	m.selected = 0
	_ = m.detailLines() // should not panic
}

func SkipTuiTestEmojiAlignment(t *testing.T) {
	m := newTestModel()
	view := m.View()
	lines := strings.Split(view, "\n")
	listStart := 2
	for i := listStart; i < listStart+listHeight; i++ {
		if i >= len(lines) {
			break
		}
		l := lines[i]
		if strings.TrimSpace(l) != "" && !strings.HasPrefix(l, "  ") {
			// All list lines should start with emoji+space
			if len([]rune(l)) < 2 || l[0] == ' ' {
				t.Errorf("list line not emoji-aligned: %q", l)
			}
		}
	}
}
