package server

import (
	"crypto/tls"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Server struct {
	Config *LConfig
	CA     *CertAuthority
	pac    string
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL, r.RemoteAddr, r.Method, r.Proto)
	if s.handleInternal(w, r) {
		return
	}
	if s.isBlocked(r.URL) {
		fmt.Fprintf(w, "Blocked request: %q", html.EscapeString(r.URL.Path))
		return
	}
	// Proxy, otherwise.
	s.proxy(w, r)
}

func (s *Server) Init() error {
	pac, err := ioutil.ReadFile("./data/proxy.pac")
	if err != nil {
		return err
	}
	s.pac = string(pac)
	return nil
}

func (s *Server) Start(port string, config *tls.Config) error {
	log.Printf("|ket%s〉\n", port)
	srv := &http.Server{
		Addr: port, Handler: s,
	}
	ln, err := NewListener("tcp", srv.Addr, config)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

/*
		func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			u := url.URL{
				Scheme:   "https",
				Opaque:   r.URL.Opaque,
				User:     r.URL.User,
				Host:     addr,
				Path:     r.URL.Path,
				RawQuery: r.URL.RawQuery,
				Fragment: r.URL.Fragment,
			}
			http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
		} else {
			http.DefaultServeMux.ServeHTTP(w, r)
		}
	}
*/
func (s *Server) StartTLS(port string) error {
	log.Printf("|ket%s〉\n", port)
	srv := &http.Server{
		Addr: port, Handler: s,
		/*TLSConfig: &tls.Config{
			GetCertificate:     s.CA.Get,
			InsecureSkipVerify: true,
		},*/
	}
	return srv.ListenAndServeTLS(s.CA.CertFile, s.CA.KeyFile)
}

//------------------------------------------------------------------------------

func isInternalHost(host string) bool {
	return host == "" || host == "localhost" || host == "ket"
}

func (s *Server) handleInternal(w http.ResponseWriter, r *http.Request) bool {
	url := r.URL
	if !isInternalHost(strings.SplitN(url.Host, ":", 2)[0]) {
		return false
	}
	if url.Path == "/proxy.pac" {
		log.Println("PAC requested!")
		if url.RawQuery != "txt" {
			w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
		}
		fmt.Fprint(w, s.pac)
		return true
	}

	// Try to serve content from local file system.
	data := s.Config.Data
	for _, dir := range data.Dirs {
		prefix := dir.Url
		if !strings.HasPrefix(url.Path, prefix) {
			continue
		}
		// TODO: fix annoying bug with missing '/' at the end of path!
		//return http.StripPrefix(prefix, http.FileServer(http.Dir(dir.FPath)))
		url.Path = strings.TrimPrefix(url.Path, prefix)
		handler := http.FileServer(http.Dir(dir.FPath))
		handler.ServeHTTP(w, r)
		return true
	}

	// TODO: make redirect to index url.
	fmt.Fprintf(w, "Hello, %q, %q", html.EscapeString(url.Path), url.RawQuery)
	return true
}

func (s *Server) isBlocked(url *url.URL) bool {
	data := s.Config.Data
	for _, block := range data.Block {
		// TODO: implement full url match!
		if strings.HasPrefix(url.Path, block) {
			return true
		}
	}
	return false
}

func (s *Server) proxy(w http.ResponseWriter, r *http.Request) {
	// Based on: https://golang.org/src/net/http/httputil/reverseproxy.go
	out := new(http.Request)
	*out = *r

	out.Proto = "HTTP/1.1"
	out.ProtoMajor = 1
	out.ProtoMinor = 1
	out.Close = false

	transport := http.DefaultTransport
	if out.URL.Scheme == "" {
		out.URL.Scheme = "http"
		if strings.HasSuffix(out.URL.Host, ":443") {
			out.URL.Scheme = "https"
			transport = &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
		}
	}
	res, err := transport.RoundTrip(out)
	if err != nil {
		log.Println("Proxy error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if out.Method == "CONNECT" {
		w.WriteHeader(http.StatusOK)
		return
	}
	copyHeader(w.Header(), res.Header)
	w.WriteHeader(res.StatusCode)
	defer res.Body.Close()
	_, err = io.Copy(w, res.Body)
	if err != nil {
		log.Println("io.Copy:", err)
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
