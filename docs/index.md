---
page_title: "Pritunl Provider"
description: |-
  The Pritunl provider is used to manage Pritunl VPN server resources.
---

# Pritunl Provider

The Pritunl provider is used to manage resources on a [Pritunl](https://pritunl.com/) VPN server. It supports managing organizations, servers, routes, users, and reading host information.

## Authentication

The provider authenticates using an API token and secret. These can be generated in the Pritunl web console under Users > API Keys.

## Example Usage

```terraform
provider "pritunl" {
  url      = "https://vpn.example.com"
  token    = var.pritunl_api_token
  secret   = var.pritunl_api_secret
  insecure = false
}

resource "pritunl_organization" "engineering" {
  name = "Engineering"
}

resource "pritunl_server" "vpn" {
  name     = "Production"
  network  = "10.10.0.0/24"
  port     = 1194
  protocol = "udp"
  ipv6     = true

  attached_organization_ids = [pritunl_organization.engineering.id]
}
```

## Schema

### Optional

- `url` (String) URL of the Pritunl server. Can also be set via the `PRITUNL_URL` environment variable.
- `token` (String, Sensitive) API token for Pritunl authentication. Can also be set via the `PRITUNL_TOKEN` environment variable.
- `secret` (String, Sensitive) API secret for Pritunl authentication. Can also be set via the `PRITUNL_SECRET` environment variable.
- `insecure` (Boolean) Skip TLS certificate verification. Defaults to `false`. Can also be set via the `PRITUNL_INSECURE` environment variable.
