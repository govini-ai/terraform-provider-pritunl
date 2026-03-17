---
page_title: "pritunl_hosts Data Source - Pritunl"
description: |-
  Fetches all Pritunl hosts.
---

# pritunl_hosts (Data Source)

Fetches all Pritunl hosts.

## Example Usage

```terraform
data "pritunl_hosts" "all" {}
```

## Schema

### Read-Only

- `hosts` (List of Object) List of hosts. Each object contains:
  - `id` (String) Host ID.
  - `name` (String) Host name.
  - `hostname` (String) Host hostname.
  - `status` (String) Host status.
  - `public_addr` (String) Public IPv4 address.
  - `public_addr6` (String) Public IPv6 address.
  - `local_addr` (String) Local IPv4 address.
  - `local_addr6` (String) Local IPv6 address.
  - `cpu_usage` (Number) CPU usage percentage.
  - `mem_usage` (Number) Memory usage percentage.
  - `version` (String) Pritunl version.
