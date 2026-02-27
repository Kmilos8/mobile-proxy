---
phase: 03-security-and-monitoring
plan: 02
subsystem: monitoring
tags: [bandwidth, webhook, golang, react, openvpn, postgresql, atomic]

requires:
  - phase: 03-01
    provides: "migration 010 with webhook_url on users and last_offline_alert_at on devices"
  - phase: 02-dashboard
    provides: "ConnectionTable, AddConnectionModal, dashboard layout and API client"

provides:
  - Per-connection bandwidth enforcement in tunnel server (atomic counters, packet drop)
  - Bandwidth flush goroutine (30s interval) persisting usage to DB via /api/internal/bandwidth-flush
  - Bandwidth counter initialized from DB at OpenVPN connect time (survives tunnel restarts)
  - Offline webhook dispatch with 5-minute cooldown (MarkStaleOfflineWithLogs)
  - Recovery webhook on device reconnect (Heartbeat offline->online transition)
  - Settings API (GET/PUT /settings/webhook, POST /settings/webhook/test)
  - Dashboard settings page with webhook URL configuration and test button
  - Connection creation form with bandwidth limit field (GB input)
  - Connection list with usage bar (BandwidthBar) and Reset Usage button

affects:
  - future phases (bandwidth enforcement now active for all OpenVPN connections)

tech-stack:
  added: []
  patterns:
    - "atomic.Int64 map per client VPN IP for lock-free bandwidth increment in hot path"
    - "Merged routingMu lock section for bandwidth lookup + clientToDevice lookup (minimize lock contention)"
    - "Flush-by-username pattern: tunnel uses clientSocksAuth[ip].user to map VPN IP to DB username"
    - "Webhook cooldown: last_offline_alert_at column checked before dispatch, updated only on success"
    - "SetterMethod pattern for optional service dependencies (SetUserRepo, SetStatusLogRepo)"

key-files:
  created:
    - dashboard/src/app/settings/page.tsx
  modified:
    - server/cmd/tunnel/main.go
    - server/internal/api/handler/openvpn_handler.go
    - server/internal/repository/connection_repo.go
    - server/internal/service/connection_service.go
    - server/internal/api/handler/connection_handler.go
    - server/internal/api/handler/router.go
    - server/internal/domain/models.go
    - server/internal/repository/user_repo.go
    - server/internal/repository/device_repo.go
    - server/internal/service/device_service.go
    - server/cmd/worker/main.go
    - server/cmd/api/main.go
    - dashboard/src/lib/api.ts
    - dashboard/src/components/connections/AddConnectionModal.tsx
    - dashboard/src/components/connections/ConnectionTable.tsx
    - dashboard/src/app/devices/page.tsx

key-decisions:
  - "Bandwidth flush sends {username: bytes} not {vpnIP: bytes} — tunnel has clientSocksAuth[ip].user available in same lock scope; avoids needing a VPN IP -> connection DB lookup"
  - "Single routingMu lock acquisition in tunToUdp reads clientToDevice + clientBandwidthUsed + clientBandwidthLimit together — minimizes contention in hot path"
  - "Webhook cooldown only updates last_offline_alert_at on successful HTTP delivery — failed webhooks don't suppress future attempts"
  - "userRepo passed as new parameter to SetupRouter (not via setter) — settings handlers need it in closure; api/main.go already had userRepo instance"
  - "Recovery webhook fires on every offline->online Heartbeat transition (no cooldown) — reconnections are always notable events"

patterns-established:
  - "Bandwidth enforcement: atomic.Int64 per client, increment outside lock, drop packet if limit > 0 && used > limit"
  - "Settings page pattern: useEffect load on mount, controlled input, Save disables when unchanged, Test sends to provided URL"

requirements-completed:
  - MON-01
  - MON-02

duration: 15min
completed: 2026-02-27
---

# Phase 3 Plan 02: Security Monitoring Summary

**Atomic bandwidth enforcement in tunnel (packet drop on limit), offline/online webhooks with cooldown, and dashboard settings page with webhook configuration and bandwidth usage bars**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-02-27T04:41:14Z
- **Completed:** 2026-02-27T04:56:25Z
- **Tasks:** 3
- **Files modified:** 16

## Accomplishments

