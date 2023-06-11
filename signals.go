// SPDX-FileCopyrightText: © 2021 Ulrich Kunitz
//
// SPDX-License-Identifier: BSD-3-Clause

package cli

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// WithTermSignals creates a context that is canceled, when the a termination
// signal is received.
func WithTermSignals(parent context.Context) (ctx context.Context, cancel context.CancelFunc) {
	ctx, cancel = context.WithCancel(parent)
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
		select {
		case <-ctx.Done():
		case sig := <-ch:
			log.Printf("received signal %s", sig)
			cancel()
		}
	}()
	return
}
