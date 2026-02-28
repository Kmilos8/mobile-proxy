# Phase 6: Multi-Tenant Isolation - Research

**Researched:** 2026-02-28
**Domain:** Multi-tenant data scoping, permission systems, JWT role claims, Go/Gin middleware, PostgreSQL, Next.js
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Device Assignment Model**
- Customers self-onboard their own devices (not operator-assigned)
- Ownership is at the device level — one device = one owner, all connections inherit the owner's customer_id
- Ownership is permanent — no transfers between customers
- Admin can see and manage all devices across all customers but cannot reassign ownership

**Device Sharing & Permissions**
- Device owner can share their device with other customers
- Sharing grants granular, toggleable permissions:
  - Rename — change device name/description
  - Manage ports — add/remove proxy connections
  - Download configs — generate and download .ovpn files
  - Rotate IP + view usage — trigger IP rotation and view bandwidth stats
- Baseline: shared user can always view the device and its connections
- Owner can revoke access or change individual permissions at any time
- Changes take effect immediately

**Admin Dashboard Changes**
- Add customer filter/dropdown to existing device list (filter by customer or view all)
- New customer management page listing all customers
- Customer detail shows: email, signup date, verification status, device count, active shares, last login, total bandwidth, account status (active/suspended)
- Admin can suspend a customer account — all their devices go offline and all outgoing shares are paused until reactivated
- Admin cannot reassign device ownership — view only + disable/suspend

**Account & Role Model**
- Admin is a separate system account (god-mode oversight, does not own devices in the customer sense)
- Customer accounts are the standard account type — they own and manage devices
- Existing devices migrate to a new customer account for the current operator/owner (not to the admin account)
- Devices without a valid customer_id are rejected at the tunnel connection level (hard enforcement)

**Role Separation in Dashboard**
- Same dashboard app, different views based on role (not separate apps/URLs)
- Customer lands on filtered version of existing device list (shows only their devices + shared devices)
- Same sidebar structure — admin-only items are hidden (not rendered) for customer role
- Customers cannot see: other customers' data, system metrics/monitoring, customer management pages
- Suspended account: all shares from that account become inaccessible to shared users

### Claude's Discretion
- Database schema design for customer_id, sharing, and permissions tables
- JWT structure for customer_id and role claims
- API middleware approach for tenant filtering
- Migration strategy for existing devices
- Permission check implementation pattern (middleware vs per-handler)

### Deferred Ideas (OUT OF SCOPE)
- Customer self-service QR code generation — Customer generates QR from their portal to onboard a device. Belongs in Phase 7 (Customer Portal).
- Balance/payment on connection creation — Connections require payment from pre-loaded balance. Beyond v2.0 scope (payment handled externally via Stripe per requirements).
- Device transfer between customers — Ownership is permanent for now. Could be added as a future capability if needed.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| TENANT-01 | Customer sees only their own assigned devices and connections | `device_ownership` table + WHERE customer_id filter in ListByCustomer repo methods; shared devices appear via JOIN on `device_shares` |
| TENANT-02 | Operator (admin) can assign devices to customers | In this phase the "assignment" is self-onboarding via pairing code claim (customer_id stamped at pairing time); admin gets a customer management page with suspend/activate; admin sees all devices with customer filter dropdown |
| TENANT-03 | Customer JWT carries customer_id; all portal queries filter by it | `generateCustomerJWT` already sets `UserID = customer.ID` and `Role = "customer"`; a new `CustomerAuthMiddleware` extracts `customer_id` from claims and injects it into gin context; all protected handlers read it |
</phase_requirements>

---

## Summary

Phase 5 delivered a complete customer authentication system: signup, login, email verification, password reset, and Google OAuth. Customers can now register and receive JWTs. Phase 6 builds on that foundation to make the data layer multi-tenant: every device and connection is owned by a specific customer, and every API response is filtered to the caller's customer_id.

