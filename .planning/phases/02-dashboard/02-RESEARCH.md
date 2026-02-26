# Phase 2: Dashboard - Research

**Researched:** 2026-02-26
**Domain:** Next.js 14 frontend redesign — no new backend; wire existing API to a new UI
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Device Overview (Home Page)
- Table/list layout — dense rows, not cards
- Minimal default columns: device name, status, IP, connection count
- Full device details (carrier, battery, signal) live on the device detail page, not the table
- Offline devices are visually dimmed/grayed out
- Sortable columns (click header) + filter controls (e.g., status filter for online/offline)
- Auto-refresh via polling (10-30 second interval)
- Summary stat bar above table: total devices, online count, offline count — devices only, no connection counts
- Clicking a device row navigates to a dedicated device detail page

#### Visual Style & Navigation
- Dark theme only (no light/toggle)
- Vercel/Linear aesthetic: minimal, monochrome, sharp typography, generous whitespace
- Left sidebar navigation, collapsible
- Sidebar sections: Devices only — connections are not a top-level nav item
- Connections are scoped inside each device's detail page

#### Credential Presentation
- All credentials always visible in plain text — no masking, no reveal toggle
- Per-field copy buttons on each credential (host, port, username, password)
- "Copy All" button that produces URL format: `protocol://username:password@host:port`
- OpenVPN connections show a download button for the .ovpn file (no config preview/expansion)

#### Connection Management Flow
- "Add Connection" button on device detail page opens a modal dialog
- Modal has protocol selector (HTTP, SOCKS5, OpenVPN) — submit creates the connection
- After creation: modal closes, new connection appears in the device's connection list
- Connection list inside device detail: simple table (type, port, username, status)
- Delete connection: confirmation dialog before removal
- Connection list only visible on device detail page, not on home/overview

### Claude's Discretion
- Click behavior on connection row (expand inline vs navigate to detail page)
- Exact polling interval for auto-refresh
- Sidebar collapsed/expanded default state
- Empty states (no devices, device with no connections)
- Loading states and error handling
- Exact column widths and responsive breakpoints
- Typography scale and spacing values

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| DASH-01 | Full UI redesign with modern component library (shadcn/ui) across all existing pages | shadcn/ui installation pattern for existing Next.js 14 app; migration from raw Tailwind to shadcn primitives |
| DASH-02 | Connection creation UI on dashboard (currently API-only) | `POST /api/connections` endpoint exists and accepts `device_id`, `username`, `password`, `proxy_type`; OpenVPN is a special case — same endpoint but `proxy_type` handling needs backend change (see Open Questions) |
| DASH-03 | Connection detail page for viewing/managing individual proxy ports | Per-connection credential display with copy; `.ovpn` download via `GET /api/connections/:id/ovpn`; delete via `DELETE /api/connections/:id` |
| DASH-04 | Responsive layout for desktop and tablet | Tailwind breakpoints `md:` (768px) and `lg:` (1024px); sidebar collapse behavior on smaller screens |
</phase_requirements>

---

## Summary

The dashboard is a Next.js 14 app using the App Router, Tailwind CSS, and a hand-rolled component library (no shadcn/ui today). The stack is already set up with a dark zinc-based color scheme, Inter font, and a brand color scale (emerald-green). All required API endpoints already exist in the Go backend — this phase is a pure frontend rebuild.

The main work is three-fold: (1) install shadcn/ui and migrate the existing component surface to its primitives, (2) redesign the home page (`/devices`) into the new dense-table + stat-bar layout, (3) rebuild the device detail page to include a proper connection management section (Add Connection modal with protocol selector, connection table with per-field copy, .ovpn download, delete with confirm dialog).

One backend gap exists: `proxy_type` on `POST /api/connections` is validated to only accept `"http"` or `"socks5"`. OpenVPN connections are conceptually the same credential set (username + password) used against the OpenVPN server, not through the DNAT proxy — they share the same `ProxyConnection` row. To show OpenVPN as a distinct type in the UI, the backend must either accept `"openvpn"` as a valid `proxy_type` or the UI must treat any connection as potentially usable for OpenVPN (every connection already has `.ovpn` download available via the API). The simplest path: add `"openvpn"` as a valid `proxy_type` in the connection service.

