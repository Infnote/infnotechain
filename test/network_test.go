package test

import (
	"encoding/json"
	"fmt"
	"github.com/Infnote/infnotechain/network"
	"github.com/Infnote/infnotechain/protocol"
	"log"
	"net/url"
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
		log.Fatal(err)
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

// TODO:
func TestURLParse(t *testing.T) {
	raw := "wss://1.2.3.4/websocket"

	obj, err := url.Parse(raw)
	if err != nil {
		log.Fatal(err)
	}

	port := obj.Port()
	if len(port) == 0 {
		if obj.Scheme == "ws" {
			port = "80"
		} else if obj.Scheme == "wss" {
			port = "443"
		}
	}

	result := obj.Scheme + "://" + obj.Hostname() + ":" + port + obj.RequestURI()

	fmt.Println(result)
}
