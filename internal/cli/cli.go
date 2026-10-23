package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"terminal_chess/internal/game"
)

type Color string

const (
	Black Color = "black"
	White Color = "white"
)

// implement the flag.Value interface with String() and Set()
func (t *Color) String() string {
	if *t == "" {
		return string(White) // default fallback represenation for usage output
	}
	return string(*t)
}

func (t *Color) Set(value string) error {
	normalized := Color(strings.ToLower(value))
	switch normalized {
	case Black, White:
		*t = normalized
		return nil
	default:
		return fmt.Errorf("Invalid team %q: must be either 'black' or 'white'", value)
	}
}

func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	// TODO: implement this

	// set up cli flags
	flagSet := flag.NewFlagSet("cli", flag.ContinueOnError)

	flagSet.SetOutput(stderr)

	var difficulty string
	flagSet.StringVar(&difficulty, "difficulty", "medium", "Computer opponent difficulty level")

	var timed bool
	flagSet.BoolVar(&timed, "timed", false, "Enable timed play mode")

	var timeLimit int
	flagSet.IntVar(&timeLimit, "time-limit", 30, "Time limit for player turn in seconds (only valid if --timed is true)")

	var playerColor Color = White // set default value
	flagSet.Var(&playerColor, "color", "Select a color to play as: black or white. White moves first.")

	flagSet.Usage = func() {
		fmt.Fprint(stderr, "Terminal Chess usage:\n")
		flagSet.PrintDefaults()
	}

	// parse flags to set up game config
	if err := flagSet.Parse(args); err != nil {
		// user inputting -h or --help is not an error and does not require an exit
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	// validate time-limit arg
	if timed {
		if timeLimit <= 0 {
			return fmt.Errorf("Invalid time-limit %d: must be greater than 0 when --timed is true", timeLimit)
		}
		fmt.Fprintf(stdout, "Game mode: Timed (%d seconds)\n", timeLimit)
	}

	isPlayerTurn := (playerColor == White) //  White goes first

	// initialize game state
	var state = game.NewGameState()

	// start game loop
	fmt.Fprintln(stdout, "Starting game...")
	for {

		// check context for interrupt
		select {
		case <-ctx.Done():
			fmt.Fprintln(stdout, "\nGame interrupted. Exiting...")
			return ctx.Err()
		default:
			// do nothing and continue
		}

		// check game state for end conditions
		if state.IsGameOver() {
			fmt.Fprintln(stdout, "Game over!")
			break
		}

		var err error
		if isPlayerTurn {
			state, err = game.HandlePlayerTurn(ctx, state, stdin, stdout, stderr)
			if err != nil {
				return err
			}
		} else {
			state, err = game.HandleComputerTurn(ctx, state, stdout)
			if err != nil {
				return err
			}
		}

		// rotate turn
		isPlayerTurn = !isPlayerTurn

	}
	return nil
}
