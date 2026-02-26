# Roadmap: PocketProxy

## Overview

PocketProxy has a working infrastructure (Go backend, Android tunnel, PostgreSQL, HTTP/SOCKS5 proxies, Docker Compose) but two things block production use: OpenVPN throughput is effectively broken, and the dashboard is a scaffold with no functional UI. This roadmap fixes the blocker first, builds the operator interface second, then hardens security and adds monitoring before exposing the product to customers.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: OpenVPN Throughput** - Fix the broken customer VPN path so all three protocols work at usable speed
- [ ] **Phase 2: Dashboard** - Build the operator UI so devices and proxy ports can be managed without hitting the API directly
- [ ] **Phase 3: Security and Monitoring** - Harden credentials, enforce bandwidth limits, and alert on device offline before customer exposure

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
**Plans**: TBD

Plans:
- [ ] 02-01: Initialize shadcn/ui and TanStack Query; rebuild device list page with live status cards
- [ ] 02-02: Build connection creation dialog (react-hook-form + zod) and connection detail page
- [ ] 02-03: Apply responsive layout across all pages; add one-click credential copy

### Phase 3: Security and Monitoring
**Goal**: Credentials are not stored or transmitted in plaintext, bandwidth limits are enforced, and operators are alerted when devices go offline
**Depends on**: Phase 2
**Requirements**: SEC-01, MON-01, MON-02
**Success Criteria** (what must be TRUE):
  1. The PasswordPlain field is no longer populated in the database; OpenVPN auth uses bcrypt comparison
  2. A proxy connection with a 1 GB bandwidth limit stops passing traffic after 1 GB is consumed
  3. An operator receives an email or webhook notification within 5 minutes of a device going offline
**Plans**: TBD

Plans:
- [ ] 03-01: Remove PasswordPlain storage; migrate existing records; update openvpn_handler.go auth to use bcrypt.CompareHashAndPassword
- [ ] 03-02: Enforce bandwidth limit in the proxy routing layer; implement device offline notification via email or webhook

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. OpenVPN Throughput | 2/2 | Complete | 2026-02-26 |
| 2. Dashboard | 0/3 | Not started | - |
| 3. Security and Monitoring | 0/2 | Not started | - |
