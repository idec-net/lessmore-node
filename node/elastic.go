package node

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"fmt"

	"github.com/Jeffail/gabs"
	log "github.com/Sirupsen/logrus"
)

const (
	echoAgg = "echo_uniq"
)

// MakePlainTextMessage ...
func MakePlainTextMessage(hit interface{}) string {

	h := make(map[string]interface{})
	h = hit.(map[string]interface{})
	s := make(map[string]interface{})
	s = h["_source"].(map[string]interface{})

	m := []string{"ii/ok", s["echo"].(string), s["date"].(string), s["author"].(string), "null", s["to"].(string), s["subg"].(string), "", s["message"].(string)}

	return strings.Join(m, "\n")
}

// GetPlainTextMessage ...
func (es ESConf) GetPlainTextMessage(msgid string) []byte {
	var message []byte

	searchURI := strings.Join([]string{es.Host, es.Index, es.Type, "_search"}, "/")
	searchQ := []byte(strings.Join([]string{
		`{"query": {"match": {"_id": "`, msgid, `"}}}`}, ""))

	req, err := http.NewRequest("POST", searchURI, bytes.NewBuffer(searchQ))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return message
	}

	esresp, err := gabs.ParseJSON(body)
	if err != nil {
		panic(err)
	}

	hits, _ := esresp.Path("hits.hits").Data().([]interface{})

	return []byte(MakePlainTextMessage(hits[0]))
}

// GetEchoMessageHashes ...
func (es ESConf) GetEchoMessageHashes(echo string) []string {
	var hashes []string

	searchURI := strings.Join([]string{es.Host, es.Index, es.Type, "_search"}, "/")
	searchQ := []byte(strings.Join([]string{
		`{"sort": [
            {"date":{ "order": "desc" }},{ "_score":{ "order": "desc" }}],
          "query": {"query_string" : {"fields": ["msgid", "echo"], "query":"`, echo, `"}}, "size": 500}`}, ""))

	req, err := http.NewRequest("POST", searchURI, bytes.NewBuffer(searchQ))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return hashes
	}

	esresp, err := gabs.ParseJSON(body)
	if err != nil {
		panic(err)
	}

	hits, _ := esresp.Path("hits.hits").Data().([]interface{})
	for _, hit := range hits {
		h := make(map[string]interface{})
		h = hit.(map[string]interface{})
		source := make(map[string]interface{})
		source = h["_source"].(map[string]interface{})
		hashes = append(hashes, source["msgid"].(string))
	}

	return hashes
}

// GetLimitedEchoMessageHashes ...
func (es ESConf) GetLimitedEchoMessageHashes(echo string, offset int, limit int) []string {
	var hashes []string

	// Check offset
	var order string
	if offset <= 0 {
		order = "desc"
	} else {
		order = "asc"
	}

	l := strconv.Itoa(limit)

	searchURI := strings.Join([]string{es.Host, es.Index, es.Type, "_search"}, "/")
	searchQ := []byte(strings.Join([]string{
		`{"sort": [
            {"date":{ "order": "`, order, `" }},{ "_score":{ "order": "`, order, `" }}],
          "query": {"query_string" : {"fields": ["msgid", "echo"], "query":"`, echo, `"}}, "size":`, l, `}`}, ""))

	req, err := http.NewRequest("POST", searchURI, bytes.NewBuffer(searchQ))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return hashes
	}

	esresp, err := gabs.ParseJSON(body)
	if err != nil {
		panic(err)
	}

	hits, _ := esresp.Path("hits.hits").Data().([]interface{})
	for _, hit := range hits {
		h := make(map[string]interface{})
		h = hit.(map[string]interface{})
		source := make(map[string]interface{})
		source = h["_source"].(map[string]interface{})
		hashes = append(hashes, source["msgid"].(string))
	}

	return hashes
}

// GetUEchoMessageHashes ...
func (es ESConf) GetUEchoMessageHashes(echoes string) []string {
	var echohashes []string
	// First get echoes list
	el := strings.Split(echoes, "/")

	// Check offset and limit
	var offset int
	var limit int
	withOL := false
	if strings.Contains(el[len(el)-1], ":") {
		oflim := strings.Split(el[len(el)-1], ":")
		o, err := strconv.Atoi(oflim[0])
		l, err := strconv.Atoi(oflim[1])
		if err != nil {
			log.Print(err)
		} else {
			offset = o
			limit = l
			withOL = true
		}
	}

	eh := make(map[string][]string)
	var curEcho string
	for i, echo := range el {
		if echo == "" {
			continue
		}

		if !strings.Contains(echo, ":") {
			curEcho = echo
		}

		if withOL {
			recEcho := es.GetLimitedEchoMessageHashes(curEcho, offset, limit)
			eh[curEcho] = make([]string, len(curEcho))
			eh[curEcho] = append(eh[curEcho], recEcho...)

		} else {
			recEcho := es.GetEchoMessageHashes(curEcho)
			eh[curEcho] = make([]string, len(recEcho))
			eh[curEcho] = append(eh[curEcho], recEcho...)
		}
		if i == len(el) {
			break
		}
	}

	// Make standard output:
	// echo.name
	// Some20SimbolsHash333
	for k, v := range eh {
		echohashes = append(echohashes, k)
		if k == "" {
			continue
		}
		for _, e := range v {
			if e == "" {
				continue
			}
			echohashes = append(echohashes, e)
		}
	}

	return echohashes
}

// GetListTXT ...
func (es ESConf) GetListTXT() []byte {
	var searchURI string
	if es.Index != "" && es.Type != "" {
		searchURI = strings.Join([]string{es.Host, es.Index, es.Type, "_search"}, "/")
	} else {
		searchURI = strings.Join([]string{es.Host, "search"}, "/")
	}
	searchQ := []byte(`{                                                   
 "size": 0,
 "aggs": {
   "uniqueEcho": {
     "cardinality": {
       "field": "echo"
     }
   },
   "echo": {
     "terms": {
       "field": "echo",
       "size": 1000
     }
   }
 }
}`)
	log.Print("Search URI: ", searchURI)

	req, err := http.NewRequest("POST", searchURI, bytes.NewBuffer(searchQ))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
	}

	defer resp.Body.Close()

	var esr EchoAggregations
	err = json.NewDecoder(resp.Body).Decode(&esr)
	if err != nil {
		log.Error(err.Error())
	}
	log.Infof("%+v", esr)

	var echoes []string
	for _, bucket := range esr.EchoAgg["echo"].Buckets {
		echoes = append(echoes, fmt.Sprintf("%s:%d:", bucket.Key, bucket.DocCount))
	}
	log.Print("Getting ", len(echoes), " echoes")

	return []byte(strings.Join(echoes, "\n"))
}
