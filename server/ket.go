package server

import (
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Server struct {
	Config *LConfig
	pac    []byte
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//log.Println("host is", r.URL.Host)
	if s.handleInternal(w, r) {
		return
	}
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
	pac, err := ioutil.ReadFile("./data/proxy.pac")
	if err != nil {
		log.Fatal(err)
	}
	s.pac = pac
	log.Printf("|ket%s〉\n", port)
	log.Fatal(http.ListenAndServe(port, s))
}

//------------------------------------------------------------------------------
func (s *Server) handleInternal(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Host != "" {
		return false
	}
	if r.URL.Path == "/proxy.pac" {
		log.Println("PAC requested!")
		w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
		fmt.Fprint(w, string(s.pac))
		return true
	}
	return false
}

func (s *Server) isBlocked(r *http.Request) bool {
	data := s.Config.Data
	for _, block := range data.Block {
		if strings.HasPrefix(r.URL.Path, block) {
			return true
		}
	}
	return false
}

func (s *Server) getStaticHandler(r *http.Request) http.Handler {
	if r.URL.Host != "" && r.URL.Host != "ket" {
		return nil
	}
	data := s.Config.Data
	for _, dir := range data.Dirs {
		if strings.HasPrefix(r.URL.Path, dir.Url) {
			return http.StripPrefix(dir.Url, http.FileServer(http.Dir(dir.FPath)))
			// Fix urls in reply.
			//r.URL.Path = strings.TrimPrefix(r.URL.Path, dir.Url)
			//return http.FileServer(http.Dir(dir.FPath))
		}
	}
	return nil
}
