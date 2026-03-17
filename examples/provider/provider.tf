terraform {
  required_providers {
    pritunl = {
      source  = "registry.terraform.io/govini-ai/pritunl"
      version = "~> 0.1"
    }
  }
}

provider "pritunl" {
  url      = var.pritunl_url
  token    = var.pritunl_api_token
  secret   = var.pritunl_api_secret
  insecure = true # Skip TLS verification for self-signed certs
}

variable "pritunl_url" {
  description = "Pritunl server URL"
  type        = string
}

variable "pritunl_api_token" {
  description = "Pritunl API token"
  type        = string
  sensitive   = true
}

variable "pritunl_api_secret" {
  description = "Pritunl API secret"
  type        = string
  sensitive   = true
}
