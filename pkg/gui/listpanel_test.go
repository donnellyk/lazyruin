package gui

import "testing"

func TestListMove(t *testing.T) {
	tests := []struct {
		name        string
		startIdx    int
		count       int
		delta       int
		wantIdx     int
		wantChanged bool
	}{
		{"move down", 0, 5, 1, 1, true},
		{"move up", 2, 5, -1, 1, true},
		{"at bottom boundary", 4, 5, 1, 4, false},
		{"at top boundary", 0, 5, -1, 0, false},
		{"empty list down", 0, 0, 1, 0, false},
		{"empty list up", 0, 0, -1, 0, false},
		{"single item down", 0, 1, 1, 0, false},
		{"single item up", 0, 1, -1, 0, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			idx := tc.startIdx
			changed := listMove(&idx, tc.count, tc.delta)
			if idx != tc.wantIdx {
				t.Errorf("index = %d, want %d", idx, tc.wantIdx)
			}
			if changed != tc.wantChanged {
				t.Errorf("changed = %v, want %v", changed, tc.wantChanged)
			}
		})
	}
}

func TestListPanel_Down(t *testing.T) {
	idx := 0
	rendered, previewed := false, false
	lp := &listPanel{
		selectedIndex: &idx,
		itemCount:     func() int { return 5 },
		render:        func() { rendered = true },
		updatePreview: func() { previewed = true },
	}

	err := lp.listDown(nil, nil)
	if err != nil {
		t.Fatalf("listDown error: %v", err)
	}
	if idx != 1 {
		t.Errorf("index = %d, want 1", idx)
	}
	if !rendered {
		t.Error("render was not called")
	}
	if !previewed {
		t.Error("updatePreview was not called")
	}
}

func TestListPanel_Down_AtBoundary(t *testing.T) {
	idx := 4
	rendered := false
	lp := &listPanel{
		selectedIndex: &idx,
		itemCount:     func() int { return 5 },
		render:        func() { rendered = true },
		updatePreview: func() {},
	}

	lp.listDown(nil, nil)
	if idx != 4 {
		t.Errorf("index = %d, want 4 (should not change)", idx)
	}
	if rendered {
		t.Error("render should not be called when at boundary")
	}
}

func TestListPanel_Up(t *testing.T) {
	idx := 3
	rendered := false
	lp := &listPanel{
		selectedIndex: &idx,
		itemCount:     func() int { return 5 },
		render:        func() { rendered = true },
		updatePreview: func() {},
	}

	lp.listUp(nil, nil)
	if idx != 2 {
		t.Errorf("index = %d, want 2", idx)
	}
	if !rendered {
		t.Error("render was not called")
	}
}

func TestListPanel_Up_AtBoundary(t *testing.T) {
	idx := 0
	rendered := false
	lp := &listPanel{
		selectedIndex: &idx,
		itemCount:     func() int { return 5 },
		render:        func() { rendered = true },
		updatePreview: func() {},
	}

	lp.listUp(nil, nil)
	if idx != 0 {
		t.Errorf("index = %d, want 0 (should not change)", idx)
	}
	if rendered {
		t.Error("render should not be called when at boundary")
	}
}

func TestListPanel_Top(t *testing.T) {
	idx := 3
	rendered, previewed := false, false
	lp := &listPanel{
		selectedIndex: &idx,
		itemCount:     func() int { return 5 },
		render:        func() { rendered = true },
		updatePreview: func() { previewed = true },
	}

	lp.listTop(nil, nil)
	if idx != 0 {
		t.Errorf("index = %d, want 0", idx)
	}
	if !rendered {
		t.Error("render was not called")
	}
	if !previewed {
		t.Error("updatePreview was not called")
	}
}

func TestListPanel_Bottom(t *testing.T) {
	idx := 0
	rendered, previewed := false, false
	lp := &listPanel{
		selectedIndex: &idx,
		itemCount:     func() int { return 5 },
		render:        func() { rendered = true },
		updatePreview: func() { previewed = true },
	}

	lp.listBottom(nil, nil)
	if idx != 4 {
		t.Errorf("index = %d, want 4", idx)
	}
	if !rendered {
		t.Error("render was not called")
	}
	if !previewed {
		t.Error("updatePreview was not called")
	}
}

func TestListPanel_Bottom_EmptyList(t *testing.T) {
	idx := 0
	rendered := false
	lp := &listPanel{
		selectedIndex: &idx,
		itemCount:     func() int { return 0 },
		render:        func() { rendered = true },
		updatePreview: func() {},
	}

	lp.listBottom(nil, nil)
	if idx != 0 {
		t.Errorf("index = %d, want 0 (should not change for empty list)", idx)
	}
	if rendered {
		t.Error("render should not be called for empty list")
	}
}
