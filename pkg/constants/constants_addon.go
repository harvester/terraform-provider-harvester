package constants

const (
	ResourceTypeAddon       = "harvester_addon"
	FieldAddonEnabled       = "enabled"
	FieldAddonValuesContent = "values_content"
	FieldAddonRepo          = "repo"
	FieldAddonChart         = "chart"
	FieldAddonVersion       = "version"

	// StateAddonInit represents the CRD AddonInitState, whose value is the
	// empty string and therefore cannot be used directly as a
	// retry.StateChangeConf state.
	StateAddonInit = "AddonInit"
)
