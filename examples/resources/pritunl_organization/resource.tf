# Create an organization
resource "pritunl_organization" "engineering" {
  name = "Engineering"
}

# Output the organization ID
output "organization_id" {
  value = pritunl_organization.engineering.id
}
