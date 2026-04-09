---
page_title: "pritunl_servers Data Source - Pritunl"
description: |-
  Fetches all Pritunl servers.
---

# pritunl_servers (Data Source)

Fetches all Pritunl servers.

## Example Usage

```terraform
data "pritunl_servers" "all" {}
```

## Schema

### Read-Only

- `servers` (List of Object) List of servers. Each object contains:
  - `id` (String) Server ID.
  - `name` (String) Server name.
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
