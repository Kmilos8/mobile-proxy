---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: SaaS Platform
status: unknown
last_updated: "2026-02-28T08:10:28.796Z"
progress:
  total_phases: 6
  completed_phases: 5
  total_plans: 15
  completed_plans: 13
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-27)

**Core value:** Customers can reliably route traffic through real mobile devices via HTTP, SOCKS5, or OpenVPN, managed through a clean dashboard.
**Current focus:** Phase 6 — Tenant Isolation (next, after Phase 5 Auth Foundation complete)

## Current Position

Phase: 6 of 9 (Phase 6: Multi-Tenant Isolation)
Plan: 1 of 3 in current phase (06-01 complete)
Status: Phase 6 in progress — plan 1 (data layer) complete, plans 2-3 remaining
Last activity: 2026-02-28 — 06-01-PLAN.md complete: 2 tasks done. Migration 012 adds customer_id to devices/pairing_codes, device_shares table, operator seed, data backfill. Domain models and customer-scoped repository methods implemented.

Progress: [████░░░░░░] 44% (v2.0 phases — Phase 5 complete, Phase 6 plan 1 of 3 done)

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
- Error sentinel string "email_not_verified" drives 403 vs 401 in Login handler — avoids custom error types for a single case
- Dev-mode fallbacks for EmailService (no API key) and Turnstile (no secret key) allow local testing without external credentials
- GoogleCallback redirect pattern: returns JWT via /login?token=X&google=true so frontend can store token then navigate
- Turnstile widget conditionally rendered via NEXT_PUBLIC_TURNSTILE_SITE_KEY — no widget in dev, forms still work (backend dev-mode auto-passes)
- Single login form with customer-first fallback to admin login — no separate operator login URL needed
- Two-step email verification: GET check on mount, user clicks Verify button, POST confirms — prevents link scanner token consumption
- Generic forgot-password success regardless of error type (except 429) — email enumeration prevention
- [Phase 06-01]: Customer_id nullable on devices/pairing_codes — backfill from admin user email match
- [Phase 06-01]: UNION query for ListByCustomer avoids duplicate rows when device is both owned and shared
- [Phase 06-01]: device_shares UNIQUE(device_id, shared_with) enforces one share record per device-customer pair at DB level
- [Phase 06-01]: OR EXISTS sub-select in GetByIDForCustomer keeps scan arity consistent — avoids JOIN-induced column count changes

### Blockers/Concerns

- Google OAuth: Google Cloud Console project and OAuth redirect URI must be registered before Phase 6 coding starts — day-one prerequisite
- Resend: sending domain must be verified in Resend dashboard before email delivery can be tested
- Cloudflare Turnstile: site key + secret key must be created before Turnstile widget can be tested
- kylemanna/openvpn image is unmaintained (OpenVPN 2.4.9); defer image replacement until needed

### Pending Todos

None yet.

## Session Continuity

Last session: 2026-02-28
Stopped at: Completed 06-01-PLAN.md — Phase 6 plan 1 (data layer) done. 2 plans remain in Phase 6.
Resume file: None
