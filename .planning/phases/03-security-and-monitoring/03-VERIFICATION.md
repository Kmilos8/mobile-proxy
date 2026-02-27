---
phase: 03-security-and-monitoring
verified: 2026-02-27T05:30:00Z
status: passed
score: 16/16 must-haves verified
re_verification: false
gaps: []
human_verification:
  - test: "Log in to dashboard, create a connection, click Regenerate on it, and close the dialog"
    expected: "New password shown in AlertDialog with copy button and 'will not be shown again' message; dialog closes cleanly; subsequent Regenerate calls produce a different password"
    why_human: "UI interaction flow and one-time display UX cannot be verified programmatically"
  - test: "Configure a webhook URL in dashboard Settings, click Send Test"
    expected: "Test payload POSTed to the URL; success message shows HTTP status code; no credentials leaked in payload"
    why_human: "Requires live HTTP server to receive the webhook; network delivery not verifiable from code alone"
  - test: "Connect an OpenVPN client, transfer data past the configured limit"
    expected: "Packets silently dropped; OpenVPN connection stays open but no data flows after limit"
    why_human: "Bandwidth enforcement is in the tunnel hot path; requires live traffic to observe drop behaviour"
  - test: "Disconnect a device's cellular for >2 minutes, then reconnect"
    expected: "Offline webhook fires within 5 minutes; recovery webhook fires on reconnect; no duplicate offline alert within 5-minute cooldown"
    why_human: "Requires live device and real timing; cooldown window and delivery ordering not testable from grep"
---

# Phase 3: Security and Monitoring Verification Report

**Phase Goal:** Harden credentials, enforce bandwidth limits, and alert on device offline before customer exposure
**Verified:** 2026-02-27T05:30:00Z
**Status:** PASSED
**Re-verification:** No — initial verification
**Commits verified:** 7952266, 1dab3c8 (Plan 01); c675997, b1058fb, c271306 (Plan 02) — all present in repo

---

## Goal Achievement

### Observable Truths — Plan 01 (SEC-01)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | OpenVPN auth validates passwords with bcrypt.CompareHashAndPassword, not plaintext | VERIFIED | `openvpn_handler.go:57` — `bcrypt.CompareHashAndPassword([]byte(conn.PasswordHash), []byte(req.Password))` |
| 2 | New connections store only bcrypt hash; PasswordPlain not exposed in API responses | VERIFIED | `connection_service.go:85-96` — no PasswordPlain assignment; `GetByID/List` return conn directly without setting `.Password = conn.PasswordPlain` |
| 3 | Operator can regenerate a connection password and receives new plaintext once | VERIFIED | `connection_service.go:199-225` — `RegeneratePassword` generates random 16-char pass, bcrypt cost 12, returns plaintext; `connection_handler.go:116-130` — handler present |
| 4 | .ovpn download embeds inline auth-user-pass when ?password= param present; prompts otherwise | VERIFIED | `openvpn_handler.go:224-231` — `c.Query("password")` conditional block confirmed |
| 5 | SOCKS5 tunnel path uses PasswordHash as credential (both tunnel server and device) | VERIFIED | `openvpn_handler.go:113` — `"socks_pass": conn.PasswordHash`; `device_handler.go:163-167` — `conn.PasswordHash != ""` condition, `Password: conn.PasswordHash` |
| 6 | Migration nulls password_plain; adds webhook_url and last_offline_alert_at columns | VERIFIED | `010_security_monitoring.up.sql` — all four statements present and correct |

**Score Plan 01:** 6/6 truths verified

