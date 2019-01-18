package test

import (
	"github.com/Infnote/infnotechain/protocol"
	"log"
	"testing"
)

func TestNewInfo(t *testing.T) {
	log.Printf("%#v", protocol.NewInfo())
}
