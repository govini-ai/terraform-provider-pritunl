---
page_title: "pritunl_user Data Source - Pritunl"
description: |-
  Fetches a Pritunl user by ID or name.
---

# pritunl_user (Data Source)

Fetches a Pritunl user by ID or name within an organization. One of `id` or `name` must be specified.

## Example Usage

```terraform
data "pritunl_user" "developer" {
  organization_id = data.pritunl_organization.engineering.id
  name            = "john.doe"
}
```

## Schema

### Required

- `organization_id` (String) Organization ID the user belongs to.

### Optional

- `id` (String) User ID.
- `name` (String) Username.

### Read-Only

- `email` (String) User email address.
- `disabled` (Boolean) Whether the user is disabled.
- `groups` (List of String) List of groups the user belongs to.
