# Create a VPN server
resource "pritunl_server" "production" {
  name     = "Production Access"
  network  = "10.10.0.0/24"
  port     = 1194
  protocol = "udp"
  ipv6     = true 

  # Attach organizations to this server
  attached_organization_ids = [
    pritunl_organization.engineering.id,
  ]

  # Optional settings
  cipher        = "aes256"
  hash          = "sha256"
  inter_client  = true
  ping_interval = 10
  ping_timeout  = 60
}

# Reference an existing organization
resource "pritunl_organization" "engineering" {
  name = "Engineering"
}

# Output the server ID
output "server_id" {
  value = pritunl_server.production.id
}

output "server_status" {
  value = pritunl_server.production.status
}
