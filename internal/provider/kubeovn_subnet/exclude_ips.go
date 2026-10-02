package kubeovn_subnet

import (
	"net/netip"
	"slices"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// The kube-ovn controller rewrites spec.excludeIps on every reconcile
// (formatExcludeIPs in pkg/controller/subnet.go): the gateway address is
// reserved unless an entry already covers it, and the list is sorted.
// Planning the same value up front keeps the state free of perpetual diffs.
func planExcludeIPs(d *schema.ResourceDiff) error {
	if !d.NewValueKnown(constants.FieldKubeOVNSubnetGateway) || !excludeIPsConfigKnown(d) {
		return d.SetNewComputed(constants.FieldKubeOVNSubnetExcludeIPs)
	}
	gateway, _ := d.Get(constants.FieldKubeOVNSubnetGateway).(string)
	planned := excludeIPsFromSet(d.Get(constants.FieldKubeOVNSubnetExcludeIPs))
	normalized := withGatewayExcluded(planned, gateway)
	sort.Strings(planned)
	if slices.Equal(planned, normalized) {
		return nil
	}
	return d.SetNew(constants.FieldKubeOVNSubnetExcludeIPs, normalized)
}

// excludeIPsConfigKnown reports whether every configured entry is known at
// plan time; NewValueKnown only covers the set as a whole, not its elements.
func excludeIPsConfigKnown(d *schema.ResourceDiff) bool {
	raw := d.GetRawConfig()
	if raw.IsNull() || !raw.Type().IsObjectType() || !raw.Type().HasAttribute(constants.FieldKubeOVNSubnetExcludeIPs) {
		return true
	}
	return raw.GetAttr(constants.FieldKubeOVNSubnetExcludeIPs).IsWhollyKnown()
}

func excludeIPsFromSet(value interface{}) []string {
	set, ok := value.(*schema.Set)
	if !ok || set == nil {
		return nil
	}
	entries := make([]string, 0, set.Len())
	for _, item := range set.List() {
		if entry, ok := item.(string); ok && entry != "" {
			entries = append(entries, entry)
		}
	}
	return entries
}

// withGatewayExcluded returns the entries plus every gateway address not
// already covered by an entry, deduplicated and sorted like kube-ovn does.
func withGatewayExcluded(entries []string, gateway string) []string {
	result := make([]string, 0, len(entries)+2)
	seen := make(map[string]struct{}, len(entries)+2)
	add := func(entry string) {
		if _, duplicate := seen[entry]; duplicate {
			return
		}
		seen[entry] = struct{}{}
		result = append(result, entry)
	}
	for _, entry := range entries {
		add(entry)
	}
	for _, gw := range gatewayAddresses(gateway) {
		if !excludeIPsCover(result, gw) {
			add(gw)
		}
	}
	sort.Strings(result)
	return result
}

// gatewayAddresses splits a kube-ovn gateway, which is comma-separated for
// dual-stack subnets.
func gatewayAddresses(gateway string) []string {
	var addresses []string
	for _, gw := range strings.Split(gateway, ",") {
		if gw = strings.TrimSpace(gw); gw != "" {
			addresses = append(addresses, gw)
		}
	}
	return addresses
}

func excludeIPsCover(entries []string, ip string) bool {
	for _, entry := range entries {
		if excludeEntryCovers(entry, ip) {
			return true
		}
	}
	return false
}

// excludeEntryCovers reports whether entry, a single address or a
// "start..end" range, contains ip.
func excludeEntryCovers(entry, ip string) bool {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}
	bounds := strings.SplitN(entry, "..", 2)
	start, err := netip.ParseAddr(strings.TrimSpace(bounds[0]))
	if err != nil {
		return false
	}
	end := start
	if len(bounds) == 2 {
		if end, err = netip.ParseAddr(strings.TrimSpace(bounds[1])); err != nil {
			return false
		}
	}
	return addr.Compare(start) >= 0 && addr.Compare(end) <= 0
}
