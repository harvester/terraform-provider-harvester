package constants

const (
	ResourceTypeKubeOVNSwitchLBRule = "harvester_kubeovn_switch_lb_rule"

	FieldKubeOVNSwitchLBRuleVip             = "vip"
	FieldKubeOVNSwitchLBRuleNamespace       = "rule_namespace"
	FieldKubeOVNSwitchLBRuleSelector        = "selector"
	FieldKubeOVNSwitchLBRuleEndpoints       = "endpoints"
	FieldKubeOVNSwitchLBRuleSessionAffinity = "session_affinity"
	FieldKubeOVNSwitchLBRulePorts           = "ports"
	FieldKubeOVNSwitchLBRuleStatusPorts     = "status_ports"
	FieldKubeOVNSwitchLBRuleStatusService   = "status_service"

	FieldKubeOVNSlrPortName       = "name"
	FieldKubeOVNSlrPortPort       = "port"
	FieldKubeOVNSlrPortTargetPort = "target_port"
	FieldKubeOVNSlrPortProtocol   = "protocol"
)
