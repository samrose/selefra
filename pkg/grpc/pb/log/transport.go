package log

import (
	"time"

	"github.com/songzhibin97/gkit/coding"
	_ "github.com/songzhibin97/gkit/coding/json"
)

var jsonCoding = coding.GetCode("json")

func TransportWsMsg(rec *UploadLogStream_Request) (ret []byte, err error) {
	return jsonCoding.Marshal(&struct {
		Stage int       `json:"stage"`
		Index uint64    `json:"index"`
		Msg   string    `json:"msg"`
		Level int       `json:"level"`
		Time  time.Time `json:"time"`
	}{
		Stage: int(rec.GetStage()),
		Index: rec.GetIndex(),
		Msg:   rec.GetMsg(),
		Level: int(rec.GetLevel()),
		Time:  rec.GetTime().AsTime(),
	})
}
