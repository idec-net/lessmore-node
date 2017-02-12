package main

import (
	"flag"
	"gitea.difrex.ru/Umbrella/lessmore/node"
)

var (
	listen          string
	es              string
	esMessagesIndex string
	esMessagesType  string
)

// init ...
func init() {
	flag.StringVar(&listen, "listen", "127.0.0.1:15582", "Address to listen")
	flag.StringVar(&es, "es", "htt://127.0.0.1:9200", "ES host")
	flag.StringVar(&esMessagesIndex, "esindex", "idec3", "ES index")
	flag.StringVar(&esMessagesType, "estype", "post", "ES index type")
	flag.Parse()
}

// main ...
func main() {
	esconf := node.ESConf{}
	esconf.Host = es
	esconf.Index = esMessagesIndex
	esconf.Type = esMessagesType
	node.Serve(listen, esconf)
}
