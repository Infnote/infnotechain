package network

import (
	"github.com/gorilla/websocket"
	"log"
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
}

func NewServer() *Server {
	return &Server{
		make(map	[string]*Peer),
		make(chan *Peer),
		make(chan *Peer),
	}
}

func (s *Server) Connect(addr string) {
	conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		log.Println(err)
	}
	peer := newPeer(addr, 100)
	peer.server = s
	peer.conn = conn

	s.In <- peer

	go peer.Read()
	go peer.Write()
}

func (s *Server) Serve() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		inbound(s, writer, request)
	})

	err := http.ListenAndServe("localhost:32767", nil)
	if err != nil {
		log.Fatal(err)
	}
}
