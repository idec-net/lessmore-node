package node

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

// ListTXTHandler ...
func (es ESConf) ListTXTHandler(w http.ResponseWriter, r *http.Request) {

	ch := make(chan []byte)
	// Get echolist
	go func() {
		ch <- es.GetListTXT()
	}()

	echoes := <-ch

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