**Primary recommendation:** Install shadcn/ui into the existing app, then rebuild pages bottom-up (components first, then pages). Do not rewrite the entire app at once — migrate page by page to avoid breaking the working dashboard.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Next.js | 14.1.0 (already installed) | App Router, SSR, routing | Already in use; App Router with `'use client'` pattern established |
| React | 18.2.0 (already installed) | UI rendering | Already in use |
| Tailwind CSS | 3.4.x (already installed) | Utility-first styling | Already in use; zinc color scale matches desired aesthetic |
| shadcn/ui | latest (not yet installed) | Accessible component primitives | DASH-01 requirement; radix-ui based, dark-mode first, works with existing Tailwind setup |
| lucide-react | 0.312.0 (already installed) | Icons | Already in use across all pages |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| clsx + tailwind-merge | already installed | Conditional class merging | Already wired via `cn()` util in `lib/utils.ts` |
| @radix-ui/react-dialog | pulled in by shadcn | Modal/dialog primitive | Use for Add Connection modal and Delete confirm dialog |
| @radix-ui/react-select | pulled in by shadcn | Protocol selector | Use for HTTP/SOCKS5/OpenVPN dropdown in Add Connection modal |
| recharts | 3.7.0 (already installed) | Bandwidth charts | Already in use in UsageTab; keep as-is |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| shadcn/ui (required by DASH-01) | HeadlessUI + raw Tailwind | shadcn/ui is locked decision; do not use alternative |
| Polling (locked decision) | WebSocket only | WebSocket already connected but device_update events only; polling adds reliability for connection list updates |

**Installation:**
```bash
# From dashboard/ directory
npx shadcn@latest init
# When prompted: use existing tailwind config, dark theme, zinc base color, src/components/ui path
# Then add individual components as needed:
npx shadcn@latest add button dialog select table badge separator
```

---

## Architecture Patterns

### Recommended Project Structure

The existing structure should be preserved and extended:

```
dashboard/src/
├── app/
│   ├── devices/
│   │   ├── page.tsx              # REWRITE: home page with stat bar + table
│   │   ├── layout.tsx            # keep
│   │   └── [id]/
│   │       ├── page.tsx          # REWRITE: device detail with connection management
│   │       └── layout.tsx        # keep
│   ├── login/page.tsx            # keep (minor style update)
│   ├── overview/                 # REMOVE or REDIRECT to /devices (per nav decision)
│   ├── connections/              # REMOVE: connections are now inside device detail
│   ├── customers/                # keep or hide per sidebar decision
│   └── layout.tsx                # keep
├── components/
│   ├── ui/                       # shadcn/ui components land here automatically
│   │   ├── button.tsx            # shadcn generated
│   │   ├── dialog.tsx            # shadcn generated
│   │   ├── select.tsx            # shadcn generated
│   │   ├── table.tsx             # shadcn generated
│   │   ├── badge.tsx             # shadcn generated
│   │   ├── StatusBadge.tsx       # keep (existing)
│   │   ├── BatteryIndicator.tsx  # keep (existing)
│   │   └── StatCard.tsx          # keep or rebuild with shadcn Card
│   ├── dashboard/
│   │   ├── Sidebar.tsx           # REWRITE: collapsible, Devices-only nav
│   │   └── DashboardLayout.tsx   # minor update for sidebar collapse state
│   ├── devices/
│   │   ├── DeviceTable.tsx       # NEW: sortable/filterable device table
│   │   └── StatBar.tsx           # NEW: total/online/offline summary
│   └── connections/
│       ├── AddConnectionModal.tsx  # NEW: protocol selector + create form
│       ├── ConnectionTable.tsx     # NEW: per-device connection list
│       └── DeleteConnectionDialog.tsx  # NEW: confirm dialog
└── lib/
    ├── api.ts                    # UPDATE: add proxy_type support; ensure password returned
    ├── auth.ts                   # keep
    ├── utils.ts                  # keep
    └── websocket.ts              # keep
```

### Pattern 1: shadcn/ui Component Composition
**What:** shadcn/ui generates component source files into `src/components/ui/`. You own the code.
**When to use:** For all interactive primitives — buttons, modals, selects, tables.
**Example:**
```typescript
// shadcn Dialog pattern (from shadcn docs)
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger,
} from "@/components/ui/dialog"

export function AddConnectionModal({ deviceId, onCreated }: Props) {
  const [open, setOpen] = useState(false)
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button size="sm"><Plus className="w-4 h-4 mr-2" />Add Connection</Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>New Connection</DialogTitle>
        </DialogHeader>
        {/* form here */}
      </DialogContent>
    </Dialog>
  )
}
```

