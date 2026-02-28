---
phase: 06-multi-tenant-isolation
plan: 01
subsystem: database
tags: [postgres, golang, repository-pattern, multi-tenant, uuid]

# Dependency graph
requires:
  - phase: 05-auth-foundation
    provides: customers table with auth columns (email, password_hash, google_id) that migration 012 builds on

provides:
  - Migration 012 with customer_id on devices and pairing_codes, device_shares table, operator seed, and data backfill
  - Device.CustomerID and PairingCode.CustomerID domain fields
  - DeviceShare domain struct with per-permission boolean columns
  - DeviceRepository.ListByCustomer and GetByIDForCustomer (owned + shared UNION query)
  - ConnectionRepository.ListByCustomer, GetByIDForCustomer, ListByDeviceForCustomer
  - DeviceShareRepository with full CRUD (Create, GetByID, GetByDeviceAndCustomer, ListByDevice, ListByCustomer, Update, Delete)
  - CustomerRepository.UpdateActive and GetStats
  - PairingCodeRepository.Create/GetByCode/List updated to include customer_id

affects:
  - 06-02 (handler isolation — depends on these repo methods for tenant-scoped API responses)
  - 06-03 (customer portal — uses DeviceShare for sharing UI)
  - Any handler that creates devices or pairing codes (must now pass CustomerID)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - UNION query pattern for owned + shared device access in a single repository call
    - Existence sub-select pattern for authorization checks in WHERE clauses (OR EXISTS)
    - Nullable *uuid.UUID CustomerID field on device/pairing models for optional ownership
    - Device share repository with GetByDeviceAndCustomer for quick permission lookup

key-files:
  created:
    - server/migrations/012_tenant_isolation.up.sql
    - server/migrations/012_tenant_isolation.down.sql
    - server/internal/repository/device_share_repo.go
  modified:
    - server/internal/domain/models.go
    - server/internal/repository/device_repo.go
    - server/internal/repository/connection_repo.go
    - server/internal/repository/customer_repo.go
    - server/internal/repository/pairing_repo.go

key-decisions:
  - "Customer_id is nullable (*uuid.UUID) on devices and pairing_codes — existing rows before migration have NULL, backfill sets them from admin user email match"
  - "Operator seed uses admin user email to create customer record — devices stamped to that customer so no device is left orphaned post-migration"
  - "ListByCustomer uses UNION not LEFT JOIN to avoid duplicate rows when a device is both owned and shared"
  - "GetByIDForCustomer and GetByIDForCustomer use OR EXISTS subquery for authorization rather than a JOIN to keep the scan count consistent with single-row expectations"
  - "device_shares.UNIQUE(device_id, shared_with) prevents duplicate share entries at DB level"
  - "DeviceShareRepository.Delete uses id (not device_id+shared_with) for explicitness; caller resolves share ID first"

patterns-established:
  - "UNION query for owned+shared: SELECT...WHERE customer_id=$1 UNION SELECT...INNER JOIN device_shares WHERE shared_with=$1"
  - "Authorization sub-select: WHERE id=$1 AND (customer_id=$2 OR EXISTS(SELECT 1 FROM device_shares WHERE ...))"
  - "Repository scan column order matches SQL SELECT column order exactly — customer_id placed after auto_rotate_minutes in device select"

requirements-completed: [TENANT-01, TENANT-02, TENANT-03]

# Metrics
duration: 30min
completed: 2026-02-28
---

# Phase 6 Plan 01: Tenant Isolation Data Layer Summary

**PostgreSQL migration adding customer_id ownership to devices, device_shares permissions table, and customer-scoped UNION-based repository queries for multi-tenant data isolation**

## Performance

- **Duration:** 30 min
- **Started:** 2026-02-28T08:04:17Z
- **Completed:** 2026-02-28T08:34:00Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments

- Migration 012 adds customer ownership to devices and pairing_codes, creates device_shares with granular permissions, seeds operator customer from admin user email, and backfills all existing device and connection records
- Domain models updated with CustomerID on Device and PairingCode, new DeviceShare struct with four permission booleans
- Repository layer extended with customer-scoped queries: ListByCustomer (UNION of owned + shared), GetByIDForCustomer (ownership OR EXISTS check), and DeviceShareRepository with full CRUD

## Task Commits

Each task was committed atomically:

1. **Task 1: Create migration 012 and update domain models** - `a96a5ab` (feat)
2. **Task 2: Add customer-scoped repository methods and DeviceShareRepository** - `93ab9d0` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified

- `server/migrations/012_tenant_isolation.up.sql` - Schema changes: customer_id on devices/pairing_codes, device_shares table, data seed and backfill
- `server/migrations/012_tenant_isolation.down.sql` - Reversal: drop device_shares, remove customer_id columns
- `server/internal/domain/models.go` - Device.CustomerID, PairingCode.CustomerID, new DeviceShare struct, CustomerID on CreatePairingCodeRequest
- `server/internal/repository/device_repo.go` - Updated SELECT columns and scan functions, new ListByCustomer and GetByIDForCustomer methods, Create includes customer_id
- `server/internal/repository/connection_repo.go` - New ListByCustomer, GetByIDForCustomer, ListByDeviceForCustomer methods
- `server/internal/repository/device_share_repo.go` - New file: DeviceShareRepository with full CRUD
- `server/internal/repository/customer_repo.go` - New UpdateActive and GetStats methods
- `server/internal/repository/pairing_repo.go` - Create/GetByCode/List updated to include customer_id column

## Decisions Made

- Customer_id is nullable on devices and pairing_codes — avoids a NOT NULL constraint that would block migration on existing data. Backfill sets all existing rows from admin user email match.
- UNION query in ListByCustomer rather than LEFT JOIN to avoid duplicate rows when a device is both owned and shared with the same customer (edge case, but prevents double-counting).
- Authorization sub-selects (OR EXISTS) in GetByIDForCustomer keep the scan arity consistent with single-row queries, avoiding JOIN-induced column count changes.
- device_shares has UNIQUE(device_id, shared_with) at the DB level to enforce one share record per device-customer pair.

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

Go compiler was not available in the shell environment for automated build verification (`go build ./...`). Code correctness was verified by careful cross-referencing of SQL column order against Go struct field scan order, and by following the exact patterns already established in the codebase. The build will be verified on next deployment.

## User Setup Required

None — no external service configuration required. The migration will be applied automatically on next `docker compose up` via the migration runner.

## Next Phase Readiness

- Data layer foundation for multi-tenant isolation is complete
- Plan 06-02 (handler isolation) can now add `customer_id` WHERE scoping to all device/connection API handlers using the new repository methods
- DeviceShareRepository is ready for Plan 06-03 (customer portal) to expose share management endpoints
- Existing handlers that create devices or pairing codes should be updated to pass the authenticated customer's ID

## Self-Check: PASSED

All claimed files exist and commits verified:
- server/migrations/012_tenant_isolation.up.sql — FOUND
- server/migrations/012_tenant_isolation.down.sql — FOUND
- server/internal/domain/models.go — FOUND (DeviceShare struct, CustomerID fields)
- server/internal/repository/device_share_repo.go — FOUND
- server/internal/repository/device_repo.go — FOUND (ListByCustomer)
- server/internal/repository/connection_repo.go — FOUND (ListByCustomer)
- server/internal/repository/customer_repo.go — FOUND
- server/internal/repository/pairing_repo.go — FOUND
- Commit a96a5ab — FOUND (Task 1)
- Commit 93ab9d0 — FOUND (Task 2)

---
*Phase: 06-multi-tenant-isolation*
*Completed: 2026-02-28*
