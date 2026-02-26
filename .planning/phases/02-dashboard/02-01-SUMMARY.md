---
phase: 02-dashboard
plan: 01
subsystem: ui
tags: [nextjs, shadcn-ui, tailwindcss, react, dark-theme]

# Dependency graph
requires:
  - phase: 01-openvpn-throughput
    provides: backend connection service with openvpn proxy_type support
provides:
  - shadcn/ui component library installed with dark zinc theme
  - Collapsible sidebar with Devices-only navigation
  - /devices home page with StatBar (total/online/offline counts)
  - DeviceTable with sortable columns and status filter
  - /overview redirects to /devices
  - Backend accepts openvpn as valid proxy_type (no port allocation)
affects: [02-dashboard]

# Tech tracking
tech-stack:
  added: [shadcn/ui, @radix-ui/react-alert-dialog, @radix-ui/react-dialog, @radix-ui/react-select, @radix-ui/react-separator, @radix-ui/react-slot, class-variance-authority, tailwindcss-animate]
  patterns: [shadcn-ui components in src/components/ui/, dark-only CSS variables in :root, 'use client' for interactive components, 15s polling + WebSocket for device data]

key-files:
  created:
    - dashboard/src/components/devices/StatBar.tsx
    - dashboard/src/components/devices/DeviceTable.tsx
    - dashboard/components.json
    - dashboard/src/components/ui/button.tsx
    - dashboard/src/components/ui/dialog.tsx
    - dashboard/src/components/ui/select.tsx
    - dashboard/src/components/ui/table.tsx
    - dashboard/src/components/ui/badge.tsx
    - dashboard/src/components/ui/separator.tsx
    - dashboard/src/components/ui/alert-dialog.tsx
  modified:
    - dashboard/package.json
    - dashboard/tailwind.config.ts
    - dashboard/src/app/globals.css
    - dashboard/src/app/layout.tsx
    - dashboard/src/lib/api.ts
    - dashboard/src/components/dashboard/Sidebar.tsx
    - dashboard/src/app/overview/page.tsx
    - dashboard/src/app/devices/page.tsx
    - server/internal/service/connection_service.go

key-decisions:
  - "shadcn/ui with zinc base color and CSS variables for consistent dark theme"
  - "StatBar shows only device counts (no connection counts) per locked plan decision"
  - "DeviceTable uses dense table layout (not cards) per locked plan decision"
  - "Offline device rows use opacity-50 for visual dimming per locked plan decision"

patterns-established:
  - "shadcn UI components live in dashboard/src/components/ui/ and are generated via npx shadcn@latest add"
  - "Device-specific feature components live in dashboard/src/components/devices/"
  - "Dark theme is set via 'dark' class on html element with CSS variable definitions in :root block"
  - "Sidebar collapsed state read from localStorage only in useEffect to prevent hydration mismatch"

requirements-completed: [DASH-01, DASH-04]

# Metrics
duration: 30min
completed: 2026-02-26
---

# Phase 2 Plan 01: shadcn/ui foundation, collapsible sidebar, and sortable device table

**shadcn/ui dark theme installed, sidebar collapsed to Devices-only nav, /devices page rebuilt with StatBar + sortable/filterable DeviceTable using zinc design system**

## Performance

- **Duration:** ~30 min
- **Started:** 2026-02-26T07:30:00Z
- **Completed:** 2026-02-26T08:00:00Z
- **Tasks:** 3
- **Files modified:** 18

## Accomplishments
- shadcn/ui installed with 7 components (button, dialog, select, table, badge, separator, alert-dialog) using zinc base, dark CSS variables, preserved brand emerald colors and glow shadows
- Sidebar rewritten: Devices-only nav, collapsible (64px icon / 220px expanded), localStorage persistence without hydration mismatch, defaults collapsed at <1024px
- /devices home page rebuilt: StatBar (total/online/offline), DeviceTable with click-to-sort headers, status filter, dimmed offline rows, clickable rows navigate to /devices/{id}
- Backend already accepted openvpn proxy_type (committed in earlier session)

## Task Commits

Each task was committed atomically:

1. **Task 1: Install shadcn/ui and configure dark theme** - `6d7208f` (feat) — committed in prior session
2. **Task 2: Rewrite sidebar (collapsible, Devices-only) and redirect /overview** - `471e791` (feat)
3. **Task 3: Rebuild /devices with StatBar and sortable DeviceTable** - `7bef2d4` (feat)

**Plan metadata:** (docs commit below)

## Files Created/Modified
- `dashboard/src/components/devices/StatBar.tsx` - Three-box stat row (total/online/offline device counts)
- `dashboard/src/components/devices/DeviceTable.tsx` - Sortable/filterable dense table with status filter, offline dimming, click-to-navigate rows
- `dashboard/src/app/devices/page.tsx` - Rewritten: StatBar + DeviceTable, fetchConnections for per-device counts, PairingModal preserved
- `dashboard/src/components/dashboard/Sidebar.tsx` - Collapsible sidebar with Devices-only nav, localStorage persistence
- `dashboard/src/app/overview/page.tsx` - Simple redirect to /devices
- `dashboard/components.json` - shadcn/ui config (zinc, CSS variables, src/components/ui)
- `dashboard/src/components/ui/button.tsx` - shadcn Button with CVA variants
- `dashboard/src/components/ui/table.tsx` - shadcn Table, TableHeader, TableRow, TableCell, etc.
- `dashboard/src/components/ui/dialog.tsx` - shadcn Dialog for modals
- `dashboard/src/components/ui/select.tsx` - shadcn Select for dropdowns
- `dashboard/src/components/ui/badge.tsx` - shadcn Badge
- `dashboard/src/components/ui/separator.tsx` - shadcn Separator
- `dashboard/src/components/ui/alert-dialog.tsx` - shadcn AlertDialog for confirmations
- `dashboard/src/app/globals.css` - shadcn CSS variables for dark-only theme, brand properties preserved
- `dashboard/src/app/layout.tsx` - Added 'dark' class to html element
- `dashboard/tailwind.config.ts` - darkMode: 'class', CSS variable colors, brand scale, glow shadows, tailwindcss-animate
- `dashboard/src/lib/api.ts` - ProxyConnection.proxy_type includes 'openvpn'
- `server/internal/service/connection_service.go` - Accepts openvpn, skips port allocation and DNAT for openvpn

## Decisions Made
- Used zinc base color and CSS variable-based dark theme (shadcn default) — consistent with existing zinc-900/950 palette
- StatBar devices only (no connection counts) — per plan locked decision
- Dense table layout (not cards) for DeviceTable — per plan locked decision
- Offline rows opacity-50 — per plan locked decision

## Deviations from Plan

None — plan executed exactly as written. Tasks 1 and 2 sidebar/overview were already partially implemented in prior session (committed in `6d7208f`).

## Issues Encountered
- Task 1 (shadcn/ui install) was already committed from a prior session as `6d7208f`. Verified the commit covered all required elements and proceeded to Tasks 2 and 3.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- shadcn/ui design system foundation in place for Plan 02 (device detail page, connection management)
- DeviceTable established pattern for future data tables
- Sidebar extensible: add nav items to `navItems` array when new sections are needed
- `next build` passes clean with no TypeScript errors

---
*Phase: 02-dashboard*
*Completed: 2026-02-26*
