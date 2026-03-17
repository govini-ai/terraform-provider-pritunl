---
page_title: "pritunl_users Data Source - Pritunl"
description: |-
  Fetches all Pritunl users in an organization.
---

# pritunl_users (Data Source)

Fetches all Pritunl users in an organization.

## Example Usage

```terraform
data "pritunl_users" "engineering" {
  organization_id = data.pritunl_organization.engineering.id
}
```

## Schema

### Required

- `organization_id` (String) Organization ID to list users from.

### Read-Only

- `users` (List of Object) List of users. Each object contains:
  - `id` (String) User ID.
  - `name` (String) Username.
  - `email` (String) User email address.
  - `disabled` (Boolean) Whether the user is disabled.
  - `groups` (List of String) List of groups the user belongs to.
