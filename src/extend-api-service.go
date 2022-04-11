package main

// https://gethttpsforfree.com/

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	gjson "github.com/tbolsh/extend-go-nginx-postgres-docker/genericjson"
	persistense "github.com/tbolsh/extend-go-nginx-postgres-docker/persistense"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"
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
			`create table clients(api_key varchar(64), email varchar(256), password varchar(256), PRIMARY KEY(api_key));`,
		})
		sqlerr(err)
	}()
	rtr := mux.NewRouter()
	rtr.HandleFunc("/alive", alive).Methods("GET")
	rtr.HandleFunc("/version", version).Methods("GET")
	rtr.HandleFunc("/cards", listCards).Methods("GET")
	rtr.HandleFunc("/cards/{card:[A-z0-9\\-_]+}/transactions", listTransactions).Methods("GET")
	rtr.HandleFunc("/cards/{card:[A-z0-9\\-_]+}/transactions/{transaction:[0-9aA-z\\-_]+}", details).Methods("GET")
	// mux.HandleFunc("/cards/")
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: proxy{Handler: rtr},
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

/*
$ curl http://localhost:8008/alive
{"alive": true}
*/
func alive(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(`{"alive": true}`))
}

/*
$ curl http://localhost:8008/version
{"version": true}
*/
func version(w http.ResponseWriter, req *http.Request) {
	content, err := ioutil.ReadFile("/root/version")
	if err != nil {
		log.Println(err)
		content = []byte(fmt.Sprintf("error reading file with version information: %v", err))
	}
	w.Write([]byte(fmt.Sprintf(`{"version": "%s"}`, strings.TrimSpace(string(content)))))
}

func sqlerr(err error) {
	if err != nil {
		log.Printf("SQL Error '%v'", err)
	}
}

/*
$ curl -H "API-Key: xxx" http://localhost:8008/cards
[]
*/
func listCards(w http.ResponseWriter, req *http.Request) {
	if tok, err := signin(req); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		reqOut, _ := http.NewRequest(http.MethodGet, "https://api.paywithextend.com/virtualcards/", nil)
		reqOut.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
		if cards, err := extendAPI(reqOut); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			retval, _ := json.MarshalIndent(cards, "  ", "  ")
			w.Write(retval)
		}
	}
}

/*
$ curl -H "API-Key: xxx" http://localhost:8008/cards/XXX/transactions
[]
*/
func listTransactions(w http.ResponseWriter, req *http.Request) {
	if tok, err := signin(req); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		params := mux.Vars(req)
		reqOut, _ := http.NewRequest(http.MethodGet,
			fmt.Sprintf("https://api.paywithextend.com/virtualcards/%s/transactions", params["card"]), nil)
		reqOut.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
		if cards, err := extendAPI(reqOut); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			retval, _ := json.MarshalIndent(cards, "  ", "  ")
			w.Write(retval)
		}
	}
}

/*
$ curl -H "API-Key: xxx" http://localhost:8008/cards/XXX/transactions/YYY
[]
*/
func details(w http.ResponseWriter, req *http.Request) {
	if tok, err := signin(req); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		params := mux.Vars(req)
		reqOut, _ := http.NewRequest(http.MethodGet,
			fmt.Sprintf("https://api.paywithextend.com/transactions/%s", params["transaction"]), nil)
		reqOut.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))
		if cards, err := extendAPI(reqOut); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			retval, _ := json.MarshalIndent(cards, "  ", "  ")
			w.Write(retval)
		}
	}

}

type token struct {
	Token      string
	Expiration time.Time
	User       gjson.GenJson
}

var cache = make(map[string]token)

func signin(req *http.Request) (string, error) {
	apiKey := req.Header.Get("API-Key")
	if apiKey == "" {
		return "", errors.New("api-Key is not specified!")
	}
	t, ok := cache[apiKey]
	if !ok || t.Expiration.Before(time.Now()) {
		data, err := persistense.Query("SELECT email, password FROM clients WHERE api_key=$1", strings.TrimSpace(apiKey))
		if err != nil {
			return "", fmt.Errorf("api-Key is not found: %s", err)
		}
		if len(data) == 0 {
			return "", fmt.Errorf("api-Key is not found")
		}
		reqOut, err := http.NewRequest(http.MethodPost, "https://api.paywithextend.com/signin",
			strings.NewReader(fmt.Sprintf(`{ "email": "%s", "password": "%s" }`, data[0][0], data[0][1])))
		g, err := extendAPI(reqOut)
		if err != nil {
			return "", err
		}
		t = token{Token: g.StringOrEmpty("token"), Expiration: time.Now().Add(time.Hour), User: g.UnwindOrNil("user")}
		cache[apiKey] = t
	}
	return t.Token, nil
}

func extendAPI(reqOut *http.Request) (gjson.GenJson, error) {
	reqOut.Header.Add("Content-Type", "application/json")
	reqOut.Header.Add("Accept", "application/vnd.paywithextend.v2021-03-12+json")
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(reqOut)
	if err != nil {
		return gjson.FromGeneric(nil), fmt.Errorf("error from extend API: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return gjson.FromGeneric(nil), fmt.Errorf("error reading extend API response: %v", err)
	}
	var retval gjson.GenJson
	if err := json.Unmarshal(body, &retval); err != nil {
		return gjson.FromGeneric(nil), fmt.Errorf("error unmarshaling extend API response: %v (%s)", err, string(body))
	}
	return retval, nil
}
