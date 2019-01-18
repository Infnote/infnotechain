package database

import (
	"database/sql"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/network"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func Register() {
	db, err := sql.Open("sqlite3", SQLiteDBFile)
	if err != nil {
		log.Fatal(err)
	}
	s := &SQLiteDriver{db}
	blockchain.RegisterStorage(s)
	network.RegisterStorage(s)
}
