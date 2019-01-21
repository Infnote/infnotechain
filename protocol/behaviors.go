package protocol

import (
	"encoding/json"
	"fmt"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/network"
	"github.com/Infnote/infnotechain/utils"
	"golang.org/x/sys/unix"
	"net/url"
	"time"
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

func fixedBytesToString(data [256]byte) string {
	var result []byte
	for _, b := range data {
		if b > 0 {
			result = append(result, b)
		} else {
			break
		}
	}
	return string(result)
}

// TODO: add support for windows
// now only can get system info at *nix systems
func newSysInfo() map[string]string {
	sysinfo := &unix.Utsname{}
	err := unix.Uname(sysinfo)
	if err != nil {
		utils.L.Fatal(err)
	}

	return map[string]string{
		"system":   fixedBytesToString(sysinfo.Sysname),
		"nodename": fixedBytesToString(sysinfo.Nodename),
		"release":  fixedBytesToString(sysinfo.Release),
		"version":  fixedBytesToString(sysinfo.Version),
		"machine":  fixedBytesToString(sysinfo.Machine),
	}
}

func NewInfo() *Info {
	chains := blockchain.LoadAllChains()
	chainMap := map[string]uint64{}
	for _, chain := range chains {
		chainMap[chain.ID] = chain.Count
	}

	return &Info{
		Version:  "1.1",
		Peers:    network.SharedStorage().CountOfPeers(),
		Chains:   chainMap,
		Platform: newSysInfo(),
		FullNode: true,
	}
}

// - Serializations
func (b ResponseBlocks) Serialize() []byte {
	var blocks []json.RawMessage
	for _, block := range b.blocks {
		blocks = append(blocks, json.RawMessage(block.Serialize()))
	}
	data, err := json.Marshal(blocks)
	if err != nil {
		utils.L.Fatal(err)
	}
	return data
}

func (b BroadcastBlock) Serialize() [] byte {
	data, err := json.Marshal(map[string]json.RawMessage{"block": json.RawMessage(b.block.Serialize())})
	if err != nil {
		utils.L.Fatal(err)
	}
	return data
}

// - Validations
func (b Info) Validate() *Error {
	//version, err := strconv.ParseFloat(b.Version, 32)
	if b.Version != "1.1" {
		return IncompatibleProtocolVersionError("only accept v1.1 protocol")
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
		block, err := blockchain.DeserializeBlock(v)
		if err != nil {
			return JSONDecodeError(err.Error())
		}

		if err := block.Validate(); err != nil {
			utils.L.Debugf("a invalid block: %v", err.Error())
			return BlockValidationError(err)
		}

		chain := blockchain.LoadChain(block.ChainID())
		if chain == nil {
			return ChainNotAcceptError(fmt.Sprintf("recovered chain id: %v", block.ChainID()))
		}

		verr := chain.ValidateBlockCached(block)
		if verr != nil {
			utils.L.Debugf("%v", verr)
			return BlockValidationError(verr)
		}

		b.blocks = append(b.blocks, block)
	}
	return nil
}

func (b *BroadcastBlock) Validate() *Error {
	block, err := blockchain.DeserializeBlock(b.Block)
	if err != nil {
		return JSONDecodeError(err.Error())
	}
	b.block = block

	chain := blockchain.LoadChain(b.block.ChainID())
	if chain == nil {
		return ChainNotAcceptError(fmt.Sprintf("recovered chain id: %v", b.block.ChainID()))
	}

	verr := chain.ValidateBlock(b.block)
	if verr != nil {
		utils.L.Debugf("%v", verr)
		return BlockValidationError(verr)
	}
	return nil
}

// - Reactions
func (b Info) React() []Behavior {
	var behaviors []Behavior
	if b.Peers > 0 {
		behaviors = append(behaviors, &RequstPeers{b.Peers})
	}
	for k, v := range b.Chains {
		chain := blockchain.LoadChain(k)
		if chain == nil {
			continue
		}
		if chain.Count >= v {
			continue
		}
		behaviors = append(behaviors, &RequestBlocks{k, chain.Count, v - 1})
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
		t := time.Unix(0, 0)
		(&network.Peer{Addr: v, Rank: 100, Last: &t}).Save()
	}
	return nil
}

func (b ResponseBlocks) React() []Behavior {
	for _, v := range b.blocks {
		blockchain.LoadChain(v.ChainID()).CommitCache()
	}
	return nil
}

func (b *BroadcastBlock) React() []Behavior {
	blockchain.LoadChain(b.block.ChainID()).SaveBlock(b.block)
	return []Behavior{b}
}

// Deserialize
func DeserializeBehavior(msg *Message) (Behavior, error) {
	instance := MapBehavior(msg.Type)
	err := json.Unmarshal(msg.Data, instance)
	if err != nil {
		return nil, err
	}

	return instance, nil
}
