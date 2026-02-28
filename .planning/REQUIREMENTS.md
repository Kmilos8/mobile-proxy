# Requirements: PocketProxy

**Defined:** 2026-02-25
**Core Value:** Customers can reliably route traffic through real mobile devices via HTTP, SOCKS5, or OpenVPN, managed through a clean dashboard.

## v1 Requirements (Complete)

### Protocol Stability

- [x] **PROTO-01**: OpenVPN direct access delivers usable throughput (pages load fully, speed tests complete)
- [x] **PROTO-02**: HTTP and SOCKS5 proxies remain stable under production load

### Dashboard Redesign

- [x] **DASH-01**: Full UI redesign with modern component library (shadcn/ui) across all existing pages
- [x] **DASH-02**: Connection creation UI on dashboard (currently API-only)
- [x] **DASH-03**: Connection detail page for viewing/managing individual proxy ports
- [x] **DASH-04**: Responsive layout for desktop and tablet

### Security

- [x] **SEC-01**: Replace plaintext proxy password storage with secure approach (bcrypt hash exists but PasswordPlain also stored and used by OpenVPN auth)

### Monitoring

- [x] **MON-01**: Device offline notification via email or webhook
- [x] **MON-02**: Enforce bandwidth limits per connection (field exists but not enforced)

## v2 Requirements

Requirements for v2.0 SaaS Platform milestone. Each maps to roadmap phases.

### Authentication & Signup

- [ ] **AUTH-01**: Customer can sign up with email and password
- [ ] **AUTH-02**: Customer receives email verification after signup and must verify before accessing portal
- [ ] **AUTH-03**: Customer can sign up and log in with Google account (OAuth)
- [ ] **AUTH-04**: Customer can reset password via email link
- [ ] **AUTH-05**: Signup and login forms are protected by Cloudflare Turnstile bot protection

### Multi-Tenant Access

- [ ] **TENANT-01**: Customer sees only their own assigned devices and connections
- [ ] **TENANT-02**: Operator (admin) can assign devices to customers
- [ ] **TENANT-03**: Customer JWT carries customer_id; all portal queries filter by it

### Customer Portal

- [ ] **PORTAL-01**: Customer can view their assigned devices and connection credentials
- [ ] **PORTAL-02**: Customer can download .ovpn config files for their connections
- [ ] **PORTAL-03**: Customer can trigger IP rotation on their connections
- [ ] **PORTAL-04**: Customer can view bandwidth usage for their connections

### Landing Page

- [ ] **LAND-01**: Public marketing page with hero section, feature highlights, and signup CTA
- [ ] **LAND-02**: Landing page is mobile responsive and matches dashboard design system

### IP Whitelist

- [ ] **IPWL-01**: Operator can add IP addresses/CIDRs to a connection's whitelist
- [ ] **IPWL-02**: Whitelisted IPs can use the proxy without username/password credentials
- [ ] **IPWL-03**: Whitelist is enforced server-side using proper CIDR parsing

### API Documentation

- [ ] **APIDOC-01**: All API endpoints are documented with Swagger/OpenAPI spec
- [ ] **APIDOC-02**: Interactive Swagger UI is accessible at /swagger/

### Device Grouping

- [ ] **GROUP-01**: Operator can create named device groups and add/remove devices
- [ ] **GROUP-02**: Operator can bulk rotate IPs for all devices in a group
- [ ] **GROUP-03**: Operator can bulk enable/disable auto-rotation for a group

### Traffic Logs

- [ ] **TLOG-01**: Customer can view per-connection traffic history with time-range filtering
- [ ] **TLOG-02**: Traffic logs have a 30-day retention policy enforced automatically

## Already Working (Keep As-Is)

These features are implemented and functional. Do not modify unless explicitly requested.

- IP rotation from dashboard (manual + auto-rotation with configurable interval)
- Rotation links (public URL tokens for external tools)
- IP history (last 50 entries per device)
- Device management (name, description, commands, status tracking)
- QR code device onboarding with pairing codes
- Multiple proxy ports per device (auto-allocated)
- Bandwidth tracking (per device, hourly charts, monthly totals)
- Device metrics (battery, signal strength, carrier, network type)
- .ovpn config file download per connection
- Proxy connection CRUD via API
- Device online/offline status with 2-minute stale detection
- WebSocket real-time updates
- JWT authentication for dashboard
- Admin user passwords stored as bcrypt (secure)
- Offline/recovery webhooks
- Bandwidth enforcement per connection

## Deferred (v2.1+)

### White-Label

- **WL-01**: Custom logo upload per operator
- **WL-02**: Custom brand colors (primary, accent) per operator
- **WL-03**: Custom page title / brand name per operator

### Advanced Features

- **ADV-01**: Per-port traffic logs downloadable as CSV/export
- **ADV-02**: Customer self-service connection creation (within operator-set limits)
- **ADV-03**: Custom domain support for white-label (CNAME)

## Out of Scope

| Feature | Reason |
|---------|--------|
| iOS app | iOS background restrictions prevent reliable proxy serving |
| Payment/billing integration | Handle externally (Stripe Checkout) until billing milestone |
| Telegram bot | Dashboard + API + rotation links cover all use cases |
| Real-time traffic interception | Legal/privacy liability, storage costs |
| Shared IP pools | Defeats core value of private mobile proxies |
| Geolocation targeting | Architecture assumes specific devices, not location pools |
| SMS forwarding | Legal risk; requires deliberate TOS design |
| Passive OS fingerprint spoofing | High complexity, limited audience |
| NextAuth.js | Conflicts with Go JWT auth system — Go backend handles all auth |
| PostgreSQL RLS | App-level scoping simpler at this scale |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| AUTH-01 | Phase 5 | Pending |
| AUTH-02 | Phase 5 | Pending |
| AUTH-03 | Phase 5 | Pending |
| AUTH-04 | Phase 5 | Pending |
| AUTH-05 | Phase 5 | Pending |
| TENANT-01 | Phase 6 | Pending |
| TENANT-02 | Phase 6 | Pending |
| TENANT-03 | Phase 6 | Pending |
| PORTAL-01 | Phase 7 | Pending |
| PORTAL-02 | Phase 7 | Pending |
| PORTAL-03 | Phase 7 | Pending |
| PORTAL-04 | Phase 7 | Pending |
| LAND-01 | Phase 8 | Pending |
| LAND-02 | Phase 8 | Pending |
| IPWL-01 | Phase 8 | Pending |
| IPWL-02 | Phase 8 | Pending |
| IPWL-03 | Phase 8 | Pending |
| APIDOC-01 | Phase 9 | Pending |
| APIDOC-02 | Phase 9 | Pending |
| GROUP-01 | Phase 9 | Pending |
| GROUP-02 | Phase 9 | Pending |
| GROUP-03 | Phase 9 | Pending |
| TLOG-01 | Phase 9 | Pending |
| TLOG-02 | Phase 9 | Pending |

**Coverage:**
- v2 requirements: 24 total
- Mapped to phases: 24
- Unmapped: 0 ✓

---
*Requirements defined: 2026-02-25*
*Last updated: 2026-02-27 after v2.0 roadmap creation*
