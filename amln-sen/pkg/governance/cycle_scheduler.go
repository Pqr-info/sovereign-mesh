package governance

import (
	"context"
	"fmt"
	"sync"
)

// ScheduledAction encapsulates a pending cycle-execution target
type ScheduledAction struct {
	TenantID string
	Agent    Agent
	Action   RuntimeAction
}

// CycleScheduler coordinates execution loops, cycle-by-cycle decrement scheduling, and safety gates (RE-1)
type CycleScheduler struct {
	mu          sync.Mutex
	ree         *RuntimeExecutionEngine
	queue       []ScheduledAction
	TotalBurned uint64
}

func NewCycleScheduler(ree *RuntimeExecutionEngine) *CycleScheduler {
	return &CycleScheduler{
		ree:   ree,
		queue: make([]ScheduledAction, 0),
	}
}

// QueueAction queues an action for scheduling
func (cs *CycleScheduler) QueueAction(action ScheduledAction) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.queue = append(cs.queue, action)
}

// RunCycleStep flushes and processes one scheduled action from the queue, enforcing constitutional limits
func (cs *CycleScheduler) RunCycleStep(ctx context.Context) error {
	cs.mu.Lock()
	if len(cs.queue) == 0 {
		cs.mu.Unlock()
		return nil
	}
	next := cs.queue[0]
	cs.queue = cs.queue[1:]
	cs.mu.Unlock()

	err := cs.ree.ExecuteCycle(ctx, next.TenantID, next.Agent, next.Action)
	if err != nil {
		return fmt.Errorf("CycleScheduler execution rejected: %w", err)
	}

	cs.mu.Lock()
	cs.TotalBurned++
	cs.mu.Unlock()

	return nil
}

// Clear flushes all pending executions
func (cs *CycleScheduler) Clear() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.queue = make([]ScheduledAction, 0)
}
