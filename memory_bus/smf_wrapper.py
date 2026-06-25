import struct
import base64

def encode_vlq(value: int) -> bytes:
    """Encodes an integer to MIDI Variable-Length Quantity (VLQ) bytes."""
    if value == 0:
        return b'\x00'
    buffer = bytearray()
    while value > 0:
        val = value & 0x7F
        value >>= 7
        if buffer:
            val |= 0x80
        buffer.insert(0, val)
    return bytes(buffer)

def pack_bytes_to_7bit(data: bytes) -> bytes:
    """Base64 encodes 8-bit bytes to generate a 7-bit ASCII safe payload."""
    return base64.b64encode(data)

def unpack_7bit_to_bytes(data: bytes) -> bytes:
    """Decodes Base64 encoded 7-bit bytes back to original 8-bit bytes."""
    return base64.b64decode(data)

def build_mthd(format_type: int = 1, tracks: int = 1, division: int = 480) -> bytes:
    """Builds the 14-byte SMF Header Chunk (MThd)."""
    return b"MThd" + struct.pack(">IHHH", 6, format_type, tracks, division)

def build_mtrk(events: list) -> bytes:
    """Builds an SMF Track Chunk (MTrk) containing delta-timed events."""
    track_data = bytearray()
    for ev in events:
        delta_time = ev.get('delta_time', 0)
        track_data.extend(encode_vlq(delta_time))
        status = ev.get('status', 0xF0)
        
        if status == 0xF0:  # SysEx Event
            payload = ev.get('payload', b'')
            if not payload.endswith(b'\xF7'):
                payload = payload + b'\xF7'
            track_data.append(0xF0)
            track_data.extend(encode_vlq(len(payload)))
            track_data.extend(payload)
        elif status == 0xFF:  # Meta Event
            meta_type = ev.get('meta_type', 0x01)
            payload = ev.get('payload', b'')
            track_data.append(0xFF)
            track_data.append(meta_type)
            track_data.extend(encode_vlq(len(payload)))
            track_data.extend(payload)
            
    # Add End of Track Meta Event: delta_time=0, status=0xFF, type=0x2F, length=0
    track_data.extend(encode_vlq(0))
    track_data.extend(b'\xFF\x2F\x00')
    
    return b"MTrk" + struct.pack(">I", len(track_data)) + bytes(track_data)

def compile_smf(tracks: list[bytes], division: int = 480) -> bytes:
    """Compiles a list of MTrk track bytes into a single Format 1 SMF binary block."""
    header = build_mthd(format_type=1, tracks=len(tracks), division=division)
    return header + b"".join(tracks)
