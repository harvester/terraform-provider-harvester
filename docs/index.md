---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "harvester Provider"
subcategory: ""
description: |-
  
---

# harvester Provider



## Example Usage

```terraform
terraform {
  required_version = ">= 0.13"
  required_providers {
    harvester = {
      source  = "harvester/harvester"
      version = "0.4.0"
    }
  }
}

provider "harvester" {
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- **kubeconfig** (String) harvester kubeconfig
