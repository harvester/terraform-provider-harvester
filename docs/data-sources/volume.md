---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "harvester_volume Data Source - terraform-provider-harvester"
subcategory: ""
description: |-
  
---

# harvester_volume (Data Source)



## Example Usage

```terraform
data "harvester_volume" "mount-disk" {
  name      = "mount-disk"
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

- `access_mode` (String)
- `attached_vm` (String)
- `description` (String) Any text you want that better describes this resource
- `id` (String) The ID of this resource.
- `image` (String)
- `message` (String)
- `phase` (String)
- `size` (String)
- `state` (String)
- `storage_class_name` (String)
- `tags` (Map of String)
- `volume_mode` (String)
