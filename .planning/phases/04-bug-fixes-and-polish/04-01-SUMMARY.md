---
phase: 04-bug-fixes-and-polish
plan: 01
subsystem: api
tags: [go, react, webhook, bandwidth, openvpn, tunnel, dashboard]

# Dependency graph
requires:
  - phase: 03-security-and-monitoring
    provides: bandwidth flush infrastructure, webhook dispatch mechanism, device_service.userRepo field
provides:
  - MON-01 recovery webhook now fires on device reconnect (SetUserRepo wired in api/main.go)
  - MON-02 bandwidth reset propagated to tunnel in-memory counter via HTTP POST
  - DASH-02 OpenVPN option added to Add Connection modal protocol selector
affects: [deployment, monitoring, dashboard]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Fire-and-forget goroutine for best-effort tunnel sync after DB write"
    - "Username-based reverse lookup in tunnel (iterate clientSocksAuth map to resolve VPN IP)"

key-files:
  created: []
  modified:
    - server/cmd/api/main.go
    - server/cmd/tunnel/main.go
    - server/internal/service/connection_service.go
    - dashboard/src/components/connections/AddConnectionModal.tsx

key-decisions:
  - "resetTunnelBandwidth is fire-and-forget — DB reset is source of truth, tunnel counter is best-effort"
  - "Tunnel bandwidth reset uses username field (not client_vpn_ip) — service layer does not track VPN IPs"
  - "Single routingMu.Lock/Unlock guards both clientSocksAuth and clientBandwidthUsed access"

patterns-established:
  - "Best-effort goroutine pattern: DB op succeeds -> return nil -> goroutine fires tunnel side-effect"

requirements-completed: [MON-01, MON-02, DASH-02]

# Metrics
duration: 13min
completed: 2026-02-27
---

# Phase 4 Plan 01: Bug Fixes and Polish Summary

**Recovery webhook wired (MON-01), bandwidth reset propagated to tunnel (MON-02), and OpenVPN added to Add Connection modal (DASH-02) — three v1.0 audit gaps closed**

## Performance

- **Duration:** 13 min
- **Started:** 2026-02-27T05:44:00Z
- **Completed:** 2026-02-27T05:57:00Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments
- MON-01: `deviceService.SetUserRepo(userRepo)` wired in api/main.go — recovery webhooks now fire when devices come back online after being offline
- MON-02: `ResetBandwidth` in connection_service.go now calls tunnel push API after DB reset — 30s bandwidth flush can no longer overwrite a zeroed counter; tunnel also accepts `username` field for reverse lookup
- DASH-02: AddConnectionModal protocol selector now includes HTTP, SOCKS5, and OpenVPN — eliminates "must be http or socks5" error when creating OpenVPN connections from dashboard

## Task Commits

Each task was committed atomically:

1. **Task 1: Wire recovery webhook and extend tunnel bandwidth reset** - `ec584fd` (feat)
2. **Task 2: Propagate bandwidth reset from service to tunnel** - `6fb527f` (feat)
3. **Task 3: Add OpenVPN option to Add Connection modal** - `7389645` (feat)

## Files Created/Modified
- `server/cmd/api/main.go` - Added `deviceService.SetUserRepo(userRepo)` after SetStatusLogRepo
- `server/cmd/tunnel/main.go` - Extended `handleResetBandwidth` to accept `username` field with reverse lookup via `clientSocksAuth` map
- `server/internal/service/connection_service.go` - Replaced `ResetBandwidth` with version that calls tunnel API; added `resetTunnelBandwidth` fire-and-forget goroutine
- `dashboard/src/components/connections/AddConnectionModal.tsx` - Widened `proxyType` type union to include `'openvpn'`, added OpenVPN SelectItem

## Decisions Made
- `resetTunnelBandwidth` is fire-and-forget (goroutine): DB reset is the source of truth. If tunnel call fails, the counter becomes consistent at the next client connect event. This avoids blocking the HTTP response on a tunnel side-effect.
- Tunnel bandwidth reset uses `username` (not `client_vpn_ip`) because the service layer holds connection records by username, not by VPN IP — IP-to-username resolution is the tunnel's responsibility via reverse lookup.
- Single `routingMu.Lock()/Unlock()` covers both `clientSocksAuth` read and `clientBandwidthUsed` write — avoids nested locking as specified in plan.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- Go toolchain not installed locally on Windows machine; build verification attempted via Docker on VPS. VPS code was on an older git state that was missing `SetUserRepo` definition and had different `SetupRouter` signature — confirmed these were pre-existing in the local codebase. TypeScript compilation verified successfully locally (passes `tsc --noEmit`).

## User Setup Required
None - no external service configuration required. Changes take effect on next deploy (`git pull` + `docker compose up --build` on VPS).

## Next Phase Readiness
- All three v1.0 audit gaps (MON-01, MON-02, DASH-02) are closed
- Codebase ready for deployment to VPS to activate fixes in production
- No blockers for next plan

---
*Phase: 04-bug-fixes-and-polish*
*Completed: 2026-02-27*