### Observable Truths — Plan 02 (MON-01, MON-02)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 7  | Proxy connection exceeding bandwidth limit has packets dropped immediately in tunnel | VERIFIED | `tunnel/main.go:479-488` — atomic ctr.Add(n), `if limit > 0 && used > limit { continue }` before packet forward |
| 8  | Bandwidth counters survive tunnel restarts by loading current usage from DB at connect | VERIFIED | `tunnel/main.go:1062-1066` — `ctr.Store(req.BandwidthUsed)` in handleOpenVPNClientConnect; `openvpn_handler.go:114-115` — sends bandwidth_limit and bandwidth_used |
| 9  | Bandwidth usage flushed to DB every 30 seconds from tunnel server | VERIFIED | `tunnel/main.go:1177-1183` — `bandwidthFlushLoop` with 30s ticker; goroutine started at `startPushAPI:901`; posts to `/api/internal/bandwidth-flush` |
| 10 | Operator receives webhook POST within 5 minutes of device going offline | VERIFIED | `device_service.go:325-351` — `MarkStaleOfflineWithLogs` calls `go s.sendOfflineWebhook(ctx, d)` after marking offline; worker runs loop every 30s |
| 11 | No duplicate offline alerts within 5-minute cooldown window per device | VERIFIED | `device_service.go:420-423` — `time.Since(*lastAlert) < 5*time.Minute` check; `device_repo.go:84-93` — `GetLastOfflineAlertAt` / `SetLastOfflineAlertAt` methods present |
| 12 | Recovery webhook fires when previously-offline device reconnects | VERIFIED | `device_service.go:144-161` — `wasOffline := device.Status == domain.DeviceStatusOffline`; `go s.sendRecoveryWebhook(ctx, *device)` |
| 13 | Operator can configure webhook URL and test it from dashboard settings page | VERIFIED | `dashboard/src/app/settings/page.tsx` — full settings page with Save + Send Test buttons, loads on mount via `api.settings.getWebhook` |
| 14 | Operator can set bandwidth limit when creating a connection | VERIFIED | `AddConnectionModal.tsx:40,68-77` — `bandwidthLimitGB` state, GB-to-bytes conversion, sent as `bandwidth_limit` in create request |
| 15 | Operator can see bandwidth usage bar on each connection in dashboard | VERIFIED | `ConnectionTable.tsx:218,339` — `<BandwidthBar used={conn.bandwidth_used} limit={conn.bandwidth_limit} />`; `BandwidthBar.tsx` — substantive component with progress bar and color thresholds |
| 16 | Operator can reset bandwidth usage from dashboard | VERIFIED | `ConnectionTable.tsx:149,259,376` — `api.connections.resetBandwidth` call with RotateCcw icon button; `connection_handler.go:149-160` — handler present; `connection_repo.go:78` — `ResetBandwidthUsed` method |

**Score Plan 02:** 10/10 truths verified

**Total Score: 16/16 truths verified**

---

## Required Artifacts

### Plan 01 Artifacts

| Artifact | Expected | Exists | Substantive | Wired | Status |
|----------|----------|--------|-------------|-------|--------|
| `server/migrations/010_security_monitoring.up.sql` | Schema: nullable password_plain, webhook_url, last_offline_alert_at | Yes | Yes (4 SQL statements) | N/A — migration file | VERIFIED |
| `server/internal/api/handler/openvpn_handler.go` | bcrypt auth swap + conditional .ovpn embed | Yes | Yes (bcrypt import, CompareHashAndPassword, c.Query("password")) | Called by router at `/api/internal/openvpn/auth` | VERIFIED |
| `server/internal/api/handler/connection_handler.go` | RegeneratePassword handler | Yes | Yes (handler at line 116, full implementation) | Wired in router.go:81 `POST /connections/:id/regenerate-password` | VERIFIED |

### Plan 02 Artifacts

| Artifact | Expected | Exists | Substantive | Wired | Status |
|----------|----------|--------|-------------|-------|--------|
| `server/cmd/tunnel/main.go` | Per-connection bandwidth counters, enforcement, flush goroutine | Yes | Yes — `clientBandwidthUsed/Limit` maps, `ctr.Add`, `bandwidthFlushLoop`, `flushBandwidthToAPI` | Goroutine started in `startPushAPI:901`; initialized in `handleOpenVPNClientConnect` | VERIFIED |
| `server/cmd/worker/main.go` | Webhook dispatch integration | Yes | Yes — `userRepo` instantiated line 32, `deviceService.SetUserRepo(userRepo)` line 39 | `MarkStaleOfflineWithLogs` called in worker loop every 30s | VERIFIED |
| `server/internal/service/device_service.go` | Webhook dispatch with cooldown and recovery | Yes | Yes — `sendOfflineWebhook` (lines 412-453), `sendRecoveryWebhook` (lines 455-490), 5-min cooldown, userRepo nil guard | Called from `MarkStaleOfflineWithLogs` and `Heartbeat` | VERIFIED |
| `dashboard/src/app/settings/page.tsx` | Settings page with webhook URL input and Send Test | Yes | Yes — full page with `useEffect` load, controlled input, Save/Test buttons, success/error feedback | Accessible at `/settings`; `api.settings.*` functions wired | VERIFIED |

