package addressing

import (
	"bytes"
	"encoding/binary"
	"io"
)

func SerializeCoord(c Address5DCoord) ([]byte, error) {
	buf := &bytes.Buffer{}
	// 5 x int64 = 40 bytes
	fields := []int64{c.X, c.Y, c.Z, c.T, c.Psi}
	for _, f := range fields {
		if err := binary.Write(buf, binary.BigEndian, f); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func DeserializeCoord(r io.Reader) (Address5DCoord, error) {
	var fields [5]int64
	for i := range fields {
		if err := binary.Read(r, binary.BigEndian, &fields[i]); err != nil {
			return Address5DCoord{}, err
		}
	}
	return Address5DCoord{
		X:   fields[0],
		Y:   fields[1],
		Z:   fields[2],
		T:   fields[3],
		Psi: fields[4],
	}, nil
}
