package memory_bus

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net"
	"os"
	"syscall"
	"time"
)

const (
	PageTablePath = "/dev/shm/sovereign_page_table"
	TotalBusSize  = 16 * 1024 * 1024 // 16MB
	Magic         = 0xDEADBEEF
	Port          = "11111"
)

type Header struct {
	Magic     uint32
	PageIndex uint32
	Offset    uint32
	PageSize  uint32
	Checksum  uint32
}

func main() {
	fmt.Printf("SOVEREIGN SYSTEM - GO MEMORY BUS\n")

	// 1. Initialize Shared Memory
	f, err := os.OpenFile(PageTablePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Failed to open shared memory: %v", err)
	}
	defer f.Close()

	if err := f.Truncate(TotalBusSize); err != nil {
		log.Fatalf("Failed to truncate: %v", err)
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, TotalBusSize, syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("Mmap failed: %v", err)
	}
	defer syscall.Munmap(data)

	// 2. Start TCP Server
	ln, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatalf("Listen failed: %v", err)
	}
	fmt.Printf("Listening on port %s...\n", Port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn, data)
	}
}

func handleConnection(conn net.Conn, mmapData []byte) {
	defer conn.Close()
	startTime := time.Now()
	totalBytes := 0
	pages := 0

	// Disable Nagle's algorithm for low latency
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}

	for {
		var head Header
		err := binary.Read(conn, binary.BigEndian, &head)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error: %v\n", err)
			}
			break
		}

		if head.Magic != Magic {
			fmt.Printf("Invalid magic number!\n")
			break
		}

		payload := make([]byte, head.PageSize)
		_, err = io.ReadFull(conn, payload)
		if err != nil {
			break
		}

		// Verify Checksum
		if crc32.ChecksumIEEE(payload) != head.Checksum {
			fmt.Printf("CRC mismatch on page %d\n", head.PageIndex)
			continue
		}

		// Zero-copy write to mmap
		if int(head.Offset+head.PageSize) <= len(mmapData) {
			copy(mmapData[head.Offset:head.Offset+head.PageSize], payload)
		}

		totalBytes += int(head.PageSize)
		pages++
	}

	elapsed := time.Since(startTime)
	fmt.Printf("Session closed. Synced %d pages (%d bytes) in %v\n", pages, totalBytes, elapsed)
}
