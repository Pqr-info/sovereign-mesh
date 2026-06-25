package timemachine

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"os"
	"testing"
)

// MockRuntime implements MeshRuntime for testing
type MockRuntime struct {
	SessionStarted bool
	SessionEnded   bool
	SessionID      string
	Events         []TimelineEvent
}

func (m *MockRuntime) OnSessionStart(id string) error {
	m.SessionStarted = true
	m.SessionID = id
	return nil
}

func (m *MockRuntime) OnSessionEnd(id string) error {
	m.SessionEnded = true
	return nil
}

func (m *MockRuntime) ApplyMemoryEvent(ev TimelineEvent) error {
	m.Events = append(m.Events, ev)
	return nil
}

func encodeVLQBytes(value uint32) []byte {
	if value == 0 {
		return []byte{0}
	}
	var buf []byte
	for value > 0 {
		val := byte(value & 0x7F)
		value >>= 7
		if len(buf) > 0 {
			val |= 0x80
		}
		buf = append([]byte{val}, buf...)
	}
	return buf
}

func TestTimeMachineReplay(t *testing.T) {
	// 1. Build a mock SMF binary manually
	buf := new(bytes.Buffer)

	// MThd header
	buf.Write([]byte("MThd"))
	binary.Write(buf, binary.BigEndian, uint32(6))     // header len
	binary.Write(buf, binary.BigEndian, uint16(1))     // format
	binary.Write(buf, binary.BigEndian, uint16(2))     // 2 tracks (1 meta, 1 agent)
	binary.Write(buf, binary.BigEndian, uint16(480))   // division

	// Track 0 (Meta Track)
	t0Data := new(bytes.Buffer)
	// Session Name: Session test-session
	t0Data.Write(encodeVLQBytes(0))
	t0Data.Write([]byte{0xFF, 0x03})
	t0Data.Write(encodeVLQBytes(uint32(len("Session test-session"))))
	t0Data.Write([]byte("Session test-session"))
	// End of Track
	t0Data.Write(encodeVLQBytes(0))
	t0Data.Write([]byte{0xFF, 0x2F, 0x00})

	buf.Write([]byte("MTrk"))
	binary.Write(buf, binary.BigEndian, uint32(t0Data.Len()))
	buf.Write(t0Data.Bytes())

	// Track 1 (Agent Track)
	t1Data := new(bytes.Buffer)
	// Agent Name: Agent alpha
	t1Data.Write(encodeVLQBytes(0))
	t1Data.Write([]byte{0xFF, 0x03})
	t1Data.Write(encodeVLQBytes(uint32(len("Agent alpha"))))
	t1Data.Write([]byte("Agent alpha"))

	// SysEx memory page event at delta 960 (1 second elapsed at 960 ticks/sec)
	pageData := `{"session_id":"test-session","agent_id":"alpha","page":0,"data":"page-0-content"}`
	b64Payload := base64.StdEncoding.EncodeToString([]byte(pageData))
	t1Data.Write(encodeVLQBytes(960))
	t1Data.Write([]byte{0xF0})
	t1Data.Write(encodeVLQBytes(uint32(len(b64Payload))))
	t1Data.Write([]byte(b64Payload))

	// End of Track
	t1Data.Write(encodeVLQBytes(0))
	t1Data.Write([]byte{0xFF, 0x2F, 0x00})

	buf.Write([]byte("MTrk"))
	binary.Write(buf, binary.BigEndian, uint32(t1Data.Len()))
	buf.Write(t1Data.Bytes())

	// 2. Parse mock SMF
	session, err := ParseSMF(buf)
	if err != nil {
		t.Fatalf("failed to parse mock SMF: %v", err)
	}

	// 3. Replay timeline
	mockRuntime := &MockRuntime{}
	tm := NewTimeMachine(mockRuntime)
	err = tm.Replay(context.Background(), session, ReplayFastForward)
	if err != nil {
		t.Fatalf("replay failed: %v", err)
	}

	if session.SessionID != "test-session" {
		t.Errorf("expected session ID test-session, got %q", session.SessionID)
	}

	if !mockRuntime.SessionStarted {
		t.Error("expected session to start")
	}
	if !mockRuntime.SessionEnded {
		t.Error("expected session to end")
	}

	if len(mockRuntime.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(mockRuntime.Events))
	}

	ev := mockRuntime.Events[0]
	if ev.AgentID != "alpha" {
		t.Errorf("expected agent ID alpha, got %s", ev.AgentID)
	}
	if ev.RealTimeMs != 1000 {
		t.Errorf("expected real time 1000ms, got %dms", ev.RealTimeMs)
	}
	if ev.Data != "page-0-content" {
		t.Errorf("expected page content 'page-0-content', got %s", ev.Data)
	}
}

func TestDecodeMalformedSysEx(t *testing.T) {
	_, err := DecodeSysExPayload([]byte("invalid-base64"))
	if err == nil {
		t.Error("expected error on malformed base64 payload")
	}
}

func TestUnpack7bit(t *testing.T) {
	packed := []byte{0x01, 0x00, 0x01, 0x02}
	unpacked, err := Unpack7bit(packed)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}

	expected := []byte{0x80, 0x01, 0x02}
	if !bytes.Equal(unpacked, expected) {
		t.Fatalf("expected %x, got %x", expected, unpacked)
	}
}

// ------------------------------------------------------------
// End-to-end test: SMF → TimeMachine → MeshRuntimeMock
// ------------------------------------------------------------
func TestTimeMachineEndToEnd(t *testing.T) {
	// Load golden SMF file
	f, err := os.Open("testdata/test_session_123_memory.mid")
	if err != nil {
		t.Fatalf("failed to open test SMF: %v", err)
	}
	defer f.Close()

	// Parse mock SMF
	session, err := ParseSMF(f)
	if err != nil {
		t.Fatalf("failed to parse mock SMF: %v", err)
	}

	// Replay timeline
	mock := &MockRuntime{}
	tm := NewTimeMachine(mock)
	err = tm.Replay(context.Background(), session, ReplayFastForward)
	if err != nil {
		t.Fatalf("replay failed: %v", err)
	}

	if mock.SessionID != "test_session_123" {
		t.Errorf("expected session ID test_session_123, got %q", mock.SessionID)
	}
	if !mock.SessionStarted {
		t.Error("expected session to start")
	}
	if !mock.SessionEnded {
		t.Error("expected session to end")
	}

	// Add a Golden Snapshot Assertion
	expectedEvents := []TimelineEvent{
		{Tick: 0, AgentID: "test_agent_alpha", Page: 0, Data: "active memory page 0"},
		{Tick: 480, AgentID: "test_agent_alpha", Page: 1, Data: "active memory page 1"},
	}

	if len(mock.Events) != len(expectedEvents) {
		t.Fatalf("expected %d events, got %d", len(expectedEvents), len(mock.Events))
	}

	for i, ev := range mock.Events {
		expected := expectedEvents[i]
		if ev.Tick != expected.Tick {
			t.Errorf("event %d: expected tick %d, got %d", i, expected.Tick, ev.Tick)
		}
		if ev.AgentID != expected.AgentID {
			t.Errorf("event %d: expected agent %q, got %q", i, expected.AgentID, ev.AgentID)
		}
		if ev.Page != expected.Page {
			t.Errorf("event %d: expected page %d, got %d", i, expected.Page, ev.Page)
		}
		if ev.Data != expected.Data {
			t.Errorf("event %d: expected data %q, got %q", i, expected.Data, ev.Data)
		}
	}
}
