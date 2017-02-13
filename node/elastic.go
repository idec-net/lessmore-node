package node

import (
	"bytes"
	// "encoding/json"
	"gitea.difrex.ru/Umbrella/fetcher/i2es"
	"github.com/Jeffail/gabs"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	echoAgg = "echo_uniq"
)

// ESConf ...
type ESConf i2es.ESConf

// Bucket ...
type Bucket struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
}

// GetListTXT ...
func (es ESConf) GetListTXT() []byte {
	searchURI := strings.Join([]string{es.Host, es.Index, es.Type, "_search"}, "/")
	searchQ := []byte(`{"aggs": {"echo_uniq": { "terms": { "field": "echo","size": 1000}}}}`)
	log.Print("Search URI: ", searchURI)

	req, err := http.NewRequest("POST", searchURI, bytes.NewBuffer(searchQ))
	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte("")
	}

	esresp, err := gabs.ParseJSON(body)
	if err != nil {
		panic(err)
	}

	var uniq map[string]interface{}
	uniq, _ = esresp.Path(strings.Join([]string{"aggregations", echoAgg}, ".")).Data().(map[string]interface{})

	var echoes []string
	for _, bucket := range uniq["buckets"].([]interface{}) {
		b := make(map[string]interface{})
		b = bucket.(map[string]interface{})
		count := int(b["doc_count"].(float64))
		c := strconv.Itoa(count)
		echostr := strings.Join([]string{b["key"].(string), ":", c, ":"}, "")
		echoes = append(echoes, echostr)
	}

	log.Print("Getting ", len(echoes), " echoes")

	return []byte(strings.Join(echoes, "\n"))
}
