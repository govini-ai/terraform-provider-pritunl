# Add routes to a server
resource "pritunl_route" "vpc" {
  server_id = pritunl_server.production.id
  network   = "10.0.0.0/16"
  comment   = "Production VPC"
}

resource "pritunl_route" "private_subnet" {
  server_id = pritunl_server.production.id
  network   = "192.168.0.0/24"
  comment   = "Private subnet"
  nat       = true
}

# Reference an existing server
resource "pritunl_server" "production" {
  name    = "Production Access"
  network = "10.10.0.0/24"
  port    = 1194
}
