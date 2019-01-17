package blockchain

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mr-tron/base58"
	"log"
	"os"
	"sync"
)

const SQLiteDBFile = "/usr/local/var/infnote/data.db"

type Storage interface {
	GetChain(string, *int64, *string, *uint64) bool
	GetAllChains(func(ref int64, id string, wif string, height uint64))
	GetBlock(int64, uint64) *Block
	GetBlockByHash(int64, string) *Block
	GetBlocks(int64, uint64, uint64) []*Block
	SaveChain(*Chain) error
	IncreaseHeight(*Chain)
	SaveBlock(Block, int64)
	Migrate()
	Prune()
}

type SQLiteDriver struct {
	db *sql.DB
}

var instance Storage
var once sync.Once

func ShardStorage() Storage {
	once.Do(func() {
		db, err := sql.Open("sqlite3", SQLiteDBFile)
		if err != nil {
			log.Fatal(err)
		}
		instance = &SQLiteDriver{db}
	})
	return instance
}

func (s SQLiteDriver) GetChain(chainID string, ref *int64, wif *string, height *uint64) bool {
	query := `SELECT id, wif, height FROM chains WHERE chain_id = ? LIMIT 1`
	rows, err := s.db.Query(query, chainID)
	if err != nil {
		log.Println(err)
		return false
	}

	if !rows.Next() {
		return false
	}

	err = rows.Scan(ref, wif, height)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func (s SQLiteDriver) GetAllChains(yield func(ref int64, id string, wif string, height uint64)) {
	query := `SELECT id, chain_id, wif, height FROM chains`
	rows, err := s.db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var id, height int64
		var chainID, wif string
		err = rows.Scan(&id, &chainID, &wif, &height)
		if err != nil {
			log.Fatal(err)
		}
		yield(id, chainID, wif, uint64(height))
	}
}

// TODO: returning an instance of Block may not be an good practice
func (s SQLiteDriver) GetBlock(id int64, height uint64) *Block {
	query := `SELECT height, time, hash, prev_hash, signature, payload FROM blocks 
			  WHERE height = ? AND chain_id = ? LIMIT 1`
	rows, err := s.db.Query(query, height, id)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		block := &Block{}
		var payload string
		err = rows.Scan(&block.Height, &block.Time, &block.Hash, &block.PrevHash, &block.Signature, &payload)
		if err != nil {
			log.Fatal(err)
		}
		block.Payload, err = base58.Decode(payload)
		if err != nil {
			log.Fatal(err)
		}
		return block
	}
	return nil
}

func (s SQLiteDriver) GetBlockByHash(id int64, hash string) *Block {
	query := `SELECT height, time, hash, prev_hash, signature, payload FROM blocks 
			  WHERE hash = ? AND chain_id = ? LIMIT 1`
	rows, err := s.db.Query(query, hash, id)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		block := &Block{}
		var payload string
		err = rows.Scan(&block.Height, &block.Time, &block.Hash, &block.PrevHash, &block.Signature, &payload)
		if err != nil {
			log.Fatal(err)
		}
		block.Payload, err = base58.Decode(payload)
		if err != nil {
			log.Fatal(err)
		}
		return block
	}
	return nil
}

// Get blocks of specific internal id by two heights
// 'from' and 'to' are both included
func (s SQLiteDriver) GetBlocks(id int64, from uint64, to uint64) []*Block {
	query := `SELECT height, time, hash, prev_hash, signature, payload FROM blocks 
			  WHERE chain_id = ? AND height >= ? AND height <= ?`
	rows, err := s.db.Query(query, id, from, to)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	var blocks []*Block

	for rows.Next() {
		block := &Block{}
		var payload string
		err = rows.Scan(&block.Height, &block.Time, &block.Hash, &block.PrevHash, &block.Signature, &payload)
		if err != nil {
			log.Fatal(err)
		}
		block.Payload, _ = base58.Decode(payload)
		blocks = append(blocks, block)
		log.Println(block)
	}
	return blocks
}

func (s SQLiteDriver) SaveChain(chain *Chain) error {
	query := `
		INSERT INTO chains (chain_id, wif)
		VALUES (?, ?)
	`
	result, err := s.db.Exec(query, chain.ID, chain.key.ToWIF())
	if err != nil {
		return err
	}
	chain.id, err = result.LastInsertId()
	return err
}

func (s SQLiteDriver) IncreaseHeight(chain *Chain) {
	query := `UPDATE chains SET height=height+1 WHERE id = ?`
	_, err := s.db.Exec(query, chain.id)
	if err != nil {
		log.Fatal(err)
	}
	chain.Height += 1
}

func (s SQLiteDriver) SaveBlock(block Block, id int64) {
	query := `
		INSERT INTO blocks (height, time, hash, prev_hash, signature, payload, chain_id) 
		VALUES (?, ?, ?, ?, ?, ?, ?) 
	`

	_, err := s.db.Exec(
		query,
		block.Height,
		block.Time,
		block.Hash,
		block.PrevHash,
		block.Signature,
		base58.Encode(block.Payload),
		id)
	if err != nil {
		log.Fatal(err)
	}
}

func (s SQLiteDriver) Migrate() {
	err := os.MkdirAll("/usr/local/var/infnote/", 0755)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite3", SQLiteDBFile)
	if err != nil {
		log.Fatal(err)
	}

	query := `
		CREATE TABLE chains (
			id			INTEGER PRIMARY KEY,
			chain_id 	TEXT NOT NULL,
			wif      	TEXT NOT NULL,
			height		INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE blocks (
			height 		INTEGER NOT NULL,
			time 		INTEGER NOT NULL,
			hash 		TEXT NOT NULL,
			prev_hash 	TEXT NOT NULL,
			signature 	TEXT NOT NULL,
			payload 	TEXT NOT NULL,
			chain_id	INTEGER NOT NULL,
			FOREIGN KEY (chain_id) REFERENCES chains(id)
		);
		CREATE UNIQUE INDEX chains_chain_id ON chains(chain_id);
		CREATE INDEX blocks_height ON blocks(height);
		CREATE INDEX blocks_hash ON blocks(hash);
	`
	_, err = db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func (s SQLiteDriver) Prune() {
	_ = os.Remove(SQLiteDBFile)
}
