package node

import (
	"log"
	"time"

	mrand "math/rand"

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
