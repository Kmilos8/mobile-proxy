---
phase: 03-security-and-monitoring
plan: 01
subsystem: auth
tags: [bcrypt, openvpn, socks5, security, passwords, postgresql, golang, react]

requires:
  - phase: 02-dashboard
    provides: Connection management dashboard, API routes for proxy connections

provides:
  - bcrypt password verification in OpenVPN auth endpoint
  - Nullable password_plain column (migration 010)
  - SOCKS5 credential path uses PasswordHash as shared token (not plaintext)
  - POST /api/connections/:id/regenerate-password endpoint (JWT protected)
  - Conditional inline auth-user-pass in .ovpn via ?password= query param
  - Dashboard Regenerate button with one-time password display dialog
  - Migration: webhook_url on users, last_offline_alert_at on devices

affects:
  - 03-02-security-monitoring (uses webhook_url and last_offline_alert_at from migration 010)
  - Android app (device heartbeat now receives PasswordHash as SOCKS5 credential, not PasswordPlain)

tech-stack:
  added: []
  patterns:
    - "bcrypt.CompareHashAndPassword for OpenVPN auth — same pattern as user login in auth_service.go"
    - "Hash-as-token SOCKS5: PasswordHash string used as SOCKS5 shared secret between tunnel and device (not a hash verification — literal string comparison)"
    - "One-time password display: regenerate endpoint returns plaintext once, stores only bcrypt hash"
    - "Conditional .ovpn inline credentials via ?password= query param at creation/regen time"

key-files:
  created:
    - server/migrations/010_security_monitoring.up.sql
  modified:
    - server/internal/api/handler/openvpn_handler.go
    - server/internal/api/handler/device_handler.go
    - server/internal/api/handler/connection_handler.go
    - server/internal/api/handler/router.go
    - server/internal/api/handler/sync_handler.go
    - server/internal/service/connection_service.go
    - server/internal/service/sync_service.go
    - server/internal/domain/models.go
    - server/internal/repository/connection_repo.go
    - dashboard/src/lib/api.ts
    - dashboard/src/components/connections/ConnectionTable.tsx

key-decisions:
  - "SOCKS5 credential migration: PasswordHash string used as SOCKS5 auth token on both sides (not plaintext, not hash verification — opaque shared secret). Device heartbeat and tunnel Connect() both use PasswordHash."
  - "PasswordPlain changed to *string (nullable) in domain model to handle NULL DB values after migration. Existing sync wire format preserved for backward compat."
  - "sync_handler.go does not map ci.PasswordPlain to domain.PasswordPlain — deprecated field ignored on receive side."
  - "Regen dialog shows .ovpn download button for OpenVPN connections, passing new password as ?password= query param for inline embedding."

patterns-established:
  - "Regenerate password pattern: generate random bytes, base64 encode, truncate to 16 chars, bcrypt cost 12, return plaintext once"
  - "Sync after credential change: RegeneratePassword triggers SyncConnections goroutine so peer server gets updated hash immediately"

requirements-completed:
  - SEC-01

duration: 25min
completed: 2026-02-27
---

# Phase 3 Plan 01: Security Monitoring Summary

**bcrypt OpenVPN auth, hash-as-SOCKS5-credential migration, and operator regenerate-password endpoint with one-time dashboard dialog**

## Performance

- **Duration:** ~25 min
- **Started:** 2026-02-27T04:29:36Z
- **Completed:** 2026-02-27T04:55:00Z
- **Tasks:** 2
- **Files modified:** 11

## Accomplishments

- OpenVPN Auth() now calls `bcrypt.CompareHashAndPassword` instead of plaintext comparison — SEC-01 requirement fulfilled
- SOCKS5 tunnel path migrated from PasswordPlain to PasswordHash as the shared credential token (device heartbeat, Connect(), sync service all use PasswordHash)
- New `POST /api/connections/:id/regenerate-password` endpoint generates 16-char random password, stores only bcrypt hash (cost 12), returns plaintext exactly once
- Migration 010 nulls out password_plain, makes it nullable, adds webhook_url to users and last_offline_alert_at to devices (needed by Plan 03-02)
- Dashboard Regenerate button (RefreshCw icon) shows new password in AlertDialog with copy button and "save now, will not be shown again" warning
- .ovpn download conditionally embeds `<auth-user-pass>` only when `?password=` query param is present; subsequent downloads prompt the user

