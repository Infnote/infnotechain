package test

import (
	"github.com/Infnote/infnotechain/utils"
	"testing"
)

func TestPrintString(t *testing.T) {
	utils.L.Debugf("%v", "a string")
}
