package v1

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/sunilkumarmohanty/tictactoe/repository"
)

const (
	msgInternalServerError = "Internal server error"
	msgResourceNotFound    = "Resource not found"
)

const (
	xMark = "X"
	oMark = "O"
	fMark = "-"
)

// Handlers represent the game handlers
type Handlers struct {
	repo IRepository
}

// New initialises the handlers struct
func New() *Handlers {
	logger.Info("Creating handlers")
	sqlConn := os.Getenv("SQL_CONN")
	if len(sqlConn) == 0 {
		logger.Fatal("SQL_CONN not set")
	}
	repo, err := repository.New(sqlConn)
	if err != nil {
		logger.Fatal("Unable to create repository")
	}
	return &Handlers{
		repo: repo,
	}
}

// GetAllGamesHandler returns all the games stored in the database
func (h *Handlers) GetAllGamesHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	games, err := h.repo.GetGames()
	if err != nil {
		logger.Error("unable to get game", zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(rw).Encode(games)
}

// GetGameHandler returns an instance of a single game
func (h *Handlers) GetGameHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	game, err := h.repo.GetGame(vars["game_id"])
	if err != nil {
		logger.Error("unable to get game", zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	if game == nil {
		logger.Error("game not found", zap.String("gameid", vars["game_id"]))
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(rw).Encode(game)
}

// CreateGameHandler creates a new game
func (h *Handlers) CreateGameHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	newGame := &Game{}
	err := json.NewDecoder(r.Body).Decode(newGame)
	if err != nil {
		logger.Error("invalid body while creating new game", zap.Error(err))

		sendJSONError(rw, http.StatusBadRequest, "invalid request body")
		return
	}
	if computerMark, ok := newGame.validateNewGame(); ok {
		newGame.play(computerMark)
		gameID, err := h.repo.NewGame(computerMark, newGame.Board)
		if err != nil {
			logger.Error("game creation failed", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp := newGameResponse{
			Location: gameID,
		}
		rw.WriteHeader(http.StatusCreated)
		json.NewEncoder(rw).Encode(resp)
		return
	}
	sendJSONError(rw, http.StatusBadRequest, "invalid new board")
}

// UpdateGameHandler handles a move made by human and if required makes the computer move. It also saves the result in db
func (h *Handlers) UpdateGameHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	gameID := mux.Vars(r)["game_id"]
	curGame := &Game{}
	err := json.NewDecoder(r.Body).Decode(curGame)
	if err != nil {
		logger.Error("invalid body while updating game", zap.Error(err))
		sendJSONError(rw, http.StatusBadRequest, "invalid request body")
		return
	}
	//Validate board
	if !curGame.validateBoard() {
		logger.Error("invalid board", zap.Any("game", curGame))
		sendJSONError(rw, http.StatusBadRequest, "invalid board")
		return
	}
	//Check if game exists
	storedState, err := h.repo.GetGame(gameID)
	if err != nil {
		logger.Error("unable to get game", zap.Error(err), zap.String("gameid", gameID))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	if storedState == nil {
		logger.Error("game not found", zap.String("gameid", gameID))
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	if storedState.Status != gameStatusRunning {
		logger.Error("game already over", zap.Error(err))
		sendJSONError(rw, http.StatusBadRequest, "game already over")
		return
	}
	//Check if play is valid
	playStatus := curGame.validatePlay(&Game{
		Board: storedState.Board,
	}, storedState.ComputerMark)
	if playStatus == 0 {
		logger.Error("no move made by opponent", zap.String("gameid", gameID))
		sendJSONError(rw, http.StatusBadRequest, "no move made")
		return
	}
	if playStatus == -1 {
		logger.Error("game state mismatch", zap.String("gameid", gameID))
		sendJSONError(rw, http.StatusBadRequest, "game state mismatch")
		return
	}

	// If game is in RUNNING state then make our move.
	status := curGame.getStatus()
	if status != gameStatusRunning {
		dbGame := &repository.Game{
			ID:     gameID,
			Board:  curGame.Board,
			Status: status,
		}
		_, err := h.repo.UpdateGame(dbGame)
		if err != nil {
			logger.Error("game update failed", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(rw).Encode(dbGame)
		return
	}
	curGame.play(storedState.ComputerMark)
	status = curGame.getStatus()
	dbGame := &repository.Game{
		ID:     gameID,
		Board:  curGame.Board,
		Status: status,
	}
	_, err = h.repo.UpdateGame(dbGame)
	if err != nil {
		logger.Error("game update failed", zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(rw).Encode(dbGame)
	return
}

// DeleteGameHandler deletes a game from the db
func (h *Handlers) DeleteGameHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	rowsAffected, err := h.repo.DeleteGame(vars["game_id"])
	if err != nil {
		logger.Error("game deletion failed", zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		logger.Error("game not found", zap.String("gameid", vars["game_id"]))
		rw.WriteHeader(http.StatusNotFound)
		return
	}
}

// Change this to send response
// For Bad Request create a struct
func sendJSONError(rw http.ResponseWriter, code int, reason string) {
	errorResp := map[string]string{"reason": reason}
	rw.WriteHeader(code)
	json.NewEncoder(rw).Encode(errorResp)
}
