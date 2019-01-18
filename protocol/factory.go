package protocol


func HandleMessage(msg *Message) []*Message {
	behavior := MapBehavior(msg.Type)
	if behavior == nil {
		err := InvalidMessageError("invalid type of message")
		NewMessage(err)
	}

	return nil
}
