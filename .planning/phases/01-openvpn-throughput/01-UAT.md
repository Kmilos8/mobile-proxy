---
status: diagnosed
phase: 01-openvpn-throughput
source: [01-01-SUMMARY.md, 01-02-SUMMARY.md]
started: 2026-02-26T01:30:00Z
updated: 2026-02-26T01:42:00Z
---

## Current Test

[testing complete]

## Tests

### 1. .ovpn Download Works Without Manual Edits
expected: Download a .ovpn file from the dashboard. Open it in a text editor and confirm it contains `sndbuf 0` and `rcvbuf 0` (not `524288`). Import it into an OpenVPN client — it should import cleanly without needing any manual modifications.
result: issue
reported: "they have sndbuf 524288 rcvbuf 524288"
severity: major

### 2. Webpage Loads Through VPN Tunnel
expected: Connect to the VPN using the downloaded .ovpn file. Open a browser and navigate to any website (e.g., example.com). The page should fully load — HTML, CSS, images, all resources — without timeouts or partial renders.
result: issue
reported: "It doesn't work"
severity: major

### 3. Speed Test Completes With Measurable Throughput
expected: While connected to the VPN, run a speed test (e.g., fast.com or speedtest.net). The test should complete and report measurable download/upload speeds — not timeout or stall. Throughput should be noticeably better than before (previously capped around ~5 Mbps due to fixed buffers).
result: issue
reported: "speed test page won't fully load in order to perform it"
severity: major

### 4. VPN Client Rejected Cleanly on API Failure
expected: If the tunnel API is down or unreachable when a VPN client connects, the client should be rejected (connection refused) rather than silently admitted with no internet. The client's OpenVPN app should show a connection failure, not a successful connection that can't route traffic.
result: skipped
reason: Hard to simulate API failure in current environment

### 5. HTTP and SOCKS5 Proxies Stable During VPN Activity
expected: While an OpenVPN client connects and disconnects, test that HTTP and SOCKS5 proxy ports still respond correctly. Send a request through each proxy — they should work normally, unaffected by VPN client connect/disconnect events.
result: issue
reported: "tested http port and it doesn't work now as well.."
severity: major

## Summary

total: 5
passed: 0
issues: 4
pending: 0
skipped: 1

## Gaps

- truth: "Generated .ovpn files contain sndbuf 0 and rcvbuf 0 instead of fixed 524288"
  status: failed
  reason: "User reported: they have sndbuf 524288 rcvbuf 524288"
  severity: major
  test: 1
  root_cause: "Stale .ovpn files downloaded before fix commit 1843118 (files dated 2026-02-25 03:17/03:52 EST, fix landed 20:07 EST). Source code is correct (sndbuf 0). Server binary may not be rebuilt/redeployed yet."
  artifacts:
    - path: "server/internal/api/handler/openvpn_handler.go"
      issue: "Code is correct (sndbuf 0) — but binary may not be redeployed"
  missing:
    - "Rebuild and redeploy server binary on VPS"
    - "Re-download .ovpn files after redeployment"
  debug_session: ".planning/debug/ovpn-sndbuf-stale-files.md"
- truth: "A webpage fully loads through the VPN tunnel without timeouts"
  status: failed
  reason: "User reported: It doesn't work"
  severity: major
  test: 2
  root_cause: "Device is offline (VpnIP empty). API returns 503 → client-connect-ovpn.sh exit 1 → OpenVPN rejects client → no tunnel established. Before Phase 1, || true swallowed the 503 and admitted client with no internet. Now it correctly rejects, but device must be online first."
  artifacts:
    - path: "server/internal/api/handler/openvpn_handler.go"
      issue: "Returns 503 when device.VpnIP is empty (device offline)"
    - path: "server/deployments/openvpn/client-connect-ovpn.sh"
      issue: "Correctly exits 1 on 503 — device must be online for VPN to work"
  missing:
    - "Ensure device is connected to tunnel (VpnIP assigned) before VPN client connects"
    - "Consider surfacing device online/offline status in dashboard"
  debug_session: ".planning/debug/vpn-webpages-not-loading.md"
- truth: "Speed test completes and reports measurable throughput through VPN"
  status: failed
  reason: "User reported: speed test page won't fully load in order to perform it"
  severity: major
  test: 3
  root_cause: "Same as Test 2 — VPN tunnel not established because device is offline. Speed test requires working VPN connection."
  artifacts: []
  missing:
    - "Resolve Test 2 (device online + redeployment) first"
  debug_session: ".planning/debug/vpn-webpages-not-loading.md"
- truth: "HTTP and SOCKS5 proxies respond correctly during VPN client connect/disconnect"
  status: failed
  reason: "User reported: tested http port and it doesn't work now as well.."
  severity: major
  test: 5
  root_cause: "DNAT rules (external port → device VPN IP) are set up in notifyConnected() in the device tunnel path. If device has no VPN IP (offline), DNAT rules were never created. Both VPN transparent proxy and HTTP proxy depend on the device being online in the tunnel."
  artifacts:
    - path: "server/cmd/tunnel/main.go"
      issue: "DNAT rules require device VPN IP — not installed when device offline"
  missing:
    - "Ensure device is connected to tunnel before testing proxy ports"
  debug_session: ".planning/debug/vpn-webpages-not-loading.md"
