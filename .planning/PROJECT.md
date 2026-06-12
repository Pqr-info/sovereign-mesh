# Project: Sovereign Mesh

## Vision

Decentralized sovereign mesh network with a 128-agent swarm operating on the PQR blockchain. The system provides neural inference via the Sovereign-27 model, cross-platform clients (web portal, Android app, MUDD console), and a high-throughput gRPC control bus with shared-memory data plane. Designed for absolute portability — all Go binaries are statically compiled with `CGO_ENABLED=0` to run on any Linux environment including Termux on Android.

## Team

- **Developer**: aellok (solo)
- **GitHub**: thealanphipps-del/sovereign-mesh
- **Development Environment**: Linux (WSL2 + native)

## Current Milestone

**v1.0 — Production Launch**

Target: Ship a stable, deployable sovereign mesh with working Android client, polished web frontend, green CI pipeline, and hardened backend.

## Tech Stack

### Core Runtime
| Component | Technology | Notes |
|-----------|-----------|-------|
| Language | Go 1.26+ | Strict `CGO_ENABLED=0` static linking |
| Module | `github.com/pqr-info/sovereign-mesh` | |
| Web Embed | Vanilla HTML5/CSS3/JS via `go:embed` | GIN-embedded static assets |
| Frontend | React 19 + TypeScript 6 + Vite 8 | Separate SPA in `frontend/` |
| Android | Kotlin/Java + Gradle | In `/home/aellok/triplehelix/android_app` |

### Database Layer
| Component | Technology | Notes |
|-----------|-----------|-------|
| Primary | CockroachDB / PostgreSQL | `github.com/lib/pq`, port 26257 |
| Fallback | Pure-Go SQLite | `modernc.org/sqlite`, auto-routed on timeout |

### Neural Layer
| Component | Technology | Notes |
|-----------|-----------|-------|
| Model | Sovereign-27 (300M params) | Cloud Run + Nvidia L4 GPU |
| Local AI | Hot-swap LLM client | LM Studio / Ollama endpoints via `pkg/llm` |

### Infrastructure
| Component | Technology | Notes |
|-----------|-----------|-------|
| RPC | gRPC (protobuf) | Control bus on `localhost:1111` |
| Memory Bus | Shared memory `/dev/shm/` | 16MB page table, 252 MB/s throughput |
| Cloud | GCP (Cloud Run, Artifact Registry) | Project `fast-web-496805-k0` |
| CI/CD | GitHub Actions | `ci.yml` + `deploy.yml` exist (partial) |
| Tunnels | Cloudflare Tunnels | For public service exposure |

## Architecture Patterns

### 1. Zero-CGO Static Portability
All Go packages compiled with `CGO_ENABLED=0`. Pure-Go SQLite (`modernc.org/sqlite`) replaces standard CGO-dependent SQLite. Guarantees deployment on any Linux including Termux.

### 2. Lock-Free Atomic Reasoning Queues
The 8×8 Strategy Swarm (128 agents) uses lock-free queues to avoid memory allocation and GC pressure during concurrent task processing.

### 3. Dynamic Database Routing Fallback
`pkg/tickets` and `pkg/rtgo` modules monitor PostgreSQL with tight timeouts. On failure, operations are dynamically translated and executed against local SQLite (`local_ledger.db`), ensuring zero downtime.

### 4. Hot-Swap Binary Upgrade (AtomicSwap)
Zero-latency binary upgrades transfer live socket descriptors and execution contexts between process versions without service interruption.

## Go Binaries (cmd/)

| Binary | Entry Point | Purpose |
|--------|-------------|---------|
| `mesh_server` | `cmd/mesh_server/main.go` | Core mesh network server |
| `mgsh_cli` | `cmd/mgsh_cli/main.go` | Mesh shell CLI |
| `sovereign-cli` | `cmd/sovereign-cli/main.go` | Sovereign command CLI |
| `sovereign-auto` | `cmd/sovereign-auto/main.go` | Automated sovereign operations |
| `mint_swarm` | `cmd/mint_swarm/main.go` | Genesis agent minting (128 agents) |
| `serve` | `cmd/serve/serve.go` | Web portal server |
| `test_auth` | `cmd/test_auth/test_auth.go` | Auth testing utility |

## Active Concerns

### RESOLVED
- **PostgreSQL JSON Concatenation**: JSONB `||` operators now dynamically translated for SQLite fallback via serializing parser.

### ACTIVE — HIGH PRIORITY
- **GCP Artifact Registry Billing**: Requires active billing on `fast-web-496805-k0` before container uploads for GPU deployments work.

### WARNING
- **LLM Knowledge Cutoffs**: LLM judges flag correct live data as "beyond training cutoff." Mitigate with shape-based rubrics instead of strict fact matching.

### UNTRACKED (Identified During Bootstrap)
- **Credential Files in Repo**: `gcp-key.json`, `gcp_key.json`, `gcp_loader_key.json`, `fast-web-key.json`, `.cloudflare_token` committed to repository. Must be rotated and moved to secrets management before production.
- **Go Test Coverage**: Only 4 test files exist (`ledger_test.go`, `memory_test.go`, `stratum_test.go`, `models/flashlite/flashlite_test.go`). Critical paths lack test coverage.
- **CI Pipeline Gaps**: Existing `ci.yml` only lints Python/shell — no Go build, Go test, or Android build steps.

## Codebase Scale

- **655 files**, ~2.76M words
- **Knowledge Graph**: 7,664 nodes, 8,541 edges, 980 communities
- **Protobuf**: `sync.proto` (17KB) + `mesh_proto.proto` (1.2KB)
- **Frontend**: React 19 SPA with Vite 8, TypeScript 6
