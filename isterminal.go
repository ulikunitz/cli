// SPDX-FileCopyrightText: © 2021 Ulrich Kunitz
//
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin || dragonfly || freebsd || (linux && !appengine) || netbsd || openbsd
// +build darwin dragonfly freebsd linux,!appengine netbsd openbsd

package cli

import (
	"syscall"
	"unsafe"
)

// IsTerminal returns true if the given file descriptor is a terminal.
func IsTerminal(fd uintptr) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd,
		ioctlGetTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}
