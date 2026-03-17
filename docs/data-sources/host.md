---
page_title: "pritunl_host Data Source - Pritunl"
description: |-
  Fetches a Pritunl host by ID or name.
---

# pritunl_host (Data Source)

Fetches a Pritunl host by ID or name. One of `id` or `name` must be specified.

## Example Usage

```terraform
data "pritunl_host" "primary" {
  name = "vpn-host-1"
}
```

## Schema

### Optional

- `id` (String) Host ID.
- `name` (String) Host name.

### Read-Only

- `hostname` (String) Host hostname.
- `status` (String) Host status.
- `public_addr` (String) Public IPv4 address.
- `public_addr6` (String) Public IPv6 address.
- `local_addr` (String) Local IPv4 address.
- `local_addr6` (String) Local IPv6 address.
- `cpu_usage` (Number) CPU usage percentage.
- `mem_usage` (Number) Memory usage percentage.
- `version` (String) Pritunl version.
