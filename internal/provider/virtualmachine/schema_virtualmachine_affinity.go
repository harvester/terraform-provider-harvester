// Package virtualmachine provides the Terraform resource and data source
// implementations for Harvester virtual machines.
package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// This file contains schema definitions for VM scheduling affinity rules.
// These schemas map to Kubernetes affinity structures:
// - NodeAffinity: corev1.NodeAffinity
// - PodAffinity: corev1.PodAffinity
// - PodAntiAffinity: corev1.PodAntiAffinity
//
// Reference: https://docs.harvesterhci.io/v1.7/vm/index/#node-scheduling
// Kubernetes docs: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/

// resourceNodeSelectorRequirementSchema returns the schema for a node selector requirement (key/operator/values)
func resourceNodeSelectorRequirementSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldExpressionKey: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The label key that the selector applies to",
		},
		constants.FieldExpressionOperator: {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				"In", "NotIn", "Exists", "DoesNotExist", "Gt", "Lt",
			}, false),
			Description: "Operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist, Gt, and Lt",
		},
		constants.FieldExpressionValues: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "Values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty",
		},
	}
}

// resourceNodeSelectorTermSchema returns the schema for a node selector term (match_expressions/match_fields)
func resourceNodeSelectorTermSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldMatchExpressions: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "A list of node selector requirements by node's labels",
			Elem: &schema.Resource{
				Schema: resourceNodeSelectorRequirementSchema(),
			},
		},
		constants.FieldMatchFields: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "A list of node selector requirements by node's fields",
			Elem: &schema.Resource{
				Schema: resourceNodeSelectorRequirementSchema(),
			},
		},
	}
}

// resourcePreferredSchedulingTermSchema returns the schema for a preferred scheduling term (weight/preference)
func resourcePreferredSchedulingTermSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldPreferredWeight: {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(1, 100),
			Description:  "Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100",
		},
		constants.FieldPreferredPreference: {
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Description: "A node selector term, associated with the corresponding weight",
			Elem: &schema.Resource{
				Schema: resourceNodeSelectorTermSchema(),
			},
		},
	}
}

// resourceNodeAffinitySchema returns the schema for node affinity (required/preferred)
func resourceNodeAffinitySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldNodeAffinityRequired: {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldNodeSelectorTerm: {
						Type:        schema.TypeList,
						Required:    true,
						MinItems:    1,
						Description: "Required. A list of node selector terms. The terms are ORed",
						Elem: &schema.Resource{
							Schema: resourceNodeSelectorTermSchema(),
						},
					},
				},
			},
		},
		constants.FieldNodeAffinityPreferred: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field",
			Elem: &schema.Resource{
				Schema: resourcePreferredSchedulingTermSchema(),
			},
		},
	}
}

// resourceLabelSelectorSchema returns the schema for a label selector (match_labels/match_expressions)
func resourceLabelSelectorSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldMatchLabels: {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "A map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is \"key\", the operator is \"In\", and the values array contains only \"value\"",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldMatchExpressions: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "A list of label selector requirements. The requirements are ANDed",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldExpressionKey: {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The label key that the selector applies to",
					},
					constants.FieldExpressionOperator: {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice([]string{
							"In", "NotIn", "Exists", "DoesNotExist",
						}, false),
						Description: "Operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist",
					},
					constants.FieldExpressionValues: {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Description: "Values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty",
					},
				},
			},
		},
	}
}

// resourcePodAffinityTermSchema returns the schema for a pod affinity term
func resourcePodAffinityTermSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldLabelSelector: {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "A label query over a set of resources, in this case pods",
			Elem: &schema.Resource{
				Schema: resourceLabelSelectorSchema(),
			},
		},
		constants.FieldNamespaces: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldNamespaceSelector: {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field",
			Elem: &schema.Resource{
				Schema: resourceLabelSelectorSchema(),
			},
		},
		constants.FieldTopologyKey: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "TopologyKey is the key of node labels. Nodes that have a label with this key and identical values are considered to be in the same topology",
		},
	}
}

// resourceWeightedPodAffinityTermSchema returns the schema for a weighted pod affinity term
func resourceWeightedPodAffinityTermSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldPreferredWeight: {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(1, 100),
			Description:  "Weight associated with matching the corresponding podAffinityTerm, in the range 1-100",
		},
		constants.FieldPodAffinityTerm: {
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Description: "Required. A pod affinity term, associated with the corresponding weight",
			Elem: &schema.Resource{
				Schema: resourcePodAffinityTermSchema(),
			},
		},
	}
}

// resourcePodAffinitySchema returns the schema for pod affinity (required/preferred)
// This is reused for both pod_affinity and pod_anti_affinity
func resourcePodAffinitySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldPodAffinityRequired: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node",
			Elem: &schema.Resource{
				Schema: resourcePodAffinityTermSchema(),
			},
		},
		constants.FieldPodAffinityPreferred: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field",
			Elem: &schema.Resource{
				Schema: resourceWeightedPodAffinityTermSchema(),
			},
		},
	}
}
