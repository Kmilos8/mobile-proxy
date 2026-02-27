# Phase 4: Bug Fixes and Polish - Research

**Researched:** 2026-02-27
**Domain:** Go backend bug fixes + React/TypeScript dashboard UI polish
**Confidence:** HIGH

---

## Summary

Phase 4 addresses six concrete, bounded fixes identified through a v1.0 milestone audit and roadmap additions. Every change maps to code that already exists — no new architecture, no new libraries. The audit found two wiring gaps (MON-01 recovery webhook silently dropped; MON-02 reset bandwidth not propagated to tunnel) and four UI improvements (OpenVPN creation bug, device search bar, auto-rotation column, connection ID column).

All six items are small and independently deployable. The largest single change is the bandwidth reset fix, which requires coordinating: `connection_service.ResetBandwidth` → tunnel push API → `handleResetBandwidth`. The remainder are one-to-three line Go fixes or React component additions that follow patterns already established in Phases 2 and 3.

No new dependencies are required. The dashboard already uses React state with `useState`/`useEffect` patterns. The Go services follow setter-injection and inline HTTP-post-to-tunnel patterns established in Phase 3.

**Primary recommendation:** Group all six items into a single plan. Each item is small enough that splitting across two plans would add overhead without benefit. Execute in dependency order: backend fixes first (MON-01, MON-02), then OpenVPN bug fix (DASH-02 backend), then UI additions (DASH-02 frontend, search bar, auto-rotation column, connection ID).

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| MON-01 | Device offline notification via webhook — specifically the **recovery** webhook (offline → online) | Gap: `api/main.go` does not call `deviceService.SetUserRepo(userRepo)` — 1-line fix identified |
| MON-02 | Enforce bandwidth limits: **Reset Usage must also reset tunnel in-memory counter** | Gap: `connService.ResetBandwidth` calls only DB reset, not tunnel push API — 20-line fix identified |
| DASH-02 | Connection creation UI — specifically **OpenVPN type must be selectable** without "must be http or socks5" error | Bug: `AddConnectionModal` state type and Select options exclude 'openvpn' |
</phase_requirements>

---

## Standard Stack

No new libraries required. All tooling is already installed.

### Core (unchanged)

| Layer | Technology | Version | Notes |
|-------|-----------|---------|-------|
| Backend API | Go + Gin | existing | Standard handler/service/repo pattern |
| Tunnel binary | Go (standalone) | existing | Push API handles HTTP calls on port 8081 |
| Dashboard | Next.js + React + TypeScript | existing | App router, `useState`/`useEffect` pattern |
| UI components | shadcn/ui (zinc dark theme) | existing | `Table`, `Badge`, `Select`, `Input` all present |
| Icons | lucide-react | existing | `Search`, `RotateCw`, `Hash` available |

### No Installation Required

All packages already present in `package.json` and `go.mod`. No `npm install` or `go get` needed.

---

## Architecture Patterns

### Pattern 1: One-Line API Process Wiring (Go setter injection)

The project uses a setter-method pattern for optional dependencies. `deviceService` already has `SetUserRepo` and `SetStatusLogRepo`. The worker (`worker/main.go`) correctly calls `deviceService.SetUserRepo(userRepo)` at line 39. The API process (`api/main.go`) calls `deviceService.SetStatusLogRepo(statusLogRepo)` at line 43 but does **not** call `SetUserRepo`.

**Fix:**
```go
// server/cmd/api/main.go — after line 43 (SetStatusLogRepo)
deviceService.SetUserRepo(userRepo)
```

`userRepo` is already instantiated at line 26. This single line closes the MON-01 recovery webhook gap.

### Pattern 2: Service → Tunnel Push API (HTTP POST)

Established in Phase 3 via `refreshDNAT`, `teardownDNAT`, and `openvpn-client-connect`. The pattern:
1. Service method gets connection from DB to find `DeviceID`
2. Look up device to resolve `TunnelPushURL`
3. HTTP POST to `tunnelURL + "/openvpn-client-reset-bandwidth"` with JSON body
4. Ignore tunnel errors (fire-and-forget, tunnel may not have the client connected)

