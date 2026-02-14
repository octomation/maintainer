//go:build unix

package web

import (
	"io/fs"
	"os"
	"syscall"
)

// ownedByUser reports whether the file belongs to the user that runs the
// process.
func ownedByUser(info fs.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && int(stat.Uid) == os.Getuid()
}
