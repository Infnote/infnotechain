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
	id    int64
}

var loadedChains = map[string]*Chain{}

// Create a chain object with genesis block payload
func CreateChain(payload []byte) *Chain {
	key := crypto.NewKey()
	chain := &Chain{
		ID:  key.ToAddress(),
		key: key,
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
	return &Chain{ID: key.ToAddress(), key: key}
}

func NewReadonlyChain(id string) *Chain {
	return &Chain{ID: id}
}

func LoadChain(id string) *Chain {
	chain, ok := loadedChains[id]
	if ok {
		return chain
	}

	chain = &Chain{}
	var wif string
	if !SharedStorage().GetChain(id, &chain.id, &wif, &chain.Count) {
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
	s.GetAllChains(func(ref int64, id string, wif string, height uint64) {
		chain := &Chain{ID: id, Count: height, id: ref}
		if len(wif) > 0 {
			chain.key, _ = crypto.FromWIF(wif)
		}
		chains = append(chains, chain)
	})
	return chains
}

func (c Chain) InternalID() int64 {
	return c.id
}

func (c Chain) SetInternalID(id int64) {
	c.id = id
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
	return SharedStorage().GetBlock(c.id, height)
}

func (c Chain) GetBlocks(from uint64, to uint64) []*Block {
	return SharedStorage().GetBlocks(c.id, from, to)
}

func (c Chain) CreateBlock(payload []byte) *Block {
	if !c.IsOwner() {
		return nil
	}

	block := &Block{Height: c.Count, Time: uint64(time.Now().Unix()), Payload: payload}
	if c.Count > 0 {
		block.PrevHash = c.GetBlock(c.Count - 1).Hash
	}

	hash := utils.SHA256(block.DataForHashing())
	block.Hash = base58.Encode(hash)
	block.Signature = base58.Encode(c.key.Sign(block.DataForHashing()))

	return block
}

// TODO: may need to cache database query result
func (c Chain) ValidateBlock(block *Block) BlockValidationError {
	if !block.IsValid() {
		return InvalidBlockError{block, "block hash or signature is invalid"}
	}

	if block.ChainID() != c.ID {
		return MismatchedIDError{block, "the block id mismatch chain id"}
	}

	b := c.GetBlock(block.Height)

	if b != nil {
		if b.Hash != block.Hash || b.PrevHash != block.PrevHash {
			return ForkError{b, block, "block on height %v already exist"}
		}
		return ExistBlockError{b, "block already exist"}
	}

	prev := SharedStorage().GetBlockByHash(c.id, block.PrevHash)
	if prev == nil {
		return DangledBlockError{b, "previous block is not exist"}
	}

	return nil
}

func (c *Chain) SaveBlock(block *Block) {
	if c.ValidateBlock(block) == nil {
		SharedStorage().SaveBlock(*block, c.id)
		c.Count += 1
		SharedStorage().IncreaseCount(c)
		utils.L.Debugf("new block saved: %v", utils.Dump(block))
	}
}

func (c *Chain) Sync() {
	err := SharedStorage().SaveChain(c)
	if err != nil {
		var wif string
		SharedStorage().GetChain(c.ID, &c.id, &wif, &c.Count)

		c.key, err = crypto.FromWIF(wif)
		if err != nil {
			utils.L.Fatal(err)
		}
	}
}
