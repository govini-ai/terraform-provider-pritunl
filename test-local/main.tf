terraform {
  required_providers {
    pritunl = {
      source  = "registry.terraform.io/govini-ai/pritunl"
      version = "~> 0.1"
    }
  }
}

provider "pritunl" {
  url      = "https://pritunl-gov-web-alb-1671640305.us-gov-west-1.elb.amazonaws.com"
  token    = var.pritunl_token
  secret   = var.pritunl_secret
  insecure = true # ALB with self-signed or internal cert
}

variable "pritunl_token" {
  type      = string
  sensitive = true
}

variable "pritunl_secret" {
  type      = string
  sensitive = true
}

# Fetch all organizations
data "pritunl_organizations" "all" {}

# Fetch all servers
data "pritunl_servers" "all" {}

# Fetch all hosts
data "pritunl_hosts" "all" {}

output "organizations" {
  value = [for org in data.pritunl_organizations.all.organizations : org.name]
}

output "servers" {
  value = [for server in data.pritunl_servers.all.servers : {
    name   = server.name
    status = server.status
  }]
}

output "hosts" {
  value = [for host in data.pritunl_hosts.all.hosts : {
    name   = host.name
    status = host.status
  }]
}
