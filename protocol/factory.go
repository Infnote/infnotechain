package protocol

import (
	"github.com/Infnote/infnotechain/utils"
)

func serialize(behaviors ...Behavior) [][]byte {
	var result [][]byte
	for _, v := range behaviors {
		utils.L.Debugf("made behavior:\n%v", v)
		result = append(result, NewMessage(v).Serialize())
	}
	return result
}

func HandleJSONData(sender interface{}, data []byte) [][]byte {
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

	b, ok := behavior.(*BroadcastBlock)
	if ok {
		b.ID = msg.ID
		b.Sender = sender
	}

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
