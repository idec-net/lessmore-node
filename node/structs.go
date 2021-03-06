package node

import "gitea.difrex.ru/Umbrella/fetcher/i2es"

// PointRequest with message
type PointRequest struct {
	Pauth string `json:"pauth"`
	Tmsg  string `json:"tmsg"`
}

// ESConf ...
type ESConf i2es.ESConf

// Bucket ...
type Bucket struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
}

// {"took":467,"timed_out":false,"_shards":{"total":5,"successful":5,"skipped":0,"failed":0},"hits":{"total":89333,"max_score":0.0,"hits":[]}}
type ESSearchResp struct {
	Took           int64 `json:"took"`
	TimedOut       bool  `json:"timed_out"`
	ESSearchShards `json:"_shards"`
	Hits           `json:"hits"`
}

type Hits struct {
	Total    int64   `json:"total"`
	MaxScore float64 `json:"max_score"`
	Hits     []Hit   `json:"hits"`
}

// {"_index":"idec5","_type":"post","_id":"aAjSbXS5XeNF6lVaPh5A","_score":1.0,"_source"
type Hit struct {
	Index  string     `json:"_index"`
	Type   string     `json:"_type"`
	ID     string     `json:"_id"`
	Source i2es.ESDoc `json:"_source"`
}

type UserHits struct {
	Total    int64     `json:"total"`
	MaxScore float64   `json:"max_score"`
	Hits     []UserHit `json:"hits"`
}

// { "_index":".lessmore_users","_type":"user","_id":"1","_score":1.0,
//   "_source":{"id": 1, "address": "dynamic", "name": "name", "authString": "thisIsAtest"}}
type UserHit struct {
	Index  string  `json:"_index"`
	Type   string  `json:"_type"`
	ID     string  `json:"_id"`
	Score  float64 `json:"_score"`
	Source User    `json:"_source"`
}

// "aggregations":{"echo":{"doc_count_error_upper_bound":2406,"sum_other_doc_count":76555,"buckets":[{"key":"bash.rss","doc_count":12779}]},"uniqueEcho":{"value":121}}}
type EchoAggregations struct {
	EchoAgg  map[string]EchoAgg `json:"aggregations"`
	UniqEcho map[string]Uniq    `json:"uniqueEcho"`
}

type MaxIdAggregation struct {
	Hits  UserHits        `json:"hits"`
	MaxID map[string]Uniq `json:"aggregations"`
}

type EchoAgg struct {
	DocCountErrorUpperBound int64    `json:"doc_count_error_upper_bound"`
	SumOtherDocCount        int64    `json:"sum_other_doc_count"`
	Buckets                 []Bucket `json:"buckets"`
}

type EchoBucket struct {
	Key   string `json:"key"`
	Count int64  `json:"doc_count"`
}

type Uniq struct {
	Value float64 `json:"value"`
}

type ESSearchShards struct {
	Total      int64 `json:"total"`
	Successful int64 `json:"successful"`
	Skipped    int64 `json:"skipped"`
	Failed     int64 `json:"failed"`
}
