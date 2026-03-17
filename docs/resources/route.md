---
page_title: "pritunl_route Resource - Pritunl"
description: |-
  Manages a route on a Pritunl VPN server.
---

# pritunl_route (Resource)

Manages a route on a Pritunl VPN server.

## Example Usage

```terraform
resource "pritunl_route" "vpc" {
  server_id = pritunl_server.vpn.id
  network   = "10.0.0.0/16"
  comment   = "VPC access"
}
```

## Schema

### Required

- `server_id` (String) Server ID this route belongs to. Changing this forces a new resource.
- `network` (String) Network CIDR for the route. Changing this forces a new resource.

### Optional

- `comment` (String) Comment/description for the route.
- `nat` (Boolean) Enable NAT for this route.

### Read-Only

- `id` (String) Route ID.

## Import

Routes can be imported using `server_id/route_id`:

```shell
terraform import pritunl_route.example <server_id>/<route_id>
```
