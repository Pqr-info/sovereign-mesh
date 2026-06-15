package addressing

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"time"
)

type Address5DCoord struct {
	X   int64
	Y   int64
	Z   int64
	T   int64 // nanoseconds since epoch or mesh epoch
	Psi int64 // phase / hyperplane index
}

func (c Address5DCoord) String() string {
	return fmt.Sprintf("X=%d Y=%d Z=%d T=%d Ψ=%d", c.X, c.Y, c.Z, c.T, c.Psi)
}

type Address5D struct {
	NodeID    string `json:"node_id"`    // D1
	RoleID    string `json:"role_id"`    // D2
	LineageID string `json:"lineage_id"` // D3
	BlockID   string `json:"block_id"`   // D4
	ThreadID  string `json:"thread_id"`  // D5
}

func (a Address5D) String() string {
	return a.NodeID + "|" + a.RoleID + "|" + a.LineageID + "|" +
		a.BlockID + "|" + a.ThreadID
}

func NewAddress5D() *Address5D {
	return &Address5D{}
}

// Resolve maps an asset identifier + timestamp into a bounded 5-D coordinate.
func (a *Address5D) Resolve(assetID string, ts time.Time) Address5DCoord {
	t := ts.UnixNano()

	// Hash assetID for stable high-entropy mapping
	h := sha1.Sum([]byte(assetID))
	v := binary.BigEndian.Uint64(h[:8])

	return Address5DCoord{
		X:   int64(v % 1024),       // stable per asset
		Y:   int64((t / 1e6) % 1024), // ms bucket
		Z:   int64((t / 1e3) % 1024), // µs bucket
		T:   t,                      // raw nanoseconds
		Psi: int64(v % 27),         // stable phase index
	}
}
