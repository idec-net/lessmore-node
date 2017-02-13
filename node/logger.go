package node

import (
	"log"
	"net/http"
	"strings"
)

// LogRequest ...
func LogRequest(r *http.Request) {
	logString := strings.Join([]string{r.Method, string(r.ContentLength), r.RequestURI, r.RemoteAddr}, " ")
	log.Print("[API] ", logString)
}
