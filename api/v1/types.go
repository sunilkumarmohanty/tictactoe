package v1

import "github.com/sunilkumarmohanty/tictactoe/repository"

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
