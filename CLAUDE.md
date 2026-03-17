# Terraform Provider for Pritunl

## Project Structure

```
├── main.go                    # Provider entry point, serves the provider via gRPC
├── internal/
│   ├── provider/provider.go   # Provider configuration (url, token, secret, insecure)
│   ├── client/                # Pritunl API client with HMAC-SHA256 authentication
│   ├── resources/             # Terraform resources (organization, server, route, user)
│   └── datasources/           # Terraform data sources (organization(s), server(s), host(s), user(s))
├── examples/                  # Example Terraform configurations
├── test-local/                # Local development Terraform config
├── .goreleaser.yml            # Release build configuration
├── .github/workflows/         # CI/CD (release on tag push)
└── terraform-registry-manifest.json  # Required by Terraform Registry
```

## Build and Test

```bash
make build      # Compile the provider binary
make install    # Install to ~/.terraform.d/plugins/ for local testing
make test       # Run unit tests
make testacc    # Run acceptance tests (needs PRITUNL_URL, PRITUNL_TOKEN, PRITUNL_SECRET)
make fmt        # go fmt
make lint       # golangci-lint
```

## Provider Details

- **Registry address**: `registry.terraform.io/govini-ai/pritunl`
- **Go module**: `github.com/govini-ai/terraform-provider-pritunl`
- **Framework**: Terraform Plugin Framework (not the older SDKv2)
- **Auth**: HMAC-SHA256 signatures over `token&timestamp&nonce&METHOD&path`

## Releasing

Push a `v*` tag to trigger the GitHub Actions release workflow. GoReleaser builds multi-platform binaries, signs checksums with GPG, and publishes to GitHub Releases. The Terraform Registry syncs automatically.

Required GitHub secrets: `GPG_PRIVATE_KEY`, `PASSPHRASE`.