The tunnel's existing `handleResetBandwidth` accepts `{"client_vpn_ip": "10.9.0.x"}`. The problem: `connection_service.ResetBandwidth` only has a connection `id` (UUID) — it does not know the client VPN IP. The tunnel doesn't expose a lookup-by-username endpoint.

**Two implementation options for MON-02:**

**Option A (preferred — modify tunnel to accept username):**
- Extend `handleResetBandwidth` to accept `{"username": "..."}` in addition to `client_vpn_ip`
- Tunnel resolves username → client VPN IP via `clientSocksAuth` map (already keyed by VPN IP, values contain `.user` field)
- Add reverse lookup: iterate `clientSocksAuth` to find the VPN IP whose `.user` matches the username
- Service posts `{"username": conn.Username}` to tunnel
- **Advantage:** Service only needs `conn.Username` (already available after fetching by ID)

**Option B (store client VPN IP in DB):**
- Add a `client_vpn_ip` column to `proxy_connections`
- Populate at OpenVPN connect time
- Service looks it up and passes it through
- **Disadvantage:** Requires migration, VPN IP is ephemeral and changes on reconnect — adds complexity

**Use Option A.** No migration needed. The tunnel's `clientSocksAuth` already maps `vpnIP → {user, pass}`, so the reverse lookup is a simple linear scan (max ~10 entries in practice).

**Code change in tunnel/main.go:**
```go
// Modified handleResetBandwidth — accepts username OR client_vpn_ip
func (s *tunnelServer) handleResetBandwidth(w http.ResponseWriter, r *http.Request) {
    var req struct {
        ClientVPNIP string `json:"client_vpn_ip"`
        Username    string `json:"username"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    s.routingMu.Lock()
    targetIP := req.ClientVPNIP
    if targetIP == "" && req.Username != "" {
        // Resolve username -> VPN IP via reverse lookup
        for ip, auth := range s.clientSocksAuth {
            if auth.user == req.Username {
                targetIP = ip
                break
            }
        }
    }
    if targetIP != "" {
        if ctr, ok := s.clientBandwidthUsed[targetIP]; ok {
            ctr.Store(0)
        }
    }
    s.routingMu.Unlock()

    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"ok":true}`))
}
```

**Code change in connection_service.go:**
```go
func (s *ConnectionService) ResetBandwidth(ctx context.Context, id uuid.UUID) error {
    if err := s.connRepo.ResetBandwidthUsed(ctx, id); err != nil {
        return err
    }
    // Also reset tunnel in-memory counter so 30s flush doesn't overwrite the zero
    conn, err := s.connRepo.GetByID(ctx, id)
    if err != nil {
        return nil // DB reset succeeded; tunnel reset is best-effort
    }
    device, err := s.deviceRepo.GetByID(ctx, conn.DeviceID)
    if err != nil {
        return nil
    }
    tunnelURL := s.getTunnelPushURL(ctx, device)
    if tunnelURL != "" {
        go s.resetTunnelBandwidth(tunnelURL, conn.Username)
    }
    return nil
}

func (s *ConnectionService) resetTunnelBandwidth(tunnelURL, username string) {
    body, _ := json.Marshal(map[string]string{"username": username})
    client := &http.Client{Timeout: 3 * time.Second}
    resp, err := client.Post(tunnelURL+"/openvpn-client-reset-bandwidth", "application/json", strings.NewReader(string(body)))
    if err != nil {
        log.Printf("[reset-bandwidth] tunnel call failed for user %s: %v", username, err)
        return
    }
    resp.Body.Close()
}
```

### Pattern 3: OpenVPN Type in AddConnectionModal (DASH-02 bug fix)

