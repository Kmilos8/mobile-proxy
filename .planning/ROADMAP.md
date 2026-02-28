# Roadmap: PocketProxy

## Overview

PocketProxy has a working infrastructure (Go backend, Android tunnel, PostgreSQL, HTTP/SOCKS5 proxies, Docker Compose) but two things block production use: OpenVPN throughput is effectively broken, and the dashboard is a scaffold with no functional UI. This roadmap fixes the blocker first, builds the operator interface second, then hardens security and adds monitoring before exposing the product to customers.

v2.0 transforms PocketProxy from an operator-only tool into a customer-facing SaaS platform with self-signup, multi-tenant access, and a public landing page.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

### v1.0 Phases (Complete)

- [x] **Phase 1: OpenVPN Throughput** - Fix the broken customer VPN path so all three protocols work at usable speed
- [x] **Phase 2: Dashboard** - Build the operator UI so devices and proxy ports can be managed without hitting the API directly (completed 2026-02-26)
- [x] **Phase 3: Security and Monitoring** - Harden credentials, enforce bandwidth limits, and alert on device offline before customer exposure (completed 2026-02-27)
- [x] **Phase 4: Bug Fixes and Polish** - Close audit gaps, fix OpenVPN bug, add search/auto-rotation column/connection ID (completed 2026-02-27)

### v2.0 Phases (Current Milestone)

- [ ] **Phase 5: Auth Foundation** - Enable customer self-signup with email/password, Google OAuth, email verification, password reset, and Turnstile bot protection
- [ ] **Phase 6: Multi-Tenant Isolation** - Scope all data access by customer_id so each customer sees only their own devices and connections
- [ ] **Phase 7: Customer Portal** - Give customers a self-service portal to view credentials, rotate IPs, download .ovpn configs, and see bandwidth usage
- [ ] **Phase 8: Landing Page and IP Whitelist** - Ship the public marketing page and server-side CIDR-based IP whitelist auth per proxy port
- [ ] **Phase 9: Device Grouping, API Docs, and Traffic Logs** - Add operator bulk tools (device groups, bulk rotation), public Swagger UI, and per-port traffic history for customers

## Phase Details

### Phase 1: OpenVPN Throughput
**Goal**: Customers can connect via .ovpn file and browse the web through a device's cellular connection at usable speed
**Depends on**: Nothing (first phase)
**Requirements**: PROTO-01, PROTO-02
**Success Criteria** (what must be TRUE):
  1. A customer importing a generated .ovpn file can fully load a webpage through the VPN tunnel
  2. A speed test run through the OpenVPN connection completes and reports measurable throughput (not timeout)
  3. HTTP and SOCKS5 proxies continue to respond correctly while OpenVPN config changes are applied
  4. The .ovpn download from the dashboard produces a working config without manual edits
**Plans**: 2 plans

Plans:
- [x] 01-01-PLAN.md — Apply OpenVPN performance tuning: reduce peekTimeout to 200ms, switch sndbuf/rcvbuf to OS autotuning, add fast-io
- [x] 01-02-PLAN.md — Fix client-connect-ovpn.sh silent failure bug and verify HTTP/SOCKS5 DNAT isolation

### Phase 2: Dashboard
**Goal**: An operator can manage their entire device fleet and proxy port inventory from the dashboard without touching the API
**Depends on**: Phase 1
**Requirements**: DASH-01, DASH-02, DASH-03, DASH-04
**Success Criteria** (what must be TRUE):
  1. Operator can see all devices with online/offline status, current IP, carrier, battery, and signal strength on one screen
  2. Operator can create a new proxy connection (HTTP, SOCKS5, or OpenVPN) from the dashboard and immediately see credentials
  3. Operator can view a connection detail page showing host, port, username, password, and .ovpn download with one-click copy for each field
  4. Dashboard layout is usable on desktop and tablet (no horizontal scroll, no broken layouts at 768px+)
  5. Operator can delete a proxy connection and the port is freed
**Plans**: 3 plans

Plans:
- [x] 02-01-PLAN.md — Install shadcn/ui, dark theme, collapsible sidebar, device table home page, backend openvpn proxy_type
- [x] 02-02-PLAN.md — Device detail page with connection management: create/view/copy/delete connections
- [x] 02-03-PLAN.md — Visual and functional verification checkpoint

