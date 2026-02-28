# Milestones

## v1.0 MVP (Shipped: 2026-02-28)

**Phases:** 1-4 | **Plans:** 9 | **Timeline:** 2026-02-25 → 2026-02-27
**Codebase:** 7,864 LOC Go + 6,731 LOC TypeScript

**Key accomplishments:**
- Fixed OpenVPN throughput (peekTimeout, OS autotuning, fast-io, client-connect retry logic)
- Full dashboard redesign with shadcn/ui (device table, connection CRUD, device detail pages)
- Security hardening (bcrypt auth for OpenVPN/SOCKS5, PasswordPlain removed, password regeneration)
- Bandwidth enforcement (per-connection limits in tunnel, atomic counters, 30s flush to DB)
- Monitoring (device offline/recovery webhooks, settings page, bandwidth reset propagation)
- Polish (device search, auto-rotation column, connection ID, OpenVPN in Add Connection modal)

---

