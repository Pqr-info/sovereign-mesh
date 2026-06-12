package addressing

import (
	"errors"
	"strings"
)

const Alphabet27 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ_"

var (
	base = int64(len(Alphabet27))
)

func EncodeBase27(n int64) string {
	if n == 0 {
		return string(Alphabet27[0])
	}
	var b strings.Builder
	for n > 0 {
		r := n % base
		b.WriteByte(Alphabet27[r])
		n /= base
	}
	s := b.String()
	// reverse
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func DecodeBase27(s string) (int64, error) {
	var n int64
	for _, ch := range s {
		idx := strings.IndexRune(Alphabet27, ch)
		if idx < 0 {
			return 0, errors.New("invalid base27 character")
		}
		n = n*base + int64(idx)
	}
	return n, nil
}
