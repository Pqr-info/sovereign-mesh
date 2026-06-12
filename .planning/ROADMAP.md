# Roadmap — v1.0 Production Launch

## Overview

Six phases ordered by dependency chain and risk. Security first (unblocks CI), then backend hardening (unblocks everything), then parallel client work, and documentation last.

---

## Phase 1: Security Hardening & Credential Purge

- **Status**: NOT STARTED
- **Requirements**: R5
- **Estimated Effort**: 1 day
- **Risk**: HIGH — credentials are live in git history

### Deliverables
1. Remove all credential files from tracking (`gcp-key.json`, `gcp_key.json`, `gcp_loader_key.json`, `fast-web-key.json`, `fast-web-key-new.json`, `.cloudflare_token`)
2. Update `.gitignore` with comprehensive secret patterns (`*.json` keys, tokens, `.env`)
3. Purge credentials from git history (BFG Repo Cleaner or `git filter-repo`)
4. Rotate all compromised keys in GCP and Cloudflare
5. Migrate secrets to GitHub Actions secrets for CI/CD use
6. Add pre-commit hook or CI check to prevent future credential leaks

### Exit Criteria
- `git log --all -p | grep -i "private_key"` returns zero matches
- All GCP service accounts re-keyed
- CI workflows reference `${{ secrets.* }}` instead of file paths

---

## Phase 2: Go Backend Hardening

- **Status**: NOT STARTED
- **Requirements**: R3, R8
- **Estimated Effort**: 3–5 days
- **Risk**: MEDIUM — limited test coverage, potential CGO leakage in dependencies

### Deliverables
1. Audit all `go.mod` dependencies for CGO requirements — replace any that break `CGO_ENABLED=0`
2. Write tests for critical paths:
   - gRPC handler registration and 128-agent handshake (`grpc.go`)
   - Ledger operations: mint, transfer, balance query (`ledger.go`)
   - Database fallback routing: PostgreSQL → SQLite switchover (`pkg/tickets`, `pkg/rtgo`)
   - Memory bus: shared memory read/write/concurrency (`memory.go`)
   - Stratum mining operations (`stratum.go`)
3. Run `go vet ./...` and fix all warnings
4. Run `staticcheck` or `golangci-lint` for deeper analysis
5. Load test gRPC bus: 128-agent concurrent handshake, 1K RPC/s sustained
6. Verify all 7 `cmd/` binaries build cleanly: `CGO_ENABLED=0 go build ./cmd/...`

### Exit Criteria
- `CGO_ENABLED=0 go test ./... -count=1` — zero failures
- `go vet ./...` — zero issues
- gRPC sustains 128 concurrent agents without error
- All binaries produce static ELF executables (verified with `file` and `ldd`)

---

## Phase 3: Android App — Fix & Ship

- **Status**: NOT STARTED
- **Requirements**: R1, R7
- **Estimated Effort**: 3–5 days
- **Risk**: MEDIUM — requires local blockchain endpoint (R7) for integration testing
- **Location**: `/home/aellok/triplehelix/android_app`

### Deliverables
1. Fix Gradle build errors — `./gradlew assembleDebug` must pass
2. Replace mock/stub blockchain calls with real Web3j integration
3. Configure app to connect to local PQR chain endpoint (Phase 5 dependency, can use stub initially)
4. Implement core screens: wallet dashboard, agent status grid, transaction history
5. UI polish: Material Design 3, dark/light theme, responsive layouts
6. Generate signed APK for testing

### Exit Criteria
- `./gradlew assembleDebug` — zero errors
- App installs and runs on Android emulator (API 34+)
- Wallet shows balances from live/local chain
- Agent status grid shows 128 agents with sync state

---

## Phase 4: CI/CD Pipeline

- **Status**: PARTIAL
- **Requirements**: R4
- **Estimated Effort**: 2 days
- **Risk**: LOW — workflows exist, need expansion
- **Dependencies**: Phase 1 (secrets in GitHub Actions), Phase 2 (Go tests pass)

### Deliverables
1. Expand `ci.yml` to include:
   - Go 1.26 setup + `CGO_ENABLED=0 go build ./cmd/...`
   - `go test ./...` + `go vet ./...`
   - React frontend `npm ci && npm run build`
   - Protobuf compilation validation
2. Add Android build job:
   - JDK 17 setup + Gradle cache
   - `./gradlew assembleDebug` (from triplehelix/android_app)
3. Enhance `deploy.yml`:
   - Build and publish Go binaries as release assets
   - Build and deploy React frontend to Cloudflare Pages
   - Create GitHub Release with changelog
4. Add credential leak scanner (`trufflehog` or `gitleaks`) to CI

### Exit Criteria
- Push to `main` triggers full green CI (Go + React + Proto + Android)
- Tag push creates GitHub Release with all artifacts
- Secret scanner blocks PRs with leaked credentials

---

## Phase 5: Web Frontend Deploy

- **Status**: NOT STARTED
- **Requirements**: R2
- **Estimated Effort**: 2–3 days
- **Risk**: LOW — React app exists with Vite build tooling
- **Dependencies**: Phase 2 (stable backend API)

### Deliverables
1. Fix any TypeScript compilation errors in `frontend/src/App.tsx` (35KB single-file — may need decomposition)
2. Configure API proxy to live `mesh_server` backend
3. Implement dashboard views:
   - 128-agent swarm status grid (live gRPC data)
   - PQR chain balances (PQR, RTGO, SOV, SOV2, LOMALO)
   - Memory bus throughput metrics
   - Ticket management interface
4. Deploy to Cloudflare Pages (or Workers)
5. Configure custom domain + SSL
6. Add error boundaries and loading states

### Exit Criteria
- `npm run build` — zero errors, zero warnings
- Live deployment accessible via public URL
- Dashboard renders real-time data from `mesh_server`
- Lighthouse score ≥ 80 (performance + accessibility)

---

## Phase 6: Documentation & Release Prep

- **Status**: NOT STARTED
- **Requirements**: R6, R7
- **Estimated Effort**: 2 days
- **Risk**: LOW
- **Dependencies**: All other phases substantially complete

### Deliverables
1. Rewrite `README.md`:
   - Project overview with architecture diagram (Mermaid)
   - Quickstart (clone, build, run mesh)
   - Environment requirements (Go 1.26, Node 22, Python 3.11)
2. API documentation:
   - gRPC service reference generated from `sync.proto`
   - REST endpoint docs (if any via web server)
3. Deployment guide:
   - Local development setup
   - Staging deployment (Cloudflare + GCP)
   - Production deployment checklist
4. Review and validate man pages (`.1`, `.7`, `.8` files)
5. Set up local PQR chain dev endpoint with documented bootstrap script
6. Create `CHANGELOG.md` for v1.0 release
7. Tag `v1.0.0` release

### Exit Criteria
- New developer can clone, build, and run the mesh following only README instructions
- All man pages render correctly via `man`
- gRPC API docs cover every RPC in `sync.proto`
- `v1.0.0` tag exists with GitHub Release + all artifacts
