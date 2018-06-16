package v1

import (
	"math/rand"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Game represent the tic tac toe game
type Game struct {
	Board string `json:"board,omitempty"`
}

const (
	gameStatusXWon    = "X_WON"
	gameStatusOWon    = "O_WON"
	gameStatusRunning = "RUNNING"
	gameStatusDraw    = "DRAW"
)

func (g *Game) validateNewGame() (string, bool) {
	//check for the length of the board
	if len(g.Board) != 9 {
		logger.Error("invalid board", zap.String("board", g.Board))

		return "", false
	}
	xMoves, oMoves := 0, 0
	moves := strings.Split(g.Board, "")
	// The number of moves should be less than equal to 1
	// Only X, O and - allowed in the board
	for _, move := range moves {
		switch move {
		case xMark:
			if xMoves+oMoves == 1 {
				logger.Error("more than one move made", zap.String("board", g.Board))
				return "", false
			}
			xMoves++
		case oMark:
			if xMoves+oMoves == 1 {
				logger.Error("more than one move made", zap.String("board", g.Board))
				return "", false
			}
			oMoves++
		case "-":
		default:
			logger.Error("invalid move", zap.String("move", move))
			return "", false
		}

	}
	if xMoves == 1 {
		return oMark, true
	}
	return xMark, true
}

func (g *Game) validateBoard() bool {
	// check for length of board
	if len(g.Board) != 9 {
		logger.Error("invalid board", zap.String("board", g.Board))
		return false
	}
	//check for invalid mark in the board
	moves := strings.Split(g.Board, "")
	for _, move := range moves {
		switch move {
		case xMark, oMark, fMark:
		default:
			logger.Error("invalid board", zap.String("move", g.Board))
			return false
		}
	}
	return true
}

// validatePlay validates if the player made exactly one move and if the move is valid
// Returns
//	0 if no move made
//  1 if 1 correct move made
//  -1 invalid move (or state of board) detected
func (g *Game) validatePlay(prevState *Game, curPlayerMark string) int {
	curMoves := strings.Split(g.Board, "")
	prevMoves := strings.Split(prevState.Board, "")
	opponentMark := findOpponentMark(curPlayerMark)
	diffs := 0
	for indx, move := range curMoves {
		if move != prevMoves[indx] {
			// check if opponent has made a valid move with respect to mark and number of moves
			if move != opponentMark || diffs == 1 {
				//the play does not complement to its previous state if there are more than 1 diff
				return -1
			}
			diffs++
		}
	}
	return diffs
}

// play checks for blank positions and randomly selects a blank space and makes it move
func (g *Game) play(mark string) {
	moves := strings.Split(g.Board, "")
	validPositions := findBlankPositions(moves)
	// make move only when valid position found
	if len(validPositions) > 0 {
		// randomly select a blank valid position
		rand := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomMovePosition := validPositions[rand.Intn(len(validPositions))]
		moves[randomMovePosition] = mark
		g.Board = strings.Join(moves, "")
	}
}

func (g *Game) getStatus() string {
	moves := strings.Split(g.Board, "")
	var winner string

	for i := 0; i < 3; i++ {
		//horizontal check
		if moves[i*3] == moves[i*3+1] && moves[i*3] == moves[i*3+2] && moves[i*3] != fMark {
			winner = moves[i*3]
			break
		}
		//vertical check
		if moves[i] == moves[i+3] && moves[i+6] == moves[i] && moves[i] != fMark {
			winner = moves[i]
			break
		}
	}
	//diagonal check
	if moves[0] == moves[4] && moves[0] == moves[8] && moves[0] != fMark {
		winner = moves[0]
	}
	//diagonal check
	if moves[2] == moves[4] && moves[2] == moves[6] && moves[2] != fMark {
		winner = moves[2]
	}
	if winner == xMark {
		return gameStatusXWon
	}
	if winner == oMark {
		return gameStatusOWon
	}
	//check if game ended
	if len(findBlankPositions(moves)) == 0 {
		return gameStatusDraw
	}
	return gameStatusRunning
}

func findBlankPositions(moves []string) []int {
	var validPositions []int
	for indx, move := range moves {
		if move == fMark {
			validPositions = append(validPositions, indx)
		}
	}
	return validPositions
}

func findOpponentMark(mark string) string {
	switch mark {
	case xMark:
		return oMark
	case oMark:
		return xMark
	default:
		return ""
	}
}
