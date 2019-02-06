package blockchain

type Storage interface {
	GetChain(chainID string, ref *int64, wif *string, count *uint64) bool
	GetAllChains(func(ref int64, id string, wif string, count uint64))
	GetBlock(id int64, height uint64) *Block
	GetBlockByHash(id int64, hash string) *Block
	GetBlocks(id int64, from uint64, to uint64) []*Block
	SaveChain(chain *Chain) error
	IncreaseCount(chain *Chain)
	SaveBlock(id int64, block *Block)
	CleanChain(chain *Chain)
}

var instance Storage

func RegisterStorage(s Storage) {
	instance = s
}

func SharedStorage() Storage {
	return instance
}
