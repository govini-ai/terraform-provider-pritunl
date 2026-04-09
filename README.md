# Terraform Provider for Pritunl

A Terraform provider for managing [Pritunl](https://pritunl.com/) VPN server resources.

## Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    pritunl = {
      source  = "govini-ai/pritunl"
      version = "~> 0.1"
    }
  }
}
```

Then run `terraform init`.

## Configuration

```hcl
provider "pritunl" {
  url      = "https://vpn.example.com"
  token    = var.pritunl_api_token
  secret   = var.pritunl_api_secret
  insecure = false
}
```

All attributes can also be set via environment variables:

| Attribute  | Environment Variable |
|------------|---------------------|
| `url`      | `PRITUNL_URL`       |
| `token`    | `PRITUNL_TOKEN`     |
| `secret`   | `PRITUNL_SECRET`    |
| `insecure` | `PRITUNL_INSECURE`  |

## Resources

| Resource | Description |
|----------|-------------|
| `pritunl_organization` | Manage organizations |
| `pritunl_server` | Manage VPN servers |
| `pritunl_route` | Manage routes on servers |
| `pritunl_user` | Manage users |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `pritunl_organization` / `pritunl_organizations` | Read organizations |
| `pritunl_server` / `pritunl_servers` | Read servers |
| `pritunl_host` / `pritunl_hosts` | Read hosts |
| `pritunl_user` / `pritunl_users` | Read users |

## Example

```hcl
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

resource "pritunl_route" "vpc" {
  server_id = pritunl_server.vpn.id
  network   = "10.0.0.0/16"
  comment   = "VPC access"
}

resource "pritunl_user" "developer" {
  organization_id = pritunl_organization.engineering.id
  name            = "john.doe"
  email           = "john@example.com"
}
```

See the `examples/` directory for more usage examples.

## Development

```bash
make build    # Build provider
make install  # Install locally
make test     # Run tests
make testacc  # Run acceptance tests (requires PRITUNL_URL, PRITUNL_TOKEN, PRITUNL_SECRET)
make fmt      # Format code
make lint     # Run linter
```

### Releasing

Releases are automated via GitHub Actions. To publish a new version:

1. Tag a commit: `git tag v0.1.0`
2. Push the tag: `git push origin v0.1.0`
3. The release workflow builds, signs, and publishes artifacts to GitHub Releases.
4. The Terraform Registry picks up the new release automatically.

## License

Apache-2.0 - see [LICENSE](LICENSE) for details.