---

## Key Link Verification

### Plan 01 Key Links

| From | To | Via | Status | Evidence |
|------|----|-----|--------|----------|
| `openvpn_handler.go` | bcrypt | `CompareHashAndPassword` in Auth() | WIRED | Line 57 — `bcrypt.CompareHashAndPassword([]byte(conn.PasswordHash), []byte(req.Password))` |
| `router.go` | `connection_handler.go` | POST `/connections/:id/regenerate-password` route | WIRED | `router.go:81` — `dashboard.POST("/connections/:id/regenerate-password", connHandler.RegeneratePassword)` |
| `ConnectionTable.tsx` | `api.ts` | `regeneratePassword` API call | WIRED | `ConnectionTable.tsx:117` — `api.connections.regeneratePassword(token, conn.id)`; `api.ts:230-231` — function defined |

### Plan 02 Key Links

| From | To | Via | Status | Evidence |
|------|----|-----|--------|----------|
| `tunnel/main.go` | `/api/internal/bandwidth-flush` | periodic HTTP POST every 30s | WIRED | `flushBandwidthToAPI:1206` — `client.Post(apiURL+"/api/internal/bandwidth-flush", ...)`; `router.go:177` — internal route registered |
| `tunnel/main.go tunToUdp` | `clientBandwidthUsed` map | `ctr.Add` on each forwarded packet | WIRED | Lines 475-483 — counter read inside `routingMu.Lock()`, `ctr.Add(int64(n))` outside lock, then `if limit > 0 && used > limit { continue }` |
| `device_service.go` | `user webhook_url column` | `GetWebhookURLForDevice` repo call | WIRED | `device_service.go:425` — `s.userRepo.GetWebhookURLForDevice(ctx, d.ID)`; `user_repo.go:55` — method present |
| `settings/page.tsx` | `/api/settings/webhook` | PUT to save webhook URL | WIRED | `settings/page.tsx:36` — `api.settings.setWebhook(token, webhookUrl)`; `api.ts:238-239` — PUT request defined |
| `router.go` | `user_repo.go` | PUT `/settings/webhook` calls `userRepo.UpdateWebhookURL` | WIRED | `router.go:127` — `userRepo.UpdateWebhookURL(c.Request.Context(), uid, urlPtr)`; `user_repo.go:47` — method present |

---

## Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| SEC-01 | 03-01 | Replace plaintext proxy password storage with secure approach | SATISFIED | bcrypt.CompareHashAndPassword in Auth(); password_plain nulled in migration 010; PasswordPlain changed to *string in domain; SOCKS5 path uses PasswordHash as shared token |
| MON-01 | 03-02 | Device offline notification via email or webhook | SATISFIED | sendOfflineWebhook dispatched from MarkStaleOfflineWithLogs with 5-min cooldown; sendRecoveryWebhook on Heartbeat offline→online; Settings page for URL config |
| MON-02 | 03-02 | Enforce bandwidth limits per connection | SATISFIED | atomic.Int64 counters per client VPN IP; packet drop in tunToUdp hot path; counters loaded from DB at connect time; 30s flush to DB; BandwidthBar in dashboard |

All three requirements declared in plan frontmatter are accounted for and satisfied. No orphaned requirements found — REQUIREMENTS.md confirms SEC-01, MON-01, MON-02 all map to Phase 3 with status "Complete".

---

## Anti-Patterns Found

No blockers or stubs detected.

