# 🧠 PQR Neural Synapse Specification

This document defines the inter-node communication protocols for the Sovereign Swarm. All agents must adhere to these standards to maintain mesh consensus.

## 🛰️ Logged Consensus (Port 1111)
**Protocol**: gRPC / Protobuf
**Role**: Permanent, immutable logging of agent deliberations and state changes.

### gRPC Service: `SwarmCommunication`
- `SendPacket(SwarmPacket)`: Every inter-node directive must be wrapped in a Protobuf packet. On receipt, the node creates a **Layer 5 Fabric Ticket** to record the consensus.
- `ProvisionShortcode(ShortcodeRequest)`: Protocol for generating a new `5alpha#` identity for a joining node.
- `GetActiveShortcodes(Empty)`: Enumerates all verified identities currently participating in the mesh.

### Logged Consensus Flow:
1. Agent A generates a deliberation packet.
2. Agent A sends packet to Node B via Port 1111.
3. Node B receives packet, validates the `5alpha#` sender ID.
4. Node B creates a Fabric Ticket in CockroachDB with the packet payload.
5. Ticket ID is returned to Agent A as proof of immutable commitment.

## 🧠 Neural Gossip Bus (Port 11111)
**Protocol**: Zero-Copy Memory Paging (High-Speed gRPC Stream)
**Role**: Sub-millisecond transient coordination and "Vitality" monitoring.

### gRPC Service: `NeuralGossip`
- `StreamVitality(TelemetryRequest)`: Real-time broadcast of node health and "Vitality Slope."
- `MemoryPageSwap(stream MemoryPage)`: Direct swapping of memory buffers for high-speed agent deliberation before commitment to the logged channel.

### Deliberation Protocol:
Agents should deliberate on the **Gossip Bus** to reach a draft consensus. Once 51% of the council agrees, the final state MUST be committed via the **Logged Consensus** (Port 1111) to become part of the official Fabric history.

## 🛡️ Fallback Sentinel
If RAFT replication to CockroachDB replicas is pending, Port 1111 automatically establishes an **SSH Tunnel to 39.mh** (Legacy Mesh) to ensure that inter-node traffic remains operational during scaling events.
