package helper

import "strings"

const (
	IPv4LinkLocalCIDRPrefix = "169.254."
	IPv6LinkLocalCIDRPrefix = "fe80:"
)

/* Returns true if the argument `ip` is a link-local IPv4 address in the common
 * CIDR notation
 *
 * Link-local IPv4 addresses are 169.254.0.0/16 except 169.254.0.0/24 and
 * 169.254.255.0/24 according to RFC 3927
 */
func IsIPv4LinkLocal(ip string) bool {
	if strings.HasPrefix(strings.ToLower(ip), IPv4LinkLocalCIDRPrefix) {
		if !strings.HasPrefix(strings.ToLower(ip), IPv4LinkLocalCIDRPrefix+"0.") &&
			!strings.HasPrefix(strings.ToLower(ip), IPv4LinkLocalCIDRPrefix+"255.") {
			return true
		}
	}
	return false
}

/* Returns true if the argument `ip` is a link-local IPv6 address in the common
 * CIDR notation
 *
 * Link-local IPv6 addresses are fe80::/64
 */
func IsIPv6LinkLocal(ip string) bool {
	return strings.HasPrefix(strings.ToLower(ip), IPv6LinkLocalCIDRPrefix)
}
