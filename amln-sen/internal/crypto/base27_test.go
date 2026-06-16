package crypto

import (
	"testing"
)

func TestBase27Encoding(t *testing.T) {
	val := uint64(123456789012345)
	encoded := EncodeUint64(val)
	if len(encoded) != 13 {
		t.Errorf("expected length 13, got %d", len(encoded))
	}

	decoded, err := DecodeToUint64(encoded)
	if err != nil {
		t.Fatalf("failed to decode Base-27 string: %v", err)
	}

	if decoded != val {
		t.Errorf("decoded value %d doesn't match original %d", decoded, val)
	}
}

func TestSymbolicTritCollapse(t *testing.T) {
	sc := SymbolicChar{
		Glyph:      'A',
		Superposed: true,
	}

	collapsed := sc.Collapse([]byte("arbitrary-seeding-vector"))
	if sc.Superposed {
		t.Error("expected character to be collapsed")
	}

	if collapsed != TritDiscontinuity && collapsed != TritGround && collapsed != TritResolution {
		t.Errorf("unexpected collapsed trit state: %v", collapsed)
	}
}

func TestEncodeTrajectory(t *testing.T) {
	traj := EncodeTrajectory(1.0, 2.0, 3.0, 0.5, 0.9)
	if len(traj) != 27 {
		t.Errorf("expected 27-char trajectory hash, got: %d (%s)", len(traj), traj)
	}
}
