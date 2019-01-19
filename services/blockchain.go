package services

import (
	"github.com/Infnote/infnotechain/network"
	"github.com/Infnote/infnotechain/protocol"
	"github.com/Infnote/infnotechain/utils"
)

var ConnectedPeers = map[*network.Peer]bool{}

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
			utils.L.Infof("incoming peer: %v", peer.Addr)
			ConnectedPeers[peer] = true
			peer.Send <- protocol.NewMessage(protocol.NewInfo()).Serialize()
			go handleMessages(peer)
		case peer := <-server.Out:
			utils.L.Infof("outcoming peer: %v", peer.Addr)
			delete(ConnectedPeers, peer)
		}
	}
}

// TODO: recover service after any error occurred
func PeerService() {
	server := network.NewServer()
	go handlePeers(server)

	utils.L.Info("network service start")
	server.Serve()
}
