package protocol

import (
	"encoding/json"
	"github.com/Infnote/infnotechain/utils"
	"github.com/mr-tron/base58"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
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

var MessageTypeMap = map[string]reflect.Type{
	"info":            reflect.TypeOf(Info{}),
	"error":           reflect.TypeOf(Error{}),
	"request:blocks":  reflect.TypeOf(RequestBlocks{}),
	"request:peers":   reflect.TypeOf(RequstPeers{}),
	"response:blocks": reflect.TypeOf(ResponseBlocks{}),
	"response:peers":  reflect.TypeOf(ResponsePeers{}),
	"broadcast:block": reflect.TypeOf(BroadcastBlock{}),
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func MapBehavior(t string) Behavior {
	cls, exist := MessageTypeMap[t]
	if !exist {
		return nil
	}
	return reflect.New(cls).Interface().(Behavior)
}

func MapType(i Behavior) string {
	names := strings.Split(reflect.TypeOf(i).String(), ".")
	name := names[len(names)-1]

	name = matchFirstCap.ReplaceAllString(name, "${1}:${2}")
	name = matchAllCap.ReplaceAllString(name, "${1}:${2}")
	return strings.ToLower(name)
}

func NewMessage(data Behavior) *Message {
	var d []byte
	var err error
	switch s := data.(type) {
	case Serializable:
		d = s.Serialize()
		if !json.Valid(d) {
			utils.L.Fatal("not a valid json raw message")
		}
	default:
		d, err = json.Marshal(s)
		if err != nil {
			utils.L.Fatal(err)
		}
	}

	id := make([]byte, 8)
	for i := range id {
		id[i] = byte(rand.Intn(90) + 32)
	}

	// Convert nested struct need to precompute the nest value first
	return &Message{base58.Encode(id), MapType(data), json.RawMessage(d)}
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
		utils.L.Fatal(err)
	}

	return data
}
