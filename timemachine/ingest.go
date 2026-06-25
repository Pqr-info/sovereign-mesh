package timemachine

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

func readBytes(r io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(r, buf)
	return buf, err
}

func readVLQ(r io.Reader) (uint32, error) {
	var value uint32
	for i := 0; i < 4; i++ {
		b := make([]byte, 1)
		if _, err := r.Read(b); err != nil {
			return 0, err
		}
		value = (value << 7) | uint32(b[0]&0x7F)
		if b[0]&0x80 == 0 {
			break
		}
	}
	return value, nil
}

func ParseSMF(r io.Reader) (*ReplaySession, error) {
	headerID, err := readBytes(r, 4)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(headerID, []byte("MThd")) {
		return nil, errors.New("invalid SMF: missing MThd header")
	}

	headerLenBytes, err := readBytes(r, 4)
	if err != nil {
		return nil, err
	}
	headerLen := binary.BigEndian.Uint32(headerLenBytes)
	if headerLen != 6 {
		return nil, fmt.Errorf("invalid MThd length: %d", headerLen)
	}

	headerData, err := readBytes(r, 6)
	if err != nil {
		return nil, err
	}

	format := binary.BigEndian.Uint16(headerData[0:2])
	if format != 1 {
		return nil, fmt.Errorf("SMF format must be 1, got %d", format)
	}

	trackCount := binary.BigEndian.Uint16(headerData[2:4])
	division := binary.BigEndian.Uint16(headerData[4:6])

	session := &ReplaySession{
		Division: division,
		Tracks:   make([]ReplayTrack, 0, trackCount),
	}

	for t := 0; t < int(trackCount); t++ {
		trackID, err := readBytes(r, 4)
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(trackID, []byte("MTrk")) {
			return nil, fmt.Errorf("invalid SMF: missing MTrk header at track %d", t)
		}

		trackLenBytes, err := readBytes(r, 4)
		if err != nil {
			return nil, err
		}
		trackLen := binary.BigEndian.Uint32(trackLenBytes)

		trackData, err := readBytes(r, int(trackLen))
		if err != nil {
			return nil, err
		}

		tr := ReplayTrack{
			Index:  t,
			Events: make([]ReplayEvent, 0),
		}

		buf := bytes.NewReader(trackData)
		var absTick uint64

		for buf.Len() > 0 {
			delta, err := readVLQ(buf)
			if err != nil {
				return nil, err
			}
			absTick += uint64(delta)

			evType, err := buf.ReadByte()
			if err != nil {
				return nil, err
			}

			switch {
			case evType == 0xFF:
				metaType, err := buf.ReadByte()
				if err != nil {
					return nil, err
				}

				length, err := readVLQ(buf)
				if err != nil {
					return nil, err
				}

				data, err := readBytes(buf, int(length))
				if err != nil {
					return nil, err
				}

				tr.Events = append(tr.Events, ReplayEvent{
					Tick:       absTick,
					DeltaTicks: delta,
					Raw:        append([]byte{metaType}, data...),
					Kind:       EventMetaText,
				})

				if metaType == 0x2F {
					tr.Events = append(tr.Events, ReplayEvent{
						Tick:       absTick,
						DeltaTicks: delta,
						Raw:        nil,
						Kind:       EventEndOfTrack,
					})
					goto TrackDone
				}

			case evType == 0xF0 || evType == 0xF7:
				length, err := readVLQ(buf)
				if err != nil {
					return nil, err
				}

				data, err := readBytes(buf, int(length))
				if err != nil {
					return nil, err
				}

				tr.Events = append(tr.Events, ReplayEvent{
					Tick:       absTick,
					DeltaTicks: delta,
					Raw:        data,
					Kind:       EventSysEx,
				})

			default:
				return nil, fmt.Errorf("unsupported MIDI event 0x%X in track %d", evType, t)
			}
		}

	TrackDone:
		session.Tracks = append(session.Tracks, tr)
	}

	return session, nil
}
