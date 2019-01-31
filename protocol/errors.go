package protocol

import "github.com/Infnote/infnotechain/blockchain"

type Error struct {
	Code string `json:"code"`
	Desc string `json:"desc"`
}

func (e Error) Validate() *Error {
	return nil
}

func (e Error) React() []Behavior {
	return nil
}

func InvalidMessageError(err string) *Error {
	return &Error{"InvalidMessageError", err}
}

func InvalidBehaviorError(err string) *Error {
	return &Error{"InvalidBehaviorError", err}
}

func IncompatibleProtocolVersionError(err string) *Error {
	return &Error{"IncompatibleProtocolVersionError", err}
}

func BadRequestError(err string) *Error {
	return &Error{"BadRequestError", err}
}

func JSONDecodeError(err string) *Error {
	return &Error{"JSONDecodeError", err}
}

func ChainNotAcceptError(err string) *Error {
	return &Error{"ChainNotAcceptError", err}
}

func BlockValidationError(err blockchain.BlockValidationError) *Error {
	return &Error{"BlockValidationError", err.Code()}
}

func InvalidURLError(err string) *Error {
	return &Error{"InvalidURLError", err}
}

func DuplicateBroadcastError(err string) *Error {
	return &Error{"DuplicateBroadcastError", err}
}
