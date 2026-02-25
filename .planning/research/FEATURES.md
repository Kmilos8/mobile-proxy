# Feature Research

**Domain:** Mobile Proxy Platform (device-to-proxy, operator dashboard, direct proxy access)
**Researched:** 2026-02-25
**Confidence:** MEDIUM — competitor features verified via official sites and help docs; some nuances are from third-party reviews

---

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Proxy credentials display (host:port:user:pass) | Every proxy product shows credentials; users copy-paste into tools | LOW | Standard format; support for multiple copy formats (IP vs hostname) is a differentiator |
| HTTP and SOCKS5 proxy protocol support | Industry standard; tools expect one or both | LOW | Already implemented; both must be stable and concurrently usable |
| OpenVPN config (.ovpn) generation and download | Tunnel-based access method expected for VPN-style use cases | MEDIUM | Already partially working; throughput must be production-quality |
| Device online/offline status | Operators need to know if a device is serving traffic | LOW | Real-time or near-real-time; the core operational signal |
| Manual IP rotation trigger | Users want to rotate IPs on demand from dashboard | LOW | Single-click or API call; table stakes for all proxy products |
| Proxy port on/off toggle | Users need to temporarily disable a proxy without deleting it | LOW | Soft disable; distinct from deletion |
| Current IP address display | Users need to know what exit IP a proxy is currently using | LOW | Shown per proxy port, updated after each rotation |
| IP history (last N IPs) | Users verify rotation is working; troubleshoot sticky IPs | LOW | iProxy shows last 3 IPs; storing 5-10 is sufficient |
| Basic traffic / bandwidth tracking | Users monitor data usage to stay within plan limits | MEDIUM | Cumulative bytes in/out per device or per port; monthly reset |
| Uptime / offline notifications | Operators lose money when a device drops; alerts are critical | MEDIUM | Email or webhook on state change; Telegram is a competitor differentiator |
| Multiple proxy ports per device | A single Android phone can serve multiple simultaneous proxy endpoints | MEDIUM | iProxy supports up to 15 ports per device; critical for density economics |
| QR code device onboarding | Scan-to-connect is the expected Android pairing UX | LOW | Dashboard generates QR; app scans it; avoids manual token entry |
| Device nickname / label | Operators manage tens or hundreds of devices; names are essential | LOW | Freetext label stored per device |
| Proxy access deletion | Users must be able to remove ports they no longer need | LOW | Hard delete with confirmation |
| IP whitelist authentication | Alternative to user/pass; many automation tools prefer whitelist | MEDIUM | Per-port whitelist; CIDR support preferred |

