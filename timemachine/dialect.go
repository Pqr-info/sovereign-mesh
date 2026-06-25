package timemachine

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// memoryPage matches your SysEx JSON schema
type memoryPage struct {
	SessionID           string    `json:"session_id"`
	AgentID             string    `json:"agent_id"`
	Page                int       `json:"page"`
	Data                string    `json:"data"`
	Coords5D            *Coords5D `json:"coords5d,omitempty"`
	EvolutionaryVersion int       `json:"evolutionary_version,omitempty"`
}

// DecodeSysExPayload decodes a single SysEx event:
//   1. Takes raw SysEx bytes (already stripped of 7-bit packing)
//   2. Base64-decodes them
//   3. JSON-unmarshals into memoryPage
func DecodeSysExPayload(raw []byte) (*memoryPage, error) {
	if len(raw) == 0 {
		return nil, errors.New("empty SysEx payload")
	}

	// In some MIDI files, SysEx payload ends with F7 (handled/stripped in parsing or wrapper).
	// Strip trailing F7 if present
	if raw[len(raw)-1] == 0xF7 {
		raw = raw[:len(raw)-1]
	}

	decoded, err := base64.StdEncoding.DecodeString(string(raw))
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	var mp memoryPage
	if err := json.Unmarshal(decoded, &mp); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	return &mp, nil
}

// DecodeDialect converts raw ReplaySession -> semantic AgentEvents.
// It extracts:
//   - Session ID from Track 0 meta-text
//   - Agent IDs from Track N meta-text
//   - Memory pages from SysEx JSON payloads (supporting both 7-bit packed and plain Base64 payloads)
func DecodeDialect(session *ReplaySession) ([]AgentEvent, error) {
	if len(session.Tracks) == 0 {
		return nil, errors.New("no tracks in session")
	}

	var events []AgentEvent

	// -----------------------------------------
	// 1. Parse Track 0 for session metadata
	// -----------------------------------------
	track0 := &session.Tracks[0]

	for _, ev := range track0.Events {
		if ev.Kind != EventMetaText {
			continue
		}

		raw := ev.Raw
		if len(raw) < 1 {
			continue
		}

		metaType := raw[0]
		data := raw[1:]

		// Meta Type 0x03 = Track Name / Text
		if metaType == 0x03 {
			txt := string(data)

			if strings.HasPrefix(txt, "Session ") {
				session.SessionID = strings.TrimPrefix(txt, "Session ")
			}
		}
	}

	if session.SessionID == "" {
		return nil, errors.New("missing Session ID in Track 0")
	}

	// -----------------------------------------
	// 2. Parse each agent track
	// -----------------------------------------
	for ti := 1; ti < len(session.Tracks); ti++ {
		track := &session.Tracks[ti]

		// Extract Agent ID from meta-text
		for _, ev := range track.Events {
			if ev.Kind != EventMetaText {
				continue
			}

			raw := ev.Raw
			if len(raw) < 1 {
				continue
			}

			metaType := raw[0]
			data := raw[1:]

			if metaType == 0x03 { // Track Name / Text
				txt := string(data)
				if strings.HasPrefix(txt, "Agent ") {
					track.AgentID = strings.TrimPrefix(txt, "Agent ")
				}
			}
		}

		if track.AgentID == "" {
			return nil, fmt.Errorf("track %d missing Agent ID", ti)
		}

		// -----------------------------------------
		// 3. Decode SysEx memory pages
		// -----------------------------------------
		for _, ev := range track.Events {
			if ev.Kind != EventSysEx {
				continue
			}

			// Attempt 1: Try unpacking 7-bit MIDI-safe payload first
			var mp *memoryPage
			var err error
			
			unpacked, unpackErr := Unpack7bit(ev.Raw)
			if unpackErr == nil {
				mp, err = DecodeSysExPayload(unpacked)
			}
			
			// Attempt 2: Fall back to direct Base64 decode if 7-bit unpack failed or produced invalid base64/JSON
			if unpackErr != nil || err != nil {
				mp, err = DecodeSysExPayload(ev.Raw)
				if err != nil {
					return nil, fmt.Errorf("track %d SysEx decode error (both 7-bit unpack and direct decode failed): %w", ti, err)
				}
			}

			// Validate agent consistency
			if mp.AgentID != track.AgentID {
				return nil, fmt.Errorf(
					"SysEx agent mismatch: track=%s payload=%s",
					track.AgentID, mp.AgentID,
				)
			}

			// Validate session consistency
			if mp.SessionID != session.SessionID {
				return nil, fmt.Errorf(
					"SysEx session mismatch: session=%s payload=%s",
					session.SessionID, mp.SessionID,
				)
			}

			events = append(events, AgentEvent{
				Tick:                ev.Tick,
				AgentID:             mp.AgentID,
				Page:                mp.Page,
				Data:                mp.Data,
				SessionID:           mp.SessionID,
				Coords5D:            mp.Coords5D,
				EvolutionaryVersion: mp.EvolutionaryVersion,
			})
		}
	}

	return events, nil
}
