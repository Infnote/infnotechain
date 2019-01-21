package test

import (
	"fmt"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/database"
	"github.com/kr/pretty"
	"github.com/mr-tron/base58"
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

func TestBlock(t *testing.T) {
	payload, err := base58.Decode("5k1XmJn4556WCM")
	if err != nil {
		log.Fatal(err)
	}
	block := &blockchain.Block{
		Height:    1,
		Time:      0,
		PrevHash:  "DiuvcftK8K51umFQpFY71ipefjxMQ1dRyYsDyNrUozbP",
		Hash:      "dRN1qDV3uvu55Rk6DWURfNXgrpG6iQXQuS6Xb1VTVBR",
		Signature: "3pz7sD9QeHLLbb7E6Dzr6CA5EZwarXTm8epSiXLzgddgpJR4n3ALGXsPHccZjZMy2MFmUa7Tki1GN9RtRKnu1KTFC",
		Payload:   payload,
	}

	fmt.Println(pretty.Sprint(block))
	fmt.Println(string(block.DataForHashing()))
	fmt.Println(block.ChainID())
	fmt.Println(block.Validate())
}
