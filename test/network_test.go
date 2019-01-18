package test

import (
	"encoding/json"
	"fmt"
	"github.com/Infnote/infnotechain/network"
	"github.com/Infnote/infnotechain/protocol"
	"log"
	"testing"
)

func TestMessage(t *testing.T) {
	data := protocol.NewInfo()
	msg := protocol.NewMessage(data)
	fmt.Println(string(msg.Serialize()))

	msg2, err := protocol.DeserializeMessage(msg.Serialize())
	if err != nil {
		log.Fatal(err)
	}
	data2 := &protocol.Info{}
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
		peer.Send <- protocol.NewMessage(protocol.NewInfo()).Serialize()
		fmt.Println(<- peer.Recv)
	}()
	<- s.Out
}
