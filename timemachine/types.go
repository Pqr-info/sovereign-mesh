package timemachine

type ReplayEventKind int

const (
	EventMetaText ReplayEventKind = iota
	EventSysEx
	EventEndOfTrack
)

type ReplaySession struct {
	SessionID string
	Division  uint16
	Tracks    []ReplayTrack
}

type ReplayTrack struct {
	Index   int
	AgentID string
	Events  []ReplayEvent
}

type ReplayEvent struct {
	Tick       uint64
	DeltaTicks uint32
	Raw        []byte
	Kind       ReplayEventKind
}

type Coords5D struct {
	X1 float64 `json:"x1"`
	X2 float64 `json:"x2"`
	X3 float64 `json:"x3"`
	X4 float64 `json:"x4"`
	X5 float64 `json:"x5"`
}

type AgentEvent struct {
	Tick                uint64
	AgentID             string
	Page                int
	Data                string
	SessionID           string
	Coords5D            *Coords5D
	EvolutionaryVersion int
}

type TimelineEvent struct {
	Tick                uint64    `json:"tick"`
	RealTimeMs          uint64    `json:"realTimeMs"`
	SessionID           string    `json:"sessionId"`
	AgentID             string    `json:"agentId"`
	Page                int       `json:"page"`
	Data                string    `json:"data"`
	Coords5D            *Coords5D `json:"coords5d,omitempty"`
	EvolutionaryVersion int       `json:"evolutionary_version,omitempty"`
}
