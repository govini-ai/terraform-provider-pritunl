---
page_title: "pritunl_organizations Data Source - Pritunl"
description: |-
  Fetches all Pritunl organizations.
---

# pritunl_organizations (Data Source)

Fetches all Pritunl organizations.

## Example Usage

```terraform
data "pritunl_organizations" "all" {}
```

## Schema

### Read-Only

- `organizations` (List of Object) List of organizations. Each object contains:
  - `id` (String) Organization ID.
  - `name` (String) Organization name.
