---
status: diagnosed
trigger: "Investigate why webpages don't load through the VPN tunnel after Phase 1 changes"
created: 2026-02-25T00:00:00Z
updated: 2026-02-25T00:00:00Z
---

## Current Focus

hypothesis: The API handler (openvpn_handler.go Connect()) omits proxy_port from the tunnel notification payload, causing the tunnel server to default to port 8080 for HTTP proxying — which was always the case. The scripts look correct. The real root cause is different.
test: Traced full call chain: script -> API -> tunnel server -> iptables
expecting: One specific step in the chain is the point of failure
next_action: DIAGNOSED — return root cause

## Symptoms

expected: Webpages load through VPN tunnel; HTTP proxy port works
actual: Webpages do not load; HTTP proxy port also broken
errors: Not specified — user reports both VPN transparent proxy and HTTP proxy port fail
reproduction: Connect a client via OpenVPN on port 1195
started: After Phase 1 changes

## Eliminated

- hypothesis: peekTimeout 200ms causes connection drops
  evidence: Logic is correct — peeked data is preserved in bufio.Reader; if peek times out, peeked is empty and the proxy falls back to raw IP; connections are not dropped, just hostname extraction is skipped. Not a blocker.
  timestamp: 2026-02-25

- hypothesis: client-connect-ovpn.sh retry logic exits 1 and rejects client
  evidence: The retry logic is correct. MAX_RETRIES=2, loops attempt 1 and 2, exits 0 on success, exits 1 only after both attempts fail. The script itself is sound. The || true removal is intentional — failing fast is correct behavior.
  timestamp: 2026-02-25

- hypothesis: client-disconnect-ovpn.sh exits 1 and causes a problem
  evidence: OpenVPN ignores exit code from disconnect scripts. Exit 1 only affects logging. Not a traffic blocker.
  timestamp: 2026-02-25

- hypothesis: fast-io in openvpn conf causes issues
  evidence: fast-io is a valid UDP optimization. proto is udp. No problem here.
  timestamp: 2026-02-25

- hypothesis: sndbuf/rcvbuf 0 causes kernel default regression
  evidence: sndbuf/rcvbuf 0 means "use kernel default" — this is a valid and common setting. Not a blocker.
  timestamp: 2026-02-25

## Evidence

- timestamp: 2026-02-25
  checked: server/deployments/openvpn/client-connect-ovpn.sh
  found: Script POSTs to $API_URL/internal/openvpn/connect with {username, vpn_ip}. Retry and exit-1 logic is correct.
  implication: Script itself is not the source of failure. If API call succeeds, OpenVPN admits client.

- timestamp: 2026-02-25
  checked: server/internal/api/handler/openvpn_handler.go Connect()
  found: Handler builds the tunnel notification body as:
    {"client_vpn_ip": req.VpnIP, "device_vpn_ip": device.VpnIP, "socks_port": 1080, "socks_user": conn.Username, "socks_pass": conn.PasswordPlain}
  found: The field "proxy_port" is ABSENT from this payload. Only "socks_port" is sent.
  implication: The tunnel server's handleOpenVPNClientConnect reads "proxy_port" (json:"proxy_port"). Since it is absent, req.ProxyPort == 0, which then defaults to 8080. So the proxy endpoint is device_vpn_ip:8080.

- timestamp: 2026-02-25
  checked: server/cmd/tunnel/main.go handleOpenVPNClientConnect
  found: Struct field is `ProxyPort int \`json:"proxy_port"\`` and `SOCKSPort int \`json:"socks_port"\``. Only proxy_port is used for the endpoint. socks_port is received but unused (labelled "legacy, unused").
  implication: The API sends socks_port=1080 but the tunnel server reads proxy_port which is absent -> defaults to 0 -> becomes 8080. This has always been the case — it predates Phase 1.

- timestamp: 2026-02-25
  checked: server/cmd/tunnel/main.go handleOpenVPNClientConnect iptables rules
  found: Three iptables rules are added in order:
    1. -t nat -A PREROUTING -s {clientIP} -p tcp -d 10.9.0.0/24 -j RETURN
    2. -t nat -A PREROUTING -s {clientIP} -p tcp -d 192.168.255.0/24 -j RETURN
    3. -t nat -A PREROUTING -s {clientIP} -p tcp -j REDIRECT --to-port 12345
  found: RETURN rules use -A (append) not -I (insert). REDIRECT also uses -A.
  implication: If there are pre-existing rules in PREROUTING (e.g. UFW rules), the RETURN and REDIRECT rules are appended AFTER them. If a UFW DROP/REJECT rule matches first, traffic never reaches the REDIRECT rule. However this was also true before Phase 1.

- timestamp: 2026-02-25
  checked: Phase 1 change — client-connect-ovpn.sh removed || true
  found: Previously the script used || true, meaning API failures were silently swallowed and OpenVPN admitted the client anyway. Now the script exits 1 on API failure, rejecting the client.
  implication: If the API server is temporarily unavailable at connect time (e.g. startup race), the client is now REJECTED by OpenVPN rather than admitted without iptables rules. This is intentional hardening but changes behavior: before, a client would connect but have no internet; now, they cannot connect at all. NOT the cause of "webpages don't load" — that implies the client connected but traffic fails.

