package crypto

import (
	"testing"
	"time"
)

func TestPQR273GenerationAndVerification(t *testing.T) {
	nodeAddr := "addr12345678901234567890123456789012345678901234567890123456789012345678901234567"
	p := &PQR273{
		Index:             1,
		Timestamp:         time.Now().UTC(),
		NodeAddress:       nodeAddr,
		PrevHash:          "PQR-GENESIS-ANCHOR-BASE",
		BasisOfBehavior:   0.8,
		TensionOfBehavior: 0.1,
		AngleOfBehavior:   0.5,
		AgenticWeight:     0.9,
	}

	hash, err := ComputePQR273(p)
	if err != nil {
		t.Fatalf("failed to compute PQR-273 hash: %v", err)
	}

	if len(hash) != 27 {
		t.Errorf("expected 27-character trajectory hash slice, got length %d (%s)", len(hash), hash)
	}

	p.Hash = hash
	valid, err := VerifyPQR273(p)
	if err != nil {
		t.Fatalf("verification execution errored: %v", err)
	}
	if !valid {
		t.Error("expected computed PQR-273 hash to verify successfully")
	}

	// Tamper with the lineage properties to verify failure detection
	p.BasisOfBehavior = 0.0
	validTampered, _ := VerifyPQR273(p)
	if validTampered {
		t.Error("expected tampered lineage verification to fail, but it verified successfully")
	}
}
