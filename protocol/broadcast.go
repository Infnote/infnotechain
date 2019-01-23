package protocol

import (
	"encoding/json"
	"fmt"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/utils"
)

// TODO: clear every 1 hour
var broadcastKeys = map[string]bool{}

// TODO: may need to find a better way to save the channel
var BroadcastChannel = make(chan *BroadcastBlock)

type BroadcastBlock struct {
	Block  json.RawMessage `json:"block"`
	block  *blockchain.Block
	ID     string
	Sender interface{}
}

func (b *BroadcastBlock) SetBlock(block *blockchain.Block) {
	b.block = block
}

func (b *BroadcastBlock) Message() *Message {
	msg := NewMessage(b)
	if len(b.ID) > 0 {
		msg.ID = b.ID
	}
	broadcastKeys[b.ID] = true
	return msg
}

func (b BroadcastBlock) Serialize() [] byte {
	data, err := json.Marshal(map[string]json.RawMessage{"block": json.RawMessage(b.block.Serialize())})
	if err != nil {
		utils.L.Fatal(err)
	}
	return data
}

func (b *BroadcastBlock) Validate() *Error {
	block, err := blockchain.DeserializeBlock(b.Block)
	if err != nil {
		return JSONDecodeError(err.Error())
	}
	b.block = block

	if broadcastKeys[b.ID] {
		return DuplicateBroadcastError(b.ID)
	}

	chain := blockchain.LoadChain(b.block.ChainID())
	if chain == nil {
		return ChainNotAcceptError(fmt.Sprintf("recovered chain ID: %v", b.block.ChainID()))
	}

	verr := chain.ValidateBlock(b.block)
	if verr != nil {
		utils.L.Debugf("%v", verr)
		return BlockValidationError(verr)
	}

	return nil
}

func (b *BroadcastBlock) React() []Behavior {
	blockchain.LoadChain(b.block.ChainID()).SaveBlock(b.block)
	broadcastKeys[b.ID] = true
	go func() {
		BroadcastChannel <- b
	}()
	return nil
}
