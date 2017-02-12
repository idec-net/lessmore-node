package node

import (
	"bytes"
	// "encoding/json"
	"gitea.difrex.ru/Umbrella/fetcher/i2es"
	"github.com/Jeffail/gabs"
	"io/ioutil"
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
func (es ESConf) GetListTXT() ([]byte, error) {
	searchURI := strings.Join([]string{es.Host, es.Index, es.Type, "_search"}, "/")
	searchQ := []byte(`{"aggs": {"echo_uniq": { "terms": { "field": "echo","size": 1000}}}}`)

	req, err := http.NewRequest("POST", searchURI, bytes.NewBuffer(searchQ))
	client := &http.Client{}
	resp, err := client.Do(req)

	// defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}

	esresp, err := gabs.ParseJSON(body)
	if err != nil {
		panic(err)
	}
	var buckets []Bucket
	buckets, _ = esresp.Path("aggregations.echo_uniq.buckets").Data().([]Bucket)

	var echoes []string
	for _, bucket := range buckets {
		c := strconv.Itoa(bucket.DocCount)
		echostr := strings.Join([]string{bucket.Key, ":", c, ":"}, "")
		echoes = append(echoes, echostr)
	}

	return []byte(strings.Join(echoes, "\n")), nil
}
