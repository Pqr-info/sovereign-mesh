//go:build js && wasm

package main

import (
	"fmt"
	"time"
	"math"
    "bytes"
    "encoding/json"
    "syscall/js"

    "github.com/pqr-info/substrate/timemachine"
)

var (
    events      []timemachine.TimelineEvent
    agentsState = map[string]timemachine.TimelineEvent{}
    tickCounter uint64
)


// --- Ternary core ---
type Ternary int8

const (
	TernaryNeg  Ternary = -1
	TernaryZero Ternary = 0
	TernaryPos  Ternary = 1
)

type Tensor27 struct {
	W, H, D int
	Cells   []Ternary
}

func NewTensor27(w, h, d int) *Tensor27 {
	return &Tensor27{
		W:     w,
		H:     h,
		D:     d,
		Cells: make([]Ternary, w*h*d),
	}
}

func (t *Tensor27) idx(x, y, z int) int {
	return (z*t.H + y)*t.W + x
}

func ternaryActivation(sum int) Ternary {
	switch {
	case sum > 0:
		return TernaryPos
	case sum < 0:
		return TernaryNeg
	default:
		return TernaryZero
	}
}

func (t *Tensor27) Step() {
	next := make([]Ternary, len(t.Cells))

	for z := 0; z < t.D; z++ {
		for y := 0; y < t.H; y++ {
			for x := 0; x < t.W; x++ {
				sum := 0
				for dz := -1; dz <= 1; dz++ {
					zz := z + dz
					if zz < 0 || zz >= t.D {
						continue
					}
					for dy := -1; dy <= 1; dy++ {
						yy := y + dy
						if yy < 0 || yy >= t.H {
							continue
						}
						for dx := -1; dx <= 1; dx++ {
							xx := x + dx
							if xx < 0 || xx >= t.W {
								continue
							}
							sum += int(t.Cells[t.idx(xx, yy, zz)])
						}
					}
				}
				next[t.idx(x, y, z)] = ternaryActivation(sum)
			}
		}
	}

	t.Cells = next
}

func (t *Tensor27) Reduce() (pos, neg, zero int32) {
	for _, c := range t.Cells {
		switch c {
		case TernaryPos:
			pos++
		case TernaryNeg:
			neg++
		default:
			zero++
		}
	}
	return
}

var tensor27 = NewTensor27(32, 32, 3)
var tensor27Pos, tensor27Neg, tensor27Zero int32

func sampleDensityAt(a *timemachine.Coords5D) float32 {
	return float32(a.X1 + a.X2) // Mock
}

func sampleDivergenceAt(a *timemachine.Coords5D) float32 {
	return float32(a.X3 - a.X4) // Mock
}

func sampleRhythmAt(a *timemachine.Coords5D) float32 {
	return float32(a.X5) // Mock
}

func toTernaryFromDensity(v float32) Ternary {
	switch {
	case v > 0.66:
		return TernaryPos
	case v < 0.33:
		return TernaryNeg
	default:
		return TernaryZero
	}
}

func toTernaryFromSigned(v float32, eps float32) Ternary {
	switch {
	case v > eps:
		return TernaryPos
	case v < -eps:
		return TernaryNeg
	default:
		return TernaryZero
	}
}

func toTernaryFromRhythm(v float32) Ternary {
	switch {
	case v > 0.66:
		return TernaryPos
	case v < 0.33:
		return TernaryNeg
	default:
		return TernaryZero
	}
}

func updateTensor27FromAgents() {
	if tensor27 == nil || len(agentsState) == 0 {
		return
	}

	w := tensor27.W
	h := tensor27.H

    // Flatten agents state into grid heuristically 
    // In a real scenario we map Coords5D spatial projection to 32x32 lattice
    // Here we will just fill it by iterating agents
	for _, ev := range agentsState {
        if ev.Coords5D == nil { continue }
        
        // Mock projection of X1, X2 into 0..31
        x := int((ev.Coords5D.X1 + 1.0) * 16.0)
        y := int((ev.Coords5D.X2 + 1.0) * 16.0)
        if x < 0 { x = 0 }
        if x >= w { x = w - 1 }
        if y < 0 { y = 0 }
        if y >= h { y = h - 1 }

        tensor27.Cells[tensor27.idx(x, y, 0)] = toTernaryFromDensity(sampleDensityAt(ev.Coords5D))
        tensor27.Cells[tensor27.idx(x, y, 1)] = toTernaryFromSigned(sampleDivergenceAt(ev.Coords5D), 0.05)
        tensor27.Cells[tensor27.idx(x, y, 2)] = toTernaryFromRhythm(sampleRhythmAt(ev.Coords5D))
	}

	tensor27.Step()
	tensor27Pos, tensor27Neg, tensor27Zero = tensor27.Reduce()
}

