package test

import (
	"fmt"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/database"
	"log"
	"testing"
)

var chain *blockchain.Chain

func init() {
	database.Register()
	//chain = blockchain.LoadAllChains()[0]
}

func TestCreateBlock(t *testing.T) {
	block := chain.CreateBlock([]byte("Test Block"))
	if block.ChainID() != chain.ID {
		t.Fail()
	}
}

func TestSyncChain(t *testing.T) {
	fmt.Println(chain)
	chain.Sync()
	fmt.Println(chain)
}

func TestGetBlock(t *testing.T) {
	chain.Sync()
	block := chain.GetBlock(0)
	fmt.Println(chain.ID)
	fmt.Println(string(block.Serialize()))
	fmt.Println(chain.ValidateBlock(block))
}

func TestCreateChain(t *testing.T) {
	chain = blockchain.CreateChain([]byte("Test Chain"))
	log.Println(chain)
	log.Println(chain.GetBlock(0))
}
