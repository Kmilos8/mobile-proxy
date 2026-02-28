# Phase 6: Multi-Tenant Isolation - Context

**Gathered:** 2026-02-28
**Status:** Ready for planning

<domain>
## Phase Boundary

Scope all data access by customer_id so each customer sees only their own devices and connections. Customers own and onboard their own devices. Device owners can share devices with other customers, granting granular permissions. Admin is a separate system account with full visibility but no device ownership. No cross-customer data is accessible through any portal endpoint.

Note: This phase expands beyond the original TENANT-01/02/03 requirements to include the device sharing/permissions model, as decided during context gathering.

</domain>

<decisions>
## Implementation Decisions

### Device Assignment Model
- Customers self-onboard their own devices (not operator-assigned)
- Ownership is at the device level — one device = one owner, all connections inherit the owner's customer_id
- Ownership is permanent — no transfers between customers
- Admin can see and manage all devices across all customers but cannot reassign ownership

### Device Sharing & Permissions
- Device owner can share their device with other customers
- Sharing grants granular, toggleable permissions:
  - **Rename** — change device name/description
  - **Manage ports** — add/remove proxy connections
  - **Download configs** — generate and download .ovpn files
  - **Rotate IP + view usage** — trigger IP rotation and view bandwidth stats
- Baseline: shared user can always view the device and its connections
- Owner can revoke access or change individual permissions at any time
- Changes take effect immediately

### Admin Dashboard Changes
- Add customer filter/dropdown to existing device list (filter by customer or view all)
- New customer management page listing all customers
- Customer detail shows: email, signup date, verification status, device count, active shares, last login, total bandwidth, account status (active/suspended)
- Admin can suspend a customer account — all their devices go offline and all outgoing shares are paused until reactivated
- Admin cannot reassign device ownership — view only + disable/suspend

### Account & Role Model
- Admin is a separate system account (god-mode oversight, does not own devices in the customer sense)
- Customer accounts are the standard account type — they own and manage devices
- Existing devices migrate to a new customer account for the current operator/owner (not to the admin account)
- Devices without a valid customer_id are rejected at the tunnel connection level (hard enforcement)

### Role Separation in Dashboard
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

</decisions>

<specifics>
## Specific Ideas

- "Whoever made the connection and paid for it owns it" — ownership follows the person who created and paid for the connection
- Customer wants full control over who can do what with their devices — owner is the authority
- Admin god-mode should be for oversight and support, not for moving devices around between customers
- The sharing model is designed so that a future customer portal (Phase 7) can expose share management UI

</specifics>

<deferred>
## Deferred Ideas

- **Customer self-service QR code generation** — Customer generates QR from their portal to onboard a device. Belongs in Phase 7 (Customer Portal).
- **Balance/payment on connection creation** — Connections require payment from pre-loaded balance. Beyond v2.0 scope (payment handled externally via Stripe per requirements).
- **Device transfer between customers** — Ownership is permanent for now. Could be added as a future capability if needed.

</deferred>

---

*Phase: 06-multi-tenant-isolation*
*Context gathered: 2026-02-28*
