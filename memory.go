package sovereign

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

// Lock shared state using atomic spinlock across processes.
func (s *AgentState) Lock() {
	for !atomic.CompareAndSwapUint32(&s.Mutex, 0, 1) {
		runtime.Gosched()
	}
}

func (s *AgentState) Unlock() {
	atomic.StoreUint32(&s.Mutex, 0)
}

// GetAgentState returns a direct pointer into the memory-mapped bus.
func (c *Controller) GetAgentState(offset int) *AgentState {
	return (*AgentState)(unsafe.Pointer(&c.memoryBus[offset]))
}
