import sys
import os
import json

sys.path.append(r"c:\Users\theal\swend-mesh\SUBSTRATE\memory_bus")
import smf_wrapper

def generate_golden_mid():
    print("Generating golden SMF file for end-to-end testing...")
    
    # Track 0: System/Meta Info
    meta_events = [
        {"delta_time": 0, "status": 0xFF, "meta_type": 0x03, "payload": b"Session test_session_123"},
        {"delta_time": 0, "status": 0xFF, "meta_type": 0x01, "payload": b"Session Initialized"},
    ]
    track0 = smf_wrapper.build_mtrk(meta_events)
    
    # Track 1: test_agent_alpha
    # TimelineEvent: {Tick: 0, AgentID: "test_agent_alpha", Page: 0, Data: "genesis page"}
    alpha_evt1 = {
        "session_id": "test_session_123",
        "agent_id": "test_agent_alpha",
        "page": 0,
        "data": "genesis page",
        "coords5d": {"x1": 500, "x2": 500, "x3": 500, "x4": 500, "x5": 500},
        "evolutionary_version": 1
    }
    
    # TimelineEvent: {Tick: 480, AgentID: "test_agent_alpha", Page: 1, Data: "active memory page 1"}
    alpha_evt2 = {
        "session_id": "test_session_123",
        "agent_id": "test_agent_alpha",
        "page": 1,
        "data": "active memory page 1",
        "coords5d": {"x1": 800, "x2": 400, "x3": 200, "x4": 500, "x5": 1000},
        "evolutionary_version": 2
    }
    
    # TimelineEvent: {Tick: 240, AgentID: "test_agent_beta", Page: 0, Data: "beta genesis page"}
    beta_evt1 = {
        "session_id": "test_session_123",
        "agent_id": "test_agent_beta",
        "page": 0,
        "data": "beta genesis page",
        "coords5d": {"x1": 100, "x2": 900, "x3": 500, "x4": 200, "x5": 800},
        "evolutionary_version": 1
    }
    
    # TimelineEvent: {Tick: 720, AgentID: "test_agent_beta", Page: 1, Data: "beta active page"}
    beta_evt2 = {
        "session_id": "test_session_123",
        "agent_id": "test_agent_beta",
        "page": 1,
        "data": "beta active page",
        "coords5d": {"x1": 600, "x2": 200, "x3": 800, "x4": 900, "x5": 300},
        "evolutionary_version": 3
    }
    
    payload1 = smf_wrapper.pack_bytes_to_7bit(json.dumps(alpha_evt1).encode('utf-8'))
    payload2 = smf_wrapper.pack_bytes_to_7bit(json.dumps(beta_evt1).encode('utf-8'))
    payload3 = smf_wrapper.pack_bytes_to_7bit(json.dumps(alpha_evt2).encode('utf-8'))
    payload4 = smf_wrapper.pack_bytes_to_7bit(json.dumps(beta_evt2).encode('utf-8'))
    
    alpha_events = [
        {"delta_time": 0, "status": 0xFF, "meta_type": 0x03, "payload": b"Agent test_agent_alpha"},
        {"delta_time": 0, "status": 0xF0, "payload": payload1},
        {"delta_time": 240, "status": 0xF0, "payload": payload2},
        {"delta_time": 240, "status": 0xF0, "payload": payload3},
        {"delta_time": 240, "status": 0xF0, "payload": payload4},
    ]
    track1 = smf_wrapper.build_mtrk(alpha_events)
    
    smf_blob = smf_wrapper.compile_smf([track0, track1])
    
    out_dir = r"c:\Users\theal\swend-mesh\SUBSTRATE\timemachine\testdata"
    os.makedirs(out_dir, exist_ok=True)
    out_path = os.path.join(out_dir, "test_session_123_memory.mid")
    
    with open(out_path, "wb") as f:
        f.write(smf_blob)
    
    print(f"Golden SMF file generated at: {out_path}")

if __name__ == "__main__":
    generate_golden_mid()
