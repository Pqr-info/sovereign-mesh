package addressing

import (
	"crypto/sha1"
	"encoding/binary"
	"math"
	"math/big"
	"strings"
)

// Kyte-Doolittle Hydrophobicity values scaled from -4.5 to 4.5
var kdScale = map[rune]float64{
	'A': 1.8, 'R': -4.5, 'N': -3.5, 'D': -3.5, 'C': 2.5,
	'Q': -3.5, 'E': -3.5, 'G': -0.4, 'H': -3.2, 'I': 4.5,
	'L': 3.8, 'K': -3.9, 'M': 1.9, 'F': 2.8, 'P': -1.6,
	'S': -0.8, 'T': -0.7, 'W': -0.9, 'Y': -1.3, 'V': 4.2,
}

// PhaseFromUniProt maps a UniProt ID to a stable Psi in [0,1]
func PhaseFromUniProt(uniprotID string) float64 {
	h := sha1.Sum([]byte(uniprotID))

	// Interpret SHA-1 as big integer
	hashInt := new(big.Int).SetBytes(h[:])

	// Max value for 160-bit number: 2^160 - 1
	maxInt := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1))

	// Convert to float64 in [0,1]
	ratio, _ := new(big.Rat).SetFrac(hashInt, maxInt).Float64()

	return roundFloat(ratio, 9)
}

// PhaseRadiansFromUniProt maps UniProt ID to Psi in [0, 2π)
func PhaseRadiansFromUniProt(uniprotID string) float64 {
	return roundFloat(PhaseFromUniProt(uniprotID)*2*math.Pi, 9)
}

// MapProteinTo5D translates a protein's UniProt ID and FASTA sequence into a stable 5D Address5DCoord.
// The output coordinates are deterministic and scaled to fit the 1024-bound strategy grids.
func MapProteinTo5D(uniprotID string, sequence string) Address5DCoord {
	cleanSeq := strings.ToUpper(strings.TrimSpace(sequence))
	seqLen := len(cleanSeq)

	// Fallbacks for edge cases
	if seqLen == 0 {
		h := sha1.Sum([]byte(uniprotID))
		v := binary.BigEndian.Uint64(h[:8])
		return Address5DCoord{
			X:   0,
			Y:   0,
			Z:   0,
			T:   int64(v & 0x7FFFFFFFFFFFFFFF),
			Psi: int64(PhaseFromUniProt(uniprotID) * 26.0),
		}
	}

	// 1. Calculate X: Hydrophobicity / Charge Index (Kyte-Doolittle)
	totalHydro := 0.0
	for _, char := range cleanSeq {
		if val, ok := kdScale[char]; ok {
			totalHydro += val
		}
	}
	avgHydro := totalHydro / float64(seqLen)
	// Map Kyte-Doolittle [-4.5, 4.5] to [0, 1023]
	normHydro := (avgHydro + 4.5) / 9.0
	if normHydro < 0.0 {
		normHydro = 0.0
	} else if normHydro > 1.0 {
		normHydro = 1.0
	}
	coordX := int64(normHydro * 1023)

	// 2. Calculate Y: Fold Complexity (Sequence Length / Molecular Weight proxy)
	// Cap length at 5000 residues and scale to [0, 1023]
	coordY := int64(seqLen)
	if coordY > 1023 {
		coordY = 1023
	}

	// 3. Calculate Z: Folding Stability Approximation
	// Percentage of stabilizing residues (Proline + Cysteine + Aromatic: F, W, Y)
	stabilizingCount := 0
	for _, char := range cleanSeq {
		switch char {
		case 'P', 'C', 'F', 'W', 'Y':
			stabilizingCount++
		}
	}
	ratioStability := float64(stabilizingCount) / float64(seqLen)
	coordZ := int64(ratioStability * 1023)

	// 4. Calculate T: Epoch / Temporal alignment
	// Hash UniProt ID to generate a stable, pseudo-temporal timestamp bucket
	hID := sha1.Sum([]byte(uniprotID))
	vID := binary.BigEndian.Uint64(hID[:8])
	// Let's create a deterministic T coordinate (nanoseconds range)
	coordT := int64(vID & 0x7FFFFFFFFFFFFFFF)

	// 5. Calculate Psi: Phase angle / Hyperplane Index using the PhaseFromUniProt ratio
	// Map to [0, 26] (27-dimensional phase space)
	coordPsi := int64(PhaseFromUniProt(uniprotID) * 26.0)

	return Address5DCoord{
		X:   coordX,
		Y:   coordY,
		Z:   coordZ,
		T:   coordT,
		Psi: coordPsi,
	}
}

