package image

import (
	"fmt"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// validateSecurityParameters validates the security parameters block using ValidateFunc
func validateSecurityParameters(v interface{}, k string) (warnings []string, errors []error) {
	securityParams, ok := v.(map[string]interface{})
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be map", k)}
	}

	// If empty map, no validation needed
	if len(securityParams) == 0 {
		return nil, nil
	}

	// Check required fields
	requiredFields := map[string]struct{}{
		constants.FieldImageCryptoOperation:      {},
		constants.FieldImageSourceImageName:      {},
		constants.FieldImageSourceImageNamespace: {},
	}

	for fieldKey := range requiredFields {
		value, exists := securityParams[fieldKey]
		if !exists || value == "" {
			errors = append(errors, fmt.Errorf("%s is required in security_parameters", fieldKey))
		}
	}

	return warnings, errors
}
