package crypto

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"time"
)

// PQR273 represents a node's canonical lineage event anchor.
type PQR273 struct {
	Index             uint64    `json:"index"`
	Timestamp         time.Time `json:"timestamp"`
	NodeAddress       string    `json:"node_address"`       // 81-character target address
	PrevHash          string    `json:"prev_hash"`          // Previous PQR hash
	BasisOfBehavior   float64   `json:"basis_of_behavior"`  // BoB ordering metric
	TensionOfBehavior float64   `json:"tension_of_behavior"` // ToB entropy metric
	AngleOfBehavior   float64   `json:"angle_of_behavior"`   // AoB conformation metric
	AgenticWeight     float64   `json:"agentic_weight"`     // Alpha weight
	Signature         string    `json:"signature"`          // Verification signature
	Hash              string    `json:"hash"`               // Event PQR-273 hash identity
}

// ComputePQR273 calculates the deterministic lineage hash for a given record.
func ComputePQR273(p *PQR273) (string, error) {
	if len(p.NodeAddress) != 81 {
		return "", errors.New("node address must be exactly 81 characters")
	}

	raw, err := json.Marshal(struct {
		Index             uint64    `json:"index"`
		Timestamp         time.Time `json:"timestamp"`
		NodeAddress       string    `json:"node_address"`
		PrevHash          string    `json:"prev_hash"`
		BasisOfBehavior   float64   `json:"basis_of_behavior"`
		TensionOfBehavior float64   `json:"tension_of_behavior"`
		AngleOfBehavior   float64   `json:"angle_of_behavior"`
		AgenticWeight     float64   `json:"agentic_weight"`
	}{
		Index:             p.Index,
		Timestamp:         p.Timestamp,
		NodeAddress:       p.NodeAddress,
		PrevHash:          p.PrevHash,
		BasisOfBehavior:   p.BasisOfBehavior,
		TensionOfBehavior: p.TensionOfBehavior,
		AngleOfBehavior:   p.AngleOfBehavior,
		AgenticWeight:     p.AgenticWeight,
	})
	if err != nil {
		return "", err
	}

	// Double SHA-256 for lineage integrity protection
	h1 := sha256.Sum256(raw)
	h2 := sha256.Sum256(h1[:])

	// Return a 27-character trajectory slice from the Base-27 representation
	return EncodeBytes(h2[:])[:27], nil
}

// VerifyPQR273 verifies if a PQR-273 record hash is mathematically valid.
func VerifyPQR273(p *PQR273) (bool, error) {
	computed, err := ComputePQR273(p)
	if err != nil {
		return false, err
	}
	return computed == p.Hash, nil
}
