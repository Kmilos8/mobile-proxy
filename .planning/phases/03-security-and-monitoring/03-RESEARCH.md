# Phase 3: Security and Monitoring - Research

**Researched:** 2026-02-26
**Domain:** Go backend security (bcrypt), bandwidth enforcement (in-memory counters), webhook delivery, PostgreSQL schema migration, Next.js 14 dashboard UI
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Credential Migration:**
- Batch migration on deploy: one-time script hashes all existing PasswordPlain values to bcrypt, then nulls out the plaintext field
- Active OpenVPN sessions continue undisturbed — only new connections use bcrypt auth
- Dashboard gets a "Regenerate Password" action per connection — generates new random password, stores bcrypt hash, shows plaintext once
- Password displayed once after creation or regeneration; after navigating away, plaintext is gone — must regenerate to see again
- .ovpn file includes auth-user-pass inline at creation/regen time

**Bandwidth Enforcement:**
- Track bytes per customer connection in the tunnel server (Go, main.go) — it already handles routing, natural place to count
- Hard cutoff immediately when limit is hit — stop forwarding packets for that connection
- Per proxy connection granularity — each OpenVPN profile has its own bandwidth limit, set by operator at creation
- Manual reset by operator only — "Reset Usage" action in dashboard, no automatic reset cycle

**Offline Alerting:**
- Webhook-only notification channel — POST to operator-configured URL
- Offline detection: no heartbeat received for 2 minutes from the device tunnel
- 5-minute cooldown after sending an offline alert for the same device — prevents alert storms from flapping connections
- Recovery notification: send a second webhook when device reconnects after being offline

**Operator Configuration:**
- Bandwidth limit field on the connection create/edit form in the dashboard, stored in database alongside connection
- Per-operator webhook URL setting in dashboard — applies to all their devices
- Usage bar on each connection card in the dashboard (e.g., "750 MB / 1 GB" progress bar)
- "Send Test" button next to webhook URL field — sends sample payload for operator to verify endpoint

### Claude's Discretion
- Webhook payload format and structure
- Exact bcrypt cost factor
- Migration script error handling and rollback approach
- Usage tracking persistence strategy (in-memory vs database flush interval)

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| SEC-01 | Replace plaintext proxy password storage with secure approach (bcrypt hash exists but PasswordPlain also stored and used by OpenVPN auth) | Migration SQL + bcrypt.CompareHashAndPassword swap in openvpn_handler.go Auth(); remove PasswordPlain from Create/update paths |
| MON-01 | Device offline notification via email or webhook | Worker goroutine detects offline transitions (already calls MarkStaleOfflineWithLogs); add webhook POST in the transition path; per-operator webhook URL stored in DB |
| MON-02 | Enforce bandwidth limits per connection (field exists but not enforced) | `bandwidth_used` column exists on `proxy_connections`; add per-connection `atomic.Int64` counter in tunnel server; periodic flush to DB; check limit before forwarding in `tunToUdp` hot path |
</phase_requirements>

---

## Summary

Phase 3 has three independent workstreams that share no code paths and can be planned and executed sequentially. Each has clear insertion points in the existing codebase.

**SEC-01 (bcrypt migration)** is the most self-contained. The codebase already imports `golang.org/x/crypto/bcrypt` and uses it for admin user passwords and for storing `password_hash` on `proxy_connections`. The only gap is that `openvpn_handler.go` Auth() compares against `conn.PasswordPlain` instead of calling `bcrypt.CompareHashAndPassword`. The fix is: (1) migration SQL to hash existing plaintext values and null the column, (2) swap the comparison in Auth(), (3) update Create() and a new Regenerate endpoint to stop writing PasswordPlain, (4) store the one-time plaintext in the response only.

**MON-02 (bandwidth enforcement)** requires the most care because the enforcement point is the hot-path packet-forwarding loop (`tunToUdp` in `cmd/tunnel/main.go`). The tunnel server is a single long-running process with no database access. The right pattern is: add per-connection `atomic.Int64` counters to `tunnelServer` (keyed by client VPN IP 10.9.0.x), increment on every packet forwarded, check against limit, and periodically flush to the `bandwidth_used` column via the API. The limit value must be loaded into the tunnel's in-memory state at `handleOpenVPNClientConnect` time (API passes it alongside `socks_user`/`socks_pass`). Cutoff drops packets immediately without logging the client out of OpenVPN.

