package database

import (
	"database/sql"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/network"
	"github.com/Infnote/infnotechain/utils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mr-tron/base58"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type SQLiteDriver struct {
	db *sql.DB
}

func (s SQLiteDriver) GetChain(chainID string, ref *int64, wif *string, count *uint64) bool {
	query := `SELECT id, wif, count FROM chains WHERE chain_id = ? LIMIT 1`
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
	query := `SELECT id, chain_id, wif, count FROM chains`
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

		if payload == "*" {
			block.Payload, err = ioutil.ReadFile(viper.GetString("data.root") + block.Hash)
		} else {
			block.Payload, err = base58.Decode(payload)
		}

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

		if payload == "*" {
			block.Payload, err = ioutil.ReadFile(viper.GetString("data.root") + block.Hash)
		} else {
			block.Payload, err = base58.Decode(payload)
		}

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

		if payload == "*" {
			block.Payload, err = ioutil.ReadFile(viper.GetString("data.root") + block.Hash)
		} else {
			block.Payload, err = base58.Decode(payload)
		}

		if err != nil {
			utils.L.Fatal(err)
		}
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
	query := `UPDATE chains SET count=count+1 WHERE id = ?`
	_, err := s.db.Exec(query, chain.Ref)
	if err != nil {
		utils.L.Fatal(err)
	}
	chain.Count += 1
}

func (s SQLiteDriver) SaveBlock(id int64, block *blockchain.Block) {
	payload := "*"
	if len(block.Payload) > 1024 * 100 {
		if err := ioutil.WriteFile(viper.GetString("data.root") + block.Hash, block.Payload, 0655); err != nil {
			utils.L.Fatal("failed to write payload to file, abort.")
			return
		}

		utils.L.Debug("write big payload (size: %v) to file", len(block.Payload))
	} else {
		payload = base58.Encode(block.Payload)
	}

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
		payload,
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

func (s SQLiteDriver) scanPeers(rows *sql.Rows) []*network.Peer {
	var peers []*network.Peer
	for rows.Next() {
		var addr string
		var rank int
		var last int64
		err := rows.Scan(&addr, &rank, &last)
		if err != nil {
			utils.L.Fatal(err)
		}
		peer := network.NewPeer(addr, rank)
		peer.IsServer = true
		peer.Last = time.Unix(last, 0)
		peers = append(peers, peer)
	}

	return peers
}

func (s SQLiteDriver) GetPeer(addr string) *network.Peer {
	query := `SELECT addr, rank, last FROM peers WHERE addr = ?`
	rows, err := s.db.Query(query, addr)
	if err != nil {
		utils.L.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	peers := s.scanPeers(rows)
	if len(peers) > 0 {
		return peers[0]
	}
	return nil
}

func (s SQLiteDriver) GetPeers(count int) []*network.Peer {
	var query string
	if count == 0 {
		query = `SELECT addr, rank, last FROM peers ORDER BY rank`
	} else {
		query = `SELECT addr, rank, last FROM peers ORDER BY rank LIMIT ?`
	}

	rows, err := s.db.Query(query, count)
	if err != nil {
		utils.L.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	return s.scanPeers(rows)
}

// TODO: need a better error check
func (s SQLiteDriver) SavePeer(peer *network.Peer) {
	query := `INSERT INTO peers VALUES (?, ?, ?)`

	_, err := s.db.Exec(query, peer.Addr, peer.Rank, peer.Last.Unix())
	if err != nil {
		utils.L.Warning("failed to add a peer: %v", err)
	}

	query = `UPDATE peers SET last=? WHERE addr=?`
	_, err = s.db.Exec(query, peer.Last.Unix(), peer.Addr)
	if err != nil {
		utils.L.Warning("failed to update peer: %v", err)
	}
}

func (s SQLiteDriver) DeletePeer(peer *network.Peer) {
	query := `DELETE FROM peers WHERE addr = ?`
	_, err := s.db.Exec(query, peer.Addr)
	if err != nil {
		utils.L.Warning("failed to delete a peer, %v", err)
	}
}

func (s SQLiteDriver) CleanChain(chain *blockchain.Chain) {
	query := `DELETE FROM chains WHERE chain_id = ?`

	_, err := s.db.Exec(query, chain.ID)
	if err != nil {
		utils.L.Warning("failed to delete a chain")
	}

	query = `SELECT hash FROM blocks WHERE payload = '*'`
	rows, err := s.db.Query(query)
	if err != nil {
		utils.L.Fatal(err)
	}

	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			utils.L.Fatal(err)
		}

		if err := os.Remove(viper.GetString("data.root") + hash); err != nil {
			utils.L.Warning("%v", err)
		}
	}

	query = `DELETE FROM blocks WHERE chain_id = ?`

	_, err = s.db.Exec(query, chain.Ref)
	if err != nil {
		utils.L.Warning("failed to delete blocks of chain %v", chain.ID)
	}
}

func sqliteMigrate() {
	file := viper.GetString("data.file")
	_, err := os.Stat(file)

	// do nothing if file exists
	if err == nil {
		return
	}

	// cannot process if file with errors which is not "not exist"
	if !os.IsNotExist(err) {
		utils.L.Error("%v", err)
		return
	}

	err = os.MkdirAll(filepath.Dir(file), 0755)
	if err != nil {
		utils.L.Fatal(err)
	}

	db, err := sql.Open("sqlite3", file)
	if err != nil {
		utils.L.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	query := `
		CREATE TABLE chains (
			id			INTEGER PRIMARY KEY,
			chain_id 	TEXT NOT NULL,
			wif      	TEXT NOT NULL,
			count		INTEGER NOT NULL DEFAULT 0
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
			addr 	TEXT PRIMARY KEY,
			rank 	INTEGER,
			last 	INTEGER
		);
		CREATE UNIQUE INDEX chains_chain_id ON chains(chain_id);
		CREATE INDEX blocks_height ON blocks(height);
		CREATE INDEX blocks_hash ON blocks(hash);
	`
	_, err = db.Exec(query)
	if err != nil {
		utils.L.Warning("%v", err)
	} else {
		utils.L.Info("database created")
	}
}

func sqlitePrune() {
	_ = os.Remove(viper.GetString("data.file"))
}
