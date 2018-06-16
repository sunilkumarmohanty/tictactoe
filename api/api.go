package api

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sunilkumarmohanty/tictactoe/api/v1"
)

// Run starts the server
func Run() {
	router := mux.NewRouter().StrictSlash(false)
	v1.MakeHandlers(router)

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Println("Invalid PORT or PORT not set in environment variable")
		port = 8080
	}

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), router))
}
