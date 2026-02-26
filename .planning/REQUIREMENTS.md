# Requirements: PocketProxy

**Defined:** 2026-02-25
**Core Value:** Customers can reliably route traffic through real mobile devices via HTTP, SOCKS5, or OpenVPN, managed through a clean dashboard.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Protocol Stability

- [x] **PROTO-01**: OpenVPN direct access delivers usable throughput (pages load fully, speed tests complete)
- [ ] **PROTO-02**: HTTP and SOCKS5 proxies remain stable under production load

### Dashboard Redesign

- [ ] **DASH-01**: Full UI redesign with modern component library (shadcn/ui) across all existing pages
- [ ] **DASH-02**: Connection creation UI on dashboard (currently API-only)
- [ ] **DASH-03**: Connection detail page for viewing/managing individual proxy ports
- [ ] **DASH-04**: Responsive layout for desktop and tablet

### Security

- [ ] **SEC-01**: Replace plaintext proxy password storage with secure approach (bcrypt hash exists but PasswordPlain also stored and used by OpenVPN auth)

### Monitoring

- [ ] **MON-01**: Device offline notification via email or webhook
- [ ] **MON-02**: Enforce bandwidth limits per connection (field exists but not enforced)

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

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Advanced Features

- **ADV-01**: IP whitelist authentication per proxy port
- **ADV-02**: REST API documentation and versioning
- **ADV-03**: Device grouping with bulk actions
- **ADV-04**: Per-port traffic logs (downloadable)

### SaaS Features

- **SAAS-01**: Multi-tenant / team access with permission levels
- **SAAS-02**: White-label dashboard theming
- **SAAS-03**: Customer self-service portal

## Out of Scope

| Feature | Reason |
|---------|--------|
| iOS app | iOS background restrictions prevent reliable proxy serving |
| Payment/billing integration | Handle externally (Stripe Checkout) until SaaS milestone |
| Telegram bot | Dashboard + API + rotation links cover all use cases |
| Real-time traffic interception | Legal/privacy liability, storage costs |
| Shared IP pools | Defeats core value of private mobile proxies |
| Geolocation targeting | Architecture assumes specific devices, not location pools |
| SMS forwarding | Legal risk; requires deliberate TOS design |
| Passive OS fingerprint spoofing | High complexity, limited audience |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| PROTO-01 | Phase 1 | Complete |
| PROTO-02 | Phase 1 | Pending |
| DASH-01 | Phase 2 | Pending |
| DASH-02 | Phase 2 | Pending |
| DASH-03 | Phase 2 | Pending |
| DASH-04 | Phase 2 | Pending |
| SEC-01 | Phase 3 | Pending |
| MON-01 | Phase 3 | Pending |
| MON-02 | Phase 3 | Pending |

**Coverage:**
- v1 requirements: 9 total
- Mapped to phases: 9
- Unmapped: 0

---
*Requirements defined: 2026-02-25*
*Last updated: 2026-02-25 after roadmap creation*
