package repository

import (
	"database/sql"

	_ "github.com/golang-migrate/migrate/source/file"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

type repository struct {
	db *sql.DB
}

func New(sqlConn string) (*repository, error) {
	db, err := connectDatabase(sqlConn)
	if err != nil {
		logger.Error("error connecting to db", zap.Error(err))
		return nil, err
	}
	return &repository{
		db: db,
	}, nil
}

func (r *repository) NewGame(computerMark, board string) (string, error) {

	query := `INSERT INTO games (computer_mark, board, status) VALUES ($1, $2, $3) RETURNING id`
	result := r.db.QueryRow(query, computerMark, board, "RUNNING")
	var gameID string
	err := result.Scan(&gameID)
	if err != nil {
		logger.Error("error creating a new game", zap.String("computer_mark", computerMark), zap.String("board", board))
		return "", err
	}
	return gameID, nil
}

func (r *repository) GetGames() ([]Game, error) {
	games := []Game{}
	//paging ignored for the timebeing
	query := "SELECT id, board, status, computer_mark FROM games"
	rows, err := r.db.Query(query)

	if err != nil {
		logger.Error("failed to get games from db", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		game := Game{}
		err = rows.Scan(&game.ID, &game.Board, &game.Status, &game.ComputerMark)
		if err != nil {
			logger.Error("failed to scan game row", zap.Error(err))
			continue
		}
		games = append(games, game)
	}
	return games, nil
}
func (r *repository) GetGame(id string) (*Game, error) {
	game := Game{}
	query := "SELECT id, board, status, computer_mark FROM games WHERE id = $1"
	row := r.db.QueryRow(query, id)

	err := row.Scan(&game.ID, &game.Board, &game.Status, &game.ComputerMark)
	if err != nil {
		// game not found
		if err == sql.ErrNoRows {
			logger.Info("game not found", zap.String("id", id))
			return nil, nil
		}
		logger.Error("failed to get game from db", zap.Error(err), zap.String("id", id))
		return nil, err
	}
	return &game, nil
}

func (r *repository) UpdateGame(game *Game) (int64, error) {
	query := "UPDATE games SET board = $2, status = $3 WHERE id = $1;"
	result, err := r.db.Exec(query, game.ID, game.Board, game.Status)
	if err != nil {
		logger.Error("failed to delete game from db", zap.Error(err))
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("failed to get rows affected after deleting game", zap.Error(err))
		return 0, err
	}

	return rowsAffected, nil
}
func (r *repository) DeleteGame(id string) (int64, error) {
	query := "DELETE FROM games WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		logger.Error("failed to delete game from db", zap.Error(err))
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("failed to get rows affected after deleting game", zap.Error(err))
		return 0, err
	}
	return rowsAffected, nil
}
