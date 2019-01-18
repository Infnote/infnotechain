package protocol

func serialize(behaviors ...Behavior) [][]byte {
	var result [][]byte
	for _, v := range behaviors {
		result = append(result, NewMessage(v).Serialize())
	}
	return result
}

func HandleJSONData(data []byte) [][]byte {
	msg, err := DeserializeMessage(data)
	if err != nil {
		return serialize(InvalidMessageError("invalid format of message"))
	}

	behavior := MapBehavior(msg.Type)
	if behavior == nil {
		return serialize(InvalidMessageError("invalid type of message"))
	}

	behavior, err = DeserializeBehavior(msg)
	if err != nil {
		return serialize(InvalidBehaviorError("invalid format of message data"))
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
