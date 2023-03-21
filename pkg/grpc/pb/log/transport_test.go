package log

import (
	"testing"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTransportWsMsg(t *testing.T) {
	v, err := TransportWsMsg(&UploadLogStream_Request{
		Stage: 1,
		Index: 1,
		Msg:   "tttt",
		Level: 1,
		Time:  timestamppb.Now(),
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(v)
}