---

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Automatic IP rotation on configurable interval | Set-and-forget rotation without airplane mode tricks; iProxy requires root for true auto-rotation | HIGH | Non-root auto-rotation requires controlling the device's data toggle; architecture-dependent |
| OpenVPN direct access (full tunnel) | Lets customers use proxies from tools that don't support HTTP/SOCKS5 natively | HIGH | Already partially built; differentiates from pure HTTP/SOCKS5 providers |
| Per-device carrier and signal strength display | Operators optimizing a multi-carrier farm need carrier-level visibility | MEDIUM | Android app reports carrier name, signal bars, and network type (4G/5G/LTE) |
| Per-device battery level display | Prevents surprise dropouts on battery-powered devices | LOW | Android API provides this; battery % + charging status |
| Device grouping | Organize devices by location, carrier, or customer | MEDIUM | Groups with custom labels; bulk operations apply to group |
| Bulk actions (rotate all, reboot all) | Operators managing 50+ devices need batch controls | MEDIUM | Applies to selection or group; rotate triggers IP change on all selected |
| Rotation via unique URL (rotation link) | Lets external tools (scripts, automations) trigger rotation without API auth | LOW | One-time or per-rotation signed URL; iProxy offers this as a feature |
| API for proxy management | Developers integrating with automation frameworks need programmatic control | MEDIUM | REST API for: list ports, rotate IP, get credentials, get status |
| Proxy credential format customization | Different tools expect different formats; one-click format switching reduces support load | LOW | user:pass@host:port vs host:port:user:pass vs others |
| Per-port traffic logs downloadable | Advanced users debug issues by inspecting traffic; compliance use cases require logs | HIGH | Log storage per plan tier (7 days vs 12 weeks is iProxy's model) |
| Proxy blacklist (block specific domains at proxy level) | Users want to prevent their proxies from accessing certain sites | MEDIUM | Per-port or per-account blocklist; enforced at proxy server layer |
| Passive OS fingerprint spoofing | Prevents detection of proxy nature via TCP/IP stack fingerprint; iProxy offers this | HIGH | Requires kernel-level or deep network stack manipulation; advanced use case |
| Wi-Fi split mode (route mobile data only through proxy) | Preserves mobile data budget by routing management traffic over Wi-Fi | MEDIUM | Android app networking configuration; iProxy calls this "Wi-Fi Split" |
| Remote app update trigger | Operators can update device app without physical access | HIGH | Requires root or device owner mode on Android; rooted fleet assumption |
| SMS forwarding to external destination | Some proxy use cases involve OTP/SMS interception for account farms | HIGH | Sensitive feature; legal/TOS risk; iProxy forwards SMS to Telegram |
| Connection sharing / team access | Multiple users operate the same proxy fleet | MEDIUM | Per-connection grant; different permission levels (view vs rotate vs manage) |

---

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Automatic IP rotation without root via airplane mode toggle | Users want hands-off rotation | Airplane mode toggle causes a 5-30 second reconnect gap; unreliable carrier reconnect times create unpredictable downtime | Support manual rotation via dashboard + rotation URL; clearly document that true auto-rotation requires root or carrier cooperation |
| Full billing and payment integration | SaaS completeness expectation | Payment processing is a compliance and security surface (PCI-DSS); billing systems are a product in themselves | Use Stripe Checkout or an external billing tool; keep billing external until multi-tenant SaaS milestone |
| Real-time traffic interception / packet capture | Power users want deep inspection | Enables abuse; creates significant legal and privacy liability; massively increases storage requirements | Provide aggregate traffic stats (bytes in/out) without payload capture |
| Geolocation targeting (country/city/carrier selection at request time) | Used by large proxy providers | Requires a pool of devices in diverse locations; architecture assumption breaks (you control specific devices, not a pool) | Expose device-level metadata (carrier, location) and let users pick specific devices; do not abstract into location pools |
| Unlimited concurrent connections per port | Users want maximum throughput | A single Android device has physical limits (CPU, memory, carrier throttling); advertising "unlimited" creates support burden | Set per-port connection limits that reflect real device capacity; expose current connection count |
| iOS app support | Larger user base | iOS background app restrictions prevent reliable proxy serving; no equivalent of Android's always-on VPN service for custom proxy logic | Android-only for v1; document clearly; iOS proxy serving is architecturally impractical without jailbreak |
| Shared proxy pool (multi-tenant IP pooling) | Reduces cost perception | Shared IPs get burned faster; customers blame each other; support burden explodes; trust model collapses | Private proxies per customer device only; this is the core value proposition of mobile proxies vs datacenter proxies |
| Telegram bot as primary control interface | Competitors offer it; users ask for it | Adds a third interface to maintain; Telegram dependency introduces an external SPOF; not all operator environments allow Telegram | Dashboard + REST API cover all use cases; add Telegram bot only if customer demand is validated |

---

## Feature Dependencies

```
Device online/offline status
    └──requires──> Device-to-server heartbeat (Android app)
                       └──requires──> Stable persistent connection (WebSocket/tunnel)

Current IP address display
    └──requires──> Device reports IP on connect + after rotation
                       └──requires──> Android app reads and reports active interface IP

Multiple proxy ports per device
    └──requires──> Server-side port allocation and management
                       └──requires──> Port registry (which ports belong to which device)

Manual IP rotation
    └──requires──> Server sends rotation command to device
                       └──requires──> Bi-directional device communication channel

Automatic IP rotation (configurable interval)
    └──requires──> Server-side scheduler per port
                       └──requires──> Manual IP rotation (same underlying mechanism)

IP whitelist authentication
    └──requires──> Per-port configuration storage
    └──requires──> Proxy server enforces whitelist at connection time

API for proxy management
    └──requires──> Stable internal port management data model
    └──enhances──> Manual IP rotation (same action, different trigger)

Per-device carrier and signal display
    └──requires──> Android app reports carrier metadata on connect + periodically

Per-device battery level display
    └──requires──> Android app reports battery state periodically

Device grouping
    └──requires──> Device list and labeling
    └──enhances──> Bulk actions (apply action to group)

Bulk actions (rotate all, reboot all)
    └──requires──> Manual IP rotation (applied to N devices)

Connection sharing / team access
    └──requires──> User accounts and authentication system
    └──requires──> Per-connection permission model

OpenVPN direct access
    └──requires──> Server generates and serves .ovpn configs
    └──requires──> VPN tunnel routes through device's mobile data

Traffic / bandwidth tracking
    └──requires──> Proxy server instruments bytes per connection

Uptime / offline notifications
    └──requires──> Device online/offline status (event stream)
    └──requires──> Notification delivery mechanism (email/webhook)
```

### Dependency Notes

- **Multiple proxy ports requires port registry:** The server must track which ports are allocated to which device. This is foundational; build it before any port management UI.
- **Manual IP rotation is a prerequisite for auto-rotation:** Both go through the same command channel to the Android app; auto-rotation is the scheduler layer on top.
- **Device status signal is the root dependency for most monitoring features:** Carrier, battery, IP, uptime notifications — all require the Android app to reliably report state back. Get this right first.
- **API requires stable data model:** Don't expose an API before the internal port/device model is stable; breaking API changes destroy integrations.
- **Team access requires user auth:** User accounts must exist before permission grants. Keep this out of MVP to avoid scope inflation.

---

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed for a proxy operator to run a functional device fleet.

- [ ] Device list with online/offline status, current IP, device name — without this the operator is flying blind
- [ ] Proxy credentials display per port (HTTP host:port:user:pass, SOCKS5 host:port:user:pass) — without this customers cannot connect
- [ ] Multiple proxy ports per device (at least 3, target 10) — economics require density; single port per device is not viable
- [ ] Manual IP rotation from dashboard — every proxy product has this; no exceptions
- [ ] OpenVPN .ovpn file generation and download per port — third protocol, core requirement per PROJECT.md
- [ ] Proxy port create and delete — basic CRUD for port management
- [ ] QR code onboarding for Android devices — eliminates manual token entry friction
- [ ] Basic traffic counter per device (total bytes) — operators need to see usage
- [ ] Online/offline notification via email — operators must be alerted when devices drop

### Add After Validation (v1.x)

Features to add once core is working and operators are using it.

- [ ] Per-device carrier name, signal strength, battery level display — trigger: operator feedback that they can't diagnose device issues
- [ ] Automatic IP rotation with configurable interval — trigger: customers requesting scheduled rotation
- [ ] IP whitelist authentication — trigger: customers using automation tools that prefer whitelist over user:pass
- [ ] Rotation via unique URL — trigger: customers building external automation
- [ ] Device grouping with bulk actions — trigger: fleet size exceeds 20 devices
- [ ] REST API for proxy management — trigger: customers requesting programmatic access
- [ ] Downloadable traffic logs — trigger: compliance or debugging requests

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] Multi-tenant / team access with permission levels — defer: requires user account system overhaul; out of scope per PROJECT.md
- [ ] Proxy blacklist per port — defer: edge case; no customer demand validated
- [ ] White-label SaaS (other operators deploy their own instance) — defer: architectural work; out of scope per PROJECT.md for v1
- [ ] Passive OS fingerprint spoofing — defer: advanced use case; high complexity; limited audience
- [ ] Per-port traffic logs (payload-level) — defer: storage costs and privacy/legal concerns require deliberate design
- [ ] SMS forwarding — defer: legal risk surface; requires deliberate TOS and abuse handling design

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Device online/offline status | HIGH | LOW | P1 |
| Proxy credentials display | HIGH | LOW | P1 |
| Proxy port create/delete | HIGH | LOW | P1 |
| Multiple ports per device | HIGH | MEDIUM | P1 |
| Manual IP rotation | HIGH | MEDIUM | P1 |
| OpenVPN .ovpn generation | HIGH | HIGH | P1 |
| QR code device onboarding | HIGH | LOW | P1 |
| Basic traffic counter | MEDIUM | LOW | P1 |
| Offline/online notification (email) | HIGH | LOW | P1 |
| Current IP display | HIGH | LOW | P1 |
| IP history (last N) | MEDIUM | LOW | P2 |
| Carrier/signal/battery display | MEDIUM | LOW | P2 |
| Automatic rotation (scheduled) | HIGH | MEDIUM | P2 |
| IP whitelist auth | MEDIUM | MEDIUM | P2 |
| Rotation via URL | MEDIUM | LOW | P2 |
| Device grouping + bulk actions | MEDIUM | MEDIUM | P2 |
| REST API | HIGH | MEDIUM | P2 |
| Traffic logs download | LOW | HIGH | P3 |
| Proxy blacklist | LOW | MEDIUM | P3 |
| Team access / sharing | MEDIUM | HIGH | P3 |
| Passive OS fingerprint spoofing | LOW | HIGH | P3 |
| SMS forwarding | LOW | HIGH | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

