package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
	"terminal_chess/internal/game"
	"time"

	"github.com/corentings/chess/v2"

	"github.com/corentings/chess/v2/uci"
)

func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {

	// set up cli flags
	flagSet := flag.NewFlagSet("cli", flag.ContinueOnError)

	flagSet.SetOutput(stderr)

	var difficulty string
	flagSet.StringVar(&difficulty, "difficulty", "medium", "Computer opponent difficulty level")

	var timed bool
	flagSet.BoolVar(&timed, "timed", false, "Enable timed play mode")

	var timeLimit int
	flagSet.IntVar(&timeLimit, "time-limit", 30, "Time limit for player turn in seconds (only valid if -timed is true)")

	var color string
	flagSet.StringVar(&color, "color", "white", "Select a color to play as: [black | white]")

	var showBoard bool
	flagSet.BoolVar(&showBoard, "show-board", true, "Show board state after each move")

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
			return fmt.Errorf("Invalid time-limit %d: must be greater than 0 when -timed is true", timeLimit)
		}
	}

	// handle player color
	var playerColor chess.Color
	switch strings.ToLower(color) {
	case "white":
		playerColor = chess.White
	case "black":
		playerColor = chess.Black
	default:
		return fmt.Errorf("invalid color: %s", color)
	}

	playerTurn := (playerColor == chess.White) //  White goes first

	// initialize game state
	state := chess.NewGame()

	// set up computer opponent
	var skillLevel string
	switch strings.ToLower(difficulty) {
	case "easy":
		skillLevel = "2"
	case "medium":
		skillLevel = "10"
	case "hard":
		skillLevel = "20"
	default:
		return fmt.Errorf("inavlid difficulty %q: must be one of [easy, medium, hard]", difficulty)
	}

	eng, err := uci.New("stockfish")
	if err != nil {
		return fmt.Errorf("failed to start UCI engine: %w", err)
	}
	defer eng.Close()

	setSkill := uci.CmdSetOption{
		Name:  "Skill Level",
		Value: skillLevel,
	}
	if err := eng.Run(uci.CmdUCI, setSkill, uci.CmdIsReady, uci.CmdUCINewGame); err != nil {
		return fmt.Errorf("failed to initialize UCI engine: %w", err)
	}

	// output game settings
	fmt.Fprintf(stdout, "==Game Settings==\nPlayer Color: %v\nDifficulty:%v\nTimed: %v", color, difficulty, timed)
	if timed {
		fmt.Fprintf(stdout, "\nTime Limit: %v seconds\n", timeLimit)
	}

	// start game loop
	fmt.Fprint(stdout, "\n\n==Game Start==\n\n")
	for {

		// check game state for end conditions
		if state.Method().String() == "Checkmate" || state.Method().String() == "Stalemate" || state.Method().String() == "Resignation" {
			fmt.Fprintf(stdout, "Game over! \nOutcome: %v\nMethod: %v\n", state.Outcome(), state.Method())
			break
		}

		// set up timer for player turn, if needed
		var playerTurnCtx context.Context
		var cancel context.CancelFunc
		if timed {
			playerTurnCtx, cancel = context.WithTimeout(ctx, time.Duration(timeLimit)*time.Second)
			defer cancel()
		} else {
			playerTurnCtx = ctx
		}

		if playerTurn {
			err := game.HandlePlayerTurn(playerTurnCtx, state, playerColor, stdin, stdout, stderr)
			if err != nil {
				return err
			}
		} else {
			err := game.HandleComputerTurn(ctx, state, eng, stdout)
			if err != nil {
				return err
			}
		}

		if showBoard {
			fmt.Fprintln(stdout, state.Position().Board().Draw2(playerColor, true))
		}

		// rotate turn
		playerTurn = !playerTurn

	}
	return nil
}
