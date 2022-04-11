package main

// https://gethttpsforfree.com/

import (
	// "encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	// "os"
	"path"
	"path/filepath"
	// "strings"
	persistense "github.com/tbolsh/extend-go-nginx-postgres-docker/persistense"
)

var (
	port               = flag.Int("p", 8000, "port to listen on")
	root               = flag.String("r", "~", "base path")
	baseDir, staticDir string
	pathf              func(p string) string
)

func main() {
	flag.Parse()
	baseDir = filepath.Clean(*root)
	staticDir = path.Clean(path.Join(baseDir, "static"))
	pathf = func(p string) string { return filepath.Join(baseDir, p) }
	if *port < 1 || *port > 65535 {
		log.Fatalf("Port should be between 0 and 65536 but it is %d", port)
	}
	persistense.Initialize()
	go func() {
		err := persistense.CreateTable("clients", []string{
			`create table clients(api_key varchar(64), extend_api_key varchar(256), PRIMARY KEY(api_key, extend_api_key));`,
			`CREATE INDEX clients_idx ON clients(api_key);`,
		})
		sqlerr(err)
	}()
	mux := http.NewServeMux()
	mux.HandleFunc("/alive", alive)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: proxy{Handler: mux},
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err)
	}
}

type proxy struct{ Handler http.Handler }

func (p proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p.Handler.ServeHTTP(w, req)
	// w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
}

func alive(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(`{"alive": true}`))
}

func sqlerr(err error) {
	if err != nil {
		log.Printf("SQL Error '%v'", err)
	}
}
