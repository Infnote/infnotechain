package network

import (
	"fmt"
	"github.com/Infnote/infnotechain/utils"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"net/http"
)

type Server struct {
	Peers map[string]*Peer
	In    chan *Peer
	Out   chan *Peer
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewServer() *Server {
	return &Server{
		make(map[string]*Peer),
		make(chan *Peer),
		make(chan *Peer),
	}
}

func (s *Server) Connect(peer *Peer) {
	conn, _, err := websocket.DefaultDialer.Dial(peer.Addr, nil)
	if err != nil {
		utils.L.Warningf("failed to connect peer: %v", err)
		return
	}
	peer.server = s
	peer.conn = conn

	s.In <- peer

	go peer.read()
	go peer.write()
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