The Go backend uses Gin with a single `AuthMiddleware` that validates JWTs and sets `user_id`, `user_email`, and `user_role` in the Gin context. The JWT already includes `UserID` (which for customers is their `customer_id`) and `role = "customer"`. The key change is: the existing `dashboard` route group is guarded by admin-only logic by convention, not by code — there is no explicit role check. This must become explicit. A second middleware group (or the same group with a role gate) must enforce customer-scoped access.

The `customers` table exists and has auth columns from migration 011. The `devices` table has no `customer_id` — ownership tracking does not yet exist at the device level. The `proxy_connections` table already has a nullable `customer_id` column. Three new things are needed: (1) add `customer_id` to the `devices` table for ownership, (2) create a `device_shares` table for the sharing/permissions model, and (3) wire `customer_id` from the JWT into all repository queries as a WHERE filter for the customer-facing API surface.

**Primary recommendation:** Stamp `customer_id` onto the `devices` table (migration), add a `device_shares` table, write a `CustomerAuthMiddleware`, add `ListByCustomer` / `GetByIDForCustomer` repo methods, and reroute the existing device/connection list handlers to branch on role. The sharing permission model is a flag bitmask or a row-per-permission table on `device_shares`.

---

## Standard Stack

### Core (already in use — no new dependencies needed)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/gin-gonic/gin` | v1.x | HTTP routing + middleware | Already used; `c.Set`/`c.Get` for context propagation |
| `github.com/golang-jwt/jwt/v5` | v5.x | JWT parse/validate | Already used in `AuthService` and `CustomerAuthService` |
| `github.com/jackc/pgx/v5` | v5.x | PostgreSQL driver | Already used for all DB queries |
| `github.com/google/uuid` | v1.x | UUID type | Already used throughout domain models |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Standard `database/sql` patterns | — | Row-level filtering via parameterized SQL | All new `ListByCustomer` queries use `WHERE customer_id = $1` |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| App-level WHERE scoping | PostgreSQL RLS | RLS was explicitly excluded in STATE.md/REQUIREMENTS.md: "App-level WHERE customer_id scoping at this scale" |
| Single `device_shares` permission row per share | Bitmask column | Row per permission is more readable and individually toggleable; bitmask is more compact but requires bitwise ops in UI/backend |

**Installation:**
```bash
# No new packages needed — all dependencies already in go.mod
```

---

## Architecture Patterns

### Recommended Project Structure

New files and changes for this phase:

```
server/
├── migrations/
│   └── 012_tenant_isolation.up.sql     # customer_id on devices, device_shares table
├── internal/
│   ├── domain/
│   │   └── models.go                   # Add DeviceOwnership fields, DeviceShare model
│   ├── repository/
│   │   ├── device_repo.go              # Add ListByCustomer, GetByIDForCustomer
│   │   ├── connection_repo.go          # Add ListByCustomer
│   │   └── device_share_repo.go        # New: CRUD for device_shares
│   ├── service/
│   │   └── device_share_service.go     # New: share/unshare, permission checks
│   ├── api/
│   │   ├── middleware/
│   │   │   └── auth.go                 # Add CustomerAuthMiddleware, AdminOnlyMiddleware
│   │   └── handler/
│   │       ├── router.go               # Add customer route group; gate admin routes
│   │       ├── device_handler.go       # Branch on role: admin = unfiltered, customer = scoped
│   │       ├── connection_handler.go   # Branch on role
│   │       └── device_share_handler.go # New: share management endpoints
dashboard/
├── src/
│   ├── lib/
│   │   └── auth.ts                     # Add role helper (isAdmin, isCustomer)
│   ├── components/
│   │   └── dashboard/
│   │       └── Sidebar.tsx             # Hide admin-only nav items for customer role
│   └── app/
│       └── admin/
│           └── customers/              # New admin customer management page
```

### Pattern 1: Role-Aware Gin Middleware

**What:** A middleware that checks the `user_role` claim from the JWT. Admin requests get unfiltered DB access; customer requests have `customer_id` injected into the Gin context.

**When to use:** On every authenticated route that returns device/connection data.