**Root cause:** `AddConnectionModal.tsx` has:
```typescript
const [proxyType, setProxyType] = useState<'http' | 'socks5'>('http')
// ...
onValueChange={(val) => setProxyType(val as 'http' | 'socks5')}
```
The Select component only renders HTTP and SOCKS5 options. OpenVPN is absent.

**Fix:** Widen the type union, add the SelectItem, and handle form reset:
```typescript
const [proxyType, setProxyType] = useState<'http' | 'socks5' | 'openvpn'>('http')
// onValueChange cast: val as 'http' | 'socks5' | 'openvpn'
// Add SelectItem: <SelectItem value="openvpn">OpenVPN</SelectItem>
// Reset: setProxyType('http') unchanged (keeps default to http on close)
```

No backend change needed — `connection_service.Create` already handles `"openvpn"` type (line 75: `proxyType != "http" && proxyType != "socks5" && proxyType != "openvpn"`).

### Pattern 4: DeviceTable Search Bar

DeviceTable already has a `filtered` array derived from `devices` prop using status filter. Adding name search follows the same pattern: introduce a `searchQuery` state, filter on `d.name.toLowerCase().includes(searchQuery.toLowerCase())` before the status filter.

The search input sits in the existing controls row alongside the status filter buttons. Use a plain `<input>` styled with the project's zinc-800 input classes (matches existing inputs in `AddConnectionModal`). Consider adding a `Search` icon from lucide-react (already installed).

**No new props needed** — search state lives inside `DeviceTable` component.

```typescript
const [searchQuery, setSearchQuery] = useState('')

const filtered = devices.filter(d => {
  const matchesSearch = searchQuery === '' ||
    (d.name || '').toLowerCase().includes(searchQuery.toLowerCase())
  if (!matchesSearch) return false
  if (statusFilter === 'online') return d.status === 'online'
  if (statusFilter === 'offline') return d.status !== 'online'
  return true
})
```

### Pattern 5: Auto-Rotation Column in DeviceTable

The `Device` interface already has `auto_rotate_minutes: number` (confirmed in `api.ts:66`). The backend already returns this field in `GET /api/devices`.

Add a new table column between "IP" and "Connections". Display logic:
- `auto_rotate_minutes > 0` → show "every Xm" (e.g., "every 5m") with a `RotateCw` or `Timer` lucide icon in emerald
- `auto_rotate_minutes <= 0` → show "—" in zinc-500

The `SortKey` type can be extended: `type SortKey = 'name' | 'status' | 'cellular_ip' | 'auto_rotate' | 'connections'`. Sort numerically by `d.auto_rotate_minutes`.

Header and cell must include `hidden md:table-cell` (responsive, same as IP column). The `colSpan` in the empty state row increments from 4 to 5.

### Pattern 6: Connection ID Column in ConnectionTable

The `ProxyConnection` interface already has `id: string` (UUID). The success criterion requires "visible connection ID assigned at creation, shown in the connection table."

Display as a truncated monospace UUID (first 8 chars, e.g., `a1b2c3d4`) with a copy button using the existing `CopyButton` sub-component. Add as the first column (before "Type") or after "Status" — placing after "Type" is cleaner so Type stays leftmost.

Column header: "ID" or "Conn ID". On mobile cards, add an ID row in the same `space-y-2 text-sm` block.

No API changes needed — `id` is already returned in all connection responses.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Reverse username→IP lookup in tunnel | A separate registry/index | Linear scan of `clientSocksAuth` map | Max ~10 entries, called rarely (only on explicit reset), O(n) is fine |
| Search debouncing | `setTimeout` / custom hook | None needed | Devices list is small (<100), instant filter is appropriate |
| UUID display formatting | Custom formatter | `conn.id.slice(0, 8)` inline | Short UUID prefix is sufficient; no library needed |

---

## Common Pitfalls

### Pitfall 1: Tunnel Reset Is Best-Effort