**MON-01 (offline alerting)** hooks into the existing `MarkStaleOfflineWithLogs` loop already running every 30 seconds in the worker. A webhook URL per operator stored in the `users` table (new column) and a cooldown tracker (in-memory map or `last_alerted_at` DB column) complete the picture.

**Primary recommendation:** Implement SEC-01 first (risk-free DB migration + single-file auth change), then MON-02 (tunnel server changes with API flush endpoint), then MON-01 (worker-side webhook dispatch + dashboard settings page).

---

## Standard Stack

### Core (already in project, no new dependencies required)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `golang.org/x/crypto/bcrypt` | v0.18.0 (already in go.mod) | Password hashing and comparison | Already used for admin passwords; `DefaultCost` is 10, recommend cost 12 |
| `sync/atomic` (stdlib) | Go 1.23 | Lock-free per-connection byte counters in tunnel hot path | Zero-dependency; `atomic.Int64` type is idiomatic Go 1.19+ |
| `net/http` (stdlib) | Go 1.23 | Webhook POST delivery from worker | No external dependency needed; simple fire-and-forget with timeout |
| `pgx/v5` | v5.5.1 (already in go.mod) | DB flush of bandwidth counters and webhook URL storage | Already the project's DB driver |
| Next.js 14 + shadcn/ui + lucide-react | already in dashboard | Dashboard settings page, bandwidth bar, regenerate button | Already the project stack |

### Supporting (no new installs)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `time.NewTicker` (stdlib) | — | Periodic DB flush of bandwidth counters (every 30s) | Run in a goroutine inside tunnel server |
| `Progress` via Tailwind width style | — | Bandwidth usage bar in dashboard | Simple div with percentage width; no chart library needed |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `sync/atomic.Int64` per-connection | `sync.Map` with mutex | Atomic is faster for single counters; sync.Map needed if map key set changes dynamically — here clientToDevice map already uses routingMu, so add bandwidth counters alongside |
| In-memory counter + periodic flush | Write to DB on every packet | On every packet is catastrophically slow; flush every 30s is the right tradeoff |
| Webhook from worker process | From API server | Worker already has the offline detection loop — natural home; avoids cross-process signaling |
| Per-operator webhook URL in `users` table | Separate `operator_settings` table | Users table is small and simple; a single nullable column is the least-friction path |

---

## Architecture Patterns

### Recommended Project Structure (changes only)

```
server/
├── cmd/tunnel/main.go           — add bandwidth counters + enforcement; load limit at connect time
├── cmd/worker/main.go           — add webhook dispatch goroutine to offline detection loop
├── internal/domain/models.go   — add WebhookURL to User; remove PasswordPlain write paths
├── internal/domain/config.go   — no changes needed
├── internal/repository/
│   ├── connection_repo.go       — add UpdateBandwidthUsed(), RegeneratePassword(), ResetBandwidthUsed()
│   └── user_repo.go             — add UpdateWebhookURL(), GetWebhookURL()
├── internal/api/handler/
│   ├── openvpn_handler.go       — Auth(): swap PasswordPlain compare → bcrypt.CompareHashAndPassword
│   │                              Connect(): pass bandwidth_limit in push payload
│   │                              add POST /connections/:id/regenerate-password
│   ├── connection_handler.go    — add ResetBandwidth handler; bandwidth_limit in create request
│   └── router.go                — wire new endpoints
├── migrations/
│   └── 010_security_monitoring.up.sql   — null password_plain, add webhook_url to users,
│                                          add last_offline_alert_at to devices
dashboard/
├── src/app/settings/page.tsx    — new page: webhook URL field + Send Test button
├── src/components/connections/
│   ├── AddConnectionModal.tsx   — add bandwidth_limit field
│   └── ConnectionTable.tsx      — add usage bar column
├── src/lib/api.ts               — add regeneratePassword(), resetBandwidth(), settings endpoints
```

