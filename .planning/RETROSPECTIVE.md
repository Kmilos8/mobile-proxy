# Retrospective

## Milestone: v1.0 — MVP

**Shipped:** 2026-02-28
**Phases:** 4 | **Plans:** 9

### What Was Built
- OpenVPN throughput fixed (peekTimeout, OS autotuning, fast-io, client-connect retry)
- Full dashboard redesign with shadcn/ui (device table, connection CRUD, device detail)
- Security hardening (bcrypt auth, PasswordPlain removed, password regeneration)
- Bandwidth enforcement (per-connection limits, atomic counters, 30s flush)
- Monitoring (offline/recovery webhooks, settings page)
- Polish (search bar, auto-rotation column, connection ID)

### What Worked
- Sequential phase dependencies kept integration clean
- Checkpoint tasks (human-verify) caught real issues before moving forward
- Milestone audit identified 2 wiring bugs that were fixed in Phase 4

### What Was Inefficient
- Phase 3/4 plan checkboxes in ROADMAP.md not updated (cosmetic, didn't affect execution)
- Audit found gaps that could have been caught by more thorough verification in Phase 3

### Patterns Established
- Go handler/service/repo layering pattern
- shadcn/ui dark theme component conventions
- Tunnel push API pattern for cross-process state sync
- Atomic counter + periodic DB flush for bandwidth tracking

### Key Lessons
- Wire all dependencies in main.go immediately — nil guards silently swallow bugs
- In-memory state (tunnel counters) must be synced when DB state resets
- Human verification checkpoints are worth the pause

## Cross-Milestone Trends

| Milestone | Phases | Plans | Timeline |
|-----------|--------|-------|----------|
| v1.0 MVP  | 4      | 9     | 2 days   |
