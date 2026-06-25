package timemachine

import (
	"context"
	"time"
)

type TimeMachine struct {
	Runtime MeshRuntime
}

func NewTimeMachine(runtime MeshRuntime) *TimeMachine {
	return &TimeMachine{
		Runtime: runtime,
	}
}

func (tm *TimeMachine) Replay(ctx context.Context, session *ReplaySession, mode ReplayMode) error {
	agentEvents, err := DecodeDialect(session)
	if err != nil {
		return err
	}

	timeline := BuildTimeline(agentEvents, session.Division)

	if err := tm.Runtime.OnSessionStart(session.SessionID); err != nil {
		return err
	}

	var lastMs uint64

	for _, ev := range timeline {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		switch mode {
		case ReplayWallClock:
			if ev.RealTimeMs > lastMs {
				delta := ev.RealTimeMs - lastMs
				time.Sleep(time.Duration(delta) * time.Millisecond)
				lastMs = ev.RealTimeMs
			}
		case ReplayStep:
			// Debug step signaling or blocking can be wired here
		}

		if err := tm.Runtime.ApplyMemoryEvent(ev); err != nil {
			return err
		}
	}

	return tm.Runtime.OnSessionEnd(session.SessionID)
}