- timestamp: 2026-02-25
  checked: proxy_port field in openvpn_handler.go Connect() payload to tunnel server
  found: "proxy_port" key is completely absent from the JSON body sent to the tunnel push API. The tunnel server struct has ProxyPort with json tag "proxy_port". Missing field -> zero value -> defaults to 8080. This means the transparent proxy always targets device_vpn_ip:8080.
  implication: If the device's HTTP proxy is actually on port 8080, this works. If the device proxy is on a different port, it fails. But this was pre-existing, not a Phase 1 change.

- timestamp: 2026-02-25
  checked: What actually changed in Phase 1 that could break both transparent proxy AND HTTP proxy port
  found: The HTTP proxy port (external port -> device) goes through DNAT rules set up in notifyConnected(). The transparent proxy goes through REDIRECT rules set up by handleOpenVPNClientConnect(). These are independent paths. If BOTH fail after Phase 1, the common cause must be something that affects both paths or the device tunnel (192.168.255.x) connectivity itself.
  found: The client-connect script now exits 1 on failure. OpenVPN docs state that if the client-connect script exits non-zero, the client is DISCONNECTED. A successfully connected client (pages loaded before) won't have this issue. But a new connection attempt where the API times out (5s wget timeout) could cause repeated rejection.
  implication: If the client IS connecting (VPN tunnel up) but pages don't load AND HTTP proxy port doesn't work, the device tunnel (192.168.255.x) must be unreachable or the iptables rules are not being installed.

- timestamp: 2026-02-25
  checked: The REDIRECT iptables rule ordering — RETURN rules use -A (append)
  found: CRITICAL: The cleanup step before adding rules does:
    runCmd("iptables", "-t nat -D PREROUTING -s {IP} -p tcp -d 10.9.0.0/24 -j RETURN" ...)
    runCmd("iptables", "-t nat -D PREROUTING -s {IP} -p tcp -d 192.168.255.0/24 -j RETURN" ...)
    runCmd("iptables", "-t nat -D PREROUTING -s {IP} -p tcp -j REDIRECT --to-port 12345" ...)
  Then adds with -A. This was also pre-existing (Phase 1 did NOT touch this file).
  implication: iptables rule setup logic is unchanged from before Phase 1.

## Resolution

root_cause: |
  TWO bugs found. The PRIMARY bug causing "webpages don't load" is in openvpn_handler.go.
  The SECONDARY preexisting bug is the proxy_port mismatch (socks_port vs proxy_port field names).

  PRIMARY ROOT CAUSE — openvpn_handler.go Connect() returns HTTP 503 when device.VpnIP == "":

    if device.VpnIP == "" {
        c.JSON(http.StatusServiceUnavailable, gin.H{"error": "device not connected to tunnel"})
        return
    }

  After the Phase 1 change to client-connect-ovpn.sh, the wget call has a 5-second timeout.
  If the API responds with a non-2xx status (503 if device offline, 404 if device not found),
  wget exits non-zero. The script then exits 1 after MAX_RETRIES, and OpenVPN REJECTS the client.

  The || true removal is what made this visible: before, a 503 was silently ignored and the
  client was admitted without iptables rules (broken but connected). Now the client is rejected
  outright. The user sees "VPN connected" may still appear at the OS level briefly before
  OpenVPN tears down, or the user sees a connection failure.

  HOWEVER — the user says webpages don't load (implying the client IS connected). This means
  the API call succeeded (exit 0), but something in the proxy chain is broken.

  ACTUAL PRIMARY ROOT CAUSE (given client IS connected):
  The openvpn_handler.go Connect() payload to the tunnel server uses field name "socks_user"
  and "socks_pass" for credentials, but the tunnel server struct reads them as ProxyUser
  (json:"socks_user") and ProxyPass (json:"socks_pass") — these DO match, so credentials
  ARE passed correctly.

  The REAL issue is that the handler sends "socks_port": 1080 to the tunnel server, but the
  tunnel server's struct field that controls the proxy endpoint port is ProxyPort
  (json:"proxy_port"). The "socks_port" field maps to SOCKSPort (json:"socks_port") which is
  documented as "legacy, unused". So ProxyPort always deserializes as 0, defaults to 8080,
  and the transparent proxy always connects to device_vpn_ip:8080.

  If the device's HTTP proxy is on port 8080, this works (and was always working before Phase 1
  too, so this is NOT what Phase 1 broke).

  CONCLUSION: Phase 1 itself did not introduce a new code-level bug in the transparent proxy
  or iptables setup. The changes are:
    1. peekTimeout 200ms — safe, fallback to raw IP if no data
    2. sndbuf/rcvbuf 0 — safe kernel default
    3. fast-io — safe for UDP
    4. client-connect-ovpn.sh || true removal + exit 1 — this IS the Phase 1 behavioral change

  The Phase 1 behavioral change (exit 1 on API failure) means: if the API returns an error
  for ANY reason (device offline, API server slow, DB lookup fails), the client is now REJECTED
  instead of admitted-but-broken. The user experience changes from "connects but no internet"
  to "VPN connection fails." Both cases result in no webpage loading, but for different reasons.

  The HTTP proxy port failing separately confirms the device tunnel (192.168.255.x) is not
  reachable — which points to the DEVICE being offline from the tunnel server, causing:
    a) device.VpnIP == "" in the API -> Connect() returns 503
    b) client-connect-ovpn.sh wget gets 503 -> exits non-zero after retries -> exits 1
    c) OpenVPN rejects the client
    d) No iptables REDIRECT rules installed -> no pages load
    e) DNAT to device also broken -> HTTP proxy port also fails

fix: Not yet applied
verification: Not yet verified
files_changed: []
