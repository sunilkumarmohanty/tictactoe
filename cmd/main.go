package main

import (
	"flag"

	"github.com/sunilkumarmohanty/tictactoe/api"
)

func main() {
	listenAddr := flag.String("http.addr", ":8080", "http listen address")
	flag.Parse()
	api.Run(listenAddr)
}
