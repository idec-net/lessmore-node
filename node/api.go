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
	features := []string{"list.txt", "u/e"}

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

// UEHandler /u/e/ schema
func (es ESConf) UEHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e := vars["echoes"]

	log.Print("/u/e/ vars: ", e)

	LogRequest(r)

	ch := make(chan []string)
	// Get echolist
	go func() {
		ch <- es.GetUEchoMessageHashes(e)
	}()

	messages := <-ch

	w.WriteHeader(200)
	w.Write([]byte(strings.Join(messages, "\n")))
}

// MHandler /m/ schema
func (es ESConf) MHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	msgid := vars["msgid"]

	LogRequest(r)

	ch := make(chan []byte)
	// Get echolist
	go func() {
		ch <- es.GetPlainTextMessage(msgid)
	}()

	message := <-ch

	w.WriteHeader(200)
	w.Write(message)
}

// Serve ...
func Serve(listen string, es ESConf) {
	r := mux.NewRouter()
	r.HandleFunc("/list.txt", es.ListTXTHandler)
	r.HandleFunc("/x/features", XFeaturesHandler)

	// Standart schemas
	r.HandleFunc("/e/{echo}", es.EHandler)
	r.HandleFunc("/m/{msgid}", es.MHandler)

	// Extensions
	r.HandleFunc("/u/e/{echoes:[a-z0-9-_/.:]+}", es.UEHandler)

	http.Handle("/", r)

	srv := http.Server{
		Handler:      r,
		Addr:         listen,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Print("Listening IDEC API on ", listen)
	log.Fatal(srv.ListenAndServe())
}
