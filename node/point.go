package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"bytes"

	idec "github.com/idec-net/go-idec"
	log "github.com/sirupsen/logrus"
)

type ESDoc struct {
	Echo    string `json:"echo"`
	Subg    string `json:"subg"`
	To      string `json:"to"`
	Author  string `json:"author"`
	Message string `json:"message"`
	Date    string `json:"date"`
	MsgID   string `json:"msgid"`
	Tags    string `json:"tags"`
	Repto   string `json:"repto"`
	Address string `json:"address"`
	TopicID string `json:"topicid"`
}

// PointMessage add point message into DB
func (es ESConf) PointMessage(req PointRequest, user User) error {
	pmsg, err := idec.ParsePointMessage(req.Tmsg)
	if err != nil {
		return err
	}
	if err := pmsg.Validate(); err != nil {
		return err
	}

	bmsg, err := idec.MakeBundledMessage(pmsg)
	if err != nil {
		return err
	}

	// Make bundle ID
	id := idec.MakeMsgID(pmsg.String())
	bmsg.ID = id
	bmsg.From = user.Name
	bmsg.Address = fmt.Sprintf("%s,%d", user.Address, user.UserID)

	if err := es.IndexMessage(bmsg); err != nil {
		return err
	}
	return nil
}

func (es ESConf) getTopicID(msgid string) string {
	var topicid string
	if msgid == "" {
		return topicid
	}
	reqURL := fmt.Sprintf("%s/%s/%s/%s", es.Host, es.Index, es.Type, msgid)
	req, err := http.NewRequest("GET", reqURL, strings.NewReader(""))
	if err != nil {
		log.Error(err)
		return topicid
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return topicid
	}

	defer resp.Body.Close()

	var hit Hit
	err = json.NewDecoder(resp.Body).Decode(&hit)
	if err != nil {
		log.Error(err)
		return topicid
	}

	if hit.Source.TopicID != "" {
		topicid = hit.Source.TopicID
	} else if hit.Source.Repto != "" {
		return es.getTopicID(hit.Source.Repto)
	}

	return topicid
}

func (es ESConf) IndexMessage(msg idec.Message) error {
	tags, _ := msg.Tags.CollectTags()
	doc := ESDoc{
		Tags:    tags,
		Echo:    msg.Echo,
		Subg:    msg.Subg,
		To:      msg.To,
		Author:  msg.From,
		Message: msg.Body,
		Date:    fmt.Sprintf("%d", msg.Timestamp),
		Repto:   msg.Repto,
		Address: msg.Address,
		MsgID:   msg.ID,
		TopicID: es.getTopicID(msg.Repto),
	}
	reqURL := fmt.Sprintf("%s/%s/%s/%s", es.Host, es.Index, es.Type, msg.ID)
	bdoc, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", reqURL, bytes.NewReader(bdoc))
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
	content, _ := ioutil.ReadAll(resp.Body)
	log.Info("Message added, response: ", string(content))

	return nil
}
