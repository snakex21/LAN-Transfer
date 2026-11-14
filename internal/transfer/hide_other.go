//go:build !windows

package transfer

// hideFile is a no-op: the folder name already starts with a dot.
func hideFile(string) {}
