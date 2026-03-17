# Fetch all users in an organization
data "pritunl_users" "engineering_users" {
  organization_id = data.pritunl_organization.engineering.id
}

# Reference to organization
data "pritunl_organization" "engineering" {
  name = "Engineering"
}

output "user_names" {
  value = [for user in data.pritunl_users.engineering_users.users : user.name]
}

output "active_users" {
  value = [for user in data.pritunl_users.engineering_users.users : user.name if !user.disabled]
}

output "user_count" {
  value = length(data.pritunl_users.engineering_users.users)
}
