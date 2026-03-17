# Create a user in an organization
resource "pritunl_user" "developer" {
  organization_id = pritunl_organization.engineering.id
  name            = "john.doe"
  email           = "john.doe@example.com"
}

# Create a disabled user
resource "pritunl_user" "contractor" {
  organization_id = pritunl_organization.engineering.id
  name            = "jane.contractor"
  email           = "jane@contractor.com"
  disabled        = true
}

# Create a user with groups
resource "pritunl_user" "admin" {
  organization_id = pritunl_organization.engineering.id
  name            = "admin.user"
  email           = "admin@example.com"
  groups          = ["admins", "developers"]
}

# Reference an existing organization
resource "pritunl_organization" "engineering" {
  name = "Engineering"
}
