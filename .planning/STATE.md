---
gsd_state_version: 1.0
milestone: v2.0
milestone_name: Phases
status: unknown
last_updated: "2026-02-28T06:34:07.071Z"
progress:
  total_phases: 5
  completed_phases: 5
  total_plans: 12
  completed_plans: 12
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-27)

**Core value:** Customers can reliably route traffic through real mobile devices via HTTP, SOCKS5, or OpenVPN, managed through a clean dashboard.
**Current focus:** Phase 6 — Tenant Isolation (next, after Phase 5 Auth Foundation complete)

## Current Position

Phase: 5 of 9 (Phase 5: Auth Foundation)
Plan: 3 of 3 in current phase (05-03 complete — checkpoint:human-verify approved)
Status: Phase 5 complete, ready for Phase 6
Last activity: 2026-02-28 — 05-03-PLAN.md complete: All 3 tasks done. Six customer-facing auth pages (login extended, signup, verify-confirm, verify-email, forgot-password, reset-password) with Turnstile and Google OAuth. User visual verification approved.

Progress: [████░░░░░░] 40% (v2.0 phases — Phase 5 complete, 2 of 5 phases done)

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

### Blockers/Concerns

- Google OAuth: Google Cloud Console project and OAuth redirect URI must be registered before Phase 6 coding starts — day-one prerequisite
- Resend: sending domain must be verified in Resend dashboard before email delivery can be tested
- Cloudflare Turnstile: site key + secret key must be created before Turnstile widget can be tested
- kylemanna/openvpn image is unmaintained (OpenVPN 2.4.9); defer image replacement until needed

### Pending Todos

None yet.

## Session Continuity

Last session: 2026-02-28
Stopped at: Completed 05-03-PLAN.md — Phase 5 Auth Foundation complete (all 3 plans done)
Resume file: None