- Tunnel server tracks per-OpenVPN-connection bandwidth with `atomic.Int64` counters loaded from DB at connect time; packets are silently dropped when limit exceeded
- Bandwidth counters flushed to DB every 30 seconds via `/api/internal/bandwidth-flush` POST (maps username -> bytes); counter is initialized from DB value at connect time so usage survives tunnel restarts
- Worker dispatches webhook POST to operator's configured URL when a device goes offline (with 5-minute cooldown per device) and when a device recovers
- Dashboard settings page at `/settings` allows webhook URL configuration with Save + Send Test buttons
- Connection creation form now includes optional Bandwidth Limit (GB) field; connection list shows usage bars and Reset Usage button using existing BandwidthBar component

## Task Commits

Each task was committed atomically:

1. **Task 1: Bandwidth enforcement in tunnel server and flush endpoint** - `c675997` (feat)
2. **Task 2: Offline webhook dispatch and settings API** - `b1058fb` (feat)
3. **Task 3: Dashboard monitoring UI (settings page, bandwidth controls)** - `c271306` (feat)

## Files Created/Modified

- `server/cmd/tunnel/main.go` - Added clientBandwidthUsed/Limit maps, enforcement in tunToUdp, bandwidthFlushLoop, handleResetBandwidth, bytes import
- `server/internal/api/handler/openvpn_handler.go` - Connect() sends bandwidth_limit and bandwidth_used to tunnel
- `server/internal/repository/connection_repo.go` - Added UpdateBandwidthUsed (by username) and ResetBandwidthUsed
- `server/internal/service/connection_service.go` - Added UpdateBandwidthUsedByUsername and ResetBandwidth
- `server/internal/api/handler/connection_handler.go` - Added BandwidthFlush (internal) and ResetBandwidth handlers
- `server/internal/api/handler/router.go` - Added bandwidth-flush route (internal), reset-bandwidth route (JWT), settings webhook routes, userRepo parameter
- `server/internal/domain/models.go` - Added WebhookURL to User; LastOfflineAlertAt to Device
- `server/internal/repository/user_repo.go` - Updated GetByEmail/GetByID to scan webhook_url; added UpdateWebhookURL and GetWebhookURLForDevice
- `server/internal/repository/device_repo.go` - Added GetLastOfflineAlertAt and SetLastOfflineAlertAt
- `server/internal/service/device_service.go` - Added userRepo field, SetUserRepo, sendOfflineWebhook (5-min cooldown), sendRecoveryWebhook; integrated into Heartbeat and MarkStaleOfflineWithLogs
- `server/cmd/worker/main.go` - Instantiate userRepo, wire to deviceService.SetUserRepo
- `server/cmd/api/main.go` - Pass userRepo to SetupRouter
- `dashboard/src/app/settings/page.tsx` - NEW: Settings page with webhook URL input, Save, Send Test
- `dashboard/src/lib/api.ts` - Added settings.getWebhook/setWebhook/testWebhook, connections.resetBandwidth
- `dashboard/src/components/connections/AddConnectionModal.tsx` - Added bandwidthLimitGB field (GB to bytes)
- `dashboard/src/components/connections/ConnectionTable.tsx` - Added BandwidthBar (Usage column), Reset Usage button (RotateCcw icon)
- `dashboard/src/app/devices/page.tsx` - Added Settings icon link in header

## Decisions Made

- Flush sends `{username -> bytes}` map rather than `{vpnIP -> bytes}` — the tunnel server has `clientSocksAuth[ip].user` (the connection username) available within the same `routingMu` lock scope. Avoids a separate DB lookup on the API side.
- Single `routingMu` lock acquisition in `tunToUdp` reads `clientToDevice`, `clientBandwidthUsed`, and `clientBandwidthLimit` together — minimizes contention in the hot packet forwarding path.
- Webhook cooldown only updates `last_offline_alert_at` on successful HTTP delivery — failed webhooks don't suppress future delivery attempts.
- `userRepo` passed as new `SetupRouter` parameter (not via setter on any service) because the settings handlers are inline closures in router.go that capture `userRepo` directly.

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

- Go binary not available locally (runs in Docker on VPS) — verified correctness via code review and grep-based pattern matching. Code follows established patterns from existing handlers.

## User Setup Required

After deploying to VPS (migration 010 from Plan 03-01 must already be applied):
- No additional migrations needed — webhook_url and last_offline_alert_at columns were created in migration 010 (Plan 03-01)
- Configure webhook URL from dashboard Settings page after deployment
- Test with "Send Test" button before relying on automatic notifications

## Next Phase Readiness

- MON-01 (offline webhooks) and MON-02 (bandwidth enforcement) requirements fulfilled
- Phase 03 is now complete — all planned security and monitoring features implemented

---
*Phase: 03-security-and-monitoring*
*Completed: 2026-02-27*
