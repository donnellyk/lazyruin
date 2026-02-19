package context

import "testing"

// simpleList is a test double for ICursorList.
type simpleList struct{ n int }

func (s *simpleList) Len() int { return s.n }

func newCursor(n, idx int) *ListCursor {
	c := NewListCursor(&simpleList{n})
	c.selectedLineIdx = idx
	return c
}

func TestListCursor_GetSet(t *testing.T) {
	c := newCursor(3, 1)
	if c.GetSelectedLineIdx() != 1 {
		t.Errorf("got %d, want 1", c.GetSelectedLineIdx())
	}
	c.SetSelectedLineIdx(2)
	if c.GetSelectedLineIdx() != 2 {
		t.Errorf("got %d, want 2", c.GetSelectedLineIdx())
	}
}

func TestListCursor_MoveSelectedLine(t *testing.T) {
	tests := []struct {
		name  string
		n     int // list length
		start int // starting index
		delta int
		want  int
	}{
		{"forward normal", 5, 2, 1, 3},
		{"back normal", 5, 2, -1, 1},
		{"forward at end clamps", 5, 4, 1, 4},
		{"back at start clamps", 5, 0, -1, 0},
		{"large positive clamps", 5, 1, 100, 4},
		{"large negative clamps", 5, 3, -100, 0},
		// Empty list: MoveSelectedLine only clamps when length > 0.
		// With length == 0, no clamp is applied, so the index is allowed to be 1.
		// Callers should use ClampSelection() after refreshing data.
		{"empty list delta ignored by clamp", 0, 0, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCursor(tt.n, tt.start)
			c.MoveSelectedLine(tt.delta)
			if got := c.GetSelectedLineIdx(); got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestListCursor_ClampSelection(t *testing.T) {
	tests := []struct {
		name  string
		n     int
		start int
		want  int
	}{
		{"in range unchanged", 5, 2, 2},
		{"at last index unchanged", 5, 4, 4},
		{"over length clamps to last", 5, 10, 4},
		{"negative clamps to 0", 5, -3, 0},
		{"empty list resets to 0", 0, 3, 0},
		{"empty list negative resets to 0", 0, -1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCursor(tt.n, tt.start)
			c.ClampSelection()
			if got := c.GetSelectedLineIdx(); got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
