package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sunilkumarmohanty/tictactoe/repository"
)

type mockDB struct {
	rowsAffected int64

	gameID string

	game  *repository.Game
	games []repository.Game

	deleteErr     error // error while deleting
	newErr        error // error while inserting
	getGameErr    error // error while getting a game
	getGamesErr   error // error while getting all games
	updateGameErr error // error while updating a game
	IRepository
}

func (m *mockDB) DeleteGame(string) (int64, error) {
	return m.rowsAffected, m.deleteErr
}
func (m *mockDB) NewGame(string, string) (string, error) {
	return m.gameID, m.newErr
}
func (m *mockDB) GetGame(string) (*repository.Game, error) {
	return m.game, m.getGameErr
}

func (m *mockDB) GetGames() ([]repository.Game, error) {
	return m.games, m.getGamesErr
}

func (m *mockDB) UpdateGame(*repository.Game) (int64, error) {
	return m.rowsAffected, m.updateGameErr
}
func TestHandlers_DeleteGameHandler(t *testing.T) {

	mockHandler := &Handlers{}
	hostURL := "http://tictactoe/api/v1/games"
	_ = hostURL
	m := mux.NewRouter()
	m.HandleFunc("/api/v1/games/{game_id}", mockHandler.DeleteGameHandler)
	type fields struct {
		dbRowsAffected int64
		dbDeleteErr    error
		gameID         string
	}
	type args struct {
	}
	tests := []struct {
		name           string
		fields         fields
		wantStatusCode int
	}{
		{
			name: "Valid",
			fields: fields{
				dbRowsAffected: 1,
				gameID:         "dummy_game_id",
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "Error from DB",
			fields: fields{
				dbDeleteErr: errors.New("delete error"),
				gameID:      "dummy_game_id",
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "No records deleted",
			fields: fields{
				dbRowsAffected: 0,
				gameID:         "dummy_game_id",
			},
			wantStatusCode: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockDB{
				rowsAffected: tt.fields.dbRowsAffected,
				deleteErr:    tt.fields.dbDeleteErr,
			}
			mockHandler.repo = mockRepo

			deleteURL := fmt.Sprintf("%v/%v", hostURL, tt.fields.gameID)
			req, err := http.NewRequest("DELETE", deleteURL, nil)
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			m.ServeHTTP(recorder, req)
			if recorder.Code != tt.wantStatusCode {
				t.Errorf("status code did not match : got %v want %v", recorder.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestHandlers_CreateGameHandler(t *testing.T) {

	hostURL := "http://tictactoe"
	mockHandler := &Handlers{
		hostAddress: hostURL,
	}
	m := mux.NewRouter()
	m.HandleFunc("/api/v1/games", mockHandler.CreateGameHandler)
	type fields struct {
		body string

		dbGameID string
		dbNewErr error
	}
	type args struct {
	}
	tests := []struct {
		name             string
		fields           fields
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name: "Valid",
			fields: fields{
				dbGameID: "dummy_game_id",
				body:     `{"board": "--------X"}`,
			},
			wantStatusCode:   http.StatusCreated,
			wantResponseBody: `{"location":"` + hostURL + `/api/v1/games/dummy_game_id"}`,
		},
		{
			name: "Valid Blank Board",
			fields: fields{
				dbGameID: "dummy_game_id",
				body:     `{"board": "---------"}`,
			},
			wantStatusCode:   http.StatusCreated,
			wantResponseBody: `{"location":"` + hostURL + `/api/v1/games/dummy_game_id"}`,
		},
		{
			name: "Error from DB",
			fields: fields{
				body:     `{"board": "--------X"}`,
				dbNewErr: errors.New("delete error"),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Invalid move - more than one-X",
			fields: fields{
				body: `{"board": "-------XX"}`,
			},
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"reason":"invalid new board"}`,
		},
		{
			name: "Invalid move - more than one-O",
			fields: fields{
				body: `{"board": "-------OO"}`,
			},
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"reason":"invalid new board"}`,
		},
		{
			name: "Invalid move - more than one-OX",
			fields: fields{
				body: `{"board": "-------OX"}`,
			},
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"reason":"invalid new board"}`,
		},
		{
			name: "Invalid move - board size wrong",
			fields: fields{
				body: `{"board": "-------XX-"}`,
			},
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"reason":"invalid new board"}`,
		},
		{
			name: "Invalid move - invalid mark",
			fields: fields{
				body: `{"board": "-------V-"}`,
			},
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"reason":"invalid new board"}`,
		},
		{
			name: "Invalid json request body",
			fields: fields{
				body: `{"board": "-------XX"`,
			},
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"reason":"invalid request body"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockDB{
				gameID: tt.fields.dbGameID,
				newErr: tt.fields.dbNewErr,
			}
			mockHandler.repo = mockRepo

			req, err := http.NewRequest("PUT", hostURL+"/api/v1/games", bytes.NewBuffer([]byte(tt.fields.body)))
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			m.ServeHTTP(recorder, req)
			if recorder.Code != tt.wantStatusCode {
				t.Errorf("status code did not match : got %v want %v", recorder.Code, tt.wantStatusCode)
			}
			gotBody := strings.TrimSpace(recorder.Body.String())
			if gotBody != tt.wantResponseBody {
				t.Errorf("response body did not match : got %v want %v", gotBody, tt.wantResponseBody)
			}

		})
	}
}

func TestHandlers_GetGameHandler(t *testing.T) {
	mockHandler := &Handlers{}
	hostURL := "http://tictactoe/api/v1/games"
	_ = hostURL
	m := mux.NewRouter()
	m.HandleFunc("/api/v1/games/{game_id}", mockHandler.GetGameHandler)
	type fields struct {
		gameID       string
		dbGame       *repository.Game
		dbGetGameErr error
	}
	type args struct {
	}
	tests := []struct {
		name             string
		fields           fields
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name: "Valid",
			fields: fields{
				gameID: "dummy_game_id",
				dbGame: &repository.Game{
					ID:           "dummy_game_id",
					Board:        "X--------",
					Status:       "RUNNING",
					ComputerMark: "X",
				},
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"id":"dummy_game_id","board":"X--------","status":"RUNNING"}`,
		},
		{
			name: "Error from DB",

			fields: fields{
				gameID:       "dummy_game_id",
				dbGetGameErr: errors.New("get game error"),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "No game returned",
			fields: fields{
				gameID: "dummy_game_id",
			},
			wantStatusCode: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockDB{
				game:       tt.fields.dbGame,
				getGameErr: tt.fields.dbGetGameErr,
			}
			mockHandler.repo = mockRepo

			getGameURL := fmt.Sprintf("%v/%v", hostURL, tt.fields.gameID)
			req, err := http.NewRequest("GET", getGameURL, nil)
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			m.ServeHTTP(recorder, req)
			if recorder.Code != tt.wantStatusCode {
				t.Errorf("status code did not match : got %v want %v", recorder.Code, tt.wantStatusCode)
			}
			gotBody := strings.TrimSpace(recorder.Body.String())
			if gotBody != tt.wantResponseBody {
				t.Errorf("response body did not match : got %v want %v", gotBody, tt.wantResponseBody)
			}
		})
	}
}

func TestHandlers_GetAllGamesHandler(t *testing.T) {
	mockHandler := &Handlers{}
	hostURL := "http://tictactoe/api/v1/games"
	m := mux.NewRouter()
	m.HandleFunc("/api/v1/games", mockHandler.GetAllGamesHandler)
	type fields struct {
		dbGames       []repository.Game
		dbGetGamesErr error
	}
	type args struct {
	}
	tests := []struct {
		name             string
		fields           fields
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name: "Valid",
			fields: fields{
				dbGames: []repository.Game{
					repository.Game{
						ID:           "dummy_game_id_1",
						Board:        "X--------",
						Status:       "RUNNING",
						ComputerMark: "X",
					},
					repository.Game{
						ID:           "dummy_game_id_2",
						Board:        "X--------",
						Status:       "RUNNING",
						ComputerMark: "X",
					},
				},
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `[{"id":"dummy_game_id_1","board":"X--------","status":"RUNNING"},{"id":"dummy_game_id_2","board":"X--------","status":"RUNNING"}]`,
		},
		{
			name: "Error from DB",

			fields: fields{
				dbGetGamesErr: errors.New("get games error"),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockDB{
				games:       tt.fields.dbGames,
				getGamesErr: tt.fields.dbGetGamesErr,
			}
			mockHandler.repo = mockRepo

			req, err := http.NewRequest("GET", hostURL, nil)
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			m.ServeHTTP(recorder, req)
			if recorder.Code != tt.wantStatusCode {
				t.Errorf("status code did not match : got %v want %v", recorder.Code, tt.wantStatusCode)
			}
			gotBody := strings.TrimSpace(recorder.Body.String())
			if gotBody != tt.wantResponseBody {
				t.Errorf("response body did not match : got %v want %v", gotBody, tt.wantResponseBody)
			}
		})
	}
}

func TestHandlers_UpdateGameHandler(t *testing.T) {
	mockHandler := &Handlers{}
	hostURL := "http://tictactoe/api/v1/games"
	m := mux.NewRouter()
	m.HandleFunc("/api/v1/games/{game_id}", mockHandler.UpdateGameHandler)
	type fields struct {
		gameID string
		body   string

		dbGame          *repository.Game
		dbRowsAffected  int64
		dbGetGameErr    error
		dbUpdateGameErr error
	}
	type args struct {
	}
	tests := []struct {
		name   string
		fields fields

		wantGameStatuses    []string
		wantStatusCode      int
		wantErrResponseBody string
	}{
		{
			name: "Valid",
			fields: fields{
				gameID: "dummy_game_id",
				dbGame: &repository.Game{
					ID:           "dummy_game_id",
					Board:        "--------X",
					Status:       "RUNNING",
					ComputerMark: "X",
				},
				body: `{"board": "-------OX"}`,
			},
			wantGameStatuses: []string{gameStatusRunning},
			wantStatusCode:   http.StatusOK,
		},
		{
			name: "Valid Human Win",
			fields: fields{
				gameID: "dummy_game_id",
				dbGame: &repository.Game{
					ID:           "dummy_game_id",
					Board:        "OO-----XX",
					Status:       "RUNNING",
					ComputerMark: "X",
				},
				body: `{"board": "OOO----XX"}`,
			},
			wantGameStatuses: []string{gameStatusOWon},
			wantStatusCode:   http.StatusOK,
		},
		{
			name: "Valid Computer May Win",
			fields: fields{
				gameID: "dummy_game_id",
				dbGame: &repository.Game{
					ID:           "dummy_game_id",
					Board:        "O------XX",
					Status:       "RUNNING",
					ComputerMark: "X",
				},
				body: `{"board": "OO-----XX"}`,
			},
			wantGameStatuses: []string{gameStatusXWon, gameStatusRunning},
			wantStatusCode:   http.StatusOK,
		},
		{
			name: "Valid Draw By Human",
			fields: fields{
				gameID: "dummy_game_id",
				dbGame: &repository.Game{
					ID:           "dummy_game_id",
					Board:        "OXXXOO-OX",
					Status:       "RUNNING",
					ComputerMark: "X",
				},
				body: `{"board": "OXXXOOOOX"}`,
			},
			wantGameStatuses: []string{gameStatusDraw},
			wantStatusCode:   http.StatusOK,
		},
		{
			name: "Error from DB - Getting Old Game",
			fields: fields{
				gameID:       "dummy_game_id",
				dbGetGameErr: errors.New("error getting game"),
				body:         `{"board": "-------OX"}`,
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Error From DB - Saving New Game",
			fields: fields{
				gameID: "dummy_game_id",
				dbGame: &repository.Game{
					ID:           "dummy_game_id",
					Board:        "--------X",
					Status:       "RUNNING",
					ComputerMark: "X",
				},
				dbUpdateGameErr: errors.New("error updating game"),
				body:            `{"board": "-------OX"}`,
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Error From DB - Saving New Game - Human Won",
			fields: fields{
				gameID: "dummy_game_id",
				dbGame: &repository.Game{
					ID:           "dummy_game_id",
					Board:        "OO-----XX",
					Status:       "RUNNING",
					ComputerMark: "X",
				},
				dbUpdateGameErr: errors.New("error updating game"),
				body:            `{"board": "OOO----XX"}`,
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "No Game Returned From DB",
			fields: fields{
				gameID: "dummy_game_id",
				body:   `{"board": "-------OX"}`,
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "Invalid JSON Body",
			fields: fields{
				gameID: "dummy_game_id",
				body:   `"board": "-------OX"}`,
			},
			wantStatusCode:      http.StatusBadRequest,
			wantErrResponseBody: `{"reason":"invalid request body"}`,
		},
		{
			name: "Invalid Board",
			fields: fields{
				gameID: "dummy_game_id",
				body:   `{"board": "-------OX-"}`,
			},
			wantStatusCode:      http.StatusBadRequest,
			wantErrResponseBody: `{"reason":"invalid board"}`,
		},
		{
			name: "Invalid Mark In Board",
			fields: fields{
				gameID: "dummy_game_id",
				body:   `{"board": "-------OH"}`,
			},
			wantStatusCode:      http.StatusBadRequest,
			wantErrResponseBody: `{"reason":"invalid board"}`,
		},
		{
			name: "No Move Made",
			fields: fields{
				gameID: "dummy_game_id",
				dbGame: &repository.Game{
					ID:           "dummy_game_id",
					Board:        "--------X",
					Status:       "RUNNING",
					ComputerMark: "X",
				},
				body: `{"board": "--------X"}`,
			},
			wantStatusCode:      http.StatusBadRequest,
			wantErrResponseBody: `{"reason":"no move made"}`,
		},
		{
			name: "Game Already Over",
			fields: fields{
				gameID: "dummy_game_id",
				dbGame: &repository.Game{
					ID:           "dummy_game_id",
					Board:        "XXXOO----",
					Status:       gameStatusXWon,
					ComputerMark: "X",
				},
				body: `{"board": "XXXOOO---"}`,
			},
			wantStatusCode:      http.StatusBadRequest,
			wantErrResponseBody: `{"reason":"game already over"}`,
		},
		{
			name: "Invalid Move",
			fields: fields{
				gameID: "dummy_game_id",
				dbGame: &repository.Game{
					ID:           "dummy_game_id",
					Board:        "-------OX",
					Status:       "RUNNING",
					ComputerMark: "X",
				},
				body: `{"board": "------XOX"}`,
			},
			wantStatusCode:      http.StatusBadRequest,
			wantErrResponseBody: `{"reason":"game state mismatch"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockDB{
				getGameErr:    tt.fields.dbGetGameErr,
				updateGameErr: tt.fields.dbUpdateGameErr,
				game:          tt.fields.dbGame,
			}
			mockHandler.repo = mockRepo

			getGameURL := fmt.Sprintf("%v/%v", hostURL, tt.fields.gameID)
			req, err := http.NewRequest("POST", getGameURL, bytes.NewBuffer([]byte(tt.fields.body)))
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			m.ServeHTTP(recorder, req)
			if recorder.Code != tt.wantStatusCode {
				t.Errorf("status code did not match : got %v want %v", recorder.Code, tt.wantStatusCode)
			}
			gotBody := strings.TrimSpace(recorder.Body.String())

			if recorder.Code != http.StatusOK {
				if gotBody != tt.wantErrResponseBody {
					t.Errorf("response body did not match : got %v want %v", gotBody, tt.wantErrResponseBody)
				}
			} else {
				// Tests for 200
				var game repository.Game
				json.Unmarshal([]byte(gotBody), &game)
				statusMatchFound := false
				for _, status := range tt.wantGameStatuses {
					if status == game.Status {
						statusMatchFound = true
					}
				}

				if !statusMatchFound {
					t.Errorf("response game status did not match : got  %v want one of %v", game.Status, tt.wantGameStatuses)
				}

				if tt.fields.gameID != game.ID {
					t.Errorf("response game id did not match : got  %v want %v", game.ID, tt.fields.gameID)
				}
				// Check if computer made the correct move
				// First check if computer has to make any move
				humanGame := &Game{}
				json.Unmarshal([]byte(tt.fields.body), &humanGame)

				if humanGame.getStatus() == gameStatusRunning {
					computerGame := &Game{game.Board}
					if computerGame.validatePlay(humanGame, findOpponentMark((tt.fields.dbGame.ComputerMark))) != 1 {
						t.Errorf("invalid move made by computer")
					}
				}
			}
		})

	}
}
