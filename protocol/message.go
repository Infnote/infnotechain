package protocol

import (
	"encoding/json"
	"github.com/mr-tron/base58"
	"log"
	"math/rand"
	"time"
)

type Serializable interface {
	Serialize() []byte
}

type Message struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewMessage(t string, data interface{}) *Message {
	var d []byte
	var e error
	switch s := data.(type) {
	case Serializable:
		d = s.Serialize()
		if !json.Valid(d) {
			log.Fatal("not a valid json raw message")
		}
	default:
		d, e = json.Marshal(s)
		if e != nil {
			log.Fatal(e)
		}
	}

	id := make([]byte, 8)
	for i := range id {
		id[i] = byte(rand.Intn(90) + 32)
	}

	// Convert nested struct need to precompute the nest value first
	return &Message{base58.Encode(id), t, json.RawMessage(d)}
}

func DeserializeMessage(jsonData []byte) (*Message, error) {
	msg := &Message{}
	e := json.Unmarshal(jsonData, msg)
	if e != nil {
		return nil, e
	}
	return msg, nil
}

func (m Message) Serialize() []byte {
	data, err := json.Marshal(m)

	if err != nil {
		log.Fatal(err)
	}

	return data
}
