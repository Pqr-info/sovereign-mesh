package addressing

import (
	"bytes"
	"testing"
	"time"
)

func TestBase27RoundTrip(t *testing.T) {
	for _, v := range []int64{0, 1, 27, 12345, 987654321} {
		s := EncodeBase27(v)
		got, err := DecodeBase27(s)
		if err != nil {
			t.Fatalf("DecodeBase27(%q): %v", s, err)
		}
		if got != v {
			t.Fatalf("round-trip mismatch: want %d got %d", v, got)
		}
	}
}

func TestSerializeDeserializeCoord(t *testing.T) {
	a := NewAddress5D()
	coord := a.Resolve("TEST_ASSET", time.Unix(0, 1234567890))
	b, err := SerializeCoord(coord)
	if err != nil {
		t.Fatalf("SerializeCoord: %v", err)
	}
	got, err := DeserializeCoord(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("DeserializeCoord: %v", err)
	}
	if coord != got {
		t.Fatalf("coord mismatch: want %+v got %+v", coord, got)
	}
}

func TestMapProteinTo5D(t *testing.T) {
	// 1. Identical inputs yield identical coordinates (determinism)
	id1 := "P04637" // p53
	seq1 := "MEEPQSDPSVEPPLSQETFSDLWKLLPENNVLSPLPSQAMDDLMLSPDDIEQWFTEDPGPDEAPRMPEAAPPVAPAPAAPTPAAPAPAPSWPLSSSVPSQKTYQGSYGFRLGFLHSGTAKSVTCTYSPALNKMFCQLAKTCPVQLWVDSTPPPGTRVRAMAIYKQSQHMTEVVRRCPHHERCSDSDGLAPPQHLIRVEGNLRVEYLDDRNTFRHSVVVPYEPPEVGSDCTTIHYNYMCNSSCMGGMNRRPILTIITLEDSSGNLLGRNSFEVRVCACPGRDRRTEEENLRKKGEPHHELPPGSTKRALPNNTSSSPQPKKKPLDGEYFTLQIRGRERFEMFRELNEALELKDAQAGKEPGGSRAHSSHLKSKKGQSTSRHKKLMFKTEGPDSD"
	
	coord1 := MapProteinTo5D(id1, seq1)
	coord2 := MapProteinTo5D(id1, seq1)
	
	if coord1 != coord2 {
		t.Fatalf("Nondeterministic mapping: %+v != %+v", coord1, coord2)
	}

	// Verify value bounds
	if coord1.X < 0 || coord1.X > 1023 {
		t.Fatalf("X coordinate out of bounds: %d", coord1.X)
	}
	if coord1.Y < 0 || coord1.Y > 1023 {
		t.Fatalf("Y coordinate out of bounds: %d", coord1.Y)
	}
	if coord1.Z < 0 || coord1.Z > 1023 {
		t.Fatalf("Z coordinate out of bounds: %d", coord1.Z)
	}
	if coord1.Psi < 0 || coord1.Psi > 26 {
		t.Fatalf("Psi coordinate out of bounds: %d", coord1.Psi)
	}

	// 2. Distinct inputs yield distinct coordinates
	id2 := "P62158" // Calmodulin
	seq2 := "MADQLTEEQIAEFKEAFSLFDKDGDGTITTKELGTVMRSLGQNPTEAELQDMINEVDADGDGTIDFPEFLTMMARKMKDTDSEEEIREAFRVFDKDGNGYISAAELRHVMTNLGEKLTDEEVDEMIREADIDGDGQVNYEEFVQMMTAK"
	
	coord3 := MapProteinTo5D(id2, seq2)
	if coord1 == coord3 {
		t.Fatalf("Collision detected between different proteins: %+v", coord1)
	}

	// 3. Edge case: empty sequence
	coordEmpty := MapProteinTo5D(id1, "")
	if coordEmpty.X != 0 || coordEmpty.Y != 0 || coordEmpty.Z != 0 {
		t.Fatalf("Empty sequence did not yield zeroed coordinates: %+v", coordEmpty)
	}
}

func TestAffinityAndPartners(t *testing.T) {
	// 1. Self affinity should be perfect (1.0)
	c1 := Address5DCoord{X: 100, Y: 200, Z: 300, T: 1000, Psi: 5}
	selfScore := CalculateAffinity(c1, c1)
	if selfScore != 1.0 {
		t.Fatalf("Self affinity not 1.0: %v", selfScore)
	}

	// 2. High distance yields lower affinity
	c2 := Address5DCoord{X: 900, Y: 800, Z: 700, T: 1000, Psi: 5}
	c3 := Address5DCoord{X: 120, Y: 210, Z: 310, T: 1000, Psi: 5}
	
	scoreFar := CalculateAffinity(c1, c2)
	scoreNear := CalculateAffinity(c1, c3)
	
	if scoreFar >= scoreNear {
		t.Fatalf("Far affinity (%v) should be less than near affinity (%v)", scoreFar, scoreNear)
	}

	// 3. Phase harmonics boost
	cNearNoPsi := Address5DCoord{X: 120, Y: 210, Z: 310, T: 1000, Psi: 20}
	scoreNearNoPsi := CalculateAffinity(c1, cNearNoPsi)
	
	// scoreNear has same Psi (5 == 5 -> harmonic), scoreNearNoPsi has distant Psi (5 vs 20 -> not harmonic)
	if scoreNear <= scoreNearNoPsi {
		t.Fatalf("Harmonic phase alignment should boost affinity: %v vs %v", scoreNear, scoreNearNoPsi)
	}

	// 4. Partner Sorting
	pool := []Address5DCoord{c2, c3, cNearNoPsi}
	partners := FindBestPartners(c1, pool, 2)
	
	if len(partners) != 2 {
		t.Fatalf("Expected 2 partners, got %d", len(partners))
	}
	
	// c3 (near, same Psi) should be the best partner
	if partners[0] != c3 {
		t.Fatalf("Best partner should be c3: got %+v", partners[0])
	}
}


