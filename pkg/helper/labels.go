package helper

import "strings"

const (
	HarvesterLabelDomain = "harvesterhci.io"
)

/* Returns true if the argument `key` is a label key owned by Harvester rather
 * than by the user.
 *
 */
func IsReservedLabelKey(key string) bool {
	domain, _, found := strings.Cut(key, "/")
	if !found {
		return false
	}
	return domain == HarvesterLabelDomain || strings.HasSuffix(domain, "."+HarvesterLabelDomain)
}
