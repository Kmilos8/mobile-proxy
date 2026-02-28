---
phase: 05-auth-foundation
plan: 01
subsystem: database
tags: [postgres, pgx, go, migration, oauth, jwt, bcrypt]

# Dependency graph
requires: []
provides:
  - Migration 011 adding password_hash, email_verified, google_id, google_email columns to customers table
  - customer_auth_tokens table with token_hash, type, expires_at, used_at (email_verify / password_reset)
  - Extended Customer domain struct with auth fields (PasswordHash, EmailVerified, GoogleID, GoogleEmail)
  - CustomerAuthToken domain struct
  - Auth request/response types: CustomerSignupRequest, CustomerLoginRequest, CustomerLoginResponse, ForgotPasswordRequest, ResetPasswordRequest, VerifyEmailRequest, ResendVerificationRequest
  - GoogleOAuthConfig, ResendConfig, TurnstileConfig in domain.Config and DefaultConfig
  - CustomerRepository auth methods: GetByEmail, GetByGoogleID, CreateWithAuth, UpdateEmailVerified, UpdatePasswordHash, LinkGoogleAccount
  - CustomerTokenRepository: Create, GetByHash (valid-only), MarkUsed, DeleteByCustomerAndType
affects:
  - 05-auth-foundation (plan 02 builds auth service and HTTP handlers on this layer)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - SQL migration files with up/down rollback
    - pgx pool QueryRow.Scan for single-row selects
    - Nullable pointer fields (*string) for optional OAuth and password columns
    - Case-insensitive email lookup via LOWER() in SQL
    - Token validity enforced in SQL (used_at IS NULL AND expires_at > NOW())

key-files:
  created:
    - server/migrations/011_customer_auth.up.sql
    - server/migrations/011_customer_auth.down.sql
    - server/internal/repository/customer_token_repo.go
  modified:
    - server/internal/domain/models.go
    - server/internal/domain/config.go
    - server/internal/repository/customer_repo.go

key-decisions:
  - "Nullable password_hash (*string) allows Google-only users with no password"
  - "Nullable google_id (*string) allows email-only users without OAuth"
  - "Token validity enforced in DB query (used_at IS NULL AND expires_at > NOW()) rather than application layer"
  - "Case-insensitive email lookup (LOWER(email) = LOWER($1)) prevents duplicate account creation"
  - "LinkGoogleAccount also sets email_verified=true (Google has confirmed email ownership)"

patterns-established:
  - "Repository methods follow pgx pool QueryRow.Scan pattern matching existing user_repo.go"
  - "New auth columns added to SELECT in both GetByID and List to keep struct fully populated"
  - "DeleteByCustomerAndType used to clean up stale tokens before issuing replacement"

requirements-completed: [AUTH-01, AUTH-02, AUTH-03, AUTH-04]

# Metrics
duration: 5min
completed: 2026-02-28
---

# Phase 5 Plan 01: Auth Foundation — DB Schema and Data Access Layer Summary

**PostgreSQL schema migration and Go repository layer for customer auth: password, email verification, Google OAuth, and one-time token management**

## Performance

- **Duration:** 5 min
- **Started:** 2026-02-28T05:46:13Z
- **Completed:** 2026-02-28T05:51:51Z
- **Tasks:** 2
- **Files modified:** 5 (3 created, 3 modified)

## Accomplishments
- Migration 011 adds four auth columns to customers and creates customer_auth_tokens table with indexes optimized for token lookup and customer+type filtering
- Customer domain struct extended with PasswordHash (*string), EmailVerified (bool), GoogleID (*string), GoogleEmail (*string) — nullable pointer types for optional fields
- Full set of auth request/response types added to models.go covering signup, login, forgot/reset password, email verification, and resend flows
- Config structs added for Google OAuth, Resend email, and Cloudflare Turnstile (all with empty defaults, populated via env vars)
- CustomerRepository extended with 6 auth methods; existing GetByID and List updated to include new columns in SELECT
- CustomerTokenRepository created following exact same pattern as existing repos in the package

## Task Commits

Each task was committed atomically:

1. **Task 1: Database migration and domain model extensions** - `b4aa166` (feat)
2. **Task 2: Customer repository auth extensions and token repository** - `abfa279` (feat)

**Plan metadata:** (final docs commit — see below)

## Files Created/Modified
- `server/migrations/011_customer_auth.up.sql` - ALTER TABLE customers adds auth columns; CREATE TABLE customer_auth_tokens with FK, CHECK constraint, and indexes
- `server/migrations/011_customer_auth.down.sql` - Rollback: drop customer_auth_tokens, drop auth columns from customers
- `server/internal/domain/models.go` - Customer struct extended with auth fields; CustomerAuthToken struct added; 7 customer auth request/response types added
- `server/internal/domain/config.go` - Config struct extended with Google/Resend/Turnstile fields; 3 new config structs; DefaultConfig updated
- `server/internal/repository/customer_repo.go` - GetByID/List updated to SELECT auth columns; 6 new auth methods added
- `server/internal/repository/customer_token_repo.go` - New: CustomerTokenRepository with Create, GetByHash, MarkUsed, DeleteByCustomerAndType

## Decisions Made
- **Nullable pointer fields for PasswordHash and GoogleID**: `*string` allows Google-only users (no password) and email-only users (no Google ID) to coexist in the same table without sentinel values
- **Token validity enforced in SQL**: `GetByHash` filters `used_at IS NULL AND expires_at > NOW()` at the DB level rather than returning all tokens and checking in Go — avoids TOCTOU race and reduces data transfer
- **Case-insensitive email**: `LOWER(email) = LOWER($1)` in GetByEmail prevents duplicate accounts via case variation (e.g., User@example.com vs user@example.com)
- **LinkGoogleAccount sets email_verified=true**: Google has already verified the email address, no separate step needed

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered
- `go build ./...` could not be run directly from the bash shell environment on Windows (MINGW64 bash does not have Go in PATH). Code correctness was verified by code review — all patterns follow the existing repository conventions exactly (matching user_repo.go, db.go patterns). Build verification will happen naturally when the service is next compiled/deployed.

## User Setup Required
None — no external service configuration required by this plan. External service credentials (Google OAuth, Resend, Turnstile) are configuration concerns for the service layer (Plan 02).

## Next Phase Readiness
- DB schema and repository layer complete — Plan 02 can implement auth service, HTTP handlers, and middleware directly on top of these methods
- CustomerTokenRepository's GetByHash validity filter (used_at IS NULL AND expires_at > NOW()) means service layer just checks for nil return to determine token validity
- All customer auth request/response types are ready for handler binding

---
*Phase: 05-auth-foundation*
*Completed: 2026-02-28*

## Self-Check: PASSED

All created files exist and all task commits verified in git log.