func getTensor27Summary(this js.Value, args []js.Value) interface{} {
    return []interface{}{tensor27Pos, tensor27Neg, tensor27Zero}
}


func ingestEvent(this js.Value, args []js.Value) interface{} {
    if len(args) < 1 {
        return nil
    }
    
    uint8Array := args[0]
    length := uint8Array.Get("length").Int()
    buf := make([]byte, length)
    js.CopyBytesToGo(buf, uint8Array)

    var incoming struct {
        SessionID          string                `json:"session_id"`
        AgentID            string                `json:"agent_id"`
        Page               int                   `json:"page"`
        Data               string                `json:"data"`
        Coords5D           *timemachine.Coords5D `json:"coords5d"`
        EvolutionaryVersion int                  `json:"evolutionary_version"`
        RealTimeMs         uint64                `json:"real_time_ms"`
    }

    if err := json.Unmarshal(buf, &incoming); err != nil {
        return nil
    }

    tickCounter++

    ev := timemachine.TimelineEvent{
        Tick:                tickCounter,
        RealTimeMs:          incoming.RealTimeMs,
        SessionID:           incoming.SessionID,
        AgentID:             incoming.AgentID,
        Page:                incoming.Page,
        Data:                incoming.Data,
        Coords5D:            incoming.Coords5D,
        EvolutionaryVersion: incoming.EvolutionaryVersion,
    }

    events = append(events, ev)
    agentsState[ev.AgentID] = ev
	updateTensor27FromAgents()

    return nil
}

func getAgentsSnapshot(this js.Value, args []js.Value) interface{} {
    type AgentSnapshot struct {
        AgentID             string                `json:"agent_id"`
        Coords5D            *timemachine.Coords5D `json:"coords5d"`
        EvolutionaryVersion int                   `json:"evolutionary_version"`
        Tick                uint64                `json:"tick"`
    }

    out := make([]AgentSnapshot, 0, len(agentsState))
    for _, ev := range agentsState {
        out = append(out, AgentSnapshot{
            AgentID:             ev.AgentID,
            Coords5D:            ev.Coords5D,
            EvolutionaryVersion: ev.EvolutionaryVersion,
            Tick:                ev.Tick,
        })
    }

    b, err := json.Marshal(out)
    if err != nil {
        return nil
    }
    
    uint8Array := js.Global().Get("Uint8Array").New(len(b))
    js.CopyBytesToJS(uint8Array, b)
    return uint8Array
}


// --- Control Loop ---
type ControlRegime string

const (
	RegimeNeutral   ControlRegime = "neutral"
	RegimeExplore   ControlRegime = "explore"
	RegimeStabilize ControlRegime = "stabilize"
	RegimeSync      ControlRegime = "synchronize"
)

type ControlDecision struct {
	Regime        ControlRegime
	MutationScale float32
	CouplingGain  float32
	InputBand     string
}

type ControlState struct {
	LastDecision   ControlDecision
	LastEmitMs     int64
	MinIntervalMs  int64
	TensionThresh  int32
	DesyncThresh   float32
}

var controlState = ControlState{
	MinIntervalMs: 1000,
	TensionThresh: 200,
	DesyncThresh:  0.25,
}

var fftLow, fftMid, fftHigh float32 // mocked for now

func computeTension() int32 {
	v := tensor27Pos - tensor27Neg
	if v < 0 {
		return -v
	}
	return v
}

func max3(a, b, c float32) float32 {
	if a > b {
		if a > c { return a }
		return c
	}
	if b > c { return b }
	return c
}