| File | Pattern | Severity | Notes |
|------|---------|----------|-------|
| `sync_service.go:63,85` | `PasswordPlain` field name in wire struct still used | Info | Field is named `password_plain` for wire compatibility but value sent is `c.PasswordHash` (documented decision). Not a bug — the SUMMARY explicitly notes this decision. |
| `user_repo.go:55-60` | `GetWebhookURLForDevice` returns first user's webhook URL (single-tenant assumption) | Info | Documented in plan as intentional for MVP single-operator system. Not a stub — actual DB query present. |

No TODO/FIXME/placeholder comments found in any phase-modified files. No empty handlers. No return-null stubs.

---

## Human Verification Required

### 1. Regenerate Password Dialog UX

**Test:** Log in to dashboard, find a connection, click the Regenerate (RefreshCw icon) button.
**Expected:** AlertDialog appears showing new password with a copy button and the message "Save this password now. It will not be shown again." Clicking Done or X clears the state. A second regenerate call produces a different password.
**Why human:** UI dialog flow, one-time display guarantee, and copy-to-clipboard functionality require manual interaction.

### 2. Webhook Delivery and Settings Page

**Test:** Navigate to `/settings`, enter a webhook URL (e.g., a RequestBin or webhook.site URL), click Save, then click Send Test.
**Expected:** HTTP 200 from test endpoint; success message shows the HTTP status code received. The webhook_url persists on page reload.
**Why human:** Requires a live HTTP receiver; network delivery and persistence cannot be verified from static analysis.

### 3. Bandwidth Enforcement Under Real Traffic

**Test:** Create a connection with a 1 MB bandwidth limit. Connect an OpenVPN client and transfer data. After crossing 1 MB, attempt further transfers.
**Expected:** Packets are dropped silently (transfer hangs or fails) without terminating the OpenVPN session. `bandwidth_used` in the DB updates within 30 seconds of transfer.
**Why human:** Enforcement is in the hot-path packet loop; requires actual traffic to observe drop behaviour.

### 4. Offline Webhook with Cooldown

**Test:** Configure webhook URL in Settings. Take a device offline (disable cellular) for >2 minutes. Verify webhook received. Re-trigger offline (bring device online then offline again within 5 minutes).
**Expected:** First offline event fires webhook. Second offline event within 5-minute window is suppressed. When device reconnects, recovery webhook fires.
**Why human:** Requires live device, real timing, and an external HTTP receiver.

---

## Summary

Phase 03 goal is fully achieved. All 16 must-haves across both plans are verified at all three levels (exists, substantive, wired):

**SEC-01 (Plan 01):** OpenVPN auth now uses `bcrypt.CompareHashAndPassword` — the old plaintext `conn.PasswordPlain != req.Password` comparison is gone. Migration 010 nulls the `password_plain` column and makes it nullable. The domain model changed `PasswordPlain` from `string` to `*string` (correct nil-safe type for NULL DB values). SOCKS5 credential path (Connect, Heartbeat, sync) all use `PasswordHash` as the shared token. The regenerate-password endpoint exists at `POST /api/connections/:id/regenerate-password`, is wired in the router, and the dashboard has a Regenerate button with a one-time AlertDialog display.

**MON-01 (Plan 02):** Offline webhook dispatch is integrated into `MarkStaleOfflineWithLogs` (called by the worker every 30 seconds). A 5-minute cooldown per device is enforced via `last_offline_alert_at` column (added in migration 010). Recovery webhooks fire from the Heartbeat path when `wasOffline == true`. The user repository correctly queries `webhook_url` and the settings page is fully functional with Save + Send Test buttons.

**MON-02 (Plan 02):** Bandwidth enforcement is live in the tunnel `tunToUdp` hot path using `atomic.Int64` counters (lock-free increment, drop on `used > limit`). Counters are initialized from DB at OpenVPN client connect time (`BandwidthUsed` sent in connect payload). A 30-second flush goroutine posts `{username -> bytes}` maps to `/api/internal/bandwidth-flush`. The dashboard shows `BandwidthBar` usage bars, bandwidth limit field in `AddConnectionModal`, and Reset Usage buttons in `ConnectionTable`.

Five commits verified as present in the repository: 7952266, 1dab3c8, c675997, b1058fb, c271306.

---

_Verified: 2026-02-27T05:30:00Z_
_Verifier: Claude (gsd-verifier)_
