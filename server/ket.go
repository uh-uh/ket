package server

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
)

type Server struct {
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//log.Println("host is", r.URL.Host)
	if s.isBlocked(r) {
		fmt.Fprintf(w, "Blocked request: %q", html.EscapeString(r.URL.Path))
		return
	}
	if handler := s.getStaticHandler(r); handler != nil {
		handler.ServeHTTP(w, r)
		return
	}
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func (s *Server) Start(port string) {
	log.Printf("|ket%s〉\n", port)
	log.Fatal(http.ListenAndServe(port, s))
}

//------------------------------------------------------------------------------

func (s *Server) isBlocked(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/test/block")
}

func (s *Server) getStaticHandler(r *http.Request) http.Handler {
	if r.URL.Host != "" {
		return nil
	}
	if strings.HasPrefix(r.URL.Path, "/test/dir") {
		// Fix urls in reply.
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/test/dir")
		return http.FileServer(http.Dir("../"))
	}
	return nil
}
