package main

import (
	"github.com/Infnote/infnotechain/cmd"
	"github.com/Infnote/infnotechain/database"
)

func main() {
	database.Register()
	cmd.Execute()
}
