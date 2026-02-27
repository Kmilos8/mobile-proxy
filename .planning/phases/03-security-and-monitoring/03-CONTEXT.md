# Phase 3: Security and Monitoring - Context

**Gathered:** 2026-02-26
**Status:** Ready for planning

<domain>
## Phase Boundary

Harden credential storage (remove plaintext passwords, switch to bcrypt), enforce per-connection bandwidth limits in the proxy routing layer, and alert operators via webhook when devices go offline. Dashboard integration for configuration and visibility.

</domain>

<decisions>
## Implementation Decisions

### Credential Migration
- Batch migration on deploy: one-time script hashes all existing PasswordPlain values to bcrypt, then nulls out the plaintext field
- Active OpenVPN sessions continue undisturbed — only new connections use bcrypt auth
- Dashboard gets a "Regenerate Password" action per connection — generates new random password, stores bcrypt hash, shows plaintext once
- Password displayed once after creation or regeneration; after navigating away, plaintext is gone — must regenerate to see again
- .ovpn file includes auth-user-pass inline at creation/regen time

### Bandwidth Enforcement
- Track bytes per customer connection in the tunnel server (Go, main.go) — it already handles routing, natural place to count
- Hard cutoff immediately when limit is hit — stop forwarding packets for that connection
- Per proxy connection granularity — each OpenVPN profile has its own bandwidth limit, set by operator at creation
- Manual reset by operator only — "Reset Usage" action in dashboard, no automatic reset cycle

### Offline Alerting
- Webhook-only notification channel — POST to operator-configured URL
- Offline detection: no heartbeat received for 2 minutes from the device tunnel
- 5-minute cooldown after sending an offline alert for the same device — prevents alert storms from flapping connections
- Recovery notification: send a second webhook when device reconnects after being offline

### Operator Configuration
- Bandwidth limit field on the connection create/edit form in the dashboard, stored in database alongside connection
- Per-operator webhook URL setting in dashboard — applies to all their devices
- Usage bar on each connection card in the dashboard (e.g., "750 MB / 1 GB" progress bar)
- "Send Test" button next to webhook URL field — sends sample payload for operator to verify endpoint

### Claude's Discretion
- Webhook payload format and structure
- Exact bcrypt cost factor
- Migration script error handling and rollback approach
- Usage tracking persistence strategy (in-memory vs database flush interval)

</decisions>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 03-security-and-monitoring*
*Context gathered: 2026-02-26*
