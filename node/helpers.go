package node

import (
	"time"

	mrand "math/rand"

	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func hashAndSalt(authString []byte) string {
	hash, err := bcrypt.GenerateFromPassword(authString, bcrypt.MinCost)
	if err != nil {
		log.Fatal(err.Error())
	}
	return string(hash)
}

// genAuthString random auth string
func genAuthString() []byte {
	authString := make([]byte, AUTH_STRING_LENGTH)
	mrand.Seed(time.Now().UnixNano())

	for i := range authString {
		authString[i] = LETTERS[mrand.Intn(len(LETTERS))]
	}

	return authString
}

// parsePointBody without regexp
// it dirty hack mades because *http.Request.Form not parsed it
func parsePointBody(content string) (string, string) {
	var pauth, tmsg string
	log.Debug("Body parser: ", content)
	if strings.Contains(content, "&") {
		log.Debug("Found &")
		for _, v := range strings.Split(content, "&") {
			if strings.Contains(v, "pauth") {
				log.Debug("Found pauth")
				pauth = strings.Split(v, "pauth=")[1]
			}
			if strings.Contains(v, "tmsg") {
				log.Debug("Found tmsg")
				tmsg = strings.Split(v, "tmsg=")[1]
			}
		}
	}

	return pauth, tmsg
}

// b64replace +,/,-,_ with A and Z
func b64replace(s string) string {
	s = strings.Replace(s, "+", "A", -1)
	s = strings.Replace(s, "-", "A", -1)
	s = strings.Replace(s, "/", "Z", -1)
	s = strings.Replace(s, "_", "Z", -1)
	return s
}
