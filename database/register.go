package database

import (
	"database/sql"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/network"
	"github.com/Infnote/infnotechain/utils"
	_ "github.com/mattn/go-sqlite3"
)

func Register() {
	db, err := sql.Open("sqlite3", SQLiteDBFile)
	if err != nil {
		utils.L.Fatal(err)
	}
	s := &SQLiteDriver{db}
	blockchain.RegisterStorage(s)
	network.RegisterStorage(s)
}

func Migrate() {
	sqliteMigrate()
}

func Prune() {
	sqlitePrune()
}
