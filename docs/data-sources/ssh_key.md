---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "harvester_ssh_key Data Source - terraform-provider-harvester"
subcategory: ""
description: |-
  
---

# harvester_ssh_key (Data Source)



## Example Usage

```terraform
data "harvester_ssh_key" "mysshkey" {
  name      = "mysshkey"
  namespace = "default"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) A unique name

### Optional

- `namespace` (String)

### Read-Only

- `description` (String) Any text you want that better describes this resource
- `fingerprint` (String)
- `id` (String) The ID of this resource.
- `message` (String)
- `public_key` (String)
- `state` (String)
- `tags` (Map of String)
