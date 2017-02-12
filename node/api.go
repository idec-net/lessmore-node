package node

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

// ListTXTHandler ...
func (es ESConf) ListTXTHandler(w http.ResponseWriter, r *http.Request) {

	// Get echolist
	echoes, err := es.GetListTXT()
	if err != nil {
		w.WriteHeader(500)
	}

	w.WriteHeader(200)
	w.Write(echoes)
}

// Serve ...
func Serve(listen string, es ESConf) {
	r := mux.NewRouter()
	r.HandleFunc("/list.txt", es.ListTXTHandler)
	http.Handle("/list.txt", r)

	srv := http.Server{
		Handler:      r,
		Addr:         listen,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
