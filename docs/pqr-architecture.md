# 🛰️ PQR Sovereign Node Architecture

The PQR Sovereign Node is a self-healing, model-aware infrastructure stack designed for autonomous governance and distributed intelligence. It operates as a "Mesh" where infrastructure, identity, and AI reasoning are tightly coupled.

## 🏗️ Core Components

### 1. The Persistence Layer (CockroachDB)
- **Engine**: CockroachDB (v23.1.13)
- **Role**: The "Long-term Memory" of the swarm. It stores the **Ticketing Fabric**, forensic audit trails, and agent memories.
- **Port**: `5196` (Local), `26257` (Internal)

### 2. The Identity & Secret Vault (HashiCorp Vault)
- **Engine**: Vault (v1.13.3)
- **Role**: The "Root of Trust." Stores SAML certificates, Cloudflare Access keys, and the Gemini Emergency Key.
- **Root Token**: `pqr-vault-token`
- **Port**: `8200`

### 3. The Sovereign Server (Go)
- **Engine**: Gin-based REST 2.0 API
- **Role**: The "Central Nervous System." Orchestrates requests between the database, the AI nodes, and the identity provider.
- **Endpoint**: `https://pqr.info`

### 4. The Connectivity Bridge (Cloudflare Tunnel)
- **Engine**: `cloudflared` (Dockerized)
- **Role**: Creates an encrypted, outbound-only tunnel to Cloudflare, bypassing the need for public IPs or open firewall ports.

## 🧠 AI Mesh Orchestration
The Swarm utilizes a multi-tiered inference strategy:
- **Local (Ollama/LM Studio)**: Fast, stateless reasoning for Layer 1-4 ticketing and basic health checks.
- **Cloud (Gemini Pro)**: High-level strategic governance and "Emergency Bridge" operations.

## 🛡️ Self-Healing Loops
The **HealingService** monitors the mesh for "Drift." 
- **Layer 7**: Connectivity and Identity issues.
- **Layer 10**: Governance and Secret integrity.
- **Escalation**: If a local model fails to resolve a ticket, it is promoted through 11 layers of complexity, eventually reaching Gemini for a "Strategic Strike" resolution.

## 🏛️ Immutable Audit Protocol (Database-First)
To prevent "Cognitive Drift" and ensure full traceability:
1. **Ticket Anchor**: No file modification may be made without a corresponding PQR Ticket ID.
2. **Forensic Diffing**: Every code change must be accompanied by a `diff` output injected into the ticket's `IntentBlob`.
3. **Genesis State**: The `manifest.json` provides the hashed "Ground Truth" for the entire stack. Any deviation without an audit trail is considered a "Drift Event."
