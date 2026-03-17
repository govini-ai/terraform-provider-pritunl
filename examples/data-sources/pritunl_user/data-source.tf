# Fetch a user by name
data "pritunl_user" "developer" {
  organization_id = data.pritunl_organization.engineering.id
  name            = "john.doe"
}

# Fetch a user by ID
data "pritunl_user" "by_id" {
  organization_id = data.pritunl_organization.engineering.id
  id              = "507f1f77bcf86cd799439011"
}

# Reference to organization
data "pritunl_organization" "engineering" {
  name = "Engineering"
}

output "user_email" {
  value = data.pritunl_user.developer.email
}

output "user_disabled" {
  value = data.pritunl_user.developer.disabled
}
