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
	block *Block
	desc  string
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
	return e.desc + ":\n" + pretty.Sprint(e.block)
}

func (e ForkError) Error() string {
	result := e.desc + ":\n"
	for _, v := range pretty.Diff(e.block, e.forkBlock) {
		result += fmt.Sprintln(v)
	}
	return result
}

func (e ExistBlockError) Error() string {
	return e.desc + ":\n" + pretty.Sprint(e.block)
}

func (e MismatchedIDError) Error() string {
	return e.desc + ":\n" + pretty.Sprint(e.block)
}

func (e DangledBlockError) Error() string {
	return e.desc + ":\n" + pretty.Sprint(e.block)
}
