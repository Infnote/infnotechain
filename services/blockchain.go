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
		if info := recover(); info != nil {
			utils.L.Errorf("%#v", info)
			handleMessages(peer)
		}
	}()
	for {
		data, ok := <-peer.Recv
		if !ok {
			return
		}
		for _, v := range protocol.HandleJSONData(peer, data) {
			peer.Send <- v
		}
	}
}

func handleBroadcast() {
	for {
		broadcast := <- protocol.BroadcastChannel
		utils.L.Debugf("broadcast a block")
		for peer := range ConnectedPeers {
			if peer != broadcast.Sender {
				peer.Send <- broadcast.Message().Serialize()
			}
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
	go handleBroadcast()

	utils.L.Info("network service start")
	server.Serve()
}
