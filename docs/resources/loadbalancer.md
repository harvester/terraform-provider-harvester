---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "harvester_loadbalancer Resource - terraform-provider-harvester"
subcategory: ""
description: |-
  
---

# harvester_loadbalancer (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **listener** (Block List, Min: 1) (see [below for nested schema](#nestedblock--listener))
- **name** (String) A unique name

### Optional

- **backend_selector** (Block Set) (see [below for nested schema](#nestedblock--backend_selector))
- **description** (String) Any text you want that better describes this resource
- **healthcheck** (Block List, Max: 1) (see [below for nested schema](#nestedblock--healthcheck))
- **id** (String) The ID of this resource.
- **ipam** (String) Where the load balancer gets its IP address from. Can be `dhcp` or `pool`.
- **ippool** (String) Which IP pool to get the IP address from.
- **namespace** (String)
- **tags** (Map of String)
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **workload_type** (String) Can be `vm` or `cluster`

### Read-Only

- **message** (String)
- **state** (String)

<a id="nestedblock--listener"></a>
### Nested Schema for `listener`

Required:

- **backend_port** (Number)
- **port** (Number)
- **protocol** (String)

Optional:

- **name** (String)


<a id="nestedblock--backend_selector"></a>
### Nested Schema for `backend_selector`

Required:

- **key** (String)
- **values** (List of String)


<a id="nestedblock--healthcheck"></a>
### Nested Schema for `healthcheck`

Required:

- **port** (Number)

Optional:

- **failure_threshold** (Number)
- **period_seconds** (Number)
- **success_threshold** (Number)
- **timeout_seconds** (Number)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **default** (String)
- **delete** (String)
- **read** (String)
- **update** (String)


