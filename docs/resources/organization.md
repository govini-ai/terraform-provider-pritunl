---
page_title: "pritunl_organization Resource - Pritunl"
description: |-
  Manages a Pritunl organization.
---

# pritunl_organization (Resource)

Manages a Pritunl organization.

## Example Usage

```terraform
resource "pritunl_organization" "engineering" {
  name = "Engineering"
}
```

## Schema

### Required

- `name` (String) Organization name.

### Read-Only

- `id` (String) Organization ID.

## Import

Organizations can be imported using their ID:

```shell
terraform import pritunl_organization.example <organization_id>
```
