package cli

import (
	"context"
	"io"
)

func Run(ctx context.Context, args []string, stdin io.Reader, stout, stdeer io.Writer) error {
	// TODO: implement this

	// set up cli flags

	// parse flags to set up game config

	// start game loop

	//// take player input

	//// validate player input (NOTE: call to methods in game)

	//// progress game state

	//// handle next move (AI) (NOTE: call to methods in game)

	//// progress game state

	return nil
}
