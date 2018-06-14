package v1

import (
	"bytes"
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

	deleteErr   error // error while deleting
	newErr      error // error while inserting
	getGameErr  error // error while getting a game
	getGamesErr error // error while getting all games
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
	mockHandler := &Handlers{}
	hostURL := "http://tictactoe/api/v1/games"
	_ = hostURL
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
			wantResponseBody: `{"location":"dummy_game_id"}`,
		},
		{
			name: "Valid Blank Board",
			fields: fields{
				dbGameID: "dummy_game_id",
				body:     `{"board": "---------"}`,
			},
			wantStatusCode:   http.StatusCreated,
			wantResponseBody: `{"location":"dummy_game_id"}`,
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

			req, err := http.NewRequest("POST", hostURL, bytes.NewBuffer([]byte(tt.fields.body)))
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
	_ = hostURL
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
