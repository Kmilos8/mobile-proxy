---
phase: 06-multi-tenant-isolation
plan: 02
subsystem: api
tags: [golang, gin, middleware, multi-tenant, rbac, device-sharing]

# Dependency graph
requires:
  - phase: 06-01
    provides: customer-scoped repository methods (ListByCustomer, GetByIDForCustomer), DeviceShareRepository, CustomerRepository.UpdateActive and GetStats

provides:
  - AdminOnlyMiddleware blocking customer tokens on admin-only routes
  - CustomerSuspensionCheck middleware validating customer.Active on every authenticated request
  - DeviceShareService with CanAccess, CanDo, CreateShare, UpdateShare, DeleteShare, ListSharesForDevice
  - DeviceHandler role-branching: customer tokens get ListByCustomer, GetByIDForCustomer, share permission checks
  - ConnectionHandler role-branching: customer tokens get customer-scoped list/get/create/delete with share checks; ResetBandwidth blocked for customers
  - DeviceShareHandler CRUD endpoints: POST/GET/PUT/DELETE /api/device-shares
  - CustomerHandler Suspend and Activate endpoints plus GetDetail with aggregate stats
  - Route restructure: adminOnly sub-group with AdminOnlyMiddleware; CustomerSuspensionCheck on full dashboard group
  - Pairing flow stamps customer_id from pairing code onto newly registered device via DeviceRepository.SetCustomerID
  - OpenVPNHandler.DownloadOVPN checks download_configs permission for customer tokens

affects:
  - 06-03 (customer portal — can now call the role-branched API endpoints directly)
  - Any admin dashboard frontend that calls /api/customers (now adminOnly-gated)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Role-branching pattern: read user_role from gin context, branch on "customer" vs admin/operator
    - Service permission layer: CanAccess (any access) vs CanDo (specific perm) separation
    - AdminOnly sub-group: dashboard.Group("").Use(AdminOnlyMiddleware()) for a clean route tier
    - CustomerSuspensionCheck as chained middleware after AuthMiddleware on the dashboard group
    - SetShareService setter pattern on OpenVPNHandler to avoid circular dep at construction time

key-files:
  created:
    - server/internal/service/device_share_service.go
    - server/internal/api/handler/device_share_handler.go
  modified:
    - server/internal/api/middleware/auth.go
    - server/internal/service/device_service.go
    - server/internal/service/connection_service.go
    - server/internal/api/handler/device_handler.go
    - server/internal/api/handler/connection_handler.go
    - server/internal/api/handler/customer_handler.go
    - server/internal/api/handler/router.go
    - server/internal/api/handler/pairing_handler.go
    - server/internal/api/handler/openvpn_handler.go
    - server/internal/service/pairing_service.go
    - server/internal/repository/device_repo.go
    - server/cmd/api/main.go

key-decisions:
  - "AdminOnlyMiddleware applied to a sub-group (adminOnly := dashboard.Group('').Use(...)) — keeps device/connection routes in the parent group so customers can access them via role-branching"
  - "CustomerSuspensionCheck runs on the full dashboard group after AuthMiddleware — every authenticated request checks the active flag for customer tokens"
  - "DeviceShareService.CanAccess vs CanDo separation — CanAccess gates read-only views (ip-history, bandwidth, commands), CanDo gates write actions (rename, manage_ports, download_configs, rotate_ip)"
  - "Customer SendCommand: only rotate_ip and rotate_ip_airplane allowed; all other command types (reboot, find_phone, etc.) return 403"
  - "PairingService.CreateCode now accepts *uuid.UUID customerID — nullable so existing admin callers without customer_id still work"
  - "DeviceRepository.SetCustomerID added to stamp ownership on device claim without touching the device.Create path"
  - "OpenVPNHandler.SetShareService setter pattern — avoids passing shareService through the NewOpenVPNHandler constructor since it's not always needed"

requirements-completed: [TENANT-01, TENANT-02, TENANT-03]

# Metrics
duration: 45min
completed: 2026-02-28
---

# Phase 6 Plan 02: Handler Isolation and API Enforcement Summary

