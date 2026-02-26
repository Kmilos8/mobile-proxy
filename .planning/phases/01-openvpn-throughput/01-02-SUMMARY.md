---
phase: 01-openvpn-throughput
plan: 02
subsystem: infra
tags: [openvpn, iptables, shell, transparent-proxy, error-handling]

# Dependency graph
requires:
  - phase: 01-openvpn-throughput
    provides: "Plan 01-01 OpenVPN config tuning (peekTimeout, sndbuf/rcvbuf, fast-io)"
provides:
  - "client-connect-ovpn.sh with retry logic (2 attempts), wget timeout (5s), and exit 1 on API failure"
  - "client-disconnect-ovpn.sh with error visibility (exit 1 on failure, wget timeout)"
  - "PROTO-02 confirmed: HTTP/SOCKS5 DNAT rules structurally isolated from OpenVPN REDIRECT rules"
affects: [02-dashboard-auth, deployment, openvpn]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "OpenVPN client-connect hooks must exit non-zero to reject clients — never swallow API errors with || true"
    - "iptables REDIRECT (source IP match) and DNAT (destination port match) are orthogonal — changes to one never affect the other"

key-files:
  created: []
  modified:
    - server/deployments/openvpn/client-connect-ovpn.sh
    - server/deployments/openvpn/client-disconnect-ovpn.sh

key-decisions:
  - "Retry count set to 2 (one immediate attempt + one retry with 1s delay) — covers transient API errors without excessive delay on rejection"
  - "wget --timeout=5 added to prevent indefinite hang if API is unreachable"
  - "PROTO-02 confirmed via code review: REDIRECT rules use -s (source IP), DNAT rules use --dport (destination port) — no conflict possible"

patterns-established:
  - "OpenVPN hook pattern: retry loop with exit 1 on exhaustion; never use || true"

requirements-completed: [PROTO-01, PROTO-02]

# Metrics
duration: ~2min
completed: 2026-02-26
---

# Phase 1 Plan 02: Fix OpenVPN Client-Connect Silent Failure Summary

**OpenVPN client-connect hook rewritten with retry (2 attempts, 1s delay), 5s wget timeout, and exit 1 on API failure — eliminating silent admission without routing; PROTO-02 confirmed via code review.**

## Performance

- **Duration:** ~2 min
- **Started:** 2026-02-26T01:00:00Z
- **Completed:** 2026-02-26T01:11:00Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- Removed `|| true` from both client-connect and client-disconnect hooks — API failures now propagate correctly
- Added retry loop (2 attempts, 1 second between) to client-connect-ovpn.sh to cover transient errors
- Added `--timeout=5` to all wget calls to prevent indefinite hangs when API is unreachable
- client-connect-ovpn.sh now exits 1 on exhausted retries, causing OpenVPN to reject the client (correct UX — client retries automatically)
- PROTO-02 confirmed by human review: HTTP/SOCKS5 DNAT rules (matched by `--dport`) and OpenVPN REDIRECT rules (matched by `-s`) use different iptables criteria and cannot conflict

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix client-connect-ovpn.sh to reject clients on API failure** - `b598d26` (fix)
2. **Task 2: Verify DNAT/REDIRECT isolation (PROTO-02)** - checkpoint:human-verify (approved by user, no code changes)

**Plan metadata:** (see final docs commit)

## Files Created/Modified

- `server/deployments/openvpn/client-connect-ovpn.sh` - Rewritten with retry loop, wget timeout, exit 1 on API failure
- `server/deployments/openvpn/client-disconnect-ovpn.sh` - Rewritten with wget timeout, exit 1 on failure for log visibility

## Decisions Made

- Retry count of 2 chosen: one immediate attempt plus one retry after 1 second. Enough to survive transient blips without adding significant delay on a genuine API failure.
- wget `--timeout=5` added to both scripts. Without a timeout, a hung API could block the OpenVPN handshake indefinitely.
- PROTO-02 satisfied by code review rather than live testing. The structural isolation (different iptables match criteria) is self-evident from the source code at lines 543, 569, and 885 of `server/cmd/tunnel/main.go`.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Both PROTO-01 and PROTO-02 requirements satisfied
- Phase 01-openvpn-throughput complete: config tuning (plan 01) and hook error handling (plan 02) are both deployed
- Ready to proceed to Phase 02 (dashboard auth) or deploy and validate throughput improvements

---
*Phase: 01-openvpn-throughput*
*Completed: 2026-02-26*

## Self-Check: PASSED

- FOUND: server/deployments/openvpn/client-connect-ovpn.sh
- FOUND: server/deployments/openvpn/client-disconnect-ovpn.sh
- FOUND: .planning/phases/01-openvpn-throughput/01-02-SUMMARY.md
- FOUND: commit b598d26 (Task 1)
