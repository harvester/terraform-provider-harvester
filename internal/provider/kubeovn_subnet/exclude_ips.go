package kubeovn_subnet

import (
	"net/netip"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// The kube-ovn controller rewrites spec.excludeIps on every reconcile
// (checkAndUpdateExcludeIPs in pkg/controller/subnet.go): overlapping entries
// are merged, a range of one address is written as that address, every
// gateway address that no entry covers is added, and the list is sorted.
// Planning the same value up front keeps the state free of perpetual diffs.
func planExcludeIPs(d *schema.ResourceDiff) error {
	if !d.NewValueKnown(constants.FieldKubeOVNSubnetGateway) || !excludeIPsConfigKnown(d) {
		return d.SetNewComputed(constants.FieldKubeOVNSubnetExcludeIPs)
	}
	var configured []string
	if excludeIPsConfigured(d) {
		configured = excludeIPsFromSet(d.Get(constants.FieldKubeOVNSubnetExcludeIPs).(*schema.Set))
	}
	gateway := d.Get(constants.FieldKubeOVNSubnetGateway).(string)
	return d.SetNew(constants.FieldKubeOVNSubnetExcludeIPs, normalizeExcludeIPs(configured, gateway))
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

// excludeIPsConfigured reports whether exclude_ips is set in the
// configuration. As the attribute is also computed, the SDK otherwise keeps
// the stored value when it is removed, and the extra entries would stay.
func excludeIPsConfigured(d *schema.ResourceDiff) bool {
	raw := d.GetRawConfig()
	if raw.IsNull() {
		return true
	}
	return !raw.GetAttr(constants.FieldKubeOVNSubnetExcludeIPs).IsNull()
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

type ipRange struct {
	start, end netip.Addr
}

// normalizeExcludeIPs returns entries the way kube-ovn stores them, with the
// gateway addresses (comma-separated on a dual-stack subnet) reserved.
// Entries that do not parse are kept as they are: the kube-ovn webhook
// rejects them.
func normalizeExcludeIPs(entries []string, gateway string) []string {
	var (
		ranges     []ipRange
		normalized []string
	)
	for _, entry := range entries {
		if r, ok := parseExcludeEntry(entry); ok {
			ranges = append(ranges, r)
		} else {
			normalized = append(normalized, entry)
		}
	}
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].start.Less(ranges[j].start) })
	var merged []ipRange
	for _, r := range ranges {
		last := len(merged) - 1
		if last >= 0 && r.start.Is4() == merged[last].end.Is4() && r.start.Compare(merged[last].end) <= 0 {
			if r.end.Compare(merged[last].end) > 0 {
				merged[last].end = r.end
			}
			continue
		}
		merged = append(merged, r)
	}
	for _, r := range merged {
		if r.start == r.end {
			normalized = append(normalized, r.start.String())
		} else {
			normalized = append(normalized, r.start.String()+".."+r.end.String())
		}
	}
	for _, gw := range strings.Split(gateway, ",") {
		if gw = strings.TrimSpace(gw); gw != "" && !excludeIPsCover(merged, gw) {
			normalized = append(normalized, gw)
		}
	}
	sort.Strings(normalized)
	return normalized
}

func excludeIPsCover(ranges []ipRange, ip string) bool {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}
	for _, r := range ranges {
		if addr.Compare(r.start) >= 0 && addr.Compare(r.end) <= 0 {
			return true
		}
	}
	return false
}

// parseExcludeEntry parses a single address or a "start..end" range.
func parseExcludeEntry(entry string) (ipRange, bool) {
	bounds := strings.SplitN(entry, "..", 2)
	start, err := netip.ParseAddr(strings.TrimSpace(bounds[0]))
	if err != nil {
		return ipRange{}, false
	}
	end := start
	if len(bounds) == 2 {
		if end, err = netip.ParseAddr(strings.TrimSpace(bounds[1])); err != nil {
			return ipRange{}, false
		}
	}
	if end.Less(start) || start.Is4() != end.Is4() {
		return ipRange{}, false
	}
	return ipRange{start: start, end: end}, true
}