**AdminOnlyMiddleware protecting all admin routes, role-branching in device/connection handlers for customer-scoped data access, DeviceShareService permission layer, DeviceShareHandler CRUD endpoints, customer suspension enforcement, and pairing code customer assignment**

## Performance

- **Duration:** 45 min
- **Started:** 2026-02-28T08:18:00Z
- **Completed:** 2026-02-28T09:03:00Z
- **Tasks:** 2
- **Files modified:** 12

## Accomplishments

- Two middleware functions added to auth.go: AdminOnlyMiddleware (role check, 403 for customers) and CustomerSuspensionCheck (DB lookup of customer.Active, 403 if suspended/missing)
- DeviceShareService created with CanAccess (read gating) and CanDo (per-permission gating), plus full share CRUD methods with owner validation
- Device handler now branches on user_role: customer tokens call ListByCustomer/GetByIDForCustomer; Update checks CanDo("rename"); SendCommand allows only rotate_ip for customers; all device metrics endpoints check CanAccess
- Connection handler branches on user_role: List uses customer-scoped queries when device_id present or not; GetByID uses GetByIDForCustomer; Create stamps customer_id and checks CanDo("manage_ports"); Delete and RegeneratePassword check CanDo("manage_ports"); ResetBandwidth blocks customers entirely
- DeviceShareHandler wired with CreateShare, ListShares, UpdateShare, DeleteShare endpoints at /api/device-shares
- Router restructured: admin-only sub-group with AdminOnlyMiddleware covers stats, customers, pairing-codes, relay-servers, rotation-links, settings; CustomerSuspensionCheck on full dashboard group; device-shares CRUD registered in dashboard group
- CustomerHandler enhanced with Suspend, Activate, and GetDetail (enriched response with device_count, share_count, total_bandwidth)
- PairingService.CreateCode now accepts customerID parameter and stores it on the PairingCode; ClaimCode stamps customer_id from pairing code onto newly registered device via new DeviceRepository.SetCustomerID method
- OpenVPNHandler.DownloadOVPN permission-gated for customer tokens via CanDo("download_configs")

## Task Commits

Each task was committed atomically:

1. **Task 1: Add middleware, DeviceShareService, and update handlers for role-branching** - `44dbd1d` (feat)
2. **Task 2: Create DeviceShareHandler, wire routes, and update pairing flow** - `a23b76b` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified

- `server/internal/api/middleware/auth.go` - AdminOnlyMiddleware and CustomerSuspensionCheck added after existing AuthMiddleware
- `server/internal/service/device_share_service.go` - New file: DeviceShareService with CanAccess, CanDo, CreateShare, UpdateShare, DeleteShare, ListSharesForDevice
- `server/internal/service/device_service.go` - ListByCustomer and GetByIDForCustomer thin wrappers added
- `server/internal/service/connection_service.go` - ListByCustomer, GetByIDForCustomer, ListByDeviceForCustomer thin wrappers added
- `server/internal/api/handler/device_handler.go` - DeviceHandler struct gets shareService field; NewDeviceHandler updated; role-branching in List, GetByID, Update, SendCommand, GetIPHistory, GetBandwidth, GetBandwidthHourly, GetUptime, GetCommands
- `server/internal/api/handler/connection_handler.go` - ConnectionHandler gets shareService; role-branching in List, GetByID, Create (stamps customerID), Delete, RegeneratePassword; ResetBandwidth blocks customers
- `server/internal/api/handler/customer_handler.go` - Suspend, Activate, GetDetail methods added
- `server/internal/api/handler/device_share_handler.go` - New file: DeviceShareHandler with CreateShare, ListShares, UpdateShare, DeleteShare
- `server/internal/api/handler/router.go` - adminOnly sub-group with AdminOnlyMiddleware; CustomerSuspensionCheck on dashboard group; device-shares CRUD registered; SetupRouter signature extended with deviceShareHandler, customerRepo, shareService
- `server/internal/api/handler/pairing_handler.go` - CreateCode passes req.CustomerID to service
- `server/internal/api/handler/openvpn_handler.go` - SetShareService setter; DownloadOVPN checks download_configs permission for customers
- `server/internal/service/pairing_service.go` - CreateCode accepts customerID, stores on PairingCode; ClaimCode stamps customer_id on device via SetCustomerID
- `server/internal/repository/device_repo.go` - SetCustomerID method added
- `server/cmd/api/main.go` - deviceShareRepo, deviceShareService, deviceShareHandler wired; openvpnHandler gets share service; SetupRouter call updated