---

## Competitor Feature Analysis

| Feature | iProxy.online | Proxidize | PocketProxy (current state) |
|---------|--------------|-----------|------------------------------|
| HTTP proxy | Yes | Yes | Yes (working) |
| SOCKS5 proxy | Yes | Yes | Yes (working) |
| OpenVPN access | Yes (.ovpn) | No (HTTP/SOCKS5 only) | Partial (throughput broken) |
| Ports per device | Up to 15 | Varies by plan | Unknown / not implemented |
| Dashboard | Web + mobile app | Web only | Basic scaffold only |
| IP rotation trigger | Dashboard, URL, Telegram bot, API | Dashboard, API | Not implemented |
| Auto-rotation | Yes (root required for interval) | Yes | Not implemented |
| Device status | Online/offline, uptime | Online/offline, signal, SIM info | Not implemented |
| Battery display | Not confirmed | Not confirmed | Not implemented |
| Carrier display | Not confirmed | Yes (carrier targeting) | Not implemented |
| IP history | Yes (last 3) | Yes | Not implemented |
| Traffic stats | Yes (per port, per month) | Yes (analytics) | Not implemented |
| Traffic logs download | Yes (7 days–12 weeks by plan) | Yes | Not implemented |
| QR code onboarding | Yes | Yes | Yes (qrcode.react present) |
| IP whitelist auth | Yes | Yes | Not implemented |
| Device grouping | Yes | Yes | Not implemented |
| Bulk actions | Yes (rotate, pay, notify) | Yes (rotate, reboot) | Not implemented |
| REST API | Yes | Yes (full developer API) | Not implemented |
| Telegram bot | Yes | No | Not planned |
| Team sharing | Yes | Yes (multi-user, Business plan) | Not implemented |
| Proxy format customization | Yes (multiple formats) | Yes | Not implemented |
| Rotation URL | Yes | Yes | Not implemented |
| Proxy blacklist | Yes (account/connection/port level) | Not confirmed | Not implemented |