func computeRhythmDesync() float32 {
	sum := fftLow + fftMid + fftHigh
	if sum <= 0 {
		return 0
	}
	target := sum / 3
	d0 := float32(math.Abs(float64(fftLow - target)))
	d1 := float32(math.Abs(float64(fftMid - target)))
	d2 := float32(math.Abs(float64(fftHigh - target)))
	maxDev := max3(d0, d1, d2)
	return maxDev / sum
}

func decideControl(nowMs int64) *ControlDecision {
	if nowMs-controlState.LastEmitMs < controlState.MinIntervalMs {
		return nil
	}

	tension := computeTension()
	desync := computeRhythmDesync()

	decision := ControlDecision{
		Regime:        RegimeNeutral,
		MutationScale: 1.0,
		CouplingGain:  1.0,
		InputBand:     "mid",
	}

	if tension > controlState.TensionThresh {
		decision.Regime = RegimeStabilize
		decision.MutationScale = 0.6
		decision.CouplingGain = 1.2
		decision.InputBand = "low"
	} else if desync > controlState.DesyncThresh {
		decision.Regime = RegimeSync
		decision.MutationScale = 0.9
		decision.CouplingGain = 0.9
		decision.InputBand = "mid"
	} else if tension < controlState.TensionThresh/3 && desync < controlState.DesyncThresh/2 {
		decision.Regime = RegimeExplore
		decision.MutationScale = 1.3
		decision.CouplingGain = 0.8
		decision.InputBand = "high"
	} else {
		return nil
	}

	controlState.LastDecision = decision
	controlState.LastEmitMs = nowMs
	return &decision
}

func getControlDecisionJSON(this js.Value, args []js.Value) interface{} {
	nowMs := time.Now().UnixMilli()
	decision := decideControl(nowMs)
	if decision == nil {
		return nil
	}
	payload := fmt.Sprintf(
		`{"type":"control","source":"substrate","version":1,"ts_ms":%d,"intent":"adjust_regime","payload":{"regime":"%s","mutation_scale":%.3f,"coupling_gain":%.3f,"input_band":"%s"}}`,
		nowMs,
		decision.Regime,
		decision.MutationScale,
		decision.CouplingGain,
		decision.InputBand,
	)
	return js.ValueOf(payload)
}

func parseSmfToTimeline(this js.Value, args []js.Value) interface{} {
    if len(args) < 1 {
        return js.ValueOf(map[string]interface{}{
            "error": "missing argument: bytes",
        })
    }

    uint8Array := args[0]
    length := uint8Array.Get("length").Int()
    buf := make([]byte, length)
    js.CopyBytesToGo(buf, uint8Array)

    session, err := timemachine.ParseSMF(bytes.NewReader(buf))
    if err != nil {
        return js.ValueOf(map[string]interface{}{
            "error": err.Error(),
        })
    }

    agentEvents, err := timemachine.DecodeDialect(session)
    if err != nil {
        return js.ValueOf(map[string]interface{}{
            "error": err.Error(),
        })
    }

    timeline := timemachine.BuildTimeline(agentEvents, session.Division)

    // Also populate live state if user parses a full file
    for _, ev := range timeline {
        events = append(events, ev)
        agentsState[ev.AgentID] = ev
	updateTensor27FromAgents()
        if ev.Tick > tickCounter {
            tickCounter = ev.Tick
        }
    }
    updateTensor27FromAgents()

    out, err := json.Marshal(timeline)
    if err != nil {
        return js.ValueOf(map[string]interface{}{
            "error": err.Error(),
        })
    }

    return js.ValueOf(string(out))
}

func main() {
    js.Global().Set("parseSmfToTimeline", js.FuncOf(parseSmfToTimeline))
    js.Global().Set("IngestEvent", js.FuncOf(ingestEvent))
    js.Global().Set("GetAgentsSnapshot", js.FuncOf(getAgentsSnapshot))
    js.Global().Set("GetTensor27Summary", js.FuncOf(getTensor27Summary))
    js.Global().Set("GetControlDecisionJSON", js.FuncOf(getControlDecisionJSON))
    select {} 
}