## Decisions Made

- AdminOnlyMiddleware applied to a sub-group within the dashboard group (not a separate route group) — this keeps device/connection routes in the parent group for customer access via role-branching. Clean separation without duplicating middleware setup.
- CustomerSuspensionCheck runs on every dashboard request for customer tokens — this is a single DB lookup per request and ensures suspended accounts are immediately locked out without token invalidation complexity.
- CanAccess vs CanDo separation in DeviceShareService — read-only endpoints (bandwidth, ip-history, commands) only need CanAccess; write endpoints need specific permission booleans. Owners always pass both.
- Customer SendCommand restricted to rotate_ip/rotate_ip_airplane only — all other commands (reboot, find_phone, wifi, etc.) are admin-only operations that could cause device disruption.
- PairingService.CreateCode customerID parameter is nullable to maintain backward compatibility with existing admin callers that don't specify a customer.
- DeviceRepository.SetCustomerID added as a minimal targeted update — avoids touching the device.Create path which would require re-reading and updating device registration request types.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical Functionality] Added DeviceRepository.SetCustomerID**
- **Found during:** Task 2 (pairing flow implementation)
- **Issue:** ClaimCode needed to stamp customer_id on newly registered device, but no method existed to update only the customer_id column on an existing device
- **Fix:** Added SetCustomerID(ctx, id, customerID) method to DeviceRepository; called from PairingService.ClaimCode after device registration
- **Files modified:** server/internal/repository/device_repo.go, server/internal/service/pairing_service.go
- **Commit:** a23b76b

**2. [Rule 2 - Missing Critical Functionality] Added DownloadOVPN permission check for customers**
- **Found during:** Task 2 (router restructure)
- **Issue:** Plan mentioned the download_configs permission check for DownloadOVPN but the OpenVPNHandler had no shareService reference to perform the check
- **Fix:** Added SetShareService setter to OpenVPNHandler; added role-branch check in DownloadOVPN; wired in main.go and noted in router
- **Files modified:** server/internal/api/handler/openvpn_handler.go, server/cmd/api/main.go
- **Commit:** a23b76b

## Issues Encountered

Go compiler was not available in the shell environment for automated build verification (`go build ./...`). Code correctness was verified by:
- Cross-referencing all function signatures before and after modification
- Checking all callers of modified constructors (NewDeviceHandler, SetupRouter, pairingService.CreateCode)
- Verifying import paths are consistent with existing package structure
- Confirming repository method signatures match service call sites

## User Setup Required

None — no external service configuration required. The middleware and handler changes are purely application code. Deployment requires `docker compose up --build` on the relay/dashboard VPS.

## Next Phase Readiness

- All tenant isolation enforcement is now in place at the API layer
- Plan 06-03 (customer portal) can build frontend pages that call the role-branched API endpoints directly
- The device-shares CRUD API is ready for the sharing UI in the customer portal
- Admin dashboard can call /api/customers/:id/suspend and /api/customers/:id/activate for account management

## Self-Check: PASSED

All claimed files exist:
- server/internal/api/middleware/auth.go — FOUND (AdminOnlyMiddleware, CustomerSuspensionCheck)
- server/internal/service/device_share_service.go — FOUND (new file)
- server/internal/api/handler/device_share_handler.go — FOUND (new file)
- server/internal/api/handler/router.go — FOUND (adminOnly sub-group, CustomerSuspensionCheck)
- server/internal/service/pairing_service.go — FOUND (customerID param, SetCustomerID call)
- server/internal/repository/device_repo.go — FOUND (SetCustomerID method)
- server/cmd/api/main.go — FOUND (deviceShareRepo, deviceShareService, deviceShareHandler wired)
- Commit 44dbd1d — FOUND (Task 1)
- Commit a23b76b — FOUND (Task 2)

---
*Phase: 06-multi-tenant-isolation*
*Completed: 2026-02-28*
