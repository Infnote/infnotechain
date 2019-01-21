package test

import (
	"fmt"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/database"
	"github.com/mr-tron/base58"
	"log"
	"testing"
)

var storage = blockchain.SharedStorage()
var sqlchain = blockchain.NewOwnedChain("KxUxDz8wbQbnxmnKiPUX9uquHB5tkPc8tF5U3uxmmb3yqnYf7MZb")

func init() {
	database.Register()
}

func TestMigrate(t *testing.T) {
	database.Migrate()
}

func TestPrune(t *testing.T) {
	database.Prune()
}

func TestSaveChain(t *testing.T)  {
	_ = storage.SaveChain(sqlchain)
}

func TestSaveBlock(t *testing.T) {
	payload, _ := base58.Decode("5k1XmJn4556WCM")
	blockchain.SharedStorage().SaveBlock(0, &blockchain.Block{
		Height: 0,
		Time: 0,
		Hash: "DiuvcftK8K51umFQpFY71ipefjxMQ1dRyYsDyNrUozbP",
		Signature: "3qhEnjcJJh3kRKMuB7tKWdzoCbGWvQ2i3969ipy91nGp8hHpro7PfHxhj385BaasQFSophLvDYoSpKMqMd3H7Kh9r",
		Payload: payload,
	})
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