**Example:**
```go
// server/internal/api/middleware/auth.go

// CustomerAuthMiddleware rejects requests where role != "customer".
// Sets "customer_id" (uuid.UUID) in gin context.
func CustomerAuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Assumes AuthMiddleware already ran and set user_id/user_role
        role, _ := c.Get("user_role")
        if role != "customer" {
            c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
            c.Abort()
            return
        }
        userID, _ := c.Get("user_id")
        c.Set("customer_id", userID.(uuid.UUID))
        c.Next()
    }
}

// AdminOnlyMiddleware rejects customers from admin-only routes.
func AdminOnlyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, _ := c.Get("user_role")
        if role != "admin" && role != "operator" {
            c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### Pattern 2: Role-Branching in Handlers

**What:** Each handler reads `user_role` from context. Admin → call unfiltered repo method. Customer → call `ListByCustomer(customerID)` repo method.

**When to use:** On `GET /devices`, `GET /connections`, and any endpoint returning data that must be tenant-scoped.

**Example:**
```go
func (h *DeviceHandler) List(c *gin.Context) {
    role, _ := c.Get("user_role")
    if role == "customer" {
        customerID, _ := c.Get("customer_id")
        devices, err := h.deviceService.ListByCustomer(c.Request.Context(), customerID.(uuid.UUID))
        // ...
        return
    }
    // Admin/operator: unfiltered
    devices, err := h.deviceService.List(c.Request.Context())
    // ...
}
```

### Pattern 3: Ownership Check for Write Operations

**What:** Before any mutating operation (update device name, delete connection, send command, rotate IP), verify the caller either owns the device or has the relevant share permission.

**When to use:** Every write endpoint when called by a customer role.

**Example:**
```go
func (h *DeviceHandler) Update(c *gin.Context) {
    role, _ := c.Get("user_role")
    if role == "customer" {
        customerID, _ := c.Get("customer_id")
        // Check ownership or "rename" share permission
        allowed, err := h.shareService.CanRename(c.Request.Context(), deviceID, customerID.(uuid.UUID))
        if !allowed {
            c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
            return
        }
    }
    // proceed with update
}
```

### Pattern 4: Stamping customer_id at Device Onboarding

**What:** The pairing code flow (`ClaimCode`) is the point where a new device is created. This is where `customer_id` must be stamped on the device. Currently `PairingCode` has a `CreatedBy *uuid.UUID` field pointing to the admin user who generated the code. For Phase 6, the pairing code needs to also carry the `customer_id` of the customer who will own the device.

**Decision point:** The customer generates their own pairing code from their portal (deferred to Phase 7 — QR self-service). For now, admin generates the pairing code on behalf of a customer, or the migration seeds existing devices with `customer_id` from a created operator account. For Phase 6: migration stamps existing devices → new operator/owner customer account. New devices registered without customer context → held in unclaimed pool OR still require admin to generate code for a specific customer.

**Practical approach for Phase 6:** Add `customer_id *uuid.UUID` to `pairing_codes` table. When admin creates a pairing code, they can specify a `customer_id`. When the device claims the code, `devices.customer_id` is set from `pairing_codes.customer_id`. This satisfies TENANT-02 (operator assigns device to customer via pairing code) while keeping self-service (Phase 7) deferred.

```sql
-- Migration 012
ALTER TABLE devices ADD COLUMN customer_id UUID REFERENCES customers(id) ON DELETE SET NULL;
ALTER TABLE pairing_codes ADD COLUMN customer_id UUID REFERENCES customers(id) ON DELETE SET NULL;
CREATE INDEX idx_devices_customer ON devices(customer_id);

