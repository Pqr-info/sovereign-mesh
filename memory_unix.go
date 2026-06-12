//go:build !windows

package sovereign

import (
	"os"
	"syscall"
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
