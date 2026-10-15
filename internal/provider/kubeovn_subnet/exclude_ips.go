package kubeovn_subnet

import (
	"net/netip"
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
	gateway := d.Get(constants.FieldKubeOVNSubnetGateway).(string)
	planned := excludeIPsFromSet(d.Get(constants.FieldKubeOVNSubnetExcludeIPs).(*schema.Set))
	if excludeIPsCover(planned, gateway) {
		return nil
	}
	return d.SetNew(constants.FieldKubeOVNSubnetExcludeIPs, append(planned, gateway))
}

// excludeIPsConfigKnown reports whether every configured entry is known at
// plan time; NewValueKnown only covers the set as a whole, not its elements.
func excludeIPsConfigKnown(d *schema.ResourceDiff) bool {
	raw := d.GetRawConfig()
	if raw.IsNull() {
		return true
	}
	return raw.GetAttr(constants.FieldKubeOVNSubnetExcludeIPs).IsWhollyKnown()
}

func excludeIPsFromSet(set *schema.Set) []string {
	entries := make([]string, 0, set.Len())
	for _, item := range set.List() {
		if entry := item.(string); entry != "" {
			entries = append(entries, entry)
		}
	}
	return entries
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
