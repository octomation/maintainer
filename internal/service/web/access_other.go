//go:build !unix

package web

import "io/fs"

// ownedByUser cannot check the owner here: the system has no Unix owner.
func ownedByUser(fs.FileInfo) bool { return true }
