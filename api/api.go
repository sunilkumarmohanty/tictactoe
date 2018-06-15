package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sunilkumarmohanty/tictactoe/api/v1"
)

// Run starts the server
func Run(listenAddr *string) {
	router := mux.NewRouter().StrictSlash(false)
	v1.MakeHandlers(router)
	log.Fatal(http.ListenAndServe(*listenAddr, router))
}
