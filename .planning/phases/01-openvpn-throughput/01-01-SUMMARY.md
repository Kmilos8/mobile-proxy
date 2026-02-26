---
phase: 01-openvpn-throughput
plan: 01
subsystem: infra
tags: [openvpn, tcp, udp, networking, performance, tuning]

# Dependency graph
requires: []
provides:
  - Reduced transparent proxy peekTimeout (2s -> 200ms) eliminating per-connection latency penalty
  - OS socket buffer autotuning enabled via sndbuf 0 / rcvbuf 0 in OpenVPN server config
  - OS socket buffer autotuning in generated .ovpn client files
  - fast-io directive in OpenVPN server config for reduced CPU overhead on UDP writes
affects: [02-openvpn-throughput, deploy, throughput-testing]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "OS TCP autotuning: value 0 for sndbuf/rcvbuf delegates buffer sizing to kernel"
    - "peekTimeout at 200ms: covers slow clients while eliminating speed-test blocking delays"

key-files:
  created: []
  modified:
    - server/internal/transparentproxy/proxy.go
    - server/deployments/openvpn/client-server.conf
    - server/internal/api/handler/openvpn_handler.go

key-decisions:
  - "peekTimeout set to 200ms: TLS ClientHello arrives in <10ms; 200ms is safe margin without blocking speed tests"
  - "sndbuf/rcvbuf set to 0 (OS autotuning): fixed 524288 cap was documented to limit throughput at ~5 Mbps vs 60 Mbps with autotuning"
  - "fast-io added: UDP-only optimization that skips poll/select before writes, reducing CPU overhead 5-10%"

patterns-established:
  - "OpenVPN perf tuning: OS autotuning (value 0) preferred over fixed buffer sizes for mobile proxy workloads"

requirements-completed: [PROTO-01]

# Metrics
duration: 1min
completed: 2026-02-26
---

# Phase 1 Plan 01: OpenVPN Throughput Tuning Summary

**peekTimeout cut from 2s to 200ms and socket buffers switched to OS autotuning (sndbuf/rcvbuf 0) with fast-io enabled, removing the two highest-confidence OpenVPN throughput bottlenecks**

## Performance

- **Duration:** ~1 min
- **Started:** 2026-02-26T01:06:21Z
- **Completed:** 2026-02-26T01:07:18Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Eliminated per-connection latency penalty: peekTimeout reduced from 2s to 200ms — new connections no longer block for up to 2 seconds before proxying begins
- Enabled OS TCP autotuning: sndbuf/rcvbuf changed from fixed 524288 (512KB cap) to 0 in both server config push directives and generated .ovpn files
- Reduced CPU overhead on UDP writes by adding fast-io directive to OpenVPN server config

## Task Commits

Each task was committed atomically:

1. **Task 1: Reduce peekTimeout from 2s to 200ms** - `4bfb8f2` (fix)
2. **Task 2: Switch sndbuf/rcvbuf to OS autotuning and add fast-io** - `1843118` (fix)

**Plan metadata:** (docs commit — see below)

## Files Created/Modified
- `server/internal/transparentproxy/proxy.go` - peekTimeout constant changed from 2 * time.Second to 200 * time.Millisecond with updated comment
- `server/deployments/openvpn/client-server.conf` - sndbuf/rcvbuf changed to 0 in all 4 lines (server-side + push directives), fast-io directive added after txqueuelen
- `server/internal/api/handler/openvpn_handler.go` - sndbuf/rcvbuf changed to 0 in DownloadOVPN .ovpn generation

## Decisions Made
- peekTimeout set to 200ms: TLS ClientHello arrives in under 10ms on normal connections; 200ms provides ample headroom for slow clients without the original 2s penalty that blocked speed test measurements
- sndbuf/rcvbuf value 0: delegates buffer sizing to Linux kernel autotuning; community testing documented 5 Mbps to 60 Mbps throughput improvement when removing the fixed 524288 cap
- fast-io included: applicable to UDP-only mode (which client-server.conf uses), skips poll/select before each UDP write

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required. Changes take effect on next OpenVPN server restart.

## Self-Check: PASSED

All modified files confirmed present. Both task commits (4bfb8f2, 1843118) confirmed in git log.
