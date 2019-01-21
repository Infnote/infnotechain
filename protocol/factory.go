package protocol

import (
	"github.com/Infnote/infnotechain/utils"
	"github.com/kr/pretty"
)

func serialize(behaviors ...Behavior) [][]byte {
	var result [][]byte
	for _, v := range behaviors {
		utils.L.Debugf("made behavior:\n%v", pretty.Sprint(v))
		result = append(result, NewMessage(v).Serialize())
	}
	return result
}

func HandleJSONData(data []byte) [][]byte {
	msg, err := DeserializeMessage(data)
	if err != nil {
		utils.L.Debugf("%v: %v", err, string(data))
		return serialize(InvalidMessageError("invalid format of message"))
	}

	behavior := MapBehavior(msg.Type)
	if behavior == nil {
		utils.L.Debugf("invalid message type: %v", msg.Type)
		return serialize(InvalidMessageError("invalid type of message"))
	}

	behavior, err = DeserializeBehavior(msg)
	if err != nil {
		utils.L.Debugf("%v: %+v", err, string(msg.Data))
		return serialize(InvalidBehaviorError("invalid format of message data"))
	}

	utils.L.Debugf("message data received:\n%v", pretty.Sprint(behavior))

	rerr := behavior.Validate()
	if rerr != nil {
		return serialize(rerr)
	}

	responses := behavior.React()
	if len(responses) > 0 {
		return serialize(responses...)
	}

	return nil
}