# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-25)

**Core value:** Customers can reliably route traffic through real mobile devices via HTTP, SOCKS5, or OpenVPN, managed through a clean dashboard.
**Current focus:** Phase 1 — OpenVPN Throughput

## Current Position

Phase: 1 of 3 (OpenVPN Throughput)
Plan: 1 of 2 in current phase
Status: In progress
Last activity: 2026-02-26 — Completed 01-01 (OpenVPN throughput tuning)

Progress: [█░░░░░░░░░] 17%

## Performance Metrics

**Velocity:**
- Total plans completed: 1
- Average duration: ~1 min
- Total execution time: ~1 min

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-openvpn-throughput | 1/2 | ~1 min | ~1 min |

**Recent Trend:**
- Last 5 plans: 01-01 (~1 min)
- Trend: —

*Updated after each plan completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Roadmap: Fix protocols before dashboard — stable proxy is the core product
- Research: OpenVPN bottleneck is transparent proxy + config tuning (not infrastructure); diagnose with tcpdump before applying fixes
- Research: PasswordPlain field is a pre-production blocker — must not expose to customers until removed
- 01-01: peekTimeout set to 200ms — TLS ClientHello arrives in <10ms; 200ms safe margin without blocking speed tests
- 01-01: sndbuf/rcvbuf value 0 (OS autotuning) — fixed 524288 cap documented to limit throughput at ~5 Mbps vs 60 Mbps with autotuning
- 01-01: fast-io added — UDP-only optimization that skips poll/select before each UDP write (5-10% CPU reduction)

### Pending Todos

None yet.

### Blockers/Concerns

- kylemanna/openvpn image is unmaintained (OpenVPN 2.4.9); defer image replacement until throughput is confirmed fixed

## Session Continuity

Last session: 2026-02-26
Stopped at: Completed 01-01-PLAN.md (OpenVPN throughput tuning — peekTimeout, sndbuf/rcvbuf, fast-io)
Resume file: None
