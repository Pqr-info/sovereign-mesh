# Requirements — v1.0 Production Launch

## Overview

These requirements define the minimum viable criteria for the Sovereign Mesh v1.0 production release. Each requirement is scoped to address a concrete gap identified during codebase analysis.

---

## R1: Android App Compiles and Runs with Real Blockchain Integration

- **Priority**: HIGH
- **Status**: NOT STARTED
- **Location**: `/home/aellok/triplehelix/android_app`
- **Acceptance Criteria**:
  - [ ] `./gradlew assembleDebug` completes without errors
  - [ ] App connects to real PQR chain endpoint (not mock/stub)
  - [ ] Web3j or equivalent library integrated for on-chain transactions
  - [ ] Wallet creation, balance display, and agent status functional
  - [ ] APK installable and testable on physical Android device or emulator

## R2: Web Frontend Deployed and Functional

- **Priority**: HIGH
- **Status**: NOT STARTED
- **Location**: `frontend/` (React 19 + Vite 8 + TypeScript 6)
- **Acceptance Criteria**:
  - [ ] `npm run build` produces clean production build with zero errors
  - [ ] Frontend deployed to Cloudflare Workers/Pages or equivalent CDN
  - [ ] Dashboard displays live agent status, balances, and mesh telemetry
  - [ ] Responsive design works on mobile and desktop
  - [ ] API calls connect to live `mesh_server` backend

## R3: Go Backend Passes All Tests with CGO_ENABLED=0

- **Priority**: HIGH
- **Status**: IN PROGRESS (4 test files exist, critical paths uncovered)
- **Location**: Root Go module + `cmd/`, `pkg/`
- **Acceptance Criteria**:
  - [ ] `CGO_ENABLED=0 go test ./...` passes with zero failures
  - [ ] `go vet ./...` reports zero issues
  - [ ] Test coverage for: gRPC handlers, ledger operations, memory bus, ticket CRUD, database fallback routing
  - [ ] All 7 binaries in `cmd/` build successfully with `CGO_ENABLED=0`
  - [ ] No race conditions detected under `go test -race` (where applicable)

## R4: CI/CD Pipeline for Automated Builds

- **Priority**: MEDIUM
- **Status**: PARTIAL (ci.yml exists but only lints Python/shell, no Go build or test steps)
- **Location**: `.github/workflows/`
- **Acceptance Criteria**:
  - [ ] CI runs `CGO_ENABLED=0 go build` for all binaries on every push
  - [ ] CI runs `go test ./...` and `go vet ./...`
  - [ ] CI builds React frontend (`npm run build`)
  - [ ] CI validates protobuf compilation
  - [ ] Deploy workflow produces versioned release artifacts on tag push
  - [ ] Android build step added (Gradle assembleDebug)

## R5: Security Audit — No Hardcoded Credentials in Production

- **Priority**: CRITICAL
- **Status**: NOT STARTED (credential files detected in repo)
- **Location**: Root directory
- **Acceptance Criteria**:
  - [ ] `gcp-key.json`, `gcp_key.json`, `gcp_loader_key.json`, `fast-web-key.json`, `fast-web-key-new.json` removed from tracked files
  - [ ] `.cloudflare_token` removed from tracked files
  - [ ] All secrets migrated to GitHub Actions secrets or environment variables
  - [ ] `.gitignore` updated to prevent future credential commits
  - [ ] Existing keys rotated in GCP console and Cloudflare dashboard
  - [ ] `git filter-branch` or BFG Repo Cleaner run to purge credentials from git history

## R6: Documentation Complete

- **Priority**: MEDIUM
- **Status**: PARTIAL (README.md exists, incomplete)
- **Location**: `docs/`, root `README.md`
- **Acceptance Criteria**:
  - [ ] README.md updated with: project overview, quickstart, architecture diagram, build instructions
  - [ ] API documentation for gRPC endpoints (from `sync.proto`)
  - [ ] Deployment guide covering: local dev, staging, production
  - [ ] Man pages validated (`.1`, `.7`, `.8` files exist but need review)
  - [ ] MUDD console usage documented
  - [ ] Android app build and install instructions

## R7: Local Blockchain Endpoint Operational for Development

- **Priority**: MEDIUM
- **Status**: NOT STARTED
- **Acceptance Criteria**:
  - [ ] Local PQR chain node or Ganache-equivalent running for dev/test
  - [ ] Genesis block with test accounts pre-funded
  - [ ] Android app and `mesh_server` can connect to local endpoint
  - [ ] Documented setup script for new developer onboarding
  - [ ] Chain state resettable for reproducible testing

## R8: gRPC Control Bus Stable Under Load

- **Priority**: HIGH
- **Status**: FUNCTIONAL (verified for 7 agents, untested at 128-agent scale)
- **Location**: `grpc.go`, `grpc_node/`, `proto/sync.proto`
- **Acceptance Criteria**:
  - [ ] 128-agent concurrent handshake completes without errors
  - [ ] Load test: 1000 RPC calls/sec sustained for 60 seconds
  - [ ] Graceful degradation under overload (backpressure, not crash)
  - [ ] Connection recovery after network partition
  - [ ] Telemetry: latency percentiles (p50, p95, p99) logged

---

## Dependency Map

```
R5 (Security) ──────┐
                     ├──▶ R4 (CI/CD) ──▶ R2 (Web Deploy)
R3 (Go Tests) ──────┘                 ──▶ R1 (Android)
R7 (Local Chain) ──▶ R1 (Android)
R8 (gRPC Stable) ──▶ R3 (Go Tests)
R6 (Docs) ← depends on all others being stable
```
