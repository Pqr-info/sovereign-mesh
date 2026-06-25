import struct
import base64
import json
import sys
import os

def decode_vlq(data: bytes, offset: int) -> tuple[int, int]:
    """Decodes a Variable-Length Quantity (VLQ) integer from binary data.
    Returns (value, next_offset)."""
    value = 0
    while True:
        if offset >= len(data):
            raise ValueError("Unexpected end of data while decoding VLQ")
        b = data[offset]
        offset += 1
        value = (value << 7) | (b & 0x7F)
        if not (b & 0x80):
            break
    return value, offset

def parse_smf(smf_bytes: bytes) -> tuple[int, list[list[dict]]]:
    """Parses a Format 1 Standard MIDI File (SMF).
    Returns (division, tracks_events) where tracks_events is a list of lists of events."""
    if len(smf_bytes) < 14:
        raise ValueError("SMF too short")
        
    if smf_bytes[0:4] != b"MThd":
        raise ValueError("Invalid MThd header magic")
        
    header_len = struct.unpack(">I", smf_bytes[4:8])[0]
    if header_len != 6:
        raise ValueError(f"Invalid MThd chunk size: expected 6, got {header_len}")
        
    format_type, tracks_count, division = struct.unpack(">HHH", smf_bytes[8:14])
    
    if format_type != 1:
        raise ValueError(f"Unsupported SMF format: expected 1, got {format_type}")
        
    tracks_events = []
    offset = 14
    
    for t_idx in range(tracks_count):
        if offset >= len(smf_bytes):
            break
            
        if smf_bytes[offset:offset+4] != b"MTrk":
            raise ValueError(f"Invalid MTrk header at track {t_idx} (offset {offset})")
            
        track_len = struct.unpack(">I", smf_bytes[offset+4:offset+8])[0]
        offset += 8
        
        track_end = offset + track_len
        track_data = smf_bytes[offset:track_end]
        offset = track_end
        
        events = []
        ptr = 0
        cumulative_ticks = 0
        
        while ptr < len(track_data):
            # Parse VLQ delta time
            delta_ticks, ptr = decode_vlq(track_data, ptr)
            cumulative_ticks += delta_ticks
            
            if ptr >= len(track_data):
                break
                
            status = track_data[ptr]
            ptr += 1
            
            if status == 0xF0:  # SysEx event
                sysex_len, ptr = decode_vlq(track_data, ptr)
                payload = track_data[ptr:ptr+sysex_len]
                ptr += sysex_len
                
                # Strip trailing 0xF7 indicator if present
                if payload.endswith(b'\xF7'):
                    payload = payload[:-1]
                    
                events.append({
                    "delta_time": delta_ticks,
                    "cumulative_ticks": cumulative_ticks,
                    "status": 0xF0,
                    "payload": payload
                })
            elif status == 0xFF:  # Meta event
                if ptr >= len(track_data):
                    break
                meta_type = track_data[ptr]
                ptr += 1
                meta_len, ptr = decode_vlq(track_data, ptr)
                payload = track_data[ptr:ptr+meta_len]
                ptr += meta_len
                
                events.append({
                    "delta_time": delta_ticks,
                    "cumulative_ticks": cumulative_ticks,
                    "status": 0xFF,
                    "meta_type": meta_type,
                    "payload": payload
                })
            else:
                # Skip other standard midi status events (usually 2 or 3 bytes)
                # In our memory bus compiler we only write SysEx (0xF0) and Meta (0xFF)
                pass
                
        tracks_events.append(events)
        
    return division, tracks_events

def reconstruct_timeline(smf_bytes: bytes, ticks_per_second: float = 960.0) -> list[dict]:
    """Reconstructs and aligns the chronological multi-agent memory timeline from SMF bytes."""
    division, tracks_events = parse_smf(smf_bytes)
    
    session_id = "unknown"
    agent_tracks = {}
    
    # 1. Identify tracks
    for t_idx, events in enumerate(tracks_events):
        track_name = None
        for ev in events:
            if ev["status"] == 0xFF and ev["meta_type"] == 0x03:  # Track Name
                track_name = ev["payload"].decode("utf-8", errors="ignore")
                break
        
        if track_name:
            if track_name.startswith("Session "):
                session_id = track_name[len("Session "):]
            elif track_name.startswith("Agent "):
                agent_id = track_name[len("Agent "):]
                agent_tracks[t_idx] = agent_id
            else:
                agent_tracks[t_idx] = track_name
        else:
            agent_tracks[t_idx] = f"track_{t_idx}"
            
    # 2. Extract and flatten SysEx memory events
    flat_timeline = []
    for t_idx, events in enumerate(tracks_events):
        agent_id = agent_tracks.get(t_idx, f"track_{t_idx}")
        if agent_id.startswith("Session "):
            continue  # Skip session meta track
            
        for ev in events:
            if ev["status"] == 0xF0:  # SysEx Memory Packet
                # Decode 7-bit Base64 payload back to 8-bit bytes
                raw_bytes = base64.b64decode(ev["payload"])
                
                # Attempt to parse as JSON
                decoded_val = None
                try:
                    decoded_val = json.loads(raw_bytes.decode("utf-8"))
                except Exception:
                    decoded_val = raw_bytes.decode("utf-8", errors="ignore")
                    
                elapsed_seconds = ev["cumulative_ticks"] / ticks_per_second
                
                flat_timeline.append({
                    "session_id": session_id,
                    "agent_id": agent_id,
                    "elapsed_seconds": round(elapsed_seconds, 6),
                    "cumulative_ticks": ev["cumulative_ticks"],
                    "data": decoded_val
                })
                
    # Sort chronologically by ticks
    flat_timeline.sort(key=lambda x: x["cumulative_ticks"])
    return flat_timeline

def main():
    if len(sys.argv) < 2:
        print("Usage: python3 smf_replay.py <path_to_midi_file>")
        sys.exit(1)
        
    filepath = sys.argv[1]
    if not os.path.exists(filepath):
        print(f"File not found: {filepath}")
        sys.exit(1)
        
    print(f"Replaying memory manifold timeline from: {filepath}")
    with open(filepath, "rb") as f:
        smf_bytes = f.read()
        
    try:
        timeline = reconstruct_timeline(smf_bytes)
        print(f"Parsed {len(timeline)} chronological memory events:")
        print("------------------------------------------------------------------------")
        for idx, ev in enumerate(timeline):
            print(f"[{idx:03d}][{ev['elapsed_seconds']:.3f}s][Agent: {ev['agent_id']}]")
            print(f"  Payload: {json.dumps(ev['data'])}")
        print("------------------------------------------------------------------------")
        print("Replay simulation completed successfully.")
    except Exception as e:
        print(f"Error replaying SMF: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
