package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"terminal_chess/internal/cli"
)

func main() {
	// handle graceful shutdown on keyboard interrupt
	context, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// pass context, cli args, stdin, stdout, stderr to runner to start game loop
	err := cli.Run(context, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

}
