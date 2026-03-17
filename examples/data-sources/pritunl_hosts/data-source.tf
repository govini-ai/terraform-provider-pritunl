# Fetch all hosts
data "pritunl_hosts" "all" {}

output "host_names" {
  value = [for host in data.pritunl_hosts.all.hosts : host.name]
}

output "host_ips" {
  value = [for host in data.pritunl_hosts.all.hosts : host.public_addr]
}