### Pattern 1: bcrypt Auth Swap (SEC-01)

**What:** Replace plaintext string comparison with bcrypt hash comparison in OpenVPN auth handler.

**When to use:** Any path that currently reads `conn.PasswordPlain`.

**Example:**
```go
// Before (openvpn_handler.go line 56):
if conn.PasswordPlain != req.Password {

// After:
if err := bcrypt.CompareHashAndPassword([]byte(conn.PasswordHash), []byte(req.Password)); err != nil {
```

Cost factor for new passwords: `bcrypt.Cost(12)` — 4x slower than DefaultCost(10), still <100ms on modern hardware. Existing records already hashed at DefaultCost(10) via `connection_service.go`; they remain valid since CompareHashAndPassword reads cost from the hash prefix.

### Pattern 2: Per-Connection Bandwidth Counter in Tunnel (MON-02)

**What:** Add an `atomic.Int64` counter per active OpenVPN client VPN IP. Increment on every forwarded packet. Enforce limit before forwarding. Flush to DB periodically.

**Data structure** (add to `tunnelServer` struct):
```go
// Keyed by client VPN IP string (10.9.0.x)
clientBandwidthUsed  map[string]*atomic.Int64  // protected by routingMu
clientBandwidthLimit map[string]int64           // bytes; 0 = unlimited; protected by routingMu
```

**Hot path enforcement in `tunToUdp` (after the `if mapped` block):**
```go
if mapped {
    s.routingMu.Lock()
    limit := s.clientBandwidthLimit[srcIP]
    var used int64
    if ctr, ok := s.clientBandwidthUsed[srcIP]; ok {
        used = ctr.Add(int64(n))
    }
    s.routingMu.Unlock()

    if limit > 0 && used > limit {
        // Hard cutoff: drop the packet silently
        continue
    }
    // Forward the packet
    s.mu.RLock()
    c, ok = s.clients[deviceIP]
    s.mu.RUnlock()
    if ok {
        buf[0] = TypeData
        s.udpConn.WriteToUDP(buf[:1+n], c.udpAddr)
    }
}
```

**NOTE:** `routingMu` is a `sync.Mutex`, not `sync.RWMutex`. The Add on the atomic counter can be done after releasing routingMu — only the limit read needs the lock. Restructure to minimize lock hold time in the hot path:
```go
// Read limit (needs lock) — fast
s.routingMu.Lock()
ctr := s.clientBandwidthUsed[srcIP]
limit := s.clientBandwidthLimit[srcIP]
s.routingMu.Unlock()

// Increment counter (lock-free)
var used int64
if ctr != nil {
    used = ctr.Add(int64(n))
}

if limit > 0 && used > limit {
    continue // drop
}
// forward...
```

**Periodic flush goroutine (add to `startPushAPI` or as a new goroutine in `main`):**
```go
go s.bandwidthFlushLoop()

func (s *tunnelServer) bandwidthFlushLoop() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    for range ticker.C {
        s.flushBandwidthToAPI()
    }
}

func (s *tunnelServer) flushBandwidthToAPI() {
    s.routingMu.Lock()
    snapshot := make(map[string]int64, len(s.clientBandwidthUsed))
    for ip, ctr := range s.clientBandwidthUsed {
        snapshot[ip] = ctr.Load()
    }
    s.routingMu.Unlock()

    // POST snapshot to API server for DB persistence
    // API maps 10.9.0.x -> connection ID via username lookup
    body, _ := json.Marshal(snapshot)
    http.Post(s.apiURL+"/api/internal/bandwidth-flush", "application/json", bytes.NewReader(body))
}
```

