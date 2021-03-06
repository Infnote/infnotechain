package blockchain

import (
	"fmt"
	"github.com/kr/pretty"
	"reflect"
	"strings"
)

type BlockValidationError interface {
	Code() string
	Error() string
}

type InvalidBlockError struct {
	block *Block
	desc  string
}

type ForkError struct {
	block     *Block
	forkBlock *Block
	desc      string
}

type ExistBlockError struct {
	block *Block
	desc  string
}

type MismatchedIDError struct {
	chainID string
	recoveredChainID string
}

type DangledBlockError struct {
	block *Block
	desc  string
}

func typeName(t interface{}) string {
	names := strings.Split(reflect.TypeOf(t).String(), ".")
	return names[len(names)-1]
}

func (e InvalidBlockError) Code() string {
	return typeName(e)
}

func (e ForkError) Code() string {
	return typeName(e)
}

func (e ExistBlockError) Code() string {
	return typeName(e)
}

func (e MismatchedIDError) Code() string {
	return typeName(e)
}

func (e DangledBlockError) Code() string {
	return typeName(e)
}

func (e InvalidBlockError) Error() string {
	return e.desc + ":\n" + e.block.String()
}

func (e ForkError) Error() string {
	result := e.desc + ":\n"
	for _, v := range pretty.Diff(e.block, e.forkBlock) {
		result += fmt.Sprintln(v)
	}
	return result
}

func (e ExistBlockError) Error() string {
	return e.desc + ":\n" + e.block.String()
}

func (e MismatchedIDError) Error() string {
	return fmt.Sprintf(
		"the Ref recovered from block (%v) mismatch chain Ref (%v)",
		e.recoveredChainID,
		e.chainID)
}

func (e DangledBlockError) Error() string {
	return e.desc + ":\n" + e.block.String()
}
