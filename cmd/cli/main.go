package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/k1lgor/container-diet/internal/cli"
)

func main() {
	// Handle graceful shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		fmt.Fprintf(os.Stderr, "\nReceived signal: %v. Shutting down.\n", sig)
		os.Exit(130)
	}()

	cli.Execute()
}
