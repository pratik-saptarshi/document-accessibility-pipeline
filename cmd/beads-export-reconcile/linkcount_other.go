//go:build !linux

package main

import "os"

func nlink(os.FileInfo) uint64 { return 1 }
