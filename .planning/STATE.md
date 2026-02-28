---
gsd_state_version: 1.0
milestone: v2.0
milestone_name: SaaS Platform
status: in_progress
last_updated: "2026-02-28"
progress:
  total_phases: 5
  completed_phases: 0
  total_plans: 1
  completed_plans: 1
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-27)

**Core value:** Customers can reliably route traffic through real mobile devices via HTTP, SOCKS5, or OpenVPN, managed through a clean dashboard.
**Current focus:** Phase 5 — Auth Foundation (v2.0)

## Current Position

Phase: 5 of 9 (Phase 5: Auth Foundation)
Plan: 1 of TBD in current phase (05-01 complete)
Status: In progress
Last activity: 2026-02-28 — 05-01-PLAN.md executed: DB schema + repository layer for customer auth

Progress: [█░░░░░░░░░] 10% (v2.0 phases)

## Accumulated Context

### Decisions

- v1.0 shipped fully: protocol stability, dashboard redesign, security hardening, monitoring
- v2.0 scope: SaaS transformation — customer self-signup, multi-tenant, landing page, IP whitelist, API docs, device grouping, traffic logs
- Auth is strictly sequential: Phase 6 (auth) must complete before Phase 7 (tenant isolation), which must complete before Phase 8 (portal)
- Phases 9 and 10 are largely independent after Phase 5 (landing page has zero auth dependency)
- NextAuth.js explicitly excluded — Go backend owns all auth via JWT
- PostgreSQL RLS excluded — app-level WHERE customer_id scoping at this scale
- White-label theming deferred to v2.1 (WL-01, WL-02, WL-03)
- Nullable password_hash (*string) allows Google-only users with no password
- Token validity enforced in SQL (used_at IS NULL AND expires_at > NOW()) rather than application layer
- Case-insensitive email lookup via LOWER() to prevent duplicate accounts
- LinkGoogleAccount also sets email_verified=true (Google confirms email ownership)

### Blockers/Concerns

- Google OAuth: Google Cloud Console project and OAuth redirect URI must be registered before Phase 6 coding starts — day-one prerequisite
- Resend: sending domain must be verified in Resend dashboard before email delivery can be tested
- Cloudflare Turnstile: site key + secret key must be created before Turnstile widget can be tested
- kylemanna/openvpn image is unmaintained (OpenVPN 2.4.9); defer image replacement until needed

### Pending Todos

None yet.

## Session Continuity

Last session: 2026-02-28
Stopped at: Completed 05-01-PLAN.md (DB schema + data access layer for customer auth)
Resume file: None
