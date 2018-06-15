package v1

import (
	"github.com/gorilla/mux"
)

// MakeHandlers creates all the routes and map it to respective handlers
func MakeHandlers(router *mux.Router) {
	uuidRegex := "[a-fA-F0-9]{8}-?[a-fA-F0-9]{4}-?4[a-fA-F0-9]{3}-?[8|9|aA|bB][a-fA-F0-9]{3}-?[a-fA-F0-9]{12}"
	gameHandlers := New()

	v1Router := router.PathPrefix("/api/v1").Subrouter()

	v1Router.Path("/games").Methods("GET").HandlerFunc(gameHandlers.GetAllGamesHandler)
	v1Router.Path("/games").Methods("POST").HandlerFunc(gameHandlers.CreateGameHandler)
	v1Router.Path("/games/{game_id:" + uuidRegex + "}").Methods("GET").HandlerFunc(gameHandlers.GetGameHandler)
	v1Router.Path("/games/{game_id:" + uuidRegex + "}").Methods("DELETE").HandlerFunc(gameHandlers.DeleteGameHandler)
	v1Router.Path("/games/{game_id:" + uuidRegex + "}").Methods("PUT").HandlerFunc(gameHandlers.UpdateGameHandler)
}
