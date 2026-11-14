//go:build !windows

package fsafe

import "syscall"

// FreeSpace returns the number of bytes available to the current user on the
// volume that holds dir.
func FreeSpace(dir string) (uint64, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(dir, &st); err != nil {
		return 0, err
	}
	return uint64(st.Bavail) * uint64(st.Bsize), nil
}
