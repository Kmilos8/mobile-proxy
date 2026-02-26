---
phase: 02-dashboard
plan: 02
subsystem: ui
tags: [nextjs, react, shadcn-ui, tailwindcss, proxy-connections, crud]

# Dependency graph
requires:
  - phase: 02-dashboard
    plan: 01
    provides: shadcn/ui component library (Button, Dialog, AlertDialog, Select, Table, Badge) and dark zinc theme
  - phase: 01-openvpn-throughput
    provides: backend connection API with HTTP/SOCKS5/OpenVPN proxy_type support and .ovpn download endpoint
provides:
  - ConnectionTable component: per-device unified connection list with per-field copy buttons, Copy All URL, OpenVPN download, delete
  - AddConnectionModal component: protocol selector (HTTP/SOCKS5/OpenVPN), auto-generated credentials, error handling
  - DeleteConnectionDialog component: AlertDialog confirmation for connection deletion
  - Device detail page /devices/[id]: unified connection management section, responsive mobile layout
affects: [02-dashboard, 03-customer-portal]

# Tech tracking
tech-stack:
  added: []
  patterns: [connection components in dashboard/src/components/connections/, mobile-first responsive with hidden md:block / md:hidden pattern, AddConnectionModal modal lifted to page level to avoid z-index clipping]

key-files:
  created:
    - dashboard/src/components/connections/ConnectionTable.tsx
    - dashboard/src/components/connections/AddConnectionModal.tsx
    - dashboard/src/components/connections/DeleteConnectionDialog.tsx
  modified:
    - dashboard/src/app/devices/[id]/page.tsx

key-decisions:
  - "Password displayed in plaintext (no masking) per locked plan decision — operators need raw credentials"
  - "Copy All URL format: protocol://username:password@host:port for HTTP and SOCKS5; OpenVPN uses download button instead"
  - "OpenVPN port displayed as 1195 (fixed) since conn.http_port and conn.socks5_port are null for OpenVPN connections"
  - "AddConnectionModal lifted to page level in DeviceDetailPage to prevent Dialog z-index/clipping issues with sidebar layout"
  - "Mobile responsive: sidebar tabs replaced with horizontal scrollable tab bar below md breakpoint; connection table replaced with stacked card layout"

patterns-established:
  - "Connection feature components live in dashboard/src/components/connections/"
  - "Per-field copy buttons use inline state (copiedKey) with 2s timeout showing 'Copied!' feedback"
  - "Delete dialogs receive ProxyConnection | null and open boolean — parent manages open state separately from target"
  - "Protocol Badge variants: http=default, socks5=secondary, openvpn=outline"

requirements-completed: [DASH-02, DASH-03, DASH-04]

# Metrics
duration: 4min
completed: 2026-02-26
---

# Phase 2 Plan 02: Device detail page with full connection management Summary

**Unified connection table (HTTP/SOCKS5/OpenVPN) with per-field copy, Copy All URL format, OpenVPN .ovpn download, and add/delete modals — full connection CRUD from device detail page**

## Performance

- **Duration:** ~4 min
- **Started:** 2026-02-26T07:58:32Z
- **Completed:** 2026-02-26T08:02:46Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- ConnectionTable: renders all proxy types in a unified shadcn Table with type Badge, per-field copy buttons (host/port/username/password), Copy All URL format for HTTP/SOCKS5, OpenVPN download button, delete trigger — plus responsive mobile card layout below md breakpoint
- AddConnectionModal: Dialog with protocol selector (HTTP/SOCKS5/OpenVPN), auto-generated username and password pre-filled, inline error display, loading state with disabled submit
- DeleteConnectionDialog: AlertDialog with "Delete connection?" title, descriptive warning, Cancel/Delete buttons (Delete is destructive red)
- Device detail /devices/[id] page: Primary tab rebuilt with unified ConnectionTable + Add Connection button; mobile sidebar replaced with horizontal scrollable tab bar; all other tabs (Advanced, Change IP, History, Metrics, Usage) unchanged

## Task Commits

Each task was committed atomically:

1. **Task 1: Create ConnectionTable, AddConnectionModal, and DeleteConnectionDialog components** - `a6850cd` (feat)
2. **Task 2: Rewrite device detail page with connection management section** - `7184eca` (feat)

**Plan metadata:** (docs commit below)

## Files Created/Modified
- `dashboard/src/components/connections/ConnectionTable.tsx` - Unified per-device connection list: type badge, host/port/username/password with per-field copy, Copy All URL, OpenVPN download, delete via dialog, responsive mobile cards
- `dashboard/src/components/connections/AddConnectionModal.tsx` - Dialog modal: protocol selector (HTTP/SOCKS5/OpenVPN), auto-generated credentials as default values, create via api.connections.create, inline error, loading state
- `dashboard/src/components/connections/DeleteConnectionDialog.tsx` - AlertDialog confirmation: "Delete connection?" title, port-free warning, Cancel + destructive Delete buttons
- `dashboard/src/app/devices/[id]/page.tsx` - Primary tab replaced old dual HTTP/SOCKS5 tables with ConnectionTable + Add Connection button; mobile horizontal tab bar added; AddConnectionModal and delete handler wired at page level

## Decisions Made
- Password shown in plaintext with no fallback masking — per locked plan decision (operators need real credentials)
- Copy All URL format is `protocol://username:password@host:port` for HTTP and SOCKS5; OpenVPN shows Download icon only (not applicable for URL format)
- OpenVPN port hardcoded to "1195" since `http_port`/`socks5_port` are null for OpenVPN connections on the ProxyConnection interface
- AddConnectionModal rendered at page level (not inside ConnectionTable) to prevent shadcn Dialog from being clipped by the flex/sidebar layout
- Mobile sidebar: `hidden md:block` for the vertical sidebar, horizontal `overflow-x-auto` tab bar for mobile — no collapse/hamburger, just a simple scroll row

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Connection CRUD fully operational from the dashboard for all three proxy types
- Device detail page now the primary operator interface: create connections, copy credentials, download .ovpn files, delete connections
- ConnectionTable, AddConnectionModal, and DeleteConnectionDialog are reusable components for any future page needing connection management
- `next build` passes clean with no TypeScript errors

## Self-Check: PASSED

- FOUND: dashboard/src/components/connections/ConnectionTable.tsx (297 lines, min 80)
- FOUND: dashboard/src/components/connections/AddConnectionModal.tsx (165 lines, min 60)
- FOUND: dashboard/src/components/connections/DeleteConnectionDialog.tsx (52 lines, min 30)
- FOUND: dashboard/src/app/devices/[id]/page.tsx (850 lines, min 100)
- FOUND commit: a6850cd (Task 1 — three connection components)
- FOUND commit: 7184eca (Task 2 — device detail page rewrite)

---
*Phase: 02-dashboard*
*Completed: 2026-02-26*
