package services

import (
	"github.com/Infnote/infnotechain/network"
	"github.com/Infnote/infnotechain/protocol"
)

var Peers = map[*network.Peer]bool{}

// TODO: any error occur in this function should not crash entire app
func handleMessages(peer *network.Peer) {
	defer func() {
		recover()
	}()
	for {
		data, ok := <-peer.Recv
		if !ok {
			return
		}
		for _, v := range protocol.HandleJSONData(data) {
			peer.Send <- v
		}
	}
}

func handlePeers(server *network.Server) {
	for {
		select {
		case peer := <-server.In:

			Peers[peer] = true
			peer.Send <- protocol.NewMessage(protocol.NewInfo()).Serialize()
			go handleMessages(peer)
		case peer := <-server.Out:
			delete(Peers, peer)
		}
	}
}

// TODO: recover service after any error occurred
func StartService() {
	server := network.NewServer()
	go handlePeers(server)
	server.Serve()
}
