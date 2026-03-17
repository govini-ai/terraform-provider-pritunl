# Fetch an organization by name
data "pritunl_organization" "engineering" {
  name = "Engineering"
}

# Fetch an organization by ID
data "pritunl_organization" "by_id" {
  id = "507f1f77bcf86cd799439011"
}

output "org_id" {
  value = data.pritunl_organization.engineering.id
}