### Pattern 2: Client-Side Sort/Filter (no server pagination needed)
**What:** Device list fits in memory (tens of devices, not thousands). Sort and filter in the component.
**When to use:** Home page device table — click column header to sort, filter dropdown for status.
**Example:**
```typescript
// Keep in component state — no external library needed
const [sortCol, setSortCol] = useState<'name' | 'status' | 'ip'>('name')
const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc')
const [statusFilter, setStatusFilter] = useState<'all' | 'online' | 'offline'>('all')

const displayed = useMemo(() => {
  let list = devices
  if (statusFilter !== 'all') list = list.filter(d => d.status === statusFilter)
  return [...list].sort((a, b) => {
    const cmp = a[sortCol] < b[sortCol] ? -1 : 1
    return sortDir === 'asc' ? cmp : -cmp
  })
}, [devices, sortCol, sortDir, statusFilter])
```

### Pattern 3: Collapsible Sidebar
**What:** Sidebar stores collapsed state in localStorage to persist across navigation.
**When to use:** Left sidebar with icon-only collapsed mode at ~64px, full width ~220px.
**Example:**
```typescript
const [collapsed, setCollapsed] = useState(() => {
  if (typeof window === 'undefined') return false
  return localStorage.getItem('sidebar-collapsed') === 'true'
})

function toggle() {
  const next = !collapsed
  setCollapsed(next)
  localStorage.setItem('sidebar-collapsed', String(next))
}
```

### Pattern 4: Polling with useEffect cleanup
**What:** The existing polling pattern (setInterval inside useEffect, cleared on unmount) is correct. Keep it.
**When to use:** Device list (15-20s interval), device detail page (20-30s).
**Recommended interval:** 20 seconds for home page (balance freshness vs. API load). Existing code uses 15s — acceptable, keep as-is.

### Anti-Patterns to Avoid
- **Putting connection list in global nav:** Locked decision — connections belong inside the device detail page only.
- **Masking passwords:** Locked decision — credentials always shown in plaintext.
- **Rewriting api.ts from scratch:** The existing api.ts is solid. Only add `proxy_type: 'http' | 'socks5' | 'openvpn'` to `create()` signature and ensure `password` field is returned from `ListByDevice` (it currently is — `PasswordPlain` is mapped to `Password` in `ListByDevice` service method).
- **Using shadcn DataTable (TanStack):** Overkill for device list size. Plain HTML table with sort state in React useState is sufficient and simpler.
- **Animating sidebar with CSS transitions that fight Tailwind:** Use `transition-all duration-200` on the aside width; don't use transform-based approaches that complicate child layout.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Modal/dialog with keyboard trap, aria | Custom modal div | shadcn Dialog (Radix) | Focus management, ESC handling, scroll lock, a11y |
| Protocol dropdown | `<select>` element | shadcn Select (Radix) | Consistent dark theme styling; native select is hard to style across browsers |
| Toast/feedback on copy | Custom timeout state per button | Inline state (already exists) | Simple enough for this case; shadcn Toast is overkill for copy feedback |
| Confirm dialog | window.confirm() | shadcn AlertDialog | Browser native confirm is unstyled and blocks thread |

**Key insight:** The hand-roll line is: if it needs a11y primitives (focus traps, ARIA roles, keyboard navigation), use Radix/shadcn. If it's just layout or display, raw Tailwind is fine.

---

## Common Pitfalls

### Pitfall 1: OpenVPN proxy_type rejected by backend
**What goes wrong:** The `Add Connection` modal sends `proxy_type: "openvpn"` and the API returns 500 because `connection_service.go` line 73-74 only accepts `"http"` or `"socks5"`.
**Why it happens:** OpenVPN was added as a concept after the original connection service was written.
**How to avoid:** Before implementing the modal UI, add `"openvpn"` as a valid `proxy_type` in `server/internal/service/connection_service.go`. OpenVPN connections need no port allocated (they use the shared OpenVPN server port 1195), so skip the `portService.AllocatePort()` call when `proxyType == "openvpn"`.
**Warning signs:** 500 error on connection create with type=openvpn.

### Pitfall 2: shadcn/ui init modifying tailwind.config.ts
**What goes wrong:** `npx shadcn@latest init` asks to overwrite `tailwind.config.ts`. Accepting overwrites the existing `brand` color scale and box-shadow extensions.
**Why it happens:** shadcn adds its own CSS variable-based color system.
**How to avoid:** When shadcn init prompts to update `tailwind.config.ts`, choose to inspect and merge manually. Preserve the `brand` color palette. The existing zinc-based theme is already compatible with shadcn's defaults.

