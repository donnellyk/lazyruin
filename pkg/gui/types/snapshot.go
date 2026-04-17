package types

// Snapshot is an opaque marker type for preview context snapshots.
// Each preview context defines its own snapshot type; the Navigator never
// inspects it.
type Snapshot any

// Snapshotter is implemented by preview contexts that can capture and
// restore their state for back/forward navigation.
type Snapshotter interface {
	CaptureSnapshot() Snapshot
	RestoreSnapshot(Snapshot) error
}
