package test

import (
	"fmt"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/database"
	"log"
	"testing"
)

var storage = blockchain.SharedStorage()
var sqlchain = blockchain.NewOwnedChain("KxUxDz8wbQbnxmnKiPUX9uquHB5tkPc8tF5U3uxmmb3yqnYf7MZb")

func init() {
	database.Register()
}

func TestMigrate(t *testing.T) {
	blockchain.SharedStorage().Migrate()
}

func TestPrune(t *testing.T) {
	blockchain.SharedStorage().Prune()
}

func TestSaveChain(t *testing.T)  {
	_ = storage.SaveChain(sqlchain)
}

func TestSaveBlock(t *testing.T) {

}

func TestGetAllChains(t *testing.T) {
	storage.GetAllChains(func(ref int64, id string, wif string, height uint64) {
		fmt.Println(ref, id, wif, height)
	})
}

func TestSQLGetBlock(t *testing.T) {
	log.Println(storage.GetBlock(1, 0))
}

func TestGetBlocks(t *testing.T) {
	log.Println(storage.GetBlocks(1, 0, 0))
}
