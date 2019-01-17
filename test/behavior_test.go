package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Infnote/infnotechain/protocol"
	"log"
	"os"
	"testing"
)

func printMessage(msg *protocol.Message) {
	var buffer bytes.Buffer
	err := json.Indent(&buffer, msg.Serialize(), "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	_, _ = buffer.WriteTo(os.Stdout)
}

func TestInfoReaction(t *testing.T) {
	info := protocol.Info{
		Peers: 10,
		Chains: map[string]uint64{
			"19AZfrNgBh5sxo5eVytX3K3yQvucS5vc45": 10,
		},
	}

	printMessage(protocol.NewMessage("info", info))

	fmt.Println(info.React())
}

func TestRequestPeers(t *testing.T) {
	req := protocol.RequstPeers{Count: 1}

	fmt.Println(req.React())
}

func TestRequestBlocks(t *testing.T) {
	req := protocol.RequestBlocks{
		ChainID: "19AZfrNgBh5sxo5eVytX3K3yQvucS5vc45",
		To:      10,
	}

	for _, v := range req.React() {
		msg := protocol.NewMessage("response:blocks", v)
		printMessage(msg)
	}
}