CREATE TABLE device_shares (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id   UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    owner_id    UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    shared_with UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    can_rename  BOOLEAN NOT NULL DEFAULT FALSE,
    can_manage_ports BOOLEAN NOT NULL DEFAULT FALSE,
    can_download_configs BOOLEAN NOT NULL DEFAULT FALSE,
    can_rotate_ip BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (device_id, shared_with)
);
CREATE INDEX idx_device_shares_shared_with ON device_shares(shared_with);
CREATE INDEX idx_device_shares_device ON device_shares(device_id);
```

### Pattern 5: Customer Suspension

**What:** Admin sets `customers.active = false`. This should cause all API requests from that customer to be rejected at the middleware level, effectively taking all their devices offline and pausing all shares.

**How to implement:** In `CustomerAuthMiddleware`, after validating the JWT, load the customer record from DB and check `active`. If `false`, return 403 and abort. This is the simplest approach and requires no additional state beyond what already exists.

```go
// In CustomerAuthMiddleware, after setting customer_id:
customer, err := customerRepo.GetByID(ctx, customerID)
if err != nil || !customer.Active {
    c.JSON(http.StatusForbidden, gin.H{"error": "account suspended"})
    c.Abort()
    return
}
```

**Cost:** One extra DB query per customer request. At this scale (hundreds of requests/s at most) this is acceptable. Cache could be added later if needed.

### Anti-Patterns to Avoid

- **Relying on client-side filtering only:** The frontend must NOT be the only enforcement. Backend must filter by `customer_id` in every query.
- **Using the same `user_id` claim for both admin and customer identification:** The current `JWTClaims.UserID` is `uuid.UUID`. For admin tokens it points to `users.id`; for customer tokens it points to `customers.id`. This is already correct — the `role` claim distinguishes which table to look up.
- **Forgetting to scope `GET /connections?device_id=X`:** If a customer queries connections for a device they don't own, the handler must verify ownership/share before returning results.
- **Skipping the ownership check on single-resource GET:** `GET /devices/:id` and `GET /connections/:id` must also verify the caller has access to that specific resource — not just that they are authenticated.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JWT validation | Custom token parser | Existing `AuthService.ValidateToken` + `JWTClaims` | Already works, already trusted |
| Role-based access | Custom RBAC framework | Simple `role` string check in middleware + per-handler ownership query | At this scale, RBAC libs are overkill |
| Suspension check | Event/pub-sub system | Single DB lookup of `customers.active` in middleware | Simplest correct solution |
| Share permission storage | Custom ACL engine | `device_shares` table with boolean columns | Directly maps to UI toggles, easy to query |

**Key insight:** The pattern here is "app-level scoping via WHERE clauses" — the approach explicitly chosen in REQUIREMENTS.md. Every new repo method for customer access just adds `AND customer_id = $N` or joins through `device_shares`.

---

## Common Pitfalls

### Pitfall 1: Missing ownership check on per-resource GET

**What goes wrong:** A customer calls `GET /devices/some-other-customer-device-id` and receives full device data.

**Why it happens:** The list endpoint is scoped, but the by-ID endpoint uses `deviceRepo.GetByID(id)` with no ownership check.

**How to avoid:** Add `GetByIDForCustomer(ctx, deviceID, customerID)` that returns 404 if the customer neither owns nor has a share on the device. Use it in all `/:id` handlers when `role == "customer"`.

**Warning signs:** Success criteria #3 in the phase — "manually crafting a request with another customer's device ID returns 403."

### Pitfall 2: Forgetting connections inherit device ownership

**What goes wrong:** `GET /connections/:id` is not scoped. A customer guesses a connection ID belonging to another customer and retrieves credentials.

**Why it happens:** `proxy_connections` already has `customer_id` but it is nullable. If a connection was created before Phase 6, `customer_id` is NULL.

**How to avoid:** The migration must backfill `proxy_connections.customer_id` from the owning device's `customer_id`. After migration, a connection's `customer_id` can always be trusted. For the `GetByIDForCustomer` connection check, join through devices: `WHERE c.id = $1 AND (c.customer_id = $2 OR d.customer_id = $2)`.

**Warning signs:** Connections with NULL `customer_id` after migration.

### Pitfall 3: Admin routes become accessible to customers

**What goes wrong:** A customer token is used against `/api/customers`, `/api/pairing-codes`, `/api/relay-servers`, `/api/stats/overview`. These endpoints currently only check that a valid JWT exists — not that the role is admin.

**Why it happens:** The `dashboard` route group in `router.go` uses a single `AuthMiddleware` with no role check.

**How to avoid:** Add `AdminOnlyMiddleware()` to all admin-only routes. Apply it as a group middleware to: `/api/stats`, `/api/customers`, `/api/pairing-codes`, `/api/relay-servers`, `/api/rotation-links`, `/api/settings`.

**Warning signs:** Success criteria #3 — "manually crafting a request with another customer's device ID returns 403."

### Pitfall 4: Shared device visibility leaks device list to wrong customer

**What goes wrong:** `ListByCustomer` only returns owned devices. Shared devices are invisible. Customer cannot see a device shared with them.

**Why it happens:** Simple `WHERE customer_id = $1` misses the shares JOIN.

**How to avoid:** `ListByCustomer` query must UNION or JOIN `device_shares`:
```sql
SELECT d.* FROM devices d
WHERE d.customer_id = $1
UNION
SELECT d.* FROM devices d
INNER JOIN device_shares ds ON ds.device_id = d.id
WHERE ds.shared_with = $1
ORDER BY name ASC
```

### Pitfall 5: Suspension does not immediately invalidate active sessions

**What goes wrong:** Admin suspends a customer. Customer's in-flight JWT remains valid (24h expiry). They keep making requests successfully.

**Why it happens:** JWTs are stateless — once issued, they are valid until expiry.

**How to avoid:** Check `customers.active` on every customer request in `CustomerAuthMiddleware` (Pattern 5 above). The DB lookup is the live authority; the JWT is only an authentication token, not an authorization cache.

### Pitfall 6: Migration leaves orphan devices with no customer_id

**What goes wrong:** After migration, devices without a `customer_id` are invisible to all customers but still function. Admin cannot identify who owns them. Context says "devices without a valid customer_id are rejected at the tunnel connection level."

**Why it happens:** Existing devices in the database were created before multi-tenancy existed.

**How to avoid:** Migration strategy: create a new `Customer` record for the current operator (email from the existing admin user), then `UPDATE devices SET customer_id = <new_customer_id>`. This satisfies the CONTEXT.md decision: "Existing devices migrate to a new customer account for the current operator/owner (not to the admin account)."

---

## Code Examples

### New SQL: ListByCustomer (owned + shared)

```sql
-- server/internal/repository/device_repo.go
-- ListByCustomer returns all devices a customer owns or has been shared with.
SELECT d.id, d.name, d.description, d.android_id, d.status,
       COALESCE(host(d.cellular_ip),'') as cellular_ip,
       COALESCE(host(d.wifi_ip),'') as wifi_ip,
       COALESCE(host(d.vpn_ip),'') as vpn_ip,
       d.carrier, d.network_type, d.battery_level, d.battery_charging, d.signal_strength,
       d.base_port, d.http_port, d.socks5_port, d.udp_relay_port, d.ovpn_port,
       d.last_heartbeat, d.app_version, d.device_model, d.android_version,
       d.relay_server_id, COALESCE(rs.ip, '') as relay_server_ip,
       d.auto_rotate_minutes, d.created_at, d.updated_at
