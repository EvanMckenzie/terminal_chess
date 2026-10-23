package game

import (
	"context"
	"io"
)

// TODO: figure out how I want to represent the game board
type GameState struct {
	Board  [64]int
	Winner string
}

// TODO: implement this
func NewGameState() GameState {
	return GameState{}
}

// TODO: adjust to match final board state representation
func (s GameState) IsGameOver() bool {
	if s.Winner != "" {
		return true
	}
	return false
}

// TODO: implement this
func HandlePlayerTurn(ctx context.Context, state GameState, stdin io.Reader, stdout, stderr io.Writer) (GameState, error) {
	return state, nil
}

// TODO: implement this
func HandleComputerTurn(ctx context.Context, state GameState, stdout io.Writer) (GameState, error) {
	return state, nil
}
