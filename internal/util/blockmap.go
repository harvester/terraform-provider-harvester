package util

// BlockMap returns the content of a block. Terraform passes nil for a block
// written without any attribute, read as an empty map here.
func BlockMap(i any) map[string]any {
	r, _ := i.(map[string]any)
	return r
}
