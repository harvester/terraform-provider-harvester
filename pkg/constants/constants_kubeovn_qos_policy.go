package constants

const (
	ResourceTypeKubeOVNQoSPolicy = "harvester_kubeovn_qos_policy"

	FieldKubeOVNQoSShared              = "shared"
	FieldKubeOVNQoSBindingType         = "binding_type"
	FieldKubeOVNQoSBandwidthLimitRules = "bandwidth_limit_rules"
	FieldKubeOVNQoSStatusShared        = "status_shared"
	FieldKubeOVNQoSStatusBindingType   = "status_binding_type"

	FieldKubeOVNQoSRuleName       = "name"
	FieldKubeOVNQoSRuleInterface  = "interface_name"
	FieldKubeOVNQoSRuleRateMax    = "rate_max"
	FieldKubeOVNQoSRuleBurstMax   = "burst_max"
	FieldKubeOVNQoSRulePriority   = "priority"
	FieldKubeOVNQoSRuleDirection  = "direction"
	FieldKubeOVNQoSRuleMatchType  = "match_type"
	FieldKubeOVNQoSRuleMatchValue = "match_value"
)
