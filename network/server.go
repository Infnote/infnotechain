package network

import (
	"fmt"
	"github.com/Infnote/infnotechain/utils"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

type Server struct {
	Peers map[string]*Peer
	In    chan *Peer
	Out   chan *Peer
}

const BufferSize = 1024 * 1024 * 2

var upgrader = websocket.Upgrader{
	ReadBufferSize:    BufferSize,
	WriteBufferSize:   BufferSize,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var dialer = &websocket.Dialer{
	ReadBufferSize:    BufferSize,
	WriteBufferSize:   BufferSize,
	EnableCompression: true,
	Proxy:             http.ProxyFromEnvironment,
	HandshakeTimeout:  45 * time.Second,
}

func NewServer() *Server {
	return &Server{
		map[string]*Peer{},
		make(chan *Peer),
		make(chan *Peer),
	}
}

func (s *Server) Connect(peer *Peer) error {
	conn, _, err := dialer.Dial(peer.Addr, nil)
	if err != nil {
		utils.L.Warningf("failed to connect peer: %v", err)
		return err
	}
	peer.server = s
	peer.conn = conn
	peer.Last = time.Now()
	peer.IsServer = true
	peer.Save()

	s.In <- peer

	go peer.read()
	go peer.write()

	return nil
}

func (s *Server) Serve() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		inbound(s, writer, request)
	})

	err := http.ListenAndServe(
		fmt.Sprintf(
			"%v:%v",
			viper.GetString("server.host"),
			viper.GetString("server.port")), nil)
	if err != nil {
		utils.L.Fatal(err)
	}
}
