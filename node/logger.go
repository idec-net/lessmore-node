package node

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// LogRequest ...
func LogRequest(r *http.Request) {
	logString := fmt.Sprintf("%s %d %s %s", r.Method, r.ContentLength, r.RequestURI, r.RemoteAddr)
	log.Print("[API] " + logString)
}
