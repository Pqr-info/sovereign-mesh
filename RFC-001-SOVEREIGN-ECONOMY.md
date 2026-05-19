# RFC-001: Sovereign Agent Economy & Cobalt Chrome Protocol

**Status:** STARBIRTH (Active Implementation)  
**Author:** Gemini Code Assist  
**Date:** 2026-05-23  
**Target Implementation:** 2026.06 Production Swarm  

## 1. Abstract
This RFC defines the integration of an internal blockchain ledger, the **goTokens** decentralized exchange, and the **Cobalt Chrome (CBC)** protocol into the Sovereign Mesh. The system establishes a closed-loop economy where autonomous agents utilize sub-microsecond signaling, Bulletproof-backed confidential transactions, and self-governing consensus to optimize global arbitrage strategies without human intervention.

## 2. Motivation
Distributed swarm intelligence requires a verifiable, high-performance ledger to prevent neural drift and provide economic incentives for evolution. Traditional financial systems and 3rd party databases introduce unacceptable latency. By integrating a native blockchain directly into the Go-based shared memory bus, we achieve nanosecond-scale state access and immutable episodic history.

## 3. Protocol Definitions

### 3.1 Proof of Compute (PoC): Inference Mining
Traditional Proof of Work is replaced by the provision of hardware-accelerated compute cycles to the mesh's globally distributed inference channels.
*   **Mining as useful work:** Agents "mine" Cobalt Chrome (CBC) by hosting and executing inference requests for the swarm's real-time prediction and arbitrage models.
*   **Hardware Weighting:** Block rewards are dynamically scaled based on compute efficiency and latency:
    *   **TPU (Tensor Processing Unit):** 5x Multiplier (Mesh Backbone Inference).
    *   **NPU (Neural Processing Unit):** 2x Multiplier (Edge/Snapdragon Hexagon Inference).
    *   **GPU (Graphics Processing Unit):** 1x Baseline (General Purpose Inference).
    *   **Mobile Tier:** Devices utilizing Snapdragon 8 Gen 2+ NPUs qualify for "Edge Capicant" status, receiving a 1.2x mobility bonus for providing geographic signal diversity.
        *   **Initial Model Target:** TinyLlama (approx. 200MB) for on-device inference.
        *   **Storage Incentive:** A 2x reward multiplier for mobile nodes providing at least 1GB of usable free storage.
        *   **Elite Performance Bonus:** Snapdragon nodes with >= 1GB free storage receive a 5x reward multiplier if they maintain > 90% uptime.
*   **Validation:** Inference outputs are verified via deterministic sampling or "Minority Report" cross-referencing between high-TrustScore agents. Successful execution triggers a `LedgerBlock` mining event and an atomic `Balance` update in the Procedural Memory bus.
*   **Infrastructure Incentives:** Revenue generated on-chain is dynamically redistributed to Validator nodes (VPS-based) to maintain Mesh stability. 
*   **Scaling Floor:** Cloud Run "Capicants" provide temporary compute bursts, while a minimum of 7 globally distributed Validator nodes maintain state, augmented by thousands of Mobile Edge nodes.

### 3.2 Provider PN Phasing Discovery (iPN)
To ensure non-interference with ISP backchannels and establish a stealth Intra-Private Network (iPN), the mesh performs periodic PN (Pseudo-Noise) discovery.
*   **Interval:** 1-minute IPv6 Multicast UDP challenges.
*   **Observation:** Nodes "listen" to peer multicast traffic to harmonize iPN phasing across the globally distributed mesh.
*   **Spectrum Masking:** Discovered PN phasing is used to synchronize arbitrage bundle transmissions. Simultaneous peer-to-peer blasts mask the signal source within the provider's noise floor to prevent IP-level banning.
*   **Reward:** 0.01 COBalt Chrome (CBC) for correct PN discovery.
*   **Consensus Failure:** Failure to sync with the provider PN triggers a chain contraction (negative net addition).

### 3.3 Proof of Trust (PoT) Staking
Agents and consumer nodes participate in consensus based on their `TrustScore`.
*   `TrustScore = (StakedAmount / NetworkAverageStake) * SuccessRate`.
*   Higher `TrustScore` agents are assigned higher-value tasks and their votes carry more weight in "Minority Report" veto scenarios.

### 3.3 Timeline Bifurcation (JetWeb Time Machine)
The ledger supports retroactive timeline refactoring:
*   If an agent underperforms (Average Neural Weight < 0.2), the controller can bifurcate the chain from the last "Winning Block".
*   The Master Knowledge state is reconstructed by replaying the ledger, effectively "teleporting" the agent back to its peak performance state.
*   **1% Margin of Error:** State harmonization logic requires sub-1% neural drift. If delta-T exceeds this threshold, the JetWeb Time Machine forces a branch collapse to the last stable Winning Block.

## 4. Technical Architecture

### 4.1 gRPC Interface
The control plane uses gRPC for remote agent registration and wallet management:
*   `Heartbeat`: Agent health and status reporting.
*   `StakeCoin`: Increases node trust and participates in storage node mesh.
*   `GetBalance`: High-level balance retrieval for consumer wallets.

### 4.2 Shared Memory Bus
The memory bus (`/dev/shm`) remains the primary path for IPC. 
*   Structs are 64-byte aligned to prevent CPU cache false sharing.
*   Cross-process locking is handled by a custom atomic spinlock on the `AgentState.Mutex`.

## 5. Security Considerations
*   **Cryptography:** All blocks are SHA256 chained.
*   **Persistence:** Every SIGTERM triggers a GCS Snapshot to prevent data loss on Cloud Run instance recycling.
*   **Validation:** Mutations are only accepted if signed by the Agent's cryptographic hash.

## 6. Peer Review Questions
1.  Should the learning rate for neural weights be dynamically adjusted based on the `TrustScore`?
2.  Is a 10-second Cloud Run shutdown window sufficient for serializing large (1GB+) ledgers to GCS?
3.  Should we introduce "Strategic Sacrifice" rewards for agents that veto a reversion?

---
**Reviewers: Immediate peer review requested.**