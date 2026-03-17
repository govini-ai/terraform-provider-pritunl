---
page_title: "pritunl_user Resource - Pritunl"
description: |-
  Manages a Pritunl user.
---

# pritunl_user (Resource)

Manages a Pritunl user.

## Example Usage

```terraform
resource "pritunl_user" "developer" {
  organization_id = pritunl_organization.engineering.id
  name            = "john.doe"
  email           = "john@example.com"
}
```

## Schema

### Required

- `organization_id` (String) Organization ID the user belongs to. Changing this forces a new resource.
- `name` (String) Username.

### Optional

- `email` (String) User email address.
- `disabled` (Boolean) Whether the user is disabled. Defaults to `false`.
- `groups` (List of String) List of groups the user belongs to.

### Read-Only

- `id` (String) User ID.

## Import

Users can be imported using `organization_id/user_id`:

```shell
terraform import pritunl_user.example <organization_id>/<user_id>
```
