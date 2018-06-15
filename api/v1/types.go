package v1

import "github.com/sunilkumarmohanty/tictactoe/repository"

// IRepository is used as an interface for storing record in a repository
// Using it also makes it easier to write Unit test cases
type IRepository interface {
	GetGames() ([]repository.Game, error)
	GetGame(string) (*repository.Game, error)
	NewGame(string, string) (string, error)
	UpdateGame(*repository.Game) (int64, error)
	DeleteGame(string) (int64, error)
}

type newGameResponse struct {
	Location string `json:"location,omitempty"`
}
