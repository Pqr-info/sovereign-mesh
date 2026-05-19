package sovereign

import (
	"sync"
	"testing"
)

func TestAgentStateLock(t *testing.T) {
	// Initialize a blank AgentState.
	// This simulates a struct pointer obtained via unsafe pointer casting in shared memory.
	state := &AgentState{}

	// 1. Basic Lock/Unlock sanity check
	state.Lock()
	if state.Mutex != 1 {
		t.Errorf("Expected Mutex to be 1 after Lock, got %d", state.Mutex)
	}
	state.Unlock()
	if state.Mutex != 0 {
		t.Errorf("Expected Mutex to be 0 after Unlock, got %d", state.Mutex)
	}

	// 2. High-concurrency race test
	// We use a shared counter protected by the AgentState's atomic spinlock.
	// If the lock fails to provide mutual exclusion, the final counter will be incorrect.
	var wg sync.WaitGroup
	sharedCounter := 0
	numGoroutines := 100
	iterations := 1000

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				state.Lock()
				sharedCounter++
				state.Unlock()
			}
		}()
	}

	wg.Wait()

	expected := numGoroutines * iterations
	if sharedCounter != expected {
		t.Errorf("Race condition detected! Expected counter %d, got %d", expected, sharedCounter)
	}
}

func TestGetAgentStateReflection(t *testing.T) {
	// Verify that multiple pointers to the same memory offset reflect shared state change.
	bus := make([]byte, 1024)
	c := &Controller{memoryBus: bus}

	s1 := c.GetAgentState(0)
	s2 := c.GetAgentState(0)

	s1.Active = true
	if !s2.Active {
		t.Error("Pointer reflection failed: changes to s1 were not visible to s2")
	}
}
