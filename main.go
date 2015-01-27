package main

import (
	"crypto/tls"
	"github.com/uh-uh/ket/server"
	"log"
)

func main() {
	log.SetFlags(log.Flags() | log.Llongfile)
	srvConfig, err := server.LiveConfig("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	//root := "./data"
	//certFile := filepath.Join(root, "cert.pem")
	//keyFile := filepath.Join(root, "key.pem")

	//ca := server.NewCertAuthority(certFile, keyFile)
	//err = ca.Init()
	//if err != nil {
	//	log.Fatal(err)
	//}
	srv := &server.Server{
		Config: srvConfig,
		//CA:     ca,
	}
	cert, err := tls.LoadX509KeyPair("./data/cert.pem", "./data/key.pem")
	if err != nil {
		log.Fatal(err)
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"http/1.1"},
	}
	err = srv.Init()
	if err != nil {
		log.Fatal(err)
	}
	errors := make(chan error)
	go func() {
		errors <- srv.Start(":4891", config)
	}()
	//go func() {
	//	errors <- srv.StartTLS(":4819", certFile, keyFile)
	//}()

	log.Fatal(<-errors)
}
