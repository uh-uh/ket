package main

import (
	//"fmt"
	"github.com/gophergala/ket/server"
	//"html"
	//"log"
	//"net/http"
)

func main() {
	srv := &server.Server{}
	srv.Start(":8080")
}
