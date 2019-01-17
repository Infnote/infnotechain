package protocol

import (
	"reflect"
)

var MessageDataMap = map[string]reflect.Type{
	"info":            reflect.TypeOf(Info{}),
	"error":           reflect.TypeOf(Error{}),
	"request:blocks":  reflect.TypeOf(RequestBlocks{}),
	"request:peers":   reflect.TypeOf(RequstPeers{}),
	"response:blocks": reflect.TypeOf(ResponseBlocks{}),
	"response:peers":  reflect.TypeOf(ResponsePeers{}),
	"broadcast:block": reflect.TypeOf(BroadcastBlock{}),
}

func HandleMessage(msg *Message) []*Message {
	//cls, exist := MessageDataMap[msg.Type]
	//if !exist {
	//	return []Behavior{NewError()}
	//}
	//
	//instance := reflect.New(cls).Interface()
	//err := json.Unmarshal(msg.Data, instance)
	//if err != nil {
	//	return []Behavior{NewError()}
	//}

	return nil
}
