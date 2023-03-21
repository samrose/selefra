package message

import (
	"fmt"
	"testing"
	"time"
)

func TestNewChannel(t *testing.T) {
	channel := NewChannel[string](func(index int, message string) {
		fmt.Println(message)
	})

	childChannel := channel.MakeChildChannel()
	go func() {
		defer func() {
			childChannel.SenderWaitAndClose()
		}()
		for i := 0; i < 10; i++ {
			time.Sleep(time.Second * 1)
			childChannel.Send(time.Now().String())
		}
	}()

	channel.SenderWaitAndClose()
}
