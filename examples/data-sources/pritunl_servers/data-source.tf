# Fetch all servers
data "pritunl_servers" "all" {}

output "server_names" {
  value = [for server in data.pritunl_servers.all.servers : server.name]
}

output "online_servers" {
  value = [for server in data.pritunl_servers.all.servers : server.name if server.status == "online"]
}
