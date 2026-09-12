package constants

const (
	ResourceTypeKubeOVNSecurityGroup = "harvester_kubeovn_security_group"

	FieldKubeOVNSGAllowSameGroupTraffic = "allow_same_group_traffic"
	FieldKubeOVNSGIngressRules          = "ingress_rules"
	FieldKubeOVNSGEgressRules           = "egress_rules"
	FieldKubeOVNSGStatusPortGroup       = "status_port_group"
	FieldKubeOVNSGStatusIngressMD5      = "status_ingress_md5"
	FieldKubeOVNSGStatusEgressMD5       = "status_egress_md5"
	FieldKubeOVNSGStatusIngressSynced   = "status_ingress_last_sync_success"
	FieldKubeOVNSGStatusEgressSynced    = "status_egress_last_sync_success"

	FieldKubeOVNSGRuleIPVersion           = "ip_version"
	FieldKubeOVNSGRuleProtocol            = "protocol"
	FieldKubeOVNSGRulePriority            = "priority"
	FieldKubeOVNSGRuleRemoteType          = "remote_type"
	FieldKubeOVNSGRuleRemoteAddress       = "remote_address"
	FieldKubeOVNSGRuleRemoteSecurityGroup = "remote_security_group"
	FieldKubeOVNSGRulePortRangeMin        = "port_range_min"
	FieldKubeOVNSGRulePortRangeMax        = "port_range_max"
	FieldKubeOVNSGRulePolicy              = "policy"
)
