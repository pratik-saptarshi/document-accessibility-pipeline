//go:build !linux

package main

import "os"

// Unsupported platforms report no count. Callers must fail closed instead of
// assuming that an unreadable link count means the file is singly linked.
func nlink(os.FileInfo) (uint64, bool) { return 0, false }
