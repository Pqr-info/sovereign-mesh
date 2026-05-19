# Protocol Specification: STARBIRTH (SBP-001)

**Release Version:** 2026.Q2.V1  
**Target Runlevel:** Starbirth  
**Status:** OPERATIONAL  

## 1. Protocol Objective
The Starbirth protocol defines the transition from a simulated agent environment to a production-grade autonomous swarm capable of sub-microsecond arbitrage and global state persistence. It prioritizes signal dominance and state integrity over traditional consensus latency.

## 2. Temporal Harmonization (iPN Phasing)
Nodes must achieve temporal alignment with the provider's Pseudo-Noise (PN) hopping sequence to establish the Intra-Private Network (iPN).
*   **Multicast Discovery:** Nodes join `[ff02::c0ba:11]:9999` to observe peer phasing.
*   **Hop Prediction:** 1 successful PN hit enables a 50-hop prediction horizon based on protocol-defined offsets.
*   **Synchronized Windows:** All arbitrage bundles are transmitted during the sub-microsecond spectrum reuse window of the local provider.

## 3. Signal Dominance (Loudest Mouth)
To ensure transaction inclusion and mask signal sources:
*   **Concurrency Factor:** 100x parallel UDP bursts per arbitrage event.
*   **Spectrum Masking:** Simultaneous multi-node transmission creates a uniform signal floor, preventing IP-level fingerprinting by ISPs or exchanges.

## 4. State Integrity (1% Rule)
State harmonization is strictly enforced to prevent neural drift in a high-concurrency environment.
*   **Drift Threshold:** Maximum 1.0% divergence from the last "Winning Block" weights.
*   **Fail-safe:** If drift exceeds threshold, the **JetWeb Time Machine** triggers a branch collapse and timeline reversion to the last stable state.
*   **Telemetry:** Every drift event is recorded via Go telemetry counters for forensic auditing.

## 5. Infrastructure Tiering
*   **Backbone (Validators):** Minimum 7 globally distributed nodes (e.g., $4 VPS) providing 15% incentive-backed state persistence.
*   **Edge (Capicants):** Mobile nodes leveraging **Snapdragon NPU** (Hexagon) for local ZK-proof generation and inference.
*   **Burst (Ephemeral):** Cloud Run containers dynamically allocated when validator counts drop or compute demand spikes.

## 6. Economic Incentives
*   **Discovery Reward:** 0.01 COBalt Chrome (CBC) per successful PN alignment.
*   **Mesh Penalty:** Negative net chain addition (contraction) for failed global sync rounds.
*   **Mobility Bonus:** 1.2x multiplier for NPU-backed mobile edge nodes.

---
**Approved for Starbirth Status Runlevel deployment.**