**What goes wrong:** Developer adds error return from tunnel HTTP call, causing `ResetBandwidth` to fail if the connection is not currently active in the tunnel (no active OpenVPN session).

**Why it happens:** The tunnel only has entries in `clientBandwidthUsed` for currently-connected OpenVPN clients. If no client is connected, the reset endpoint returns 200 but does nothing — which is correct behavior.

**How to avoid:** Fire the tunnel call as a goroutine (`go s.resetTunnelBandwidth(...)`). Return `nil` from `ResetBandwidth` even if tunnel call fails — the DB reset is the source of truth; the tunnel counter becomes consistent at the next connect.

### Pitfall 2: TypeScript State Type Narrowing

**What goes wrong:** The `proxyType` state typed as `'http' | 'socks5'` is passed to `api.connections.create` which accepts `proxy_type?: string`. The type won't error at compile time, but the Select `onValueChange` cast must be updated or TypeScript will infer the wrong union.

**How to avoid:** Update the `useState` type annotation, the `onValueChange` cast, and the form reset line all in one change.

### Pitfall 3: DeviceTable colSpan Off-By-One

**What goes wrong:** Adding the auto-rotation column increases the column count from 4 to 5. The empty state `<TableCell colSpan={4}` must be updated to `colSpan={5}` or the empty-state row renders incorrectly.

**How to avoid:** Search for `colSpan` in DeviceTable.tsx before committing.

### Pitfall 4: Recovery Webhook Fires on Every Heartbeat if Device Stays Online

**What goes wrong:** Misreading `wasOffline := device.Status == domain.DeviceStatusOffline`. This correctly checks the status **before** the heartbeat update, so it only fires on the `offline → online` transition. If device is already online, `wasOffline` is `false`. No change needed here — the logic is correct. The only bug is `userRepo` being nil.

**How to avoid:** Confirm the fix is only `deviceService.SetUserRepo(userRepo)` in `api/main.go` — no logic change in `device_service.go`.

### Pitfall 5: Adding routingMu.Lock Inside Already-Locked Scope

**What goes wrong:** If `handleResetBandwidth` modifications accidentally nest a `routingMu.Lock()` call inside an existing lock acquisition on the same goroutine, deadlock occurs.

**How to avoid:** The existing implementation correctly acquires `routingMu.Lock()` once at the top of the handler. The modification only extends the lock body (same single lock/unlock) — do not add a second lock call.

---

## Code Examples

### MON-01: Fix — `api/main.go`

```go
// server/cmd/api/main.go — add after line 43
deviceService.SetStatusLogRepo(statusLogRepo)
deviceService.SetUserRepo(userRepo)  // ← ADD THIS (closes recovery webhook gap)
deviceService.SetRelayServerRepo(relayServerRepo)
```

### MON-02: Fix — Tunnel `handleResetBandwidth` with username lookup

```go
// server/cmd/tunnel/main.go — replace handleResetBandwidth
func (s *tunnelServer) handleResetBandwidth(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var req struct {
        ClientVPNIP string `json:"client_vpn_ip"`
        Username    string `json:"username"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    s.routingMu.Lock()
    targetIP := req.ClientVPNIP
    if targetIP == "" && req.Username != "" {
        for ip, auth := range s.clientSocksAuth {
            if auth.user == req.Username {
                targetIP = ip
                break
            }
        }
    }
    if targetIP != "" {
        if ctr, ok := s.clientBandwidthUsed[targetIP]; ok {
            ctr.Store(0)
        }
    }
    s.routingMu.Unlock()

    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"ok":true}`))
}
```

### MON-02: Fix — `connection_service.ResetBandwidth`

