package timemachine

type MeshRuntime interface {
	OnSessionStart(id string) error
	ApplyMemoryEvent(ev TimelineEvent) error
	OnSessionEnd(id string) error
}

type ReplayMode int

const (
	ReplayFastForward ReplayMode = iota
	ReplayWallClock
	ReplayStep
)
