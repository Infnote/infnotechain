package protocol

import (
	"encoding/json"
	"fmt"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/network"
	"github.com/Infnote/infnotechain/utils"
	"github.com/spf13/viper"
	"golang.org/x/sys/unix"
	"net/url"
	"time"
)

type Behavior interface {
	// returns an Error if invalid, otherwise returns nil
	Validate() *Error

	// Call Validate() before this method
	React() []Behavior

	// Description for debuging
	String() string
}

// - Declarations
type Info struct {
	Version  string            `json:"version"`
	Peers    int               `json:"peers"`
	Chains   map[string]uint64 `json:"chains"`
	Platform map[string]string `json:"platform"`
	FullNode bool              `json:"full_node"`
}

type RequestPeers struct {
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


// - Stringer
func (b Info) String() string {
	var result string
	result += fmt.Sprintf("[Version  ] %v\n", b.Version)
	result += fmt.Sprintf("[Peers    ] %v\n", b.Peers)
	result += fmt.Sprintf("[Chains   ]\n")
	for id, count := range b.Chains {
		result += fmt.Sprintf("\t[%v] %v\n", id, count)
	}
	result += fmt.Sprintf("[Platform ] %v\n", b.Platform)
	result += fmt.Sprintf("[Full Node] %v\n", b.FullNode)
	return result
}

func (b RequestPeers) String() string {
	return fmt.Sprintf("[Count] %v", b.Count)
}

func (b RequestBlocks) String() string {
	var result string
	result += fmt.Sprintf("[Chain ID] %v\n", b.ChainID)
	result += fmt.Sprintf("[From    ] %v\n", b.From)
	result += fmt.Sprintf("[To      ] %v\n", b.To)
	return result
}

func (b ResponsePeers) String() string {
	var result string
	result += fmt.Sprintf("[Peers]\n")
	for _, peer := range b.Peers {
		result += fmt.Sprintf("%v\n", peer)
	}
	return result
}

func (b ResponseBlocks) String() string {
	var result string
	result += fmt.Sprintf("[Blocks]\n")
	for _, block := range b.blocks {
		result += utils.Intend(block.String(), 1)
	}
	return result
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
		"system":   string(sysinfo.Sysname[:]),
		"nodename": string(sysinfo.Nodename[:]),
		"release":  string(sysinfo.Release[:]),
		"version":  string(sysinfo.Version[:]),
		"machine":  string(sysinfo.Machine[:]),
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
	data, err := json.Marshal(map[string][]json.RawMessage{"blocks": blocks})
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

func (b RequestPeers) Validate() *Error {
	if b.Count < 0 {
		return BadRequestError("'count' needs to be a non-negative number")
	}
	return nil
}

func (b ResponsePeers) Validate() *Error {
	var processed []string
	for _, v := range b.Peers {
		addr, err := url.Parse(v)
		if err != nil {
			return InvalidURLError(err.Error())
		}
		if addr.Scheme != "ws" && addr.Scheme != "wss" {
			return InvalidURLError("not a websocket URL")
		}
		processed = append(processed, addr.String())
	}
	b.Peers = processed
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
			return ChainNotAcceptError(fmt.Sprintf("recovered chain ID: %v", block.ChainID()))
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

// - Reactions
func (b Info) React() []Behavior {
	var behaviors []Behavior
	if b.Peers > 0 && viper.GetBool("peer.sync") {
		behaviors = append(behaviors, &RequestPeers{b.Peers})
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

func (b RequestPeers) React() []Behavior {
	var peers []string
	for _, p := range network.SharedStorage().GetPeers(b.Count) {
		peers = append(peers, p.Addr)
	}
	return []Behavior{ResponsePeers{peers}}
}

func (b ResponsePeers) React() []Behavior {
	for _, v := range b.Peers {
		t := time.Unix(0, 0)
		(&network.Peer{Addr: v, Rank: 100, Last: t}).Save()
	}
	return nil
}

func (b ResponseBlocks) React() []Behavior {
	for _, v := range b.blocks {
		blockchain.LoadChain(v.ChainID()).CommitCache()
	}
	return nil
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
