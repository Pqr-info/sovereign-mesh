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
