package network

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type Peer struct {
	Addr   string
	Rank   int
	Recv   chan []byte
	Send   chan []byte
	server *Server
	conn   *websocket.Conn
}

func newPeer(addr string, rank int) *Peer {
	return &Peer{Addr: addr, Rank: rank, Recv: make(chan []byte), Send: make(chan []byte)}
}

// 2 MB
const MaxMessageSize = 1024 * 1024 * 2
const WriteWait = 30 * time.Second

func (c *Peer) Read() {
	defer func() {
		c.server.Out <- c
		_ = c.conn.Close()
		close(c.Recv)
		log.Printf("Peer %v reading closed", c.Addr)
	}()

	c.conn.SetReadLimit(MaxMessageSize)
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Connection closed unexpectedly: %v", err)
			}
			close(c.Send)
			return
		}
		c.Recv <- data
	}
}

func (c *Peer) Write() {
	defer func() {
		_ = c.conn.Close()
		log.Printf("Peer %v writing closed", c.Addr)
	}()
	for {
		select {
		case msg, ok := <-c.Send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(msg)
			_ = w.Close()
		}
	}
}

func inbound(server *Server, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	peer := newPeer(conn.RemoteAddr().String(), 100)
	peer.server = server
	peer.conn = conn
	peer.server.In <- peer

	go peer.Read()
	go peer.Write()
}
