---
page_title: "pritunl_server Resource - Pritunl"
description: |-
  Manages a Pritunl VPN server.
---

# pritunl_server (Resource)

Manages a Pritunl VPN server.

## Example Usage

```terraform
resource "pritunl_server" "vpn" {
  name     = "Production"
  network  = "10.10.0.0/24"
  port     = 1194
  protocol = "udp"

  attached_organization_ids = [pritunl_organization.engineering.id]
}
```

## Schema

### Required

- `name` (String) Server name.
- `network` (String) VPN network CIDR.
- `port` (Number) Server port.

### Optional

- `protocol` (String) Protocol (`udp` or `tcp`). Defaults to `"udp"`.
- `cipher` (String) Encryption cipher. Defaults to `"aes256"`.
- `hash` (String) Hash algorithm. Defaults to `"sha256"`.
- `inter_client` (Boolean) Allow inter-client routing. Defaults to `true`.
- `ping_interval` (Number) Ping interval in seconds. Defaults to `10`.
- `ping_timeout` (Number) Ping timeout in seconds. Defaults to `60`.
- `attached_organization_ids` (List of String) List of organization IDs attached to this server.

### Read-Only

- `id` (String) Server ID.
- `status` (String) Server status (`online`/`offline`).

## Import

Servers can be imported using their ID:

```shell
terraform import pritunl_server.example <server_id>
```