```go
// server/internal/service/connection_service.go — replace ResetBandwidth
func (s *ConnectionService) ResetBandwidth(ctx context.Context, id uuid.UUID) error {
    if err := s.connRepo.ResetBandwidthUsed(ctx, id); err != nil {
        return err
    }
    // Best-effort: reset tunnel in-memory counter to prevent 30s flush overwrite
    conn, err := s.connRepo.GetByID(ctx, id)
    if err != nil {
        return nil
    }
    device, err := s.deviceRepo.GetByID(ctx, conn.DeviceID)
    if err != nil {
        return nil
    }
    tunnelURL := s.getTunnelPushURL(ctx, device)
    if tunnelURL != "" {
        go s.resetTunnelBandwidth(tunnelURL, conn.Username)
    }
    return nil
}

func (s *ConnectionService) resetTunnelBandwidth(tunnelURL, username string) {
    body, _ := json.Marshal(map[string]string{"username": username})
    client := &http.Client{Timeout: 3 * time.Second}
    resp, err := client.Post(tunnelURL+"/openvpn-client-reset-bandwidth", "application/json", strings.NewReader(string(body)))
    if err != nil {
        log.Printf("[reset-bandwidth] tunnel call failed for %s: %v", username, err)
        return
    }
    resp.Body.Close()
}
```

### DASH-02: Fix — `AddConnectionModal.tsx` OpenVPN option

```typescript
// dashboard/src/components/connections/AddConnectionModal.tsx
// Line 37: widen type
const [proxyType, setProxyType] = useState<'http' | 'socks5' | 'openvpn'>('http')

// Line 105: widen cast
onValueChange={(val) => setProxyType(val as 'http' | 'socks5' | 'openvpn')}

// After SOCKS5 SelectItem (line 112):
<SelectItem value="openvpn" className="text-white focus:bg-zinc-700">OpenVPN</SelectItem>
```

### Search Bar in `DeviceTable.tsx`

```typescript
const [searchQuery, setSearchQuery] = useState('')

// In filter controls row (after status buttons):
<div className="relative ml-auto">
  <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-zinc-500 pointer-events-none" />
  <input
    type="text"
    placeholder="Search devices..."
    value={searchQuery}
    onChange={(e) => setSearchQuery(e.target.value)}
    className="pl-8 pr-3 py-1 bg-zinc-800 border border-zinc-700 text-white text-xs rounded focus:outline-none focus:border-brand-500"
  />
</div>

// filtered array:
const filtered = devices.filter(d => {
  if (searchQuery && !(d.name || '').toLowerCase().includes(searchQuery.toLowerCase())) return false
  if (statusFilter === 'online') return d.status === 'online'
  if (statusFilter === 'offline') return d.status !== 'online'
  return true
})
```

### Auto-Rotation Column in `DeviceTable.tsx`

```typescript
// In TableHeader — after IP column (hidden md:table-cell):
<TableHead className="text-zinc-400 hover:text-white h-10 px-4 hidden md:table-cell">
  Auto-Rotate
</TableHead>

// In TableRow — after IP cell:
<TableCell className="px-4 py-3 text-xs text-zinc-400 hidden md:table-cell">
  {device.auto_rotate_minutes > 0
    ? <span className="text-emerald-400">every {device.auto_rotate_minutes}m</span>
    : <span className="text-zinc-600">—</span>}
</TableCell>

// Empty state: colSpan={5}  (was 4)
```

### Connection ID Column in `ConnectionTable.tsx`

```typescript
// Desktop table header — add "ID" before "Type":
<TableHead className="text-xs font-medium text-zinc-500">ID</TableHead>

// Desktop table cell — in each row:
<TableCell className="font-mono text-xs text-zinc-500">
  {conn.id.slice(0, 8)}
  <CopyButton text={conn.id} copyKey={`${conn.id}-id`} />
</TableCell>

// Mobile card — add ID row in space-y-2 block:
<div className="flex items-center justify-between">
  <span className="text-zinc-500">ID</span>
  <span className="font-mono text-xs text-zinc-400 flex items-center">
    {conn.id.slice(0, 8)}
    <CopyButton text={conn.id} copyKey={`m-${conn.id}-id`} />
  </span>
</div>
```

---

## File Change Map

