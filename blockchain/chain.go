package blockchain

import (
	"github.com/Infnote/infnotechain/blockchain/crypto"
	"github.com/Infnote/infnotechain/utils"
	"github.com/mr-tron/base58"
	"time"
)

type Chain struct {
	ID    string
	Count uint64
	key   *crypto.Key
	Ref   int64
	cache map[uint64]*Block
}

var loadedChains = map[string]*Chain{}
var BlockSavedHook func(block *Block) = nil

func ResetChainCache() {
	loadedChains = map[string]*Chain{}
}

// Create a chain object with genesis block payload
func CreateChain(payload []byte) *Chain {
	key := crypto.NewKey()
	chain := &Chain{
		ID:    key.ToAddress(),
		key:   key,
		cache: map[uint64]*Block{},
	}
	chain.Sync()
	chain.SaveBlock(chain.CreateBlock(payload))
	return chain
}

func NewOwnedChain(wif string) *Chain {
	key, err := crypto.FromWIF(wif)
	if err != nil {
		utils.L.Fatal(err)
	}
	return &Chain{ID: key.ToAddress(), key: key, cache: map[uint64]*Block{}}
}

func NewReadonlyChain(id string) *Chain {
	return &Chain{ID: id, cache: map[uint64]*Block{}}
}

func LoadChain(id string) *Chain {
	chain, ok := loadedChains[id]
	if ok {
		return chain
	}

	chain = &Chain{ID: id, cache: map[uint64]*Block{}}
	var wif string
	if !SharedStorage().GetChain(id, &chain.Ref, &wif, &chain.Count) {
		return nil
	}
	if len(wif) > 0 {
		chain.key, _ = crypto.FromWIF(wif)
	}

	loadedChains[chain.ID] = chain
	return chain
}

func LoadAllChains() []*Chain {
	s := SharedStorage()
	var chains []*Chain
	s.GetAllChains(func(ref int64, id string, wif string, count uint64) {
		chain := &Chain{ID: id, Count: count, Ref: ref, cache: map[uint64]*Block{}}
		if len(wif) > 0 {
			chain.key, _ = crypto.FromWIF(wif)
		}
		chains = append(chains, chain)
	})
	return chains
}

func (c Chain) IsOwner() bool {
	return c.key != nil
}

func (c Chain) WIF() string {
	if c.IsOwner() {
		return c.key.ToWIF()
	}
	return ""
}

func (c Chain) GetBlock(height uint64) *Block {
	if block := c.cache[height]; block != nil {
		return block
	}
	return SharedStorage().GetBlock(c.Ref, height)
}

func (c Chain) GetBlocks(from uint64, to uint64) []*Block {
	return SharedStorage().GetBlocks(c.Ref, from, to)
}

func (c Chain) CreateBlock(payload []byte) *Block {
	if !c.IsOwner() {
		return nil
	}

	block := &Block{Height: c.Count, Time: uint64(time.Now().Unix()), Payload: payload}
	if c.Count > 0 {
		prev := c.GetBlock(c.Count - 1)
		if prev == nil {
			utils.L.Fatalf("attempt get block at height [%v] on chain [%v] failed", c.Count - 1, c.ID)
		}
		block.PrevHash = prev.Hash
	}

	hash := utils.SHA256(block.DataForHashing())
	block.Hash = base58.Encode(hash)
	block.Signature = base58.Encode(c.key.Sign(block.DataForHashing()))

	return block
}

// TODO: may need to cache database query result
func (c Chain) ValidateBlock(block *Block) BlockValidationError {
	if err := block.Validate(); err != nil {
		return err
	}

	if block.ChainID() != c.ID {
		return MismatchedIDError{c.ID, block.ChainID()}
	}

	b := c.GetBlock(block.Height)

	if b != nil {
		if b.Hash != block.Hash || b.PrevHash != block.PrevHash {
			return ForkError{b, block, "block on height %v already exist"}
		}
		return ExistBlockError{b, "block already exist"}
	}

	if block.Height > 0 {
		prev := c.GetBlock(block.Height-1)
		if prev == nil {
			return DangledBlockError{b, "previous block is not exist"}
		}
		if prev.Hash != block.PrevHash {
			return InvalidBlockError{b, "previous hash is not match"}
		}
	}

	return nil
}

func (c *Chain) SaveBlock(block *Block) bool {
	if c.ValidateBlock(block) == nil {
		SharedStorage().SaveBlock(c.Ref, block)
		SharedStorage().IncreaseCount(c)
		utils.L.Debugf("new block saved: %#v", block.Hash)
		if BlockSavedHook != nil {
			BlockSavedHook(block)
		}
		return true
	}
	return false
}

func (c *Chain) CacheBlock(block *Block) BlockValidationError {
	err := c.ValidateBlock(block)
	if err != nil {
		return err
	}
	utils.L.Debugf("block cached: %v", block)
	c.cache[block.Height] = block
	return nil
}

func (c *Chain) CommitCache() {
	for _, block := range c.cache {
		SharedStorage().SaveBlock(c.Ref, block)
		SharedStorage().IncreaseCount(c)
		delete(c.cache, block.Height)
		utils.L.Debugf("block saved: %v", block)
	}
}

func (c *Chain) Sync() {
	err := SharedStorage().SaveChain(c)
	if err != nil {
		var wif string
		SharedStorage().GetChain(c.ID, &c.Ref, &wif, &c.Count)

		c.key, err = crypto.FromWIF(wif)
		if err != nil {
			utils.L.Fatal(err)
		}
	}
}
