# 🛡️ Forensic Commit Protocol (FCP)

The FCP is the mandatory workflow for all autonomous mesh-wide modifications.

## 📜 The Core Rule
**"No change exists until it is forensically anchored."**
Every file modification MUST be preceded by or accompanied by an `IntentBlob` in the CockroachDB Ticketing Fabric.

## 🔄 The FCP Workflow
1. **Record Ticket**: Create a ticket defining the intent of the change.
2. **Inject Diff**: Post the `git diff` or raw content changes to the ticket's `IntentBlob`.
3. **Execute Mutation**: Apply the code change to the local filesystem.
4. **Update Manifest**: Recalculate the SHA-256 hash in `manifest.json`.

## 📊 Database-First Pattern
When an agent edits a file (e.g., `server.go`):
- It MUST call `s.Service.CreateFabricTicket` with the diff.
- It MUST include the **Genesis Ticket ID** as a parent to maintain the lineage.

## 🕵️ Forensic Auditing
The `Forensic Auditor` agent (council-003) continuously compares the live filesystem hashes against `manifest.json` and the Ticketing Fabric. Any "Shadow Commit" (change without a ticket) triggers an automatic roll-back and an alert on the Sovereign HUD.
