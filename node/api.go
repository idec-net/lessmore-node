package node

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
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

func (ed ESConf) BlacklistTXT(w http.ResponseWriter, r *http.Request) {
	LogRequest(r)
	w.WriteHeader(200)
	w.Write([]byte(`KNMUXRTMJA6XWHMB22CD
6AKO6DEWYF7EHXWMI2BY
qcYj9ceDYjoxt2z6qhzN
k1zaEUu1Tg0g97osDeS7
j8s7ZMdTzmHzsRJna3xb
vuMNfQWe8xonMFPZtOxP
bcp6izdkdCj9AXUOP2aT
0dIDXSJTLtR8N2KnLTl1
2QzNqUiyuoPcPn0l74KP
`))
}

// XFeaturesHandler list supported features
func XFeaturesHandler(w http.ResponseWriter, r *http.Request) {
	features := []string{"list.txt", "u/e", "u/m", "x/c"}

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

// UMHandler /u/m/ schema
func (es ESConf) UMHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e := vars["ids"]

	log.Print("/u/m/ vars: ", e)

	LogRequest(r)

	ch := make(chan []string)
	// Get echolist
	go func() {
		ch <- es.GetUMMessages(e)
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

// XCHandler /x/c schema
func (es ESConf) XCHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	echoes := vars["echoes"]

	LogRequest(r)

	ch := make(chan []string)
	// Get echolist
	go func() {
		ch <- es.GetXC(echoes)
	}()

	counts := <-ch

	w.WriteHeader(200)
	w.Write([]byte(strings.Join(counts, "\n")))
}

// UPointHandler /u/point scheme
func (es ESConf) UPointHandler(w http.ResponseWriter, r *http.Request) {
	var req PointRequest
	LogRequest(r)

	// Log request
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Fail to parse POST body: ", err.Error())
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("error: %s", err.Error())))
		return
	}
	log.Debugf("Point request is: ", string(content))

	// Get plain POST variables
	if err := r.ParseForm(); err != nil {
		log.Error("Fail to parse POST args: ", err.Error())
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("error: %s", err.Error())))
		return
	}
	pauth := r.Form.Get("pauth")
	tmsg := r.Form.Get("tmsg")
	if pauth == "" && tmsg == "" {
		log.Debug("Trying parse body request")
		pauth, tmsg = parsePointBody(string(content))
	}

	req.Pauth = pauth
	req.Tmsg = tmsg

	log.Debugf("pauth: %s\ntmsg: %s", pauth, tmsg)

	if pauth == "" {
		w.WriteHeader(403)
		w.Write([]byte("auth error"))
		return
	}
	// Authorization check
	user, ok := es.checkAuth(req)
	if !ok {
		w.WriteHeader(403)
		w.Write([]byte("auth error"))
		return
	}

	// Proccess point message
	if err := es.PointMessage(req, user); err != nil {
		log.Error("Fail to parse point message: ", err.Error())
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("error: %s", err.Error())))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("msg ok"))
}

// Serve ...
func Serve(listen string, es ESConf) {
	r := mux.NewRouter()
	r.HandleFunc("/list.txt", es.ListTXTHandler).Methods("GET")
	r.HandleFunc("/blacklist.txt", es.BlacklistTXT).Methods("GET")
	r.HandleFunc("/x/features", XFeaturesHandler).Methods("GET")

	// Standart schemas
	r.HandleFunc("/e/{echo}", es.EHandler).Methods("GET")
	r.HandleFunc("/m/{msgid}", es.MHandler).Methods("GET")

	// Extensions
	r.HandleFunc("/u/e/{echoes:[a-z0-9-_/.:]+}", es.UEHandler).Methods("GET")
	r.HandleFunc("/u/m/{ids:[a-zA-Z0-9-_/.:]+}", es.UMHandler).Methods("GET")
	r.HandleFunc("/x/c/{echoes:[a-zA-Z0-9-_/.:]+}", es.XCHandler).Methods("GET")

	// Point methods
	r.HandleFunc("/u/point", es.UPointHandler).Methods("POST")

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
