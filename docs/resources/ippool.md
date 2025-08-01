---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "harvester_ippool Resource - terraform-provider-harvester"
subcategory: ""
description: |-
  
---

# harvester_ippool (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **name** (String) A unique name
- **range** (Block List, Min: 1) IP Range belonging to this pool, can be given multiple times (see [below for nested schema](#nestedblock--range))

### Optional

- **description** (String) Any text you want that better describes this resource
- **id** (String) The ID of this resource.
- **selector** (Block List, Max: 1) (see [below for nested schema](#nestedblock--selector))
- **tags** (Map of String)
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- **message** (String)
- **state** (String)

<a id="nestedblock--range"></a>
### Nested Schema for `range`

Required:

- **end** (String)
- **gateway** (String)
- **start** (String)
- **subnet** (String)


<a id="nestedblock--selector"></a>
### Nested Schema for `selector`

Optional:

- **network** (String) Namespace/name of the VM network
- **priority** (Number) Priority of the IP pool. Large numbers have higher priority
- **scope** (Block List) Scope of the IP pool (see [below for nested schema](#nestedblock--selector--scope))

<a id="nestedblock--selector--scope"></a>
### Nested Schema for `selector.scope`

Optional:

- **guest_cluster** (String) Name of the guest cluster
- **namespace** (String) Namespace of the VMs of the guest cluster
- **project** (String) Name of the project



<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **default** (String)
- **delete** (String)
- **read** (String)
- **update** (String)


