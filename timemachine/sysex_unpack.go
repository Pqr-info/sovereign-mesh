package timemachine

// Unpack7bit reverses MIDI-safe 7-bit SysEx packing.
// Input:  N bytes of 7-bit-safe data
// Output: original 8-bit binary payload
func Unpack7bit(packed []byte) ([]byte, error) {
	if len(packed) == 0 {
		return nil, nil
	}

	out := make([]byte, 0, len(packed))

	for i := 0; i < len(packed); {
		// First byte contains MSBs for next 7 bytes
		msb := packed[i]
		i++

		// Process next 7 bytes or until end
		for bit := 0; bit < 7 && i < len(packed); bit++ {
			b := packed[i]
			i++

			// Restore MSB from msb byte
			if (msb & (1 << bit)) != 0 {
				b |= 0x80
			}

			out = append(out, b)
		}
	}

	return out, nil
}
