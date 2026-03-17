---
page_title: "pritunl_organization Data Source - Pritunl"
description: |-
  Fetches a Pritunl organization by ID or name.
---

# pritunl_organization (Data Source)

Fetches a Pritunl organization by ID or name. One of `id` or `name` must be specified.

## Example Usage

```terraform
data "pritunl_organization" "engineering" {
  name = "Engineering"
}
```

## Schema

### Optional

- `id` (String) Organization ID.
- `name` (String) Organization name.
