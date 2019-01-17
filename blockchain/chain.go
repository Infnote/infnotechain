package blockchain

import (
	"github.com/Infnote/infnotechain/blockchain/crypto"
	"github.com/Infnote/infnotechain/utils"
	"github.com/mr-tron/base58"
	"log"
	"time"
)

type Chain struct {
	ID     string
	Height uint64
	key    *crypto.Key
	id     int64
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
		log.Fatal(err)
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
	if !ShardStorage().GetChain(id, &chain.id, &wif, &chain.Height) {
		return nil
	}
	if len(wif) > 0 {
		chain.key, _ = crypto.FromWIF(wif)
	}

	loadedChains[chain.ID] = chain
	return chain
}

func LoadAllChains() []*Chain {
	s := ShardStorage()
	var chains []*Chain
	s.GetAllChains(func(ref int64, id string, wif string, height uint64) {
		chain := &Chain{ID: id, Height: height, id: ref}
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

func (c Chain) GetBlock(height uint64) *Block {
	return ShardStorage().GetBlock(c.id, height)
}

func (c Chain) GetBlocks(from uint64, to uint64) []*Block {
	return ShardStorage().GetBlocks(c.id, from, to)
}

func (c Chain) CreateBlock(payload []byte) *Block {
	if !c.IsOwner() {
		return nil
	}

	block := &Block{Height: c.Height, Time: uint64(time.Now().Unix()), Payload: payload}
	if c.Height > 0 {
		block.PrevHash = c.GetBlock(c.Height - 1).Hash
	}

	hash := utils.SHA256(block.DataForHashing())
	block.Hash = base58.Encode(hash)
	block.Signature = base58.Encode(c.key.Sign(block.DataForHashing()))

	return block
}

func (c Chain) ValidateBlock(block *Block) bool {
	if block.IsValid() && block.ChainID() == c.ID {
		b := c.GetBlock(block.Height)
		if b != nil && b.Hash == block.Hash &&
			b.PrevHash == block.PrevHash &&
			b.Signature == block.Signature {
			return true
		} else if b == nil {
			prev := ShardStorage().GetBlockByHash(c.id, block.PrevHash)
			if prev != nil {
				return true
			}
		}
	}
	return false
}

// TODO: check block duplication
func (c *Chain) SaveBlock(block *Block) {
	if c.GetBlock(block.Height) == nil && c.ValidateBlock(block) {
		ShardStorage().SaveBlock(*block, c.id)
		c.Height += 1
		ShardStorage().IncreaseHeight(c)
	}
}

func (c *Chain) Sync() {
	err := ShardStorage().SaveChain(c)
	if err != nil {
		var wif string
		ShardStorage().GetChain(c.ID, &c.id, &wif, &c.Height)

		c.key, err = crypto.FromWIF(wif)
		if err != nil {
			log.Fatal(err)
		}
	}
}
