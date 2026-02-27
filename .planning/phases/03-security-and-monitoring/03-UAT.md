---
status: complete
phase: 03-security-and-monitoring
source: 03-01-SUMMARY.md, 03-02-SUMMARY.md
started: 2026-02-27T12:00:00Z
updated: 2026-02-27T12:08:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Regenerate Password Dialog
expected: In the Connections table, each row has a Regenerate button (refresh icon). Clicking it opens an AlertDialog showing the new plaintext password with a Copy button and a "save now, will not be shown again" warning.
result: pass

### 2. Download .ovpn After Regenerate
expected: For OpenVPN connections, the Regenerate dialog includes a "Download .ovpn" button. Downloading the .ovpn file embeds the new credentials inline (auth-user-pass block).
result: pass

### 3. Bandwidth Limit on Connection Creation
expected: The Add Connection modal includes a Bandwidth Limit field where you enter a value in GB. The limit is saved with the connection.
result: pass

### 4. Bandwidth Usage Bar in Connection List
expected: The Connections table has a Usage column showing a visual bar (BandwidthBar) indicating how much bandwidth each connection has consumed relative to its limit.
result: pass

### 5. Reset Bandwidth Usage
expected: Each connection row with bandwidth usage has a Reset Usage button (rotate icon). Clicking it resets the usage counter to 0 and the usage bar clears.
result: pass

### 6. Settings Page
expected: Navigating to /settings (via a settings icon in the header) shows a page with a Webhook URL input field, a Save button, and a Send Test button.
result: pass

### 7. Save Webhook URL
expected: Entering a webhook URL and clicking Save persists the URL. Reloading the settings page shows the saved URL.
result: pass

### 8. Send Test Webhook
expected: After saving a valid webhook URL, clicking "Send Test" delivers a test POST to that URL. The button provides feedback that the test was sent.
result: skipped
reason: User unsure how to test webhook delivery at this time

## Summary

total: 8
passed: 7
issues: 0
pending: 0
skipped: 1

## Gaps

[none yet]
