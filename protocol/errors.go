package protocol

type Error struct {
	Code string
	Desc string
}

func (e Error) Validate() *Error {
	return nil
}

func (e Error) React() []Behavior {
	return nil
}
