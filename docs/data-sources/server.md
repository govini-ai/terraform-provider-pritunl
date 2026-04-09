---
page_title: "pritunl_server Data Source - Pritunl"
description: |-
  Fetches a Pritunl server by ID or name.
---

# pritunl_server (Data Source)

Fetches a Pritunl server by ID or name. One of `id` or `name` must be specified.

## Example Usage

```terraform
data "pritunl_server" "production" {
  name = "Production"
}
```

## Schema

### Optional

- `id` (String) Server ID.
- `name` (String) Server name.

### Read-Only

- `network` (String) VPN network CIDR.
- `ipv6` (Boolean) IPv6 enabled.
- `port` (Number) Server port.
- `protocol` (String) Protocol (`udp` or `tcp`).
- `cipher` (String) Encryption cipher.
- `hash` (String) Hash algorithm.
- `inter_client` (Boolean) Allow inter-client routing.
- `ping_interval` (Number) Ping interval in seconds.
- `ping_timeout` (Number) Ping timeout in seconds.
- `status` (String) Server status (`online`/`offline`).
