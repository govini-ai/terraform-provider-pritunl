# Fetch a server by name
data "pritunl_server" "production" {
  name = "Production Access"
}

# Fetch a server by ID
data "pritunl_server" "by_id" {
  id = "507f1f77bcf86cd799439011"
}

output "server_network" {
  value = data.pritunl_server.production.network
}

output "server_status" {
  value = data.pritunl_server.production.status
}
