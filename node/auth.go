package node

import (
	"fmt"

	"net/http"

	"strings"

	"io/ioutil"

	"encoding/json"

	"errors"

	"bytes"

	"time"

	log "github.com/sirupsen/logrus"
)

const (
	USERS_INDEX        = ".lessmore_points"
	USERS_DOC_TYPE     = "points"
	NODE_ADDRESS       = "dynamic"
	AUTH_STRING_LENGTH = 16
	LETTERS            = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	SALT_BYTES         = 32
	HASH_BYTES         = 64
)

// User document structure
type User struct {
	// Will be added to the address:
	// i.e. dynamic,1
	UserID int64 `json:"user_id"`
	// Will be added to the bundled message
	Name string `json:"name"`
	// Email address needs for password restore
	Email      string `json:"email"`
	AuthString string `json:"auth_string"`
	Address    string `json:"address"`
	// Created time
	Created int64 `json:"created"`
}

// checkAuth token in point request
// do a search by the auth_string field
func (es ESConf) checkAuth(r PointRequest) (User, bool) {
	reqURL := fmt.Sprintf("%s/%s/_search", es.Host, USERS_INDEX)
	query := `{"query": {"term": { "auth_string": "%s" }}}`
	query = fmt.Sprintf(query, r.Pauth)

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(query))
	if err != nil {
		log.Error("Can't prepare a request: ", err)
		return User{}, false
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Can't make a request: ", err)
		return User{}, false
	}
	defer resp.Body.Close()

	var esr MaxIdAggregation
	err = json.NewDecoder(resp.Body).Decode(&esr)
	if err != nil {
		log.Error("Can't decode a response from ES: ", err)
		return User{}, false
	}

	if len(esr.Hits.Hits) == 1 && esr.Hits.Hits[0].Source.AuthString == r.Pauth {
		return esr.Hits.Hits[0].Source, true
	}
	return User{}, false
}

// DeleteUser from users index
func DeleteUser(name string) error {
	return nil
}

// AddNewUser to the .lessmore_users index
func (es ESConf) AddNewUser(name, email string) (User, error) {
	var user User
	if err := es.checkUser(name); err != nil {
		return user, err
	}

	max, err := es.getMaxUser()
	if err != nil {
		log.Fatal(err)
	}

	user.Name = name
	user.UserID = max + 1
	user.Address = NODE_ADDRESS
	user.AuthString = string(genAuthString())
	user.Created = time.Now().Unix()

	err = es.IndexUser(user)
	if err != nil {
		return user, err
	}

	return user, nil
}

// IndexUser in `USERS_INDEX` index
func (es ESConf) IndexUser(user User) error {
	reqURL := fmt.Sprintf("%s/%s/%s/%d", es.Host, USERS_INDEX, USERS_DOC_TYPE, user.UserID)
	js, err := json.Marshal(user)
	if err != nil {
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest("PUT", reqURL, bytes.NewReader(js))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (es ESConf) checkUser(name string) error {
	reqURL := es.Host + "/" + USERS_INDEX + "/_search"
	reqName := `{"query": {"term": { "name": "%s" }}}`
	reqName = fmt.Sprintf(reqName, name)
	req, err := http.NewRequest("POST", reqURL, strings.NewReader(reqName))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var esr MaxIdAggregation
	err = json.NewDecoder(resp.Body).Decode(&esr)
	if err != nil {
		return err
	}
	if len(esr.Hits.Hits) > 0 {
		return errors.New(fmt.Sprintf("User %s alredy exists", name))
	}

	return nil
}

func (es ESConf) getMaxUser() (int64, error) {
	ok, err := es.checkIndex()
	if err != nil {
		return -1, err
	}
	if !ok {
		if err := es.createIndex(); err != nil {
			return -1, err
		}
		return 0, nil
	}

	usersSearchURL := es.Host + "/" + USERS_INDEX + "/_search"
	usersSearchReq := `
{
  "aggs": {
    "max_id": { "max": { "field": "user_id" } }
  },
  "size": 0
}
`
	client := http.Client{}
	req, err := http.NewRequest("POST", usersSearchURL, strings.NewReader(usersSearchReq))
	if err != nil {
		return -1, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return -1, err
	}

	defer resp.Body.Close()

	content, _ := ioutil.ReadAll(resp.Body)
	var esr MaxIdAggregation
	err = json.NewDecoder(strings.NewReader(string(content))).Decode(&esr)
	if err != nil {
		log.Error("Cant parse JSON")
		return -1, err
	}

	return int64(esr.MaxID["max_id"].Value), nil
}

func (es ESConf) checkIndex() (bool, error) {
	indexListURL := es.Host + "/_cat/indices"
	// Initialize http client
	client := http.Client{}
	indicesReq, err := http.NewRequest("GET", indexListURL, strings.NewReader(""))
	if err != nil {
		log.Error(err)
		return false, err
	}
	indicesResp, err := client.Do(indicesReq)
	if err != nil {
		log.Error(err)
		return false, err
	}

	defer indicesResp.Body.Close()

	list, err := ioutil.ReadAll(indicesResp.Body)
	if err != nil {
		return false, err
	}

	if strings.Contains(string(list), USERS_INDEX) {
		return true, nil
	}

	return false, nil
}

func (es ESConf) createIndex() error {
	mapping := `
{
  "mappings": {
    "%s": { 
      "properties": { 
        "user_id":     { "type": "integer"  },
        "name":        { "type": "keyword"  },
        "email":       { "type": "keyword"  },
        "address":     { "type": "keyword"  },
        "auth_string": { "type": "keyword"  },
        "created":  {
          "type":   "date",
          "format": "strict_date_optional_time||epoch_second"
        }
      }
    }
  }
}
`
	mapping = fmt.Sprintf(mapping, USERS_DOC_TYPE)

	reqURL := fmt.Sprintf("%s/%s", es.Host, USERS_INDEX)
	req, err := http.NewRequest("PUT", reqURL, strings.NewReader(mapping))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	log.Warn("Creating new users index")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Warn("Created new index mapping")
	fmt.Println(string(content))

	return nil
}
