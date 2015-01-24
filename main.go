package main

import (
	//"fmt"
	"github.com/gophergala/ket/server"
	//"html"
	"log"
	//"net/http"
)

func main() {
	config, err := server.LiveConfig("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	srv := &server.Server{Config: config}
	srv.Start(":8080")
}