FROM devices d
LEFT JOIN relay_servers rs ON d.relay_server_id = rs.id
WHERE d.customer_id = $1
UNION
SELECT d.id, d.name, d.description, d.android_id, d.status,
       COALESCE(host(d.cellular_ip),'') as cellular_ip,
       COALESCE(host(d.wifi_ip),'') as wifi_ip,
       COALESCE(host(d.vpn_ip),'') as vpn_ip,
       d.carrier, d.network_type, d.battery_level, d.battery_charging, d.signal_strength,
       d.base_port, d.http_port, d.socks5_port, d.udp_relay_port, d.ovpn_port,
       d.last_heartbeat, d.app_version, d.device_model, d.android_version,
       d.relay_server_id, COALESCE(rs.ip, '') as relay_server_ip,
       d.auto_rotate_minutes, d.created_at, d.updated_at
FROM devices d
LEFT JOIN relay_servers rs ON d.relay_server_id = rs.id
INNER JOIN device_shares ds ON ds.device_id = d.id
WHERE ds.shared_with = $1
ORDER BY name ASC
```

### New domain model: DeviceShare

```go
// server/internal/domain/models.go

type DeviceShare struct {
    ID               uuid.UUID `json:"id" db:"id"`
    DeviceID         uuid.UUID `json:"device_id" db:"device_id"`
    OwnerID          uuid.UUID `json:"owner_id" db:"owner_id"`
    SharedWith       uuid.UUID `json:"shared_with" db:"shared_with"`
    CanRename        bool      `json:"can_rename" db:"can_rename"`
    CanManagePorts   bool      `json:"can_manage_ports" db:"can_manage_ports"`
    CanDownloadConfigs bool    `json:"can_download_configs" db:"can_download_configs"`
    CanRotateIP      bool      `json:"can_rotate_ip" db:"can_rotate_ip"`
    CreatedAt        time.Time `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}
```

### Permission check helper

```go
// server/internal/service/device_share_service.go

// CanAccess returns true if customerID owns the device or has any share on it.
func (s *DeviceShareService) CanAccess(ctx context.Context, deviceID, customerID uuid.UUID) (bool, error) {
    device, err := s.deviceRepo.GetByID(ctx, deviceID)
    if err != nil {
        return false, err
    }
    if device.CustomerID != nil && *device.CustomerID == customerID {
        return true, nil // owner
    }
    share, err := s.shareRepo.GetByDeviceAndCustomer(ctx, deviceID, customerID)
    return share != nil && err == nil, nil
}

// CanDo checks a specific permission bit. Owners can always do everything.
func (s *DeviceShareService) CanDo(ctx context.Context, deviceID, customerID uuid.UUID, perm string) (bool, error) {
    device, err := s.deviceRepo.GetByID(ctx, deviceID)
    if err != nil {
        return false, err
    }
    if device.CustomerID != nil && *device.CustomerID == customerID {
        return true, nil // owner has all permissions
    }
    share, err := s.shareRepo.GetByDeviceAndCustomer(ctx, deviceID, customerID)
    if err != nil || share == nil {
        return false, nil
    }
    switch perm {
    case "rename":        return share.CanRename, nil
    case "manage_ports":  return share.CanManagePorts, nil
    case "download_configs": return share.CanDownloadConfigs, nil
    case "rotate_ip":     return share.CanRotateIP, nil
    }
    return false, nil
}
```

### Frontend: Role-aware auth helper

```typescript
// dashboard/src/lib/auth.ts

export function getRole(): string | null {
  const user = getUser()
  return user ? user.role : null
}

export function isAdmin(): boolean {
  const role = getRole()
  return role === 'admin' || role === 'operator'
}

export function isCustomer(): boolean {
  return getRole() === 'customer'
}
```

### Frontend: Sidebar role gating

```typescript
// dashboard/src/components/dashboard/Sidebar.tsx (additions)

const adminNavItems = [
  { href: '/overview', label: 'Overview', icon: BarChart },
  { href: '/devices', label: 'Devices', icon: Smartphone },
  { href: '/connections', label: 'Connections', icon: Network },
  { href: '/admin/customers', label: 'Customers', icon: Users },
]

const customerNavItems = [
  { href: '/devices', label: 'Devices', icon: Smartphone },
]

// In component:
const navItems = isAdmin() ? adminNavItems : customerNavItems
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| No tenant isolation — all authenticated users see all data | WHERE customer_id scoping in every repo query | Phase 6 | Fundamental SaaS requirement |
| Single admin JWT flow | Two JWT paths (admin in `users`, customer in `customers`) | Phase 5 | Already implemented — role claim already present |
| Devices have no owner | `devices.customer_id` FK to customers | Phase 6 (migration 012) | Enables row-level scoping |
| Single `AuthMiddleware` | `AuthMiddleware` + `CustomerAuthMiddleware` + `AdminOnlyMiddleware` | Phase 6 | Enables per-role enforcement |

**Current state confirmed from code:**
- `JWTClaims` (in `auth_service.go`) has `UserID`, `Email`, `Role` — `role = "customer"` already emitted by `generateCustomerJWT`
- `proxy_connections.customer_id` already nullable UUID FK — exists, partially used
- `devices` table has NO `customer_id` column — must add
- `device_shares` table does NOT exist — must create
- `pairing_codes` has NO `customer_id` — must add for admin-assigns-device-to-customer flow
- All dashboard routes are under one `AuthMiddleware` group — no role separation yet
- `CustomerHandler` exists but it's a thin CRUD handler; the admin customer management page is not yet built

---

## Open Questions

1. **Who generates pairing codes for Phase 6?**
   - What we know: Deferred items say "QR self-service belongs in Phase 7." But TENANT-02 requires "operator can assign devices to customers."
   - What's unclear: Does Phase 6 need the admin to pick a customer when generating a pairing code, or is it OK for existing devices to just be migrated to the operator account and new pairing codes to remain unassigned until Phase 7?
   - Recommendation: Phase 6 approach: (a) migration stamps all existing devices with a new operator-owned customer account, (b) admin UI adds an optional "assign to customer" dropdown on the pairing code creation form, (c) device registration via `ClaimCode` stamps `customer_id` from the pairing code onto the new device. This fully satisfies TENANT-02 without needing Phase 7 self-service.

2. **What happens to connections with NULL customer_id after migration?**
   - What we know: `proxy_connections.customer_id` has been nullable since migration 001. Some connections may have NULL.
   - What's unclear: Should the migration backfill them from the device's new `customer_id`, or leave them NULL?
   - Recommendation: Backfill in the migration: `UPDATE proxy_connections pc SET customer_id = d.customer_id FROM devices d WHERE pc.device_id = d.id AND pc.customer_id IS NULL`. This makes the data consistent.

3. **Does the `DeviceService` need to be split into admin/customer variants?**
   - What we know: `DeviceService.List()` returns all devices. Adding `ListByCustomer()` as a new method is cleaner than changing the existing one.
   - What's unclear: Should the handler branch (`if role == "customer"`) or should there be a `CustomerDeviceService` wrapper?
   - Recommendation: Branch in the handler (Pattern 2 above). A wrapper service is over-engineering for the current scope. The branching is a 3-line `if` block per handler — simple and explicit.

---

## Sources

### Primary (HIGH confidence)
- Direct codebase inspection: `server/internal/api/middleware/auth.go` — current middleware pattern
- Direct codebase inspection: `server/internal/service/auth_service.go` — `JWTClaims` struct
- Direct codebase inspection: `server/internal/service/customer_auth_service.go` — `generateCustomerJWT` with `role = "customer"`
- Direct codebase inspection: `server/internal/domain/models.go` — all existing model fields
- Direct codebase inspection: `server/internal/repository/device_repo.go` — current device query pattern
- Direct codebase inspection: `server/migrations/001_initial.up.sql` — full schema baseline
- Direct codebase inspection: `server/migrations/011_customer_auth.up.sql` — Phase 5 additions
- Direct codebase inspection: `server/internal/api/handler/router.go` — all routes and middleware groups

### Secondary (MEDIUM confidence)
- Pattern: "check `active` flag in middleware per request for suspension" — standard stateless JWT with live DB check. Well-established pattern for SaaS tenant suspension without token revocation infrastructure.
- UNION query pattern for "owned OR shared" device list — standard SQL, no external verification needed.

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new libraries; all patterns use existing codebase components
- Architecture: HIGH — based on direct codebase reading; patterns match existing conventions
- Pitfalls: HIGH — derived from code inspection of exact endpoints that lack role checks
- Schema design: HIGH — directly constrained by existing migrations and models

**Research date:** 2026-02-28
**Valid until:** 2026-03-30 (stable domain — schema patterns don't change rapidly)
