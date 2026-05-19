package sovereign

import (
	"os"
	"runtime"
	"sync/atomic"
	"syscall"
	"unsafe"
)

const (
	PageTablePath = "/dev/shm/sovereign_page_table"
	BusSize       = 64 * 1024 * 1024
	STMSegment    = 16 * 1024 * 1024
)

// InitMemoryBus sets up the zero-copy shared memory segment.
func (c *Controller) InitMemoryBus() error {
	f, err := os.OpenFile(PageTablePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := f.Truncate(BusSize); err != nil {
		return err
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, BusSize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}

	c.memoryBus = data
	// STM (Short Term Memory) segment partition
	c.shortTermMemory = c.memoryBus[:STMSegment]
	return nil
}

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
