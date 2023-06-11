// SPDX-FileCopyrightText: © 2021 Ulrich Kunitz
//
// SPDX-License-Identifier: BSD-3-Clause

package cli

import (
	"syscall"
	"unsafe"
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

var getConsoleMode = kernel32.NewProc("GetConsoleMode")

// IsTerminal returns true if the given file descriptor is a terminal.
func IsTerminal(fd uintptr) bool {
	var st uint32
	r, _, e := syscall.Syscall(getConsoleMode.Addr(),
		2, fd, uintptr(unsafe.Pointer(&st)), 0)
	return r != 0 && e == 0
}
