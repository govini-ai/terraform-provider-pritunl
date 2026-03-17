# Fetch all organizations
data "pritunl_organizations" "all" {}

output "organization_names" {
  value = [for org in data.pritunl_organizations.all.organizations : org.name]
}

output "organization_count" {
  value = length(data.pritunl_organizations.all.organizations)
}
