package test

import (
	"encoding/json"
	"fmt"
	"github.com/Infnote/infnotechain/network"
	"github.com/Infnote/infnotechain/protocol"
	"log"
	"testing"
)

type TestMessageData struct {
	Field1 string `json:"field_1"`
	Field2 int64  `json:"field_2"`
	Field3 bool   `json:"field_3"`
}

func (t TestMessageData) Serialize() []byte {
	r, e := json.Marshal(t)
	if e != nil {
		log.Fatal(e)
	}
	return r
}

func TestMessage(t *testing.T) {
	data := TestMessageData{"Test String", 801, true}
	msg := protocol.NewMessage("testtype", data)
	fmt.Println(string(msg.Serialize()))

	msg2, err := protocol.DeserializeMessage(msg.Serialize())
	if err != nil {
		log.Fatal(err)
	}
	data2 := TestMessageData{}
	e := json.Unmarshal(msg2.Data, &data2)
	if e != nil {
		log.Fatal(e)
	}
	fmt.Println(data2)
}

func TestEcho(t *testing.T) {
	s := network.NewServer()
	go s.Serve()
	go func() {
		peer := <- s.In
		peer.Send <- protocol.NewMessage("test", TestMessageData{"Test String", 801, true}).Serialize()
		fmt.Println(<- peer.Recv)
	}()
	<- s.Out
}
