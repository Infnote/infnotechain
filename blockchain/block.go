package blockchain

import (
	"encoding/json"
	"github.com/Infnote/infnotechain/blockchain/crypto"
	"github.com/Infnote/infnotechain/utils"
	"github.com/mr-tron/base58"
	"strconv"
)

type Block struct {
	Height    uint64	`json:"height"`
	Time      uint64	`json:"time"`
	PrevHash  string	`json:"prev_hash"`
	Hash      string	`json:"hash"`
	Signature string	`json:"signature"`
	Payload   []byte	`json:"payload"`

	data []byte
}

// Create a byte array for hashing
// height + time + prev_hash + payload
// int value should be converted to string literally
// hash string should be converted to string with base58 decoding
// TODO: store result to reduce calculations
func (b Block) DataForHashing() []byte {
	data := []byte(strconv.FormatUint(b.Height, 10) + strconv.FormatUint(b.Time, 10))
	if len(b.PrevHash) > 0 {
		prevHash, err := base58.Decode(b.PrevHash)
		if err != nil {
			utils.L.Fatal(err)
		}
		data = append(data, prevHash...)
	}
	return append([]byte(data), b.Payload...)
}

// TODO: store result to reduce calculations
func (b Block) IsValid() bool {
	if base58.Encode(utils.SHA256(b.DataForHashing())) == b.Hash && len(b.ChainID()) > 0 {
		return true
	}
	return false
}

// TODO: store result to reduce calculations
func (b Block) ChainID() string {
	sig, err := base58.Decode(b.Signature)
	if err != nil {
		utils.L.Debugf("invalid base58 string: %v", err)
		return ""
	}
	addr, err := crypto.RecoverAddress(sig, b.DataForHashing())
	if err != nil {
		utils.L.Debugf("invalid signature: %v", err)
		return ""
	}
	return addr
}

func (b Block) Size() int {
	return len(b.data)
}

func (b Block) Serialize() []byte {
	if b.data != nil {
		return b.data
	}

	type Alias Block
	data := struct {
		Alias
		Payload   string	`json:"payload"`
	}{(Alias)(b), base58.Encode(b.Payload)}

	j, err := json.Marshal(data)
	if err != nil {
		utils.L.Fatal(err)
	}

	b.data = j[:]
	return j
}