| File | Change | Gap/Feature |
|------|--------|-------------|
| `server/cmd/api/main.go` | Add `deviceService.SetUserRepo(userRepo)` (1 line) | MON-01 |
| `server/cmd/tunnel/main.go` | Extend `handleResetBandwidth` to accept `username` field | MON-02 |
| `server/internal/service/connection_service.go` | Extend `ResetBandwidth` to call tunnel reset API | MON-02 |
| `dashboard/src/components/connections/AddConnectionModal.tsx` | Add `openvpn` to type union + SelectItem | DASH-02 |
| `dashboard/src/components/devices/DeviceTable.tsx` | Add search input + auto-rotate column | Search + Auto-Rotate |
| `dashboard/src/components/connections/ConnectionTable.tsx` | Add connection ID column (desktop + mobile) | Connection ID |

**No migrations required.** All DB columns exist. No new API endpoints required. No new dependencies.

---

## Open Questions

1. **Connection ID display format**
   - What we know: The requirement says "visible connection ID assigned at creation, shown in the connection table." The `id` field is a UUID.
   - What's unclear: Full UUID (36 chars) vs. truncated prefix (8 chars). Full UUID is unwieldy in a table column.
   - Recommendation: Show `conn.id.slice(0, 8)` with a copy-full button. This matches common patterns for identifiers in dashboards (GitHub commit hashes, Stripe IDs).

2. **Auto-rotate column sortability**
   - What we know: The DeviceTable has sortable columns. Auto-rotate values are numeric.
   - What's unclear: Whether the user wants this column to be sortable.
   - Recommendation: Make it sortable (extend `SortKey` union) — costs 5 extra lines and follows the existing column pattern exactly. No reason to leave it non-sortable.

3. **Recovery webhook cooldown**
   - What we know: `sendOfflineWebhook` has a 5-minute cooldown per device. `sendRecoveryWebhook` has no cooldown.
   - What's unclear: Should recovery webhooks also have a cooldown? The Phase 3 decision was "no cooldown — reconnections are always notable events."
   - Recommendation: Keep no cooldown on recovery webhooks. The decision is locked in `03-02-SUMMARY.md`.

---

## Sources

### Primary (HIGH confidence)

- Codebase direct inspection:
  - `server/cmd/api/main.go` — confirmed `SetUserRepo` absent, `userRepo` available at line 26
  - `server/cmd/tunnel/main.go` lines 1155–1174 — `handleResetBandwidth` current implementation with `client_vpn_ip` only
  - `server/internal/service/connection_service.go` lines 249–251 — `ResetBandwidth` DB-only, no tunnel call
  - `dashboard/src/components/connections/AddConnectionModal.tsx` lines 37, 105 — type `'http' | 'socks5'` confirmed
  - `dashboard/src/components/devices/DeviceTable.tsx` — no search input, no auto-rotate column
  - `dashboard/src/components/connections/ConnectionTable.tsx` — no ID column
  - `.planning/v1.0-MILESTONE-AUDIT.md` — gap analysis with root cause and fix specifications
  - `server/cmd/tunnel/main.go` lines 88–96 — `clientSocksAuth` map structure confirmed (`socksAuth{user, pass}`)

### Secondary (MEDIUM confidence)

- Phase 3 summary decisions:
  - `03-02-SUMMARY.md` — bandwidth flush pattern, webhook cooldown decisions
  - `03-VERIFICATION.md` — confirmed MON-01/MON-02 code-level pass but live behavior gaps

---

## Metadata

**Confidence breakdown:**
- Bug root causes: HIGH — direct code inspection of all affected files
- Fix approaches: HIGH — follows established patterns in existing codebase (setter injection, tunnel HTTP post)
- UI additions: HIGH — DeviceTable and ConnectionTable patterns are fully understood
- No new libraries needed: HIGH — all required React/TS primitives already installed

**Research date:** 2026-02-27
**Valid until:** Stable — this is purely internal code, no third-party API changes relevant
