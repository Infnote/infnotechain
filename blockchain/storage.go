package blockchain

type Storage interface {
	GetChain(string, *int64, *string, *uint64) bool
	GetAllChains(func(ref int64, id string, wif string, height uint64))
	GetBlock(int64, uint64) *Block
	GetBlockByHash(int64, string) *Block
	GetBlocks(int64, uint64, uint64) []*Block
	SaveChain(*Chain) error
	IncreaseCount(*Chain)
	SaveBlock(Block, int64)
	Migrate()
	Prune()
}

var instance Storage

func RegisterStorage(s Storage) {
	instance = s
}

func SharedStorage() Storage {
	return instance
}