### Pitfall 3: shadcn Dialog not rendering in dark mode
**What goes wrong:** shadcn Dialog overlay and content use CSS variables that default to white in light mode.
**Why it happens:** shadcn's default CSS variables are in `:root {}`. The project uses a hardcoded dark background, not the `dark` class.
**How to avoid:** After `npx shadcn@latest init`, set `darkMode: 'class'` in tailwind.config.ts and add `class="dark"` to the `<html>` tag in `layout.tsx`, OR override shadcn's CSS variables directly to match the zinc dark palette. The simplest fix: override in `globals.css`:
```css
:root {
  --background: 0 0% 4%;      /* zinc-950 */
  --foreground: 0 0% 93%;     /* zinc-100 */
  --card: 0 0% 9%;            /* zinc-900 */
  --border: 0 0% 15%;         /* zinc-800 */
  --muted: 0 0% 15%;
  --muted-foreground: 0 0% 45%;
  --primary: 160 84% 39%;     /* brand-500 emerald */
  --primary-foreground: 0 0% 100%;
  --destructive: 0 84% 60%;
  --ring: 160 84% 39%;
}
```

### Pitfall 4: Stale password field after connection list refresh
**What goes wrong:** After creating a connection, the password is visible. After the 15s polling refresh hits `api.connections.list()`, the password field becomes empty/undefined.
**Why it happens:** The backend `List()` method (not `ListByDevice`) returns `PasswordPlain` mapped to `Password` — this is set in `connection_service.go:ListByDevice`. This is correct. However, the `ProxyConnection` interface in `api.ts` has `password?: string` as optional, and components may conditionally render `conn.password || '********'`. Since passwords ARE returned, this is fine — but the UI must not hide them behind a toggle.
**How to avoid:** Verify in the redesigned credential display that `conn.password` renders directly without masking guards. The existing `[id]/page.tsx` has `{conn.password || '********'}` which masks when empty — this must be removed per locked decision.

### Pitfall 5: Sidebar collapsible state causing hydration mismatch
**What goes wrong:** SSR renders the sidebar at full width (localStorage unavailable), client hydrates with collapsed state, causing React hydration mismatch warning.
**Why it happens:** localStorage is not available during SSR.
**How to avoid:** Read localStorage only in `useEffect` (after mount), not in initial state initializer. Or use `useState(false)` initially and set from localStorage in `useEffect`.

### Pitfall 6: Connection count column requires an extra API call
**What goes wrong:** The home page table needs "connection count" per device. The devices API (`GET /api/devices`) does not return connection counts.
**Why it happens:** Connection counts are not in the `Device` struct.
**How to avoid:** Two options: (a) fetch `GET /api/connections` once and compute counts client-side, or (b) accept that the connection count column requires an extra call. Recommendation: fetch all connections at page load alongside devices (one extra request, inexpensive). Store as a map `{ [deviceId]: count }` in component state.

---

## Code Examples

Verified patterns from existing codebase:

### Existing API client pattern (keep as-is, only add openvpn type)
```typescript
// Source: dashboard/src/lib/api.ts
connections: {
  create: (token: string, data: {
    device_id: string
    username: string
    password: string
    proxy_type?: 'http' | 'socks5' | 'openvpn'  // ADD openvpn
    ip_whitelist?: string[]
    bandwidth_limit?: number
  }) => request<ProxyConnection>('/connections', { method: 'POST', token, body: data }),
  // ...
}
```