**Limit loaded at connect time** — modify `handleOpenVPNClientConnect` request body to include `bandwidth_limit int64` (API sends it when calling the tunnel's `/openvpn-client-connect`). This requires `openvpn_handler.go` Connect() to read the connection's `BandwidthLimit` and pass it forward.

**Reset usage** — operator clicks "Reset Usage" in dashboard → API sets `bandwidth_used = 0` in DB AND calls a new push endpoint `/openvpn-client-reset-bandwidth` on the tunnel (so the in-memory counter is reset immediately for active sessions).

### Pattern 3: Offline Webhook Dispatch (MON-01)

**What:** In the worker's `MarkStaleOfflineWithLogs` loop, after marking a device offline, look up the owning operator's webhook URL and POST a notification. Track `last_offline_alert_at` per device to enforce the 5-minute cooldown. Send a recovery webhook in the next run when a previously-offline device comes back online (status transitions offline→online).

**Storage:** Add `webhook_url TEXT` column to `users` table. Add `last_offline_alert_at TIMESTAMPTZ` column to `devices` table.

**Worker changes:**
```go
// After s.statusLogRepo.Insert(offline transition):
go s.sendOfflineWebhook(ctx, d)

func (s *DeviceService) sendOfflineWebhook(ctx context.Context, d domain.Device) {
    // 1. Check cooldown: d.LastOfflineAlertAt must be nil or > 5 minutes ago
    // 2. Look up owning user's webhook URL
    // 3. POST JSON payload
    // 4. Update d.LastOfflineAlertAt = now in DB
}
```

**Webhook payload structure (Claude's discretion):**
```json
{
  "event": "device.offline",
  "device_id": "uuid",
  "device_name": "Pixel 8 Pro",
  "last_seen": "2026-02-26T10:30:00Z",
  "timestamp": "2026-02-26T10:32:00Z"
}
```
Recovery event uses `"event": "device.online"` with `"reconnected_at"` instead of `"last_seen"`.

**Delivery:** Fire-and-forget goroutine with `http.Client{Timeout: 10 * time.Second}`. No retry — the next 30-second worker tick will retry if the device is still offline and cooldown has expired (because `last_offline_alert_at` is only updated on successful delivery).

### Anti-Patterns to Avoid

- **Writing PasswordPlain on any new code path:** After migration, the column stays in the schema (nullable) to avoid a multi-step migration, but application code must never populate it. The `connection_service.go` Create() currently sets `PasswordPlain: req.Password` — this must be removed.
- **Checking bandwidth in the `tunToUdp` TUN read hot path while holding routingMu for the full packet duration:** The lock must be released before the UDP write. Read limit + get counter pointer under lock, then do atomic Add + UDP write outside the lock.
- **Sending webhook from the API request handler on device disconnect:** Disconnect is a polling concern (stale heartbeat), not a synchronous API event. Only the worker has reliable visibility into the 2-minute timeout.
- **Storing bandwidth counters only in the tunnel process:** The tunnel can restart; counters must be flushed to `bandwidth_used` in the DB regularly and loaded back (as a starting offset) when a client reconnects.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Password hashing | Custom hash | `bcrypt.GenerateFromPassword` / `bcrypt.CompareHashAndPassword` | Handles salt, timing-safe comparison, cost embedded in hash string |
| HTTP webhook delivery | Custom HTTP sender | `net/http` stdlib with timeout | No external dependency; fire-and-forget is sufficient per decision |
| Atomic byte counter | Mutex-protected int64 | `sync/atomic.Int64` (Go 1.19+) | Zero-overhead CAS; benchmark shows 30% faster than mutex for single-value counters |
| DB migration runner | Custom SQL exec | Plain SQL migration file (project already uses manual migrations) | Project doesn't use golang-migrate; matches existing pattern of numbered .up.sql files |

---

## Common Pitfalls

### Pitfall 1: OpenVPN Auth Timing — Existing Sessions Break

**What goes wrong:** After the migration nulls `password_plain`, any currently-connected OpenVPN session that reconnects (OpenVPN re-auth) will fail because the new auth code calls `bcrypt.CompareHashAndPassword` but the old .ovpn file has the plaintext credential that maps to the hash — this actually works correctly since the plaintext credential is what gets hashed. The only risk is if PasswordHash was never set for some records.

**Why it happens:** Migration 006 added `password_plain` with a default of `''`. If any record has a blank `password_hash`, bcrypt compare will fail against anything.

**How to avoid:** Migration script: `UPDATE proxy_connections SET password_hash = crypt(password_plain, gen_salt('bf', 12)) WHERE password_plain != '' AND (password_hash IS NULL OR password_hash = '')` — or simpler: verify in Go that every row with a non-null plain has a valid hash before nulling the plain column.

**Warning signs:** After migration, check `SELECT COUNT(*) FROM proxy_connections WHERE password_hash = '' OR password_hash IS NULL` — must be 0.

### Pitfall 2: Bandwidth Counter Lost on Tunnel Restart

**What goes wrong:** Tunnel process restarts (container recreate, deploy). All `clientBandwidthUsed` counters are zero. A customer who has already consumed 900 MB of their 1 GB limit effectively resets.

**Why it happens:** In-memory state is ephemeral.

**How to avoid:** At `handleOpenVPNClientConnect`, also receive the current `bandwidth_used` value from the API (API reads it from DB) and initialize the counter to that value, not zero.

```go
// In handleOpenVPNClientConnect request body:
BandwidthUsed  int64  `json:"bandwidth_used"`   // existing DB value — initial offset

// Initialize counter:
ctr := &atomic.Int64{}
ctr.Store(req.BandwidthUsed)
s.clientBandwidthUsed[req.ClientVPNIP] = ctr
```

### Pitfall 3: routingMu Held During UDP Write (Performance Regression)

**What goes wrong:** Holding `routingMu` (a `sync.Mutex`) while calling `s.udpConn.WriteToUDP` blocks all other goroutines from updating routing state. Under high throughput this causes latency spikes.

**Why it happens:** Natural tendency to bracket the entire if-block with a single lock.

**How to avoid:** Lock only to read the counter pointer and limit value, then release. The `atomic.Int64.Add()` and UDP write happen outside the lock.

### Pitfall 4: Webhook Flood on Flapping Device

**What goes wrong:** Device goes offline, gets alert. Comes back online briefly (recovery webhook sent). Goes offline again 10 seconds later — alert sent again. Repeat rapidly.

**Why it happens:** Per-decision: 5-minute cooldown applies to offline alerts. Recovery webhook is always sent immediately on reconnect. If the device reconnects and immediately goes offline again within 5 minutes, only one offline alert is sent (cooldown in effect). Recovery alert has no cooldown — this is correct behavior; operators want to know when it's back.

**How to avoid:** Implement cooldown correctly. `last_offline_alert_at` is set on offline alert, not on recovery. Recovery webhook is always sent. The alert storm prevention is on the offline direction only — per the user decision.

### Pitfall 5: DownloadOVPN and Connect Still Use PasswordPlain

**What goes wrong:** `openvpn_handler.go` DownloadOVPN() writes `conn.PasswordPlain` inline in the `<auth-user-pass>` block (line 224). `Connect()` sends `conn.PasswordPlain` to the tunnel as `socks_pass` (line 113). After migration these will be empty strings.

**Why it happens:** PasswordPlain was the only source of plaintext after the initial create response.

**How to avoid:** The CONTEXT.md decision says `.ovpn file includes auth-user-pass inline at creation/regen time`. This means DownloadOVPN must be changed: either (a) remove the inline `<auth-user-pass>` block and require the user to enter credentials manually when importing, or (b) keep inline credentials but only populate them right after create/regen when the plaintext is briefly known. Option (a) is simpler and safer. For `Connect()`, the `socks_pass` is used by the SOCKS5 forwarder in the tunnel to authenticate to the device's proxy — the tunnel's `clientSocksAuth` map stores these credentials. After removing plaintext, `Connect()` cannot get the plaintext from DB. **Resolution:** At create/regen time, the API must cache the plaintext in a short-lived way, OR the tunnel's SOCKS5 forwarder must use the bcrypt hash for comparison (but SOCKS5 protocol sends plaintext, so the device proxy validates it — the device gets credentials via heartbeat response `HeartbeatResponse.Credentials`). See Open Questions #1.

---

## Code Examples

### SEC-01: Migration SQL

```sql
-- 010_security_monitoring.up.sql

-- Backfill: hash any rows where password_hash is empty but password_plain is not
-- (In practice connection_service.go already hashes at cost 10; this is a safety net)
-- Skip rows already having a valid hash (starts with $2a$ or $2b$)
-- NOTE: pgcrypto is NOT available; use Go migration script instead for actual hashing.
-- The SQL migration only nulls the column after the Go script has run.

-- Add webhook_url to users (per-operator)
ALTER TABLE users ADD COLUMN IF NOT EXISTS webhook_url TEXT;

-- Add last_offline_alert_at to devices (cooldown tracking)
ALTER TABLE devices ADD COLUMN IF NOT EXISTS last_offline_alert_at TIMESTAMPTZ;

-- Null out password_plain (Go migration script must run FIRST to verify all hashes are valid)
UPDATE proxy_connections SET password_plain = NULL WHERE password_plain IS NOT NULL;

-- Make password_plain nullable (it was NOT NULL DEFAULT '')
ALTER TABLE proxy_connections ALTER COLUMN password_plain DROP NOT NULL;
ALTER TABLE proxy_connections ALTER COLUMN password_plain DROP DEFAULT;
```

**Important:** The `password_plain` column cannot be dropped in this migration because it's referenced in `scanConnection` scan and `Create` insert. Remove the column from all Go code first, then a later migration can `ALTER TABLE proxy_connections DROP COLUMN password_plain`.

### SEC-01: Regenerate Password API Handler

```go
// POST /api/connections/:id/regenerate-password
func (h *ConnectionHandler) RegeneratePassword(c *gin.Context) {
    id, _ := uuid.Parse(c.Param("id"))

    // Generate new plaintext
    newPass := generateRandomPassword(16)
    hash, _ := bcrypt.GenerateFromPassword([]byte(newPass), 12)

    // Store hash (NOT plaintext), null the plain column
    if err := h.connService.SetPassword(c.Request.Context(), id, string(hash)); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Return plaintext ONCE — client must copy it now
    c.JSON(http.StatusOK, gin.H{"password": newPass})
}
```

### MON-01: Webhook Dispatch in Worker

```go
func (s *DeviceService) sendOfflineWebhook(deviceID uuid.UUID, deviceName string, lastSeen time.Time) {
    webhookURL, err := s.userRepo.GetWebhookURLForDevice(context.Background(), deviceID)
    if err != nil || webhookURL == "" {
        return
    }

    payload, _ := json.Marshal(map[string]interface{}{
        "event":      "device.offline",
        "device_id":  deviceID.String(),
        "device_name": deviceName,
        "last_seen":  lastSeen.Format(time.RFC3339),
        "timestamp":  time.Now().UTC().Format(time.RFC3339),
    })

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(payload))
    if err != nil {
        log.Printf("[webhook] offline alert failed for device %s: %v", deviceID, err)
        return
    }
    resp.Body.Close()

    // Update last_offline_alert_at only on success
    s.deviceRepo.SetLastOfflineAlertAt(context.Background(), deviceID, time.Now().UTC())
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Compare password_plain directly | bcrypt.CompareHashAndPassword | This phase | Timing-safe; no plaintext at rest |
| bandwidth_used field exists but never updated | Atomic counter in tunnel + periodic DB flush | This phase | Actual enforcement |
| No offline alerting | Worker-dispatched webhook with 5-min cooldown | This phase | Operator awareness within 2 min detection + delivery time |

---

## Open Questions

1. **SOCKS5 credential path after PasswordPlain removal**
   - What we know: `openvpn_handler.go Connect()` currently reads `conn.PasswordPlain` and sends it to the tunnel as `socks_pass` for the SOCKS5 forwarder. The device proxy validates these credentials against the credentials it received via heartbeat response (`HeartbeatResponse.Credentials`). The device gets `{username, password}` pairs — the plaintext password.
   - What's unclear: After PasswordPlain is nulled, how does the tunnel get the plaintext credential to authenticate to the device's SOCKS5 proxy? The hash is useless for SOCKS5 auth since SOCKS5 requires plaintext.
   - **Recommendation:** The Regenerate endpoint generates a new random password and stores only the bcrypt hash. The plaintext is returned once in the API response AND used to regenerate the `.ovpn` file inline. The tunnel's SOCKS5 forwarder path requires the plaintext at Connect() time. **Solution:** At create/regen time, return the plaintext once; for the SOCKS5 path the credential is already on the device (it received it via heartbeat). The heartbeat response `Credentials` field in `HeartbeatResponse` currently populates from `conn.PasswordPlain` — this needs to change to use a separate field or regeneration clears the credential from the device too. **Simplest fix:** `HeartbeatResponse.Credentials` must continue to send plaintext to the device so it can authenticate incoming SOCKS5 connections. Store the plaintext only in the Android app (which owns the credential flow) OR use the bcrypt hash as the password for device SOCKS5 (hash the hash again). **Actual simplest fix:** The `socks_pass` sent to the tunnel is used by the SOCKS5 forwarder to connect to the device's proxy — if the device's proxy uses bcrypt comparison (configurable), it can verify. But Android SOCKS5 proxy implementation uses plaintext. **Real fix:** Store a second credential field, `proxy_credential_hash` for bcrypt auth and keep using `PasswordPlain` only for the tunnel→device SOCKS5 path — OR the simplest: keep `password_plain` populated but not exposed through the API (no `db:"-"` removal yet). This is a design tension the planner must resolve explicitly. Recommend the planner clarify this before writing tasks. **Most pragmatic:** Phase 03-01 covers bcrypt auth for OpenVPN auth only; the SOCKS5 tunnel path continues using PasswordPlain internally (not exposed in API responses), migration nulls the column only once the SOCKS5 path is also resolved. Alternatively, accept that PasswordPlain stays in DB (hidden from API) for now and SEC-01's success criterion (no plaintext in OpenVPN auth) is still met.

2. **Bandwidth counter for HTTP/SOCKS5 connections vs OpenVPN connections**
   - What we know: The decision says "track bytes per customer connection in the tunnel server." HTTP/SOCKS5 connections route via DNAT to the device's proxy, not through the tunnel's `tunToUdp` NAT path. The `clientToDevice` map and `tunToUdp` NAT path only handles OpenVPN clients (10.9.0.x).
   - What's unclear: Does bandwidth enforcement apply to HTTP/SOCKS5 connections too, or only OpenVPN?
   - **Recommendation:** `BandwidthLimit` field exists on `ProxyConnection` for all types. For phase scope, enforce only for OpenVPN connections (which transit the tunnel's `tunToUdp`). HTTP/SOCKS5 enforcement would require a different mechanism (iptables byte counting or proxy-level). Plan tasks accordingly.

---

## Sources

### Primary (HIGH confidence)
- Codebase direct inspection — `server/cmd/tunnel/main.go`, `server/internal/api/handler/openvpn_handler.go`, `server/internal/domain/models.go`, `server/internal/service/connection_service.go`, `server/internal/repository/connection_repo.go`, `server/cmd/worker/main.go`
- `golang.org/x/crypto/bcrypt` (pkg.go.dev) — `CompareHashAndPassword`, `GenerateFromPassword`, cost constants
- `sync/atomic` stdlib Go 1.23 — `atomic.Int64` type and methods

### Secondary (MEDIUM confidence)
- [golang/go Issue #54573](https://github.com/golang/go/issues/54573) — bcrypt DefaultCost should be increased to 12 (Go team proposal discussion)
- [Grab Engineering: Highly concurrent in-memory counter in GoLang](https://engineering.grab.com/highly-concurrent-in-memory-counter-in-go-lang) — atomic counter + periodic DB flush pattern achieving 68% query reduction
- [DeepSource GO-S1045](https://deepsource.com/directory/go/issues/GO-S1045) — bcrypt cost factor below 10 is flagged as insecure

### Tertiary (LOW confidence)
- Various WebSearch results on webhook design patterns and migration approach — consistent with primary sources but not independently verified against official docs.

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all libraries already in project; no new dependencies
- Architecture: HIGH — insertion points confirmed by reading actual code; known data structures
- Pitfalls: HIGH for Pitfall 5 (PasswordPlain/SOCKS5 tension confirmed by code reading); MEDIUM for others
- Open Questions: These are genuine design tensions found in the code, not speculation

**Research date:** 2026-02-26
**Valid until:** 2026-03-28 (stable ecosystem — 30 day window)
