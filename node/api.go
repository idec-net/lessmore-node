package node

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
	"time"
)

// ListTXTHandler ...
func (es ESConf) ListTXTHandler(w http.ResponseWriter, r *http.Request) {

	LogRequest(r)

	ch := make(chan []byte)
	// Get echolist
	go func() {
		ch <- es.GetListTXT()
	}()

	echoes := <-ch

	w.WriteHeader(200)
	w.Write(echoes)
}

// XFeaturesHandler list supported features
func XFeaturesHandler(w http.ResponseWriter, r *http.Request) {
	features := []string{"list.txt", "x/features"}

	LogRequest(r)

	w.WriteHeader(200)
	w.Write([]byte(strings.Join(features, "\n")))
}

// EHandler /e/ schema
func (es ESConf) EHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	echo := vars["echo"]

	LogRequest(r)

	ch := make(chan []string)
	// Get echolist
	go func() {
		ch <- es.GetEchoMessageHashes(echo)
	}()

	messages := <-ch

	w.WriteHeader(200)
	w.Write([]byte(strings.Join(messages, "\n")))
}

// Serve ...
func Serve(listen string, es ESConf) {
	r := mux.NewRouter()
	r.HandleFunc("/list.txt", es.ListTXTHandler)
	r.HandleFunc("/x/features", XFeaturesHandler)

	// Standart schemas
	r.HandleFunc("/e/{echo}", es.EHandler)

	http.Handle("/", r)

	srv := http.Server{
		Handler:      r,
		Addr:         listen,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
