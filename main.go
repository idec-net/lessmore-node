package main

import (
	"flag"
	"os"

	"gitea.difrex.ru/Umbrella/lessmore/node"
	log "github.com/Sirupsen/logrus"
)

var (
	listen          string
	es              string
	esMessagesIndex string
	esMessagesType  string
	add             string
	email           string
)

// init ...
func init() {
	flag.StringVar(&listen, "listen", "127.0.0.1:15582", "Address to listen")
	flag.StringVar(&es, "es", "http://127.0.0.1:9200", "ES host")
	flag.StringVar(&esMessagesIndex, "esindex", "idec3", "ES index")
	flag.StringVar(&esMessagesType, "estype", "post", "ES index type")
	flag.StringVar(&add, "add", "", "User to add")
	flag.StringVar(&email, "email", "", "User email address")
	flag.Parse()

	log.SetLevel(log.DebugLevel)
}

// main ...
func main() {
	esconf := node.ESConf{}
	esconf.Host = es
	esconf.Index = esMessagesIndex
	esconf.Type = esMessagesType
	if add != "" {
		addUser(add, esconf)
	}
	node.Serve(listen, esconf)
}

func addUser(name string, esconf node.ESConf) {
	user, err := esconf.AddNewUser(add, email)
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
	log.Infof("Created: %+v", user)
	os.Exit(0)
}
