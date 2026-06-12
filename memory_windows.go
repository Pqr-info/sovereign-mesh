//go:build windows

package sovereign

const (
	BusSize    = 64 * 1024 * 1024
	STMSegment = 16 * 1024 * 1024
)

// InitMemoryBus sets up the simulated zero-copy shared memory segment on Windows.
func (c *Controller) InitMemoryBus() error {
	c.memoryBus = make([]byte, BusSize)
	// STM (Short Term Memory) segment partition
	c.shortTermMemory = c.memoryBus[:STMSegment]
	return nil
}
