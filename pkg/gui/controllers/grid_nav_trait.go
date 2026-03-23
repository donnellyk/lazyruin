package controllers

// GridNavTrait provides directional grid navigation (up/down/left/right)
// parameterised by horizontal and vertical deltas and a moveDay callback.
type GridNavTrait struct {
	c       *ControllerCommon
	hDelta  int // horizontal: left = -hDelta, right = +hDelta
	vDelta  int // vertical:   up   = -vDelta, down  = +vDelta
	moveDay func(int)
}

func (t *GridNavTrait) gridLeft() error  { t.moveDay(-t.hDelta); return nil }
func (t *GridNavTrait) gridRight() error { t.moveDay(t.hDelta); return nil }
func (t *GridNavTrait) gridUp() error    { t.moveDay(-t.vDelta); return nil }
func (t *GridNavTrait) gridDown() error  { t.moveDay(t.vDelta); return nil }
