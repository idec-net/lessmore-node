package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"bytes"

	log "github.com/Sirupsen/logrus"
	idec "github.com/idec-net/go-idec"
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
	// Prevent collission via adding Timestamp
	id := idec.MakeMsgID(fmt.Sprintf("%s\n%d", pmsg.String(), bmsg.Timestamp))
	bmsg.ID = id
	bmsg.From = user.Name
	bmsg.Address = fmt.Sprintf("%s,%d", user.Address, user.UserID)

	if err := es.IndexMessage(bmsg); err != nil {
		return err
	}
	return nil
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
