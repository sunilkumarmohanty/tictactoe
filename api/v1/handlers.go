package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	fMark = "-" //blank position
)

// Handlers represent the game handlers
type Handlers struct {
	repo        IRepository
	hostAddress string
}

// New initialises the handlers struct
func New() *Handlers {
	logger.Info("Creating handlers")
	sqlConn := os.Getenv("SQL_CONN")
	if len(sqlConn) == 0 {
		logger.Fatal("SQL_CONN not set")
	}

	hostAddress := os.Getenv("HOST_ADDR")
	_, err := url.Parse(hostAddress)
	if err != nil {
		logger.Fatal("invalid HOST_ADDR in environment variable")
	}
	repo, err := repository.New(sqlConn)
	if err != nil {
		logger.Fatal("Unable to create repository")
	}

	return &Handlers{
		repo:        repo,
		hostAddress: hostAddress,
	}
}

// GetAllGamesHandler returns all the games stored in the database
func (h *Handlers) GetAllGamesHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	games, err := h.repo.GetGames()
	if err != nil {
		logger.Error("unable to get games", zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(rw).Encode(games)
}

// GetGameHandler returns an instance of a single game
func (h *Handlers) GetGameHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	game, err := h.repo.GetGame(params["game_id"])
	if err != nil {
		logger.Error("unable to get game", zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	// game not found
	if game == nil {
		logger.Error("game not found", zap.String("gameid", params["game_id"]))
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
	// Check if the new board is valid. If valid, then make a move and save the state
	if computerMark, ok := newGame.validateNewGame(); ok {
		// computer makes the move
		newGame.play(computerMark)
		// save the game
		gameID, err := h.repo.NewGame(computerMark, newGame.Board)
		if err != nil {
			logger.Error("game creation failed", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp := newGameResponse{
			Location: fmt.Sprintf("%s/%s/%s", h.hostAddress, "api/v1/games", gameID),
		}
		rw.WriteHeader(http.StatusCreated)
		json.NewEncoder(rw).Encode(resp)
		return
	}
	sendJSONError(rw, http.StatusBadRequest, "invalid new board")
}

// UpdateGameHandler handles a move made by opponent and if required makes the computer move. It also saves the result in db
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
	// Check if game is still in play as per stored state
	if storedState.Status != gameStatusRunning {
		logger.Error("game already over", zap.Error(err))
		sendJSONError(rw, http.StatusBadRequest, "game already over")
		return
	}
	//Check if play made by opponent is valid. Compare the game with the previous stat
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
	// If not running then opponent has either won or drawn
	if status != gameStatusRunning {
		dbGame := &repository.Game{
			ID:     gameID,
			Board:  curGame.Board,
			Status: status,
		}
		recordsAffected, err := h.repo.UpdateGame(dbGame)
		if err != nil {
			logger.Error("game update failed", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		// recordsAffected will be 0 only when the game gets deleted during the time the computer is checking the status. This is an extreme corner case and will occur in rare scenario.
		if recordsAffected == 0 {
			logger.Error("game not found", zap.String("gameid", gameID))
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(rw).Encode(dbGame)
		return
	}
	// game is running and now computer can make its move
	curGame.play(storedState.ComputerMark)
	status = curGame.getStatus()
	dbGame := &repository.Game{
		ID:     gameID,
		Board:  curGame.Board,
		Status: status,
	}
	recordsAffected, err := h.repo.UpdateGame(dbGame)
	if err != nil {
		logger.Error("game update failed", zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	// recordsAffected will be 0 only when the game gets deleted during the time the computer is deciding to make a move. This is a corner case and will occur in rare scenario.
	if recordsAffected == 0 {
		logger.Error("game not found", zap.String("gameid", gameID))
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(rw).Encode(dbGame)
	return
}

// DeleteGameHandler deletes a game from the db
func (h *Handlers) DeleteGameHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	rowsAffected, err := h.repo.DeleteGame(params["game_id"])
	if err != nil {
		logger.Error("game deletion failed", zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		logger.Error("game not found", zap.String("gameid", params["game_id"]))
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
