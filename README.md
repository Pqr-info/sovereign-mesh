# Sovereign Mesh: The Cobalt Chrome 2027 Masterpiece

## Vision
The Sovereign Mesh is a globally distributed agent economy. It establishes a "Mothership" (Cloud Run + Private Gemini Models) that orchestrates a swarm of autonomous agents. These agents compete and cooperate in a high-stakes environment where intelligence is rewarded with **Cobalt Chrome (CBC)** coin.

## Core Architecture

### 1. 5-Tier Memory Stack
*   **Sensory Memory:** Volatile UDP/XDP telemetry for sub-microsecond price signaling.
*   **Short-Term Memory (STM):** Hash-indexed RAM database mapped to `/dev/shm` for active working data.
*   **Procedural Memory:** Shared memory bus containing `AgentState` with atomic spinlocks for zero-copy weight updates.
*   **Episodic Memory (JetWeb Time Machine):** An internal high-speed blockchain that records every mutation and allows for timeline bifurcation/reversion.
*   **Long-Term Memory (LTM):** Distributed ledger state persisted to Google Cloud Storage snapshots.

### 2. Economic Engine (Cobalt Chrome)
*   **Proof of Work (PoW):** Agents earn CBC by completing metaprogramming and prompt optimization tasks.
*   **Proof of Trust (PoT):** Consumer nodes stake CBC to increase their `TrustScore`, granting them higher priority in task assignment and consensus voting.
*   **Consumer Wallet:** A cross-platform interface allowing users to stake coin, view agent neural weights, and contribute local storage to the mesh.

### 3. Managed Intelligence
Leverages Vertex AI's Prompt Optimizer to evolve agent personas. The "Mothership" verifies work quality before the internal blockchain mines the reward block.

## Deployment
Designed for **Cloud Run** with `CGO_ENABLED=0`. 
*   **Zero Dependencies:** No external database required; the internal blockchain is the source of truth.
*   **Snapshotting:** Automatic GCS persistence on `SIGTERM` ensures state survives container ephemerality.

## Quick Start
```go
package main

import "github.com/sovereign/mesh"

func main() {
    swarm := sovereign.NewController("my-project", "us-central1")
    swarm.InitMemoryBus()
    swarm.Start(context.Background())
}
```

## Technical Specs
*   **Language:** Pure Go 1.26+ (Transitioned from 1.16 for eBPF and modern protobuf support)
*   **Protocol:** gRPC / Protobuf
*   **IPC:** `syscall.Mmap` / `sync/atomic`
*   **Blockchain:** SHA256 linked blocks with JSON mutation payloads.

---
*Propelling the swarm into 2027.*