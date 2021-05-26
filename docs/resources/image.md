---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "harvester_image Resource - terraform-provider-harvester"
subcategory: ""
description: |-
  
---

# harvester_image (Resource)



## Example Usage

```terraform
resource "harvester_image" "k3os" {
  name         = "k3os"
  display_name = "k3os"
  url          = "https://github.com/rancher/k3os/releases/download/v0.20.6-k3s1r0/k3os-amd64.iso"
}

resource "harvester_image" "ubuntu20" {
  name         = "ubuntu20"
  display_name = "ubuntu20"
  url          = "http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **display_name** (String)
- **name** (String) A unique name
- **url** (String) supports the `raw` and `qcow2` image formats which are supported by [qemu](https://www.qemu.org/docs/master/system/images.html#disk-image-file-formats). Bootable ISO images can also be used and are treated like `raw` images.

### Optional

- **description** (String) Any text you want that better describes this resource
- **id** (String) The ID of this resource.
- **namespace** (String)
- **tags** (Map of String)

### Read-Only

- **size** (Number)
- **state** (String)

## Import

Import is supported using the following syntax:

```shell
terraform import harvester_image.foo <Namespace>/<Name>
```