---

## Sources

- iProxy.online homepage: https://iproxy.online/ (MEDIUM confidence — official site, marketing copy)
- iProxy.online personal account features: https://iproxy.online/blog/all-the-personal-account-features (MEDIUM confidence — official blog)
- iProxy.online FAQ personal area: https://iproxy.online/faq/personal-area (MEDIUM confidence — official FAQ)
- iProxy.online new features blog: https://iproxy.online/blog/new-features-of-iproxy (MEDIUM confidence — official blog)
- iProxy.online mobile proxies page: https://iproxy.online/mobile-proxies (MEDIUM confidence — official product page)
- iProxy.online Telegram bot guide: https://iproxy.online/blog/telegram-bot-iproxy (MEDIUM confidence — official docs)
- Proxidize Proxy Builder: https://proxidize.com/proxy-builder/ (MEDIUM confidence — official site)
- Proxidize device onboarding: https://help.proxidize.com/en/onboard-new-android-device-using-proxidize-android-agent (MEDIUM confidence — official help docs)
- Proxidize platform plans: https://help.proxidize.com/en/proxdize-platform-plans (MEDIUM confidence — official pricing/features page)
- Proxidize review (Dolphin Anty): https://dolphin-anty.com/blog/en/review-of-proxidize-the-all-in-one-mobile-proxy-solution/ (LOW confidence — third-party review)
- ProxyLTE via PrivateProxyReviews: https://www.privateproxyreviews.com/proxylte/ (LOW confidence — third-party review)
- Multilogin ProxyLTE review: https://multilogin.com/review/proxylte-multilogin-exclusive-offer/ (LOW confidence — partner review)
- G2 iProxy reviews: https://www.g2.com/products/iproxy-online/reviews (LOW confidence — aggregated user reviews)

---

*Feature research for: Mobile proxy platform (PocketProxy)*
*Researched: 2026-02-25*
