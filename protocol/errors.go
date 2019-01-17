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

func IncompatibleProtocolVersion(err string) *Error {
	return &Error{"IncompatibleProtocolVersion", err}
}

func BadRequestError(err string) *Error  {
	return &Error{"BadRequestError", err}
}

func JSONDecodeError(err string) *Error {
	return &Error{"JSONDecodeError", err}
}

func ChainNotAcceptError(err string) *Error {
	return &Error{"ChainNotAcceptError", err}
}

func BlockAlreadyExistError(err string) *Error {
	return &Error{"BlockAlreadyExistError", err}
}

func InvalidBlockError(err string) *Error {
	return &Error{"InvalidBlockError", err}
}

func URLError(err string) *Error {
	return &Error{"URLError", err}
}
