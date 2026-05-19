# PQR SG-DAO: Self-Autonomous Governance Layer

## The Autonomous Authority
The `pqr-info-swarm` is the primary **Self-Autonomous Governance Layer** of the SG-DAO. It is not merely a tracking system, but the active enforcement mechanism that ensures the organization's evolution adheres to its core principles without human intervention.

### The Seven Sticky Rules
1. **Strict Lineage**: No ticket can be committed without a verified parent hash.
2. **Zero Divergence**: State mutations must be validated against the Consensus Mesh.
3. **Forensic Primacy**: Audit logs are immutable and stored in the Fabric.
4. **Agent Accountability**: Every action must be signed by a valid Agent ID from Vault.
5. **Content Addressing**: All payloads are identified by their SHA-256 hashes.
6. **Layer Isolation**: Agents cannot modify layers above their current sovereignty level.
7. **Self-Healing Requirement**: Every failure must spawn a healing ticket in the Fabric.

## Forensic Auditing
The `/REST/2.0/ticket/:id/audit` endpoint provides the complete history of a ticket. This includes:
- **Who**: The Agent ID.
- **What**: The specific action (LINK, UPDATE, MUTATE).
- **Why**: The LLM's rationale (stored in the `intent_blob`).
- **When**: High-precision timestamps.

## Design by Contract (DbC)
We use Go's `reflect` package to enforce these rules at runtime. Before a ticket is finalized, the system inspects the payload to ensure all contractual pre-conditions are met.
