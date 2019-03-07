package network

import (
	"github.com/Infnote/infnotechain/utils"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type Peer struct {
	Addr     string
	Rank     int
	Last     time.Time
	Recv     chan []byte
	Send     chan []byte
	IsServer bool

	server *Server
	conn   *websocket.Conn
}

type Storage interface {
	CountOfPeers() int
	GetPeer(addr string) *Peer
	GetPeers(count int) []*Peer
	SavePeer(peer *Peer)
	DeletePeer(peer *Peer)
}

// 2 MB
const MaxMessageSize = 1024 * 1024 * 2
const WriteWait = 30 * time.Second

var instance Storage

func RegisterStorage(s Storage) {
	instance = s
}

func SharedStorage() Storage {
	return instance
}

func NewPeer(addr string, rank int) *Peer {
	return &Peer{
		Addr: addr,
		Rank: rank,
		Last: time.Now(),
		Recv: make(chan []byte),
		Send: make(chan []byte),
	}
}

func (c *Peer) Save() {
	// TODO: validate address
	instance.SavePeer(c)
}

func (c *Peer) read() {
	defer func() {
		c.server.Out <- c
		_ = c.conn.Close()
		close(c.Recv)
		//utils.L.Debugf("peer %v reading closed", c.Addr)
	}()

	//c.conn.SetReadLimit(MaxMessageSize)
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				utils.L.Debugf("connection closed unexpectedly: %v", err)
			}
			safeClose(c.Send)
			return
		}
		utils.L.Debugf("message received: %v bytes", len(data))
		c.Recv <- data
	}
}

func (c *Peer) write() {
	defer func() {
		_ = c.conn.Close()
		//utils.L.Debugf("peer %v writing closed", c.Addr)
	}()
	for {
		select {
		case msg, ok := <-c.Send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(WriteWait))

			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			utils.L.Debugf("writing message: %v bytes", len(msg))
			_, _ = w.Write(msg)
			_ = w.Close()
		}
	}
}

func safeClose(c chan []byte) {
	defer func() {
		recover()
	}()
	close(c)
}

func inbound(server *Server, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.L.Warning("%v", err)
		return
	}

	peer := NewPeer(conn.RemoteAddr().String(), 100)
	peer.server = server
	peer.conn = conn
	peer.Last = time.Now()
	peer.IsServer = false

	peer.server.In <- peer

	go peer.read()
	go peer.write()
}
