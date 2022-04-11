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
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
}

func alive(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(`{"alive"}`))
}
