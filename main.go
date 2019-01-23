package main

import (
	"github.com/Infnote/infnotechain/database"
	"github.com/Infnote/infnotechain/services/command"
)

func main() {
	database.Register()
	command.DirectExecute()
}