### Copy-all URL format (per locked decision)
```typescript
// protocol://username:password@host:port
function buildProxyUrl(conn: ProxyConnection, device: Device, serverHost: string): string {
  const port = conn.proxy_type === 'http' ? conn.http_port : conn.socks5_port
  return `${conn.proxy_type}://${conn.username}:${conn.password}@${serverHost}:${port}`
}
// For OpenVPN: no URL format applicable — show download button only
```

### Existing polling + WebSocket pattern (keep as-is)
```typescript
// Source: dashboard/src/app/devices/page.tsx (existing pattern)
useEffect(() => {
  fetchDevices()
  const unsub = addWSHandler((msg) => {
    if (msg.type === 'device_update') {
      setDevices(prev => prev.map(d => d.id === (msg.payload as Device).id ? msg.payload as Device : d))
    }
  })
  const interval = setInterval(fetchDevices, 15000)
  return () => { unsub(); clearInterval(interval) }
}, [fetchDevices])
```

### shadcn/ui AlertDialog for delete confirmation
```typescript
// Source: shadcn/ui docs — AlertDialog pattern
import { AlertDialog, AlertDialogAction, AlertDialogCancel,
  AlertDialogContent, AlertDialogDescription, AlertDialogFooter,
  AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from "@/components/ui/alert-dialog"

<AlertDialog>
  <AlertDialogTrigger asChild>
    <Button variant="ghost" size="icon"><Trash2 className="w-4 h-4" /></Button>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Delete connection?</AlertDialogTitle>
      <AlertDialogDescription>This will free the port. This action cannot be undone.</AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <AlertDialogCancel>Cancel</AlertDialogCancel>
      <AlertDialogAction onClick={() => handleDelete(conn.id)} className="bg-destructive text-destructive-foreground">Delete</AlertDialogAction>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Hand-rolled modals with fixed inset-0 | shadcn Dialog (Radix Portal) | Required by DASH-01 | Focus trap, a11y, scroll lock handled automatically |
| Connections as top-level nav item (`/connections`) | Connections scoped to `/devices/:id` | Phase 2 decision | Simplifies nav; must delete `/connections` route and redirect |
| Overview page (`/overview`) with stat cards + device table | Devices page (`/devices`) IS the home page with stat bar | Phase 2 decision | The overview page becomes redundant — redirect `/overview` → `/devices` |

**Deprecated/outdated:**
- `/connections` route: All connection display moves into device detail page; this route should be removed or redirected.
- `/overview` route: The new `/devices` page with stat bar replaces overview. Keep or redirect.
- `customers/` route: Not mentioned in phase decisions; keep but remove from sidebar if only "Devices" is in nav.

---

## Open Questions

1. **OpenVPN as proxy_type in backend**
   - What we know: `connection_service.go` rejects `proxy_type != "http" && proxy_type != "socks5"` with a 500 error.
   - What's unclear: Does OpenVPN need a separate port allocated? Current OpenVPN uses a shared server port (1195). All credentials work the same way — same username/password for both HTTP/SOCKS5 proxy auth and OpenVPN auth.
   - Recommendation: Add `"openvpn"` as accepted proxy_type, skip port allocation for it (set `BasePort = nil`), store it as a marker so the UI can show "OpenVPN" type with download button only. This is a small backend change (~5 lines) that must happen in the first implementation task.

2. **Where does `/overview` go?**
   - What we know: The sidebar will only have "Devices". `/overview` currently exists and has its own layout.
   - What's unclear: Should `/overview` redirect to `/devices` or be deleted?
   - Recommendation: Add a redirect in `app/overview/page.tsx` (`redirect('/devices')`) — keeps any bookmarked URLs working.

3. **Connection count on device table**
   - What we know: `GET /api/devices` does not return connection counts.
   - What's unclear: Is the count column high-value enough to justify an extra API call?
   - Recommendation: Fetch all connections once with `GET /api/connections` (no `device_id` filter) and compute counts client-side. This is one additional request at page load with negligible payload.

4. **Sidebar collapse on tablet (768px)**
   - What we know: DASH-04 requires usable layout at 768px+. Current sidebar is 256px wide with no collapse.
   - What's unclear: Should the sidebar auto-collapse at 768px or remain expanded but narrower?
   - Recommendation: At `md` (768px), sidebar starts collapsed (icon-only, ~64px). At `lg` (1024px), sidebar shows labels. This is Claude's discretion.

---

## Validation Architecture

> `workflow.nyquist_validation` is not present in `.planning/config.json` — this section is skipped.

---

## Sources

### Primary (HIGH confidence)
- Direct codebase inspection — `dashboard/src/` (all files read), `server/internal/` (domain models, handlers, service)
- `dashboard/package.json` — confirmed Next.js 14.1.0, React 18.2.0, Tailwind 3.4.x, lucide-react 0.312.0
- `server/internal/service/connection_service.go` — proxy_type validation and port allocation logic confirmed
- `server/internal/api/handler/router.go` — all available API endpoints confirmed

### Secondary (MEDIUM confidence)
- shadcn/ui installation for existing Next.js 14 projects: based on shadcn official docs pattern (https://ui.shadcn.com/docs/installation/next); confirmed compatible with App Router and existing Tailwind setup
- Radix UI Dialog/AlertDialog/Select: accessible primitives, pulled in automatically by shadcn

### Tertiary (LOW confidence)
- None — all research was grounded in direct codebase inspection and official library docs.

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — confirmed from package.json and existing code
- Architecture: HIGH — based on direct inspection of every existing page and component
- Pitfalls: HIGH for backend (code read directly); MEDIUM for shadcn dark mode CSS vars (based on docs pattern)

**Research date:** 2026-02-26
**Valid until:** 2026-03-28 (Next.js 14 and shadcn are stable)
