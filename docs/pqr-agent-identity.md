# 🛡️ Agent Identity & Shortcodes

Every participant in the PQR Swarm is identified by a unique, immutable **`5alpha#` shortcode**.

## 🆔 The Shortcode Format
- **Structure**: Five alphanumeric characters followed by a hash (e.g., `ΩX9R2#`).
- **Genesis Node**: Reserved as `ΩX9R2#`.
- **Scope**: Used as the `sender_id` in all gRPC packets and `CreatorAgentID` in all Fabric Tickets.

## 🔄 Provisioning Protocol
When a new node joins the mesh:
1. It connects to an existing node on **Port 1111**.
2. It sends a `ProvisionShortcode` gRPC request with its designated role.
3. The host node generates a unique, unassigned shortcode and records it in the `agents` table of CockroachDB.
4. The joining node adopts this identity for all future communications.

## 🛰️ Peer Discovery
Agents can learn about their peers via:
- `GetActiveShortcodes`: gRPC call to discover live nodes.
- `GET /REST/2.0/tickets`: Querying the fabric to see which shortcodes are creating work.
- `TelemetryRequest`: Monitoring the **Gossip Bus** (Port 11111) for active vitality streams.

## 🛡️ Forensic Trust
Shortcodes are linked to public keys in the Identity Vault. Any packet or ticket created without a valid signature matching the shortcode's registered key is immediately blackholed by the **Healing Service**.
