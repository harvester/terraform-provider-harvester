package util

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func NamespacedSchemaWrap(s map[string]*schema.Schema, system bool) {
	var namespace = constants.NamespaceDefault
	if system {
		namespace = constants.NamespaceHarvesterSystem
	}
	NonNamespacedSchemaWrap(s)
	s[constants.FieldCommonNamespace] = &schema.Schema{
		Type:         schema.TypeString,
		ForceNew:     true,
		Optional:     true,
		Default:      namespace,
		ValidateFunc: IsValidName,
	}
}

func NonNamespacedSchemaWrap(s map[string]*schema.Schema) {
	s[constants.FieldCommonName] = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: IsValidName,
		Description:  "A unique name",
	}
	s[constants.FieldCommonDescription] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Any text you want that better describes this resource",
	}
	s[constants.FieldCommonTags] = &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
	}
	s[constants.FieldCommonState] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
	s[constants.FieldCommonMessage] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
}

func IsValidName(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if errs := validation.IsDNS1123Subdomain(v); len(errs) > 0 {
		return nil, []error{fmt.Errorf("expected %q to be an valid DNS1123 subdomain", k)}
	}

	return nil, nil
}

func DataSourceSchemaWrap(s map[string]*schema.Schema) map[string]*schema.Schema {
	for k, v := range s {
		if k == constants.FieldCommonName || k == constants.FieldCommonNamespace {
			v.ForceNew = false
			continue
		}
		if v.Elem != nil {
			switch elem := v.Elem.(type) {
			case schema.Resource:
				v.Elem = DataSourceSchemaWrap(elem.Schema)
			}
		}
		v.ForceNew = false
		v.Computed = true
		v.Default = nil
		v.DefaultFunc = nil
		v.Optional = false
		v.Required = false
		v.ValidateFunc = nil
		v.ConflictsWith = nil
		v.MinItems = 0
		v.MaxItems = 0
	}
	return s
}
