# Roadmap: PocketProxy

## Milestones

- ✅ **v1.0 MVP** — Phases 1-4 (shipped 2026-02-28)
- 🚧 **v2.0 SaaS Platform** — Phases 5-9 (in progress)

## Phases

<details>
<summary>✅ v1.0 MVP (Phases 1-4) — SHIPPED 2026-02-28</summary>

- [x] Phase 1: OpenVPN Throughput (2/2 plans) — completed 2026-02-26
- [x] Phase 2: Dashboard (3/3 plans) — completed 2026-02-26
- [x] Phase 3: Security and Monitoring (2/2 plans) — completed 2026-02-27
- [x] Phase 4: Bug Fixes and Polish (2/2 plans) — completed 2026-02-27

See: `.planning/milestones/v1.0-ROADMAP.md` for full details.

</details>

### 🚧 v2.0 SaaS Platform (In Progress)

- [x] **Phase 5: Auth Foundation** - Enable customer self-signup with email/password, Google OAuth, email verification, password reset, and Turnstile bot protection (completed 2026-02-28)
- [ ] **Phase 6: Multi-Tenant Isolation** - Scope all data access by customer_id so each customer sees only their own devices and connections
- [ ] **Phase 7: Customer Portal** - Give customers a self-service portal to view credentials, rotate IPs, download .ovpn configs, and see bandwidth usage
- [ ] **Phase 8: Landing Page and IP Whitelist** - Ship the public marketing page and server-side CIDR-based IP whitelist auth per proxy port
- [ ] **Phase 9: Device Grouping, API Docs, and Traffic Logs** - Add operator bulk tools (device groups, bulk rotation), public Swagger UI, and per-port traffic history for customers

## Phase Details

### Phase 5: Auth Foundation
**Goal**: Customers can self-register and log in securely using email/password or Google, with email verification and bot protection on all public forms
**Depends on**: Phase 4
**Requirements**: AUTH-01, AUTH-02, AUTH-03, AUTH-04, AUTH-05
**Success Criteria** (what must be TRUE):
  1. A new customer can create an account with email and password and receives a verification email; they cannot access the portal until they click verify
  2. A customer who signs up with a Google account is logged in immediately with no separate verification step required
  3. A customer who forgets their password can reset it via an emailed link within 24 hours of the link being issued
  4. The signup and login forms display a Cloudflare Turnstile challenge and the Go backend rejects requests where the Turnstile token fails server-side verification
**Plans**: 3 plans

Plans:
- [x] 05-01-PLAN.md — Database migration, domain models, and repositories for customer auth
- [x] 05-02-PLAN.md — Backend auth services, handlers, and route wiring (signup, login, verify, reset, Google OAuth, Turnstile)
- [x] 05-03-PLAN.md — Frontend auth pages (login extension, signup, verify email, forgot/reset password, Turnstile widget)

### Phase 6: Multi-Tenant Isolation
**Goal**: Every customer sees only their own assigned devices and connections; no cross-customer data is accessible through any portal endpoint
**Depends on**: Phase 5
**Requirements**: TENANT-01, TENANT-02, TENANT-03
**Success Criteria** (what must be TRUE):
  1. Logging in as Customer A returns only Customer A's devices and connections — Customer B's data does not appear in any portal response
  2. An operator can assign a device to a specific customer from the admin dashboard, and that device then appears in the customer's portal
  3. All customer portal API responses are filtered by the customer_id embedded in the JWT; manually crafting a request with another customer's device ID returns 403
**Plans**: 3 plans

Plans:
- [ ] 06-01-PLAN.md — Database migration, domain models, and customer-scoped repository methods
- [ ] 06-02-PLAN.md — Backend middleware, role-branching handlers, device share service, route wiring
- [ ] 06-03-PLAN.md — Frontend role-aware UI (sidebar gating, admin customer management, pairing code assignment)

### Phase 7: Customer Portal
**Goal**: Customers can manage their assigned proxies end-to-end through a self-service portal without contacting the operator
**Depends on**: Phase 6
**Requirements**: PORTAL-01, PORTAL-02, PORTAL-03, PORTAL-04
**Success Criteria** (what must be TRUE):
  1. A logged-in customer can see all their assigned devices and the proxy credentials (host, port, username, password) for each connection
  2. A customer can download the .ovpn config file for any of their OpenVPN connections directly from the portal
  3. A customer can trigger an IP rotation on one of their connections from the portal and see the new IP in the device's IP history
  4. A customer can view the current bandwidth usage for each of their connections in the portal
**Plans**: TBD

### Phase 8: Landing Page and IP Whitelist
**Goal**: The product has a public-facing marketing page and proxy connections support CIDR-based IP authentication as an alternative to username/password
**Depends on**: Phase 5 (landing page is independent of auth; IP whitelist depends on Phase 6 for operator access)
**Requirements**: LAND-01, LAND-02, IPWL-01, IPWL-02, IPWL-03
**Success Criteria** (what must be TRUE):
  1. Visiting the root URL shows a marketing page with a hero section, feature highlights, and a signup call-to-action button that links to the signup flow
  2. The landing page renders without horizontal scroll and all content is readable on a 375px mobile viewport
  3. An operator can add one or more IP addresses or CIDR ranges to a connection's whitelist from the connection detail page
  4. A client connecting from a whitelisted IP can use the proxy without supplying username or password credentials
  5. A client connecting from a non-whitelisted IP is rejected at the server using proper CIDR parsing, not string comparison
**Plans**: TBD

### Phase 9: Device Grouping, API Docs, and Traffic Logs
**Goal**: Operators can manage devices in bulk using named groups, all API endpoints are publicly documented with a live Swagger UI, and customers can view per-connection traffic history
**Depends on**: Phase 7 (traffic logs require customer portal to exist; groups and docs are independent)
**Requirements**: GROUP-01, GROUP-02, GROUP-03, APIDOC-01, APIDOC-02, TLOG-01, TLOG-02
**Success Criteria** (what must be TRUE):
  1. An operator can create a named device group, add devices to it, and bulk-rotate IPs for all devices in the group with a single action
  2. An operator can bulk enable or disable auto-rotation for all devices in a group from one control
  3. Visiting /swagger/ shows a live Swagger UI with all API endpoints listed, request/response schemas visible, and the ability to make authenticated test requests
  4. A logged-in customer can view a traffic history chart for any of their connections, filtered by 24h, 7d, or 30d time ranges
  5. Traffic log records older than 30 days are automatically deleted; the retention policy runs without manual intervention
**Plans**: TBD

## Progress

**Execution Order:**
v1.0 phases: 1 → 2 → 3 → 4 (complete)
v2.0 phases: 5 → 6 → 7 → 8 → 9

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. OpenVPN Throughput | v1.0 | 2/2 | Complete | 2026-02-26 |
| 2. Dashboard | v1.0 | 3/3 | Complete | 2026-02-26 |
| 3. Security and Monitoring | v1.0 | 2/2 | Complete | 2026-02-27 |
| 4. Bug Fixes and Polish | v1.0 | 2/2 | Complete | 2026-02-27 |
| 5. Auth Foundation | v2.0 | 3/3 | Complete | 2026-02-28 |
| 6. Multi-Tenant Isolation | v2.0 | 0/3 | Not started | - |
| 7. Customer Portal | v2.0 | 0/TBD | Not started | - |
| 8. Landing Page and IP Whitelist | v2.0 | 0/TBD | Not started | - |
| 9. Device Grouping, API Docs, and Traffic Logs | v2.0 | 0/TBD | Not started | - |