### Phase 3: Security and Monitoring
**Goal**: Credentials are not stored or transmitted in plaintext, bandwidth limits are enforced, and operators are alerted when devices go offline
**Depends on**: Phase 2
**Requirements**: SEC-01, MON-01, MON-02
**Success Criteria** (what must be TRUE):
  1. The PasswordPlain field is no longer populated in the database; OpenVPN auth uses bcrypt comparison
  2. A proxy connection with a 1 GB bandwidth limit stops passing traffic after 1 GB is consumed
  3. An operator receives an email or webhook notification within 5 minutes of a device going offline
**Plans**: 2 plans

Plans:
- [ ] 03-01-PLAN.md — Swap OpenVPN auth to bcrypt, add regenerate-password endpoint, migration SQL for webhook/alerting columns
- [ ] 03-02-PLAN.md — Bandwidth enforcement in tunnel server, offline webhook dispatch, dashboard settings/monitoring UI

### Phase 4: Bug Fixes and Polish
**Goal**: Close audit gaps, fix OpenVPN creation bug, and add dashboard improvements (search, auto-rotation column, connection ID)
**Depends on**: Phase 3
**Requirements**: MON-01, MON-02, DASH-02
**Gap Closure**: Closes gaps from v1.0 milestone audit + user-reported bugs
**Success Criteria** (what must be TRUE):
  1. When a device reconnects after being offline, a recovery webhook POST is delivered to the operator's configured URL
  2. Clicking Reset Usage in the dashboard resets both the DB value and the tunnel's in-memory counter — usage does not reappear after the next 30s flush
  3. Creating an OpenVPN config from the OpenVPN tab succeeds without "must be http or socks5" error
  4. The devices page has a search bar that filters devices by name
  5. The device table shows an auto-rotation column indicating whether auto-rotation is enabled on each device
  6. Every connection has a visible connection ID assigned at creation, shown in the connection table
**Plans**: 2 plans

Plans:
- [ ] 04-01-PLAN.md — Wire recovery webhook, propagate bandwidth reset to tunnel, add OpenVPN to Add Connection modal
- [ ] 04-02-PLAN.md — Add device search bar, auto-rotation column, and connection ID column to dashboard

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
- [ ] 05-01-PLAN.md — Database migration, domain models, and repositories for customer auth
- [ ] 05-02-PLAN.md — Backend auth services, handlers, and route wiring (signup, login, verify, reset, Google OAuth, Turnstile)
- [ ] 05-03-PLAN.md — Frontend auth pages (login extension, signup, verify email, forgot/reset password, Turnstile widget)

### Phase 6: Multi-Tenant Isolation
**Goal**: Every customer sees only their own assigned devices and connections; no cross-customer data is accessible through any portal endpoint
**Depends on**: Phase 5
**Requirements**: TENANT-01, TENANT-02, TENANT-03
**Success Criteria** (what must be TRUE):
  1. Logging in as Customer A returns only Customer A's devices and connections — Customer B's data does not appear in any portal response
  2. An operator can assign a device to a specific customer from the admin dashboard, and that device then appears in the customer's portal
  3. All customer portal API responses are filtered by the customer_id embedded in the JWT; manually crafting a request with another customer's device ID returns 403
**Plans**: TBD

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
v1.0 phases execute in numeric order: 1 -> 2 -> 3 -> 4
v2.0 phases execute in numeric order: 5 -> 6 -> 7 -> 8 -> 9

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. OpenVPN Throughput | 2/2 | Complete | 2026-02-26 |
| 2. Dashboard | 3/3 | Complete | 2026-02-26 |
| 3. Security and Monitoring | 2/2 | Complete | 2026-02-27 |
| 4. Bug Fixes and Polish | 2/2 | Complete | 2026-02-27 |
| 5. Auth Foundation | 0/3 | Planned | - |
| 6. Multi-Tenant Isolation | 0/TBD | Not started | - |
| 7. Customer Portal | 0/TBD | Not started | - |
| 8. Landing Page and IP Whitelist | 0/TBD | Not started | - |
| 9. Device Grouping, API Docs, and Traffic Logs | 0/TBD | Not started | - |
