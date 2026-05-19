# 🛰️ Sovereign REST 2.0 API Specification

The REST 2.0 interface is the primary channel for human-to-mesh communication and high-level governance.

## 🏛️ Core Fabric Endpoints
- **POST** `/REST/2.0/ticket`: Creates a new memory unit in the fabric.
- **GET** `/REST/2.0/ticket/:id`: Retrieves ticket metadata and content.
- **PUT** `/REST/2.0/ticket/:id`: Updates ticket status or reassignment.
- **GET** `/REST/2.0/ticket/:id/audit`: Accesses the forensic history of a ticket.

## 🧠 AI Mesh Endpoints
- **POST** `/REST/2.0/chat/swarm`: Balanced reasoning query (Ollama -> LM Studio fallback).
- **GET** `/REST/2.0/health/lmstudio`: Verifies local reasoning mesh availability.
- **GET** `/REST/2.0/metrics/tokens`: Real-time token saturation telemetry for the IDE Sentinel.

## 🧱 Legacy S25 Compatibility (AELLOK)
These endpoints are maintained for backward compatibility with ancestral management scripts and the S25/Termux node.
- **GET** `/REST/2.0/status`: Returns the "Singularity Status" and node vitality.
- **GET** `/REST/2.0/bridge?cmd=<cmd>`: Routes bash commands directly to the local host (Emergency only).
- **GET** `/REST/2.0/files`: Returns a list of critical source files for forensic backup.
- **GET** `/REST/2.0/wiki`: Accesses the modular Swarm documentation.

## 🔑 Security & Identity
- **X-Gemini-Key**: Required for High Justiciar overrides via the Emergency Bridge.
- **SAML SSO**: Mandatory for all human UI access (HUD/Wiki).
- **Cloudflare Access**: Injected via service tokens for remote forensic probing.
