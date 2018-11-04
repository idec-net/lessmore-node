package node

import (
	log "github.com/Sirupsen/logrus"
	"github.com/idec-net/go-idec"
)

// PointMessage add point message into DB
func (es ESConf) PointMessage(req PointRequest) error {
	msg, err := idec.ParsePointMessage(req.Tmsg)
	if err != nil {
		return err
	}

	log.Infof("%+v", msg)
	return nil
}
