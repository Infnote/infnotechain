package database

import (
	"database/sql"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/network"
	"github.com/Infnote/infnotechain/utils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mr-tron/base58"
	"os"
)

type SQLiteDriver struct {
	db *sql.DB
}

const SQLiteDBFile = "/usr/local/var/infnote/data.db?cache=shared&mode=rwc"

func (s SQLiteDriver) GetChain(chainID string, ref *int64, wif *string, count *uint64) bool {
	query := `SELECT id, wif, height FROM chains WHERE chain_id = ? LIMIT 1`
	rows, err := s.db.Query(query, chainID)
	if err != nil {
		utils.L.Debugf("sqlite query error: ", err)
		return false
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		return false
	}

	err = rows.Scan(ref, wif, count)
	if err != nil {
		utils.L.Debugf("sqlite scan error: ", err)
		return false
	}

	return true
}

func (s SQLiteDriver) GetAllChains(yield func(ref int64, id string, wif string, count uint64)) {
	query := `SELECT id, chain_id, wif, height FROM chains`
	rows, err := s.db.Query(query)
	if err != nil {
		utils.L.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var id, count int64
		var chainID, wif string
		err = rows.Scan(&id, &chainID, &wif, &count)
		if err != nil {
			utils.L.Fatal(err)
		}
		yield(id, chainID, wif, uint64(count))
	}
}

// TODO: returning an instance of Block may not be an good practice
func (s SQLiteDriver) GetBlock(id int64, height uint64) *blockchain.Block {
	query := `SELECT height, time, hash, prev_hash, signature, payload FROM blocks 
			  WHERE height = ? AND chain_id = ? LIMIT 1`
	rows, err := s.db.Query(query, height, id)
	if err != nil {
		utils.L.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		block := &blockchain.Block{}
		var payload string
		err = rows.Scan(&block.Height, &block.Time, &block.Hash, &block.PrevHash, &block.Signature, &payload)
		if err != nil {
			utils.L.Fatal(err)
		}
		block.Payload, err = base58.Decode(payload)
		if err != nil {
			utils.L.Fatal(err)
		}
		return block
	}
	return nil
}

func (s SQLiteDriver) GetBlockByHash(id int64, hash string) *blockchain.Block {
	query := `SELECT height, time, hash, prev_hash, signature, payload FROM blocks 
			  WHERE hash = ? AND chain_id = ? LIMIT 1`
	rows, err := s.db.Query(query, hash, id)
	if err != nil {
		utils.L.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		block := &blockchain.Block{}
		var payload string
		err = rows.Scan(&block.Height, &block.Time, &block.Hash, &block.PrevHash, &block.Signature, &payload)
		if err != nil {
			utils.L.Fatal(err)
		}
		block.Payload, err = base58.Decode(payload)
		if err != nil {
			utils.L.Fatal(err)
		}
		return block
	}
	return nil
}

// Get blocks of specific internal id by two heights
// 'from' and 'to' are both included
func (s SQLiteDriver) GetBlocks(id int64, from uint64, to uint64) []*blockchain.Block {
	query := `SELECT height, time, hash, prev_hash, signature, payload FROM blocks 
			  WHERE chain_id = ? AND height >= ? AND height <= ?`
	rows, err := s.db.Query(query, id, from, to)
	if err != nil {
		utils.L.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	var blocks []*blockchain.Block

	for rows.Next() {
		block := &blockchain.Block{}
		var payload string
		err = rows.Scan(&block.Height, &block.Time, &block.Hash, &block.PrevHash, &block.Signature, &payload)
		if err != nil {
			utils.L.Fatal(err)
		}
		block.Payload, _ = base58.Decode(payload)
		blocks = append(blocks, block)
	}
	return blocks
}

func (s SQLiteDriver) SaveChain(chain *blockchain.Chain) error {
	query := `
		INSERT INTO chains (chain_id, wif)
		VALUES (?, ?)
	`
	result, err := s.db.Exec(query, chain.ID, chain.WIF())
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	chain.Ref = id
	return err
}

func (s SQLiteDriver) IncreaseCount(chain *blockchain.Chain) {
	query := `UPDATE chains SET height=height+1 WHERE id = ?`
	_, err := s.db.Exec(query, chain.Ref)
	if err != nil {
		utils.L.Fatal(err)
	}
	chain.Count += 1
}

func (s SQLiteDriver) SaveBlock(id int64, block *blockchain.Block) {
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
		utils.L.Fatal(err)
	}
}

func (s SQLiteDriver) CountOfPeers() int {
	query := `SELECT COUNT(addr) FROM peers`

	count := 0
	err := s.db.QueryRow(query).Scan(&count)
	if err != nil {
		utils.L.Fatal(err)
	}

	return count
}

func (s SQLiteDriver) GetPeers(count int) []*network.Peer {
	query := `SELECT addr, rank FROM peers ORDER BY rank LIMIT ?`

	rows, err := s.db.Query(query, count)
	if err != nil {
		utils.L.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	var peers []*network.Peer
	for rows.Next() {
		peer := &network.Peer{}
		err = rows.Scan(&peer.Addr, &peer.Rank)
		if err != nil {
			utils.L.Fatal(err)
		}
		peers = append(peers, peer)
	}

	return peers
}

func (s SQLiteDriver) SavePeer(peer *network.Peer) {
	query := `INSERT INTO peers VALUES (?, ?, ?)`

	_, err := s.db.Exec(query, peer.Addr, peer.Rank, peer.Last)
	if err != nil {
		utils.L.Fatal(err)
	}
}

func sqliteMigrate() {
	err := os.MkdirAll("/usr/local/var/infnote/", 0755)
	if err != nil {
		utils.L.Fatal(err)
	}

	db, err := sql.Open("sqlite3", SQLiteDBFile)
	if err != nil {
		utils.L.Fatal(err)
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
		CREATE TABLE peers (
			addr TEXT PRIMARY KEY,
			rank INTEGER,
			last INTEGER
		);
		CREATE UNIQUE INDEX chains_chain_id ON chains(chain_id);
		CREATE INDEX blocks_height ON blocks(height);
		CREATE INDEX blocks_hash ON blocks(hash);
	`
	_, err = db.Exec(query)
	if err != nil {

	}
}

func sqlitePrune() {
	_ = os.Remove(SQLiteDBFile)
}