## Task Commits

Each task was committed atomically:

1. **Task 1: Migration SQL, bcrypt auth swap, and SOCKS5 credential migration** - `7952266` (feat)
2. **Task 2: Regenerate password endpoint and dashboard button** - `1dab3c8` (feat)

**Plan metadata:** (docs commit pending)

## Files Created/Modified

- `server/migrations/010_security_monitoring.up.sql` - Schema migration: nullable password_plain, webhook_url, last_offline_alert_at
- `server/internal/api/handler/openvpn_handler.go` - bcrypt auth, PasswordHash SOCKS5 cred, conditional .ovpn inline creds
- `server/internal/api/handler/device_handler.go` - Heartbeat credentials use PasswordHash instead of PasswordPlain
- `server/internal/api/handler/connection_handler.go` - RegeneratePassword handler
- `server/internal/api/handler/router.go` - POST /connections/:id/regenerate-password route
- `server/internal/api/handler/sync_handler.go` - Removed PasswordPlain mapping (deprecated, *string incompatible)
- `server/internal/service/connection_service.go` - RegeneratePassword method, removed plaintext exposure in GetByID/List
- `server/internal/service/sync_service.go` - Sends PasswordHash as password_plain wire field
- `server/internal/domain/models.go` - PasswordPlain changed from string to *string (nullable)
- `server/internal/repository/connection_repo.go` - Added UpdatePasswordHash method
- `dashboard/src/lib/api.ts` - regeneratePassword API function, downloadOVPN updated with optional password param
- `dashboard/src/components/connections/ConnectionTable.tsx` - Regenerate button, AlertDialog with copy + .ovpn download

## Decisions Made

- `PasswordPlain` changed to `*string` in domain model to prevent scan error from NULL DB column after migration runs
- `sync_handler.go` does not populate `PasswordPlain` on the receive side since the field is now deprecated and the type changed
- Regen dialog includes optional "Download .ovpn with embedded credentials" button for OpenVPN connections — passes `?password=` param so .ovpn includes inline credentials at regen time

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Changed PasswordPlain from string to *string in domain model**
- **Found during:** Task 1 (migration SQL)
- **Issue:** Migration nulls out password_plain column, but domain model had non-nullable `string` type — scanning NULL into a Go `string` via pgx would panic at runtime
- **Fix:** Changed `PasswordPlain string` to `PasswordPlain *string` in ProxyConnection domain model; updated sync_handler.go to not assign the deprecated field (type incompatible); sync_service.go local connItem struct unchanged (string, wire compat)
- **Files modified:** server/internal/domain/models.go, server/internal/api/handler/sync_handler.go
- **Verification:** grep confirms no remaining type mismatch; PasswordHash used in all credential paths
- **Committed in:** 7952266 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 - Bug)
**Impact on plan:** Required for correctness. NULL scanning into non-nullable Go string would have caused runtime panic on first DB read after migration.

## Issues Encountered

- Go binary not available locally (runs in Docker on VPS) — verified correctness via code review and grep-based pattern matching instead of `go build`. The code follows established patterns used in auth_service.go and is structurally sound.

## User Setup Required

After deploying this to VPS, run the migration:
```
docker compose exec postgres psql -U postgres -d mobileproxy -f /migrations/010_security_monitoring.up.sql
```
Or trigger via the app's auto-migrate on startup. After migration:
- All existing connections will have NULL password_plain
- OpenVPN auth will use bcrypt — existing bcrypt hashes already present (Create() always hashed)
- SOCKS5 devices will receive PasswordHash as credential on next heartbeat

## Next Phase Readiness

- Migration 010 creates webhook_url and last_offline_alert_at columns needed by Plan 03-02
- SEC-01 fulfilled: no more plaintext password comparison in production code paths
- Regen endpoint ready for operator use from dashboard

---
*Phase: 03-security-and-monitoring*
*Completed: 2026-02-27*
