package main

import (
	"github.com/Infnote/infnotechain/database"
	"github.com/Infnote/infnotechain/services/command"
)

func main() {
	database.Migrate()
	database.Register()
	command.DirectExecute()
}
