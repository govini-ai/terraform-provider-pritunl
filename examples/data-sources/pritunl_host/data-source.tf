# Fetch a host by name
data "pritunl_host" "primary" {
  name = "vpn-server-1"
}

# Fetch a host by ID
data "pritunl_host" "by_id" {
  id = "507f1f77bcf86cd799439011"
}

output "host_public_ip" {
  value = data.pritunl_host.primary.public_addr
}

output "host_version" {
  value = data.pritunl_host.primary.version
}
