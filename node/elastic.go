package node

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"fmt"

	"encoding/base64"

	"gitea.difrex.ru/Umbrella/fetcher/i2es"
	log "github.com/Sirupsen/logrus"
)

const (
	echoAgg = "echo_uniq"
)

// MakePlainTextMessage ...
func MakePlainTextMessage(hit i2es.ESDoc) []byte {
	tags := "ii/ok"
	if hit.Repto != "" {
		tags += fmt.Sprintf("/repto/%s", hit.Repto)
	}
	m := []string{
		tags,
		hit.Echo,
		hit.Date,
		hit.Author,
		hit.Address,
		hit.To,
		hit.Subg,
		hit.Message,
	}

	return []byte(strings.Join(m, "\n"))
}

// GetPlainTextMessage ...
func (es ESConf) GetPlainTextMessage(msgid string) []byte {
	var searchURI string
	if es.Index != "" && es.Type != "" {
		searchURI = strings.Join([]string{es.Host, es.Index, es.Type, "_search"}, "/")
	} else {
		searchURI = strings.Join([]string{es.Host, "search"}, "/")
	}

	searchQ := []byte(strings.Join([]string{
		`{"query": {"match": {"_id": "`, msgid, `"}}}`}, ""))

	req, err := http.NewRequest("POST", searchURI, bytes.NewBuffer(searchQ))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
	}

	defer resp.Body.Close()
	var esr ESSearchResp
	err = json.NewDecoder(resp.Body).Decode(&esr)
	if err != nil {
		log.Error(err.Error())
		return []byte("")
	}

	if len(esr.Hits.Hits) > 0 {
		return MakePlainTextMessage(esr.Hits.Hits[0].Source)
	}

	return []byte("")
}

// GetEchoMessageHashes ...
func (es ESConf) GetEchoMessageHashes(echo string) []string {
	var hashes []string
	var searchURI string
	if es.Index != "" && es.Type != "" {
		searchURI = strings.Join([]string{es.Host, es.Index, es.Type, "_search"}, "/")
	} else {
		searchURI = strings.Join([]string{es.Host, "search"}, "/")
	}

	searchQ := []byte(strings.Join([]string{
		`{"sort": [
            {"date":{ "order": "desc" }},{ "_score":{ "order": "desc" }}],
          "query": {"query_string" : {"fields": ["msgid", "echo"], "query":"`, echo, `"}}, "size": 500}`}, ""))

	req, err := http.NewRequest("POST", searchURI, bytes.NewBuffer(searchQ))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return hashes
	}

	defer resp.Body.Close()

	var esr ESSearchResp
	err = json.NewDecoder(resp.Body).Decode(&esr)
	if err != nil {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Error(string(b))
		log.Error(err.Error())
		hashes = append(hashes, "error: Internal error")
		return hashes
	}

	for _, hit := range esr.Hits.Hits {
		hashes = append(hashes, hit.Source.MsgID)
	}

	return hashes
}

// GetLimitedEchoMessageHashes ...
func (es ESConf) GetLimitedEchoMessageHashes(echo string, offset int, limit int) []string {
	var hashes []string
	var searchURI string
	if es.Index != "" && es.Type != "" {
		searchURI = strings.Join([]string{es.Host, es.Index, es.Type, "_search"}, "/")
	} else {
		searchURI = strings.Join([]string{es.Host, "search"}, "/")
	}

	// Check offset
	var order string
	if offset <= 0 {
		order = "desc"
	} else {
		order = "asc"
	}

	l := strconv.Itoa(limit)

	searchQ := []byte(strings.Join([]string{
		`{"sort": [
            {"date":{ "order": "`, order, `" }},{ "_score":{ "order": "`, order, `" }}],
          "query": {"query_string" : {"fields": ["msgid", "echo"], "query":"`, echo, `"}}, "size":`, l, `}`}, ""))

	req, err := http.NewRequest("POST", searchURI, bytes.NewBuffer(searchQ))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return hashes
	}

	defer resp.Body.Close()

	var esr ESSearchResp
	err = json.NewDecoder(resp.Body).Decode(&esr)
	if err != nil {
		log.Error(err.Error())
		return hashes
	}
	for _, hit := range esr.Hits.Hits {
		hashes = append(hashes, hit.Source.MsgID)
	}

	return hashes
}

func (es ESConf) GetUMMessages(msgs string) []string {
	var encodedMessages []string

	// First get messages list
	messages := strings.Split(msgs, "/")
	var searchURI string
	if es.Index != "" && es.Type != "" {
		searchURI = strings.Join([]string{es.Host, es.Index, es.Type, "_search"}, "/")
	} else {
		searchURI = strings.Join([]string{es.Host, "search"}, "/")
	}
	query := []byte(`
{
  "query": {
    "query_string" : {
      "fields": ["msgid"],
      "query":"` + strings.Join(messages, " OR ") + `"
    }
  },
  "sort": [{"date":{ "order": "desc" }},
           { "_score":{ "order": "desc" }}
  ]
}`)
	req, err := http.NewRequest("POST", searchURI, bytes.NewBuffer(query))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return encodedMessages
	}

	defer resp.Body.Close()

	var esr ESSearchResp
	err = json.NewDecoder(resp.Body).Decode(&esr)
	if err != nil {
		log.Error(err.Error())
		return encodedMessages
	}

	for _, hit := range esr.Hits.Hits {
		m := fmt.Sprintf("%s:%s", hit.Source.MsgID, base64.StdEncoding.EncodeToString(MakePlainTextMessage(hit.Source)))
		encodedMessages = append(encodedMessages, m)
	}

	return encodedMessages
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

// GetXC implements /x/c
func (es ESConf) GetXC(echoes string) []string {
	var searchURI string
	var counts []string
	if es.Index != "" && es.Type != "" {
		searchURI = strings.Join([]string{es.Host, es.Index, es.Type, "_search"}, "/")
	} else {
		searchURI = strings.Join([]string{es.Host, "search"}, "/")
	}

	query := []byte(`
{
    "query": {
        "query_string" : {
            "fields": ["echo"],
            "query": "` + strings.Join(strings.Split(echoes, "/"), " OR ") + `"
        }
    },
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
}
`)
	req, err := http.NewRequest("POST", searchURI, bytes.NewBuffer(query))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return counts
	}

	defer resp.Body.Close()

	var esr EchoAggregations
	err = json.NewDecoder(resp.Body).Decode(&esr)
	if err != nil {
		log.Error(err.Error())
		return counts
	}
	log.Infof("%+v", esr)

	for _, hit := range esr.EchoAgg["echo"].Buckets {
		counts = append(counts, fmt.Sprintf("%s:%d", hit.Key, hit.DocCount))
	}
	return counts
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
