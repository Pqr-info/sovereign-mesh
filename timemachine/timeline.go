package timemachine

import "sort"

func BuildTimeline(events []AgentEvent, division uint16) []TimelineEvent {
	timeline := make([]TimelineEvent, len(events))

	// By default we assume a fixed division / time conversion where 960 ticks = 1 second.
	// For standard SMF division, ticks_per_second matches division.
	// In server.py, delta_ticks = (now - last_time) * 960.
	// So ticks_per_second = 960.0.
	// Let's use 960.0 as ticks_per_second.
	ticksPerSecond := uint64(960)

	for i, ev := range events {
		ms := uint64(ev.Tick) * 1000 / ticksPerSecond
		timeline[i] = TimelineEvent{
			Tick:                ev.Tick,
			RealTimeMs:          ms,
			SessionID:           ev.SessionID,
			AgentID:             ev.AgentID,
			Page:                ev.Page,
			Data:                ev.Data,
			Coords5D:            ev.Coords5D,
			EvolutionaryVersion: ev.EvolutionaryVersion,
		}
	}

	sort.Slice(timeline, func(i, j int) bool {
		if timeline[i].Tick != timeline[j].Tick {
			return timeline[i].Tick < timeline[j].Tick
		}
		// Deterministic tie-breaker
		return timeline[i].AgentID < timeline[j].AgentID
	})

	return timeline
}
