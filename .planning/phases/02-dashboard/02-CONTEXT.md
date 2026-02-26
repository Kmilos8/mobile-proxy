# Phase 2: Dashboard - Context

**Gathered:** 2026-02-26
**Status:** Ready for planning

<domain>
## Phase Boundary

Build the operator dashboard UI so devices and proxy connections can be managed without hitting the API directly. Devices are the top-level entity; connections (HTTP, SOCKS5, OpenVPN) are scoped to individual devices and managed from the device detail page. No new backend capabilities — this phase wires existing Go API endpoints to a functional frontend.

</domain>

<decisions>
## Implementation Decisions

### Device Overview (Home Page)
- Table/list layout — dense rows, not cards
- Minimal default columns: device name, status, IP, connection count
- Full device details (carrier, battery, signal) live on the device detail page, not the table
- Offline devices are visually dimmed/grayed out
- Sortable columns (click header) + filter controls (e.g., status filter for online/offline)
- Auto-refresh via polling (10-30 second interval)
- Summary stat bar above table: total devices, online count, offline count — devices only, no connection counts
- Clicking a device row navigates to a dedicated device detail page

### Visual Style & Navigation
- Dark theme only (no light/toggle)
- Vercel/Linear aesthetic: minimal, monochrome, sharp typography, generous whitespace
- Left sidebar navigation, collapsible
- Sidebar sections: Devices only — connections are not a top-level nav item
- Connections are scoped inside each device's detail page

### Credential Presentation
- All credentials always visible in plain text — no masking, no reveal toggle
- Per-field copy buttons on each credential (host, port, username, password)
- "Copy All" button that produces URL format: `protocol://username:password@host:port`
- OpenVPN connections show a download button for the .ovpn file (no config preview/expansion)

### Connection Management Flow
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

</decisions>

<specifics>
## Specific Ideas

- "I want it to feel like Vercel/Linear" — clean, monochrome, sharp typography, generous whitespace
- Connections are a sub-resource of devices — no separate connections page or top-level nav
- Summary stats track devices only (online/offline counts), not connection counts
- Copy All uses URL format: `socks5://user:pass@host:port` — ready to paste into proxy tools

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 02-dashboard*
*Context gathered: 2026-02-26*
