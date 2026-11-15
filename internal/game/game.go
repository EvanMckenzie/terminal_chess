package game

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/corentings/chess/v2"
	"github.com/corentings/chess/v2/uci"
)

func HandlePlayerTurn(ctx context.Context, state *chess.Game, playerColor chess.Color, stdin io.Reader, stdout, stderr io.Writer) error {

	for {
		// define input channel and error channels for background stdin read func
		input := make(chan string)
		errs := make(chan error)

		go func() {
			reader := bufio.NewReader(stdin)
			line, err := reader.ReadString('\n')
			if err != nil {
				errs <- err
				return
			}
			input <- strings.TrimSpace(line)
		}()

		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("turn time limit exceeded")
			}
			return ctx.Err()
		case err := <-errs:
			return err
		case playerMove := <-input:
			if strings.EqualFold(playerMove, "board") {
				fmt.Fprintln(stdout, state.Position().Board().Draw2(playerColor, true))
				continue
			} else if strings.EqualFold(playerMove, "exit") {
				return fmt.Errorf("game exited")
			} else if strings.EqualFold(playerMove, "resign") {
				state.Resign(playerColor)
				return nil
			}

			if err := state.PushNotationMove(playerMove, chess.AlgebraicNotation{}, nil); err != nil {
				fmt.Fprintf(stderr, "invalid move: %v\n", err)
				continue // let player retry
			}
			return nil
		}
	}
}

func HandleComputerTurn(ctx context.Context, state *chess.Game, eng *uci.Engine, stdout io.Writer) error {
	// use same pattern as player turn but without for loop since I don't need to prompt again

	if err := ctx.Err(); err != nil {
		return err
	}

	type result struct {
		move *chess.Move
		err  error
	}

	resultChan := make(chan result, 1)

	// run computer move in background and update results to channel
	go func() {
		// give engine current board state
		boardState := uci.CmdPosition{
			Position: state.Position(),
		}
		// set up parameters for engine to search for a move
		makeMove := uci.CmdGo{
			MoveTime: 500 * time.Millisecond,
		}

		if err := eng.Run(boardState, makeMove); err != nil {
			resultChan <- result{err: err}
		}

		resultChan <- result{move: eng.SearchResults().BestMove}
	}()

	// now select between computer move result and context
	select {
	case <-ctx.Done():
		return ctx.Err()
	case res := <-resultChan:
		if res.err != nil {
			return fmt.Errorf("computer failed to make move: %w", res.err)
		}
		if err := state.Move(res.move, nil); err != nil {
			return fmt.Errorf("failed to apply computer move: %w", err)
		}
		fmt.Fprintf(stdout, "Computer played: %s\n", res.move.String())
		return nil
	}
}
