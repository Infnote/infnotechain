package protocol

import (
	"encoding/json"
	"fmt"
	"github.com/Infnote/infnotechain/blockchain"
	"log"
	"net/url"
	"strconv"
)

type Behavior interface {
	// returns an Error if invalid, otherwise returns nil
	Validate() *Error

	// Call Validate() before this method
	React() []Behavior
}

// - Declarations
type Info struct {
	Version  string            `json:"version"`
	Peers    int               `json:"peers"`
	Chains   map[string]uint64 `json:"chains"`
	Platform map[string]string `json:"platform"`
	FullNode bool              `json:"full_node"`
}

type RequstPeers struct {
	Count int `json:"count"`
}

type RequestBlocks struct {
	ChainID string `json:"chain_id"`
	From    uint64 `json:"from"`
	To      uint64 `json:"to"`
}

type ResponsePeers struct {
	Peers []string `json:"peers"`
}

type ResponseBlocks struct {
	Blocks []json.RawMessage `json:"blocks"`
	blocks []*blockchain.Block
}

type BroadcastBlock struct {
	Block json.RawMessage `json:"block"`
	block *blockchain.Block
}

// - Serializations
func (b ResponseBlocks) Serialize() []byte {
	var blocks []json.RawMessage
	for _, block := range b.blocks {
		blocks = append(blocks, json.RawMessage(block.Serialize()))
	}
	data, err := json.Marshal(blocks)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func (b BroadcastBlock) Serialize() [] byte {
	data, err := json.Marshal(map[string]json.RawMessage{"block": json.RawMessage(b.block.Serialize())})
	if err != nil {
		log.Fatal(err)
	}
	return data
}

// - Validations
func (b Info) Validate() *Error {
	version, err := strconv.ParseFloat(b.Version, 32)
	if err != nil || version != 1.1 {
		return IncompatibleProtocolVersion("only accept v1.1 protocol")
	}

	if b.Peers < 0 {
		return BadRequestError("'peers' needs to be a non-negative number")
	}

	return nil
}

func (b RequestBlocks) Validate() *Error {
	chain := blockchain.LoadChain(b.ChainID)
	if chain == nil {
		return ChainNotAcceptError(b.ChainID)
	}
	if b.From > b.To {
		return BadRequestError("'from' must greater or equal 'to'")
	}
	if chain.Count < b.From {
		return BadRequestError("request not existed blocks")
	}
	return nil
}

func (b RequstPeers) Validate() *Error {
	if b.Count < 0 {
		return BadRequestError("'count' needs to be a non-negative number")
	}
	return nil
}

func (b ResponsePeers) Validate() *Error {
	for _, v := range b.Peers {
		addr, err := url.Parse(v)
		if err != nil {
			return URLError(err.Error())
		}
		if addr.Scheme != "ws" && addr.Scheme != "wss" {
			return URLError("not a websocket URL")
		}
	}
	return nil
}

func (b *ResponseBlocks) Validate() *Error {
	for _, v := range b.Blocks {
		block := &blockchain.Block{}
		err := json.Unmarshal(v, block)
		if err != nil {
			return JSONDecodeError(err.Error())
		}
		if !block.IsValid() {
			return InvalidBlockError("invalid block hash value")
		}

		if blockchain.LoadChain(block.ChainID()) == nil {
			return ChainNotAcceptError(block.ChainID())
		}

		b.blocks = append(b.blocks, block)
	}
	return nil
}

func (b *BroadcastBlock) Validate() *Error {
	b.block = &blockchain.Block{}
	err := json.Unmarshal(b.Block, b.block)
	if err != nil {
		return JSONDecodeError(err.Error())
	}

	chain := blockchain.LoadChain(b.block.ChainID())
	if chain == nil {
		return ChainNotAcceptError(b.block.ChainID())
	}

	if b.block.Height < chain.Count {
		return BlockAlreadyExistError(fmt.Sprintf("height %v on %v", b.block.Height, chain.ID))
	}
	return nil
}

// - Reactions
func (b Info) React() []Behavior {
	var behaviors []Behavior
	if b.Peers > 0 {
		behaviors = append(behaviors, RequstPeers{b.Peers})
	}
	for k, v := range b.Chains {
		chain := blockchain.LoadChain(k)
		if chain == nil {
			continue
		}
		if chain.Count >= v {
			continue
		}
		behaviors = append(behaviors, RequestBlocks{k, chain.Count, v - 1})
	}
	return behaviors
}

// Split blocks for every 1 MB
func (b RequestBlocks) React() []Behavior {
	chain := blockchain.LoadChain(b.ChainID)

	var behaviors []Behavior
	var blocks []*blockchain.Block
	size := 0
	for i := b.From; i <= b.To; i++ {
		block := chain.GetBlock(i)
		if block == nil {
			break
		}

		// TODO: extract magic number
		if size+block.Size() > 1024*1024 {
			behaviors = append(behaviors, &ResponseBlocks{blocks: blocks})
			blocks = []*blockchain.Block{block}
			size = block.Size()
		} else {
			blocks = append(blocks, block)
			size += block.Size()
		}
	}
	if len(blocks) > 0 {
		behaviors = append(behaviors, &ResponseBlocks{blocks: blocks})
	}
	if len(behaviors) > 0 {
		return behaviors
	}
	return nil
}

func (b RequstPeers) React() []Behavior {
	return []Behavior{ResponsePeers{[]string{"wss://chain.infnote.com:32767/"}}}
}

func (b ResponsePeers) React() []Behavior {
	for _, v := range b.Peers {
		// save as a new peer
		fmt.Println(v)
	}
	return nil
}

func (b ResponseBlocks) React() []Behavior {
	for _, v := range b.blocks {
		blockchain.LoadChain(v.ChainID()).SaveBlock(v)
	}
	return nil
}

func (b BroadcastBlock) React() []Behavior {
	blockchain.LoadChain(b.block.ChainID()).SaveBlock(b.block)
	return nil
}
