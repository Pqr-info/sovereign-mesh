package crypto

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
)

// Base27Alphabet is the canonical 27-character alphabet used in the Sovereign Mesh.
const Base27Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ_"

// Trit represents the latent quantum state in the Newtonian Trit Alphabet.
// -1 = ¡ (Discontinuity), 0 = – (Ground), 1 = 0 (Resolution)
type Trit int8

const (
	TritDiscontinuity Trit = -1 // ¡
	TritGround        Trit = 0  // –
	TritResolution    Trit = 1  // 0
)

// SymbolicChar represents a Base-27 glyph containing a latent superposed Trit state.
type SymbolicChar struct {
	Glyph      rune `json:"glyph"`
	State      Trit `json:"state"`
	Superposed bool `json:"superposed"`
}

// Collapse resolves the superposed state deterministically using a seeding byte sequence.
func (sc *SymbolicChar) Collapse(seed []byte) Trit {
	if !sc.Superposed {
		return sc.State
	}
	hash := sha256.Sum256(seed)
	// Collapse to one of the three states using modulo 3 on the digest first byte
	val := hash[0] % 3
	switch val {
	case 0:
		sc.State = TritDiscontinuity
	case 1:
		sc.State = TritGround
	case 2:
		sc.State = TritResolution
	}
	sc.Superposed = false
	return sc.State
}

// EncodeUint64 encodes a uint64 value into a 13-character Base-27 string.
func EncodeUint64(val uint64) string {
	res := make([]byte, 13)
	for i := 12; i >= 0; i-- {
		res[i] = Base27Alphabet[val%27]
		val /= 27
	}
	return string(res)
}

// DecodeToUint64 decodes a 13-character Base-27 string back to a uint64.
func DecodeToUint64(s string) (uint64, error) {
	if len(s) != 13 {
		return 0, errors.New("input string must be exactly 13 characters for uint64 decoding")
	}
	var res uint64
	for i := 0; i < 13; i++ {
		char := s[i]
		idx := -1
		for j := 0; j < 27; j++ {
			if Base27Alphabet[j] == char {
				idx = j
				break
			}
		}
		if idx == -1 {
			return 0, errors.New("invalid character in Base-27 string")
		}
		res = res*27 + uint64(idx)
	}
	return res, nil
}

// EncodeBytes encodes an arbitrary byte slice into Base-27 format.
// It splits the SHA-256 digest of the data into 8-byte chunks and encodes each chunk.
func EncodeBytes(b []byte) string {
	digest := sha256.Sum256(b)
	out := ""
	for i := 0; i < 32; i += 8 {
		chunkVal := new(big.Int).SetBytes(digest[i : i+8]).Uint64()
		out += EncodeUint64(chunkVal)
	}
	return out
}

// EncodeTrajectory converts a 5-D vertex structure into a 27-character trajectory hash.
func EncodeTrajectory(x, y, z, v, i float64) string {
	vertexStr := fmt.Sprintf("%f,%f,%f,%f,%f", x, y, z, v, i)
	rawEncoded := EncodeBytes([]byte(vertexStr))
	if len(rawEncoded) > 27 {
		return rawEncoded[:27]
	}
	return rawEncoded
}
