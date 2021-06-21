package util

import (
	"reflect"
	"strings"

	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

type Constructor interface {
	Setup() Processors
	Result() (interface{}, error)
}

type Processor struct {
	Parser   func(interface{}) error
	Field    string
	Required bool
}

type Processors []Processor

func (p Processors) String(key string, dst *string, required bool) Processors {
	return append(p, Processor{
		Field: key,
		Parser: func(i interface{}) error {
			*dst = i.(string)
			return nil
		},
		Required: required,
	})
}

func (p Processors) Bool(key string, dst *bool, required bool) Processors {
	return append(p, Processor{
		Field: key,
		Parser: func(i interface{}) error {
			*dst = i.(bool)
			return nil
		},
		Required: required,
	})
}

func (p Processors) Description(annotations *map[string]string) Processors {
	return append(p, Processor{
		Field: constants.FieldCommonDescription,
		Parser: func(i interface{}) error {
			*annotations = mapMerge(*annotations, builder.AnnotationPrefixCattleField, map[string]interface{}{
				constants.FieldCommonDescription: i,
			})
			return nil
		},
	})
}

func (p Processors) Tags(labels *map[string]string) Processors {
	for labelKey := range *labels {
		if strings.HasPrefix(labelKey, builder.LabelPrefixHarvesterTag) {
			delete(*labels, labelKey)
		}
	}
	return append(p, Processor{
		Field: constants.FieldCommonTags,
		Parser: func(i interface{}) error {
			*labels = mapMerge(*labels, builder.LabelPrefixHarvesterTag, i.(map[string]interface{}))
			return nil
		},
	})
}

func NewProcessors() Processors {
	return []Processor{}
}

func ResourceConstruct(d *schema.ResourceData, c Constructor) (interface{}, error) {
	for _, processor := range c.Setup() {
		var (
			value interface{}
			ok    bool
		)
		if processor.Required {
			value = d.Get(processor.Field)
		} else {
			value, ok = d.GetOk(processor.Field)
			if !ok {
				continue
			}
		}
		reflectVal := reflect.ValueOf(value)
		if reflectVal.Kind() == reflect.Slice {
			for _, item := range value.([]interface{}) {
				if err := processor.Parser(item); err != nil {
					return nil, err
				}
			}
		} else {
			if err := processor.Parser(value); err != nil {
				return nil, err
			}
		}
	}
	return c.Result()
}

func mapMerge(dst map[string]string, prefix string, values map[string]interface{}) map[string]string {
	if dst == nil {
		dst = map[string]string{}
	}
	for key, value := range values {
		dst[prefix+key] = value.(string)
	}
	return dst
}

func NewObjectMeta(namespace, name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        name,
		Namespace:   namespace,
		Labels:      map[string]string{},
		Annotations: map[string]string{},
	}
}

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}
