package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/Infnote/infnotechain/blockchain/crypto"
	"github.com/Infnote/infnotechain/utils"
	"github.com/mr-tron/base58"
	"strconv"
)

type Block struct {
	Height    uint64 `json:"height"`
	Time      uint64 `json:"time"`
	PrevHash  string `json:"prev_hash"`
	Hash      string `json:"hash"`
	Signature string `json:"signature"`
	Payload   []byte `json:"payload"`

	recoveredChainID string
	data             []byte
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
func (b *Block) Validate() BlockValidationError {
	if base58.Encode(utils.SHA256(b.DataForHashing())) != b.Hash {
		return &InvalidBlockError{b, "hash value not match"}
	} else if len(b.ChainID()) == 0 {
		return &InvalidBlockError{b, "cannot recover chain id"}
	}
	return nil
}

// TODO: store result to reduce calculations
func (b Block) ChainID() string {
	if len(b.recoveredChainID) > 0 {
		return b.recoveredChainID
	}

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

	b.recoveredChainID = addr
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
		Payload string `json:"payload"`
	}{(Alias)(b), base58.Encode(b.Payload)}

	j, err := json.Marshal(data)
	if err != nil {
		utils.L.Fatal(err)
	}

	b.data = j[:]
	return j
}

func (b Block) String() string {
	var result string
	result += fmt.Sprintf("[Height   ] %v\n", b.Height)
	result += fmt.Sprintf("[Time     ] %v\n", b.Time)
	result += fmt.Sprintf("[PrevHash ] %v\n", b.PrevHash)
	result += fmt.Sprintf("[Hash     ] %v\n", b.Hash)
	result += fmt.Sprintf("[Signature] %v\n", b.Signature)
	if len(b.Payload) > 100 {
		result += fmt.Sprintf("[Payload  ] (%v bytes) %v ...", len(b.Payload), string(b.Payload[:100]))
	} else {
		result += fmt.Sprintf("[Payload  ] %v", string(b.Payload))
	}
	return result
}

func DeserializeBlock(message json.RawMessage) (*Block, error) {
	data := &struct {
		Block
		Payload string `json:"payload"`
	}{}
	err := json.Unmarshal(message, data)
	if err != nil {
		return nil, err
	}

	payload, err := base58.Decode(data.Payload)
	if err != nil {
		return nil, err
	}
	return &Block{
		Height:    data.Height,
		Time:      data.Time,
		PrevHash:  data.PrevHash,
		Hash:      data.Hash,
		Signature: data.Signature,
		Payload:   payload,
	}, nil
}
