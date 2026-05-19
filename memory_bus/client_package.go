package memory_bus

import (
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"hash/crc32"
	"log"
	"net"
	"time"
)

const ClientMagic = 0xDEADBEEF

type ClientHeader struct {
	Magic     uint32
	PageIndex uint32
	Offset    uint32
	PageSize  uint32
	Checksum  uint32
}

func ClientMain() {
	host := flag.String("host", "127.0.0.1", "Target host")
	port := flag.String("port", "11111", "Target port")
	testSizeMB := flag.Int("test-size", 4, "Size in MB for synthetic test")
	pageSize := flag.Int("page-size", 4096, "Page size")
	flag.Parse()

	target := *host + ":" + *port
	conn, err := net.Dial("tcp", target)
	if err != nil {
		log.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}

	// Generate synthetic data
	dataSize := *testSizeMB * 1024 * 1024
	data := make([]byte, dataSize)
	rand.Read(data)

	totalChars := (dataSize + *pageSize - 1) / *pageSize
	start := time.Now()

	for i := 0; i < totalChars; i++ {
		offset := i * (*pageSize)
		end := offset + (*pageSize)
		if end > dataSize {
			end = dataSize
		}

		chunk := data[offset:end]
		// Padding
		if len(chunk) < *pageSize {
			padding := make([]byte, *pageSize-len(chunk))
			chunk = append(chunk, padding...)
		}

		head := Header{
			Magic:     Magic,
			PageIndex: uint32(i),
			Offset:    uint32(offset),
			PageSize:  uint32(*pageSize),
			Checksum:  crc32.ChecksumIEEE(chunk),
		}

		// Write Header
		binary.Write(conn, binary.BigEndian, head)
		// Write Payload
		conn.Write(chunk)
	}

	elapsed := time.Since(start)
	throughput := float64(*testSizeMB) / elapsed.Seconds()
	fmt.Printf("Transmission Complete!\n")
	fmt.Printf("Time: %v\n", elapsed)
	fmt.Printf("Throughput: %.2f MB/s\n", throughput)
}
