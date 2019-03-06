package database

import (
	"database/sql"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/network"
	"github.com/Infnote/infnotechain/utils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

func Register() {
	db, err := sql.Open("sqlite3", viper.GetString("data.file"))
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
