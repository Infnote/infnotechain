package services

import (
	"bytes"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/network"
	"github.com/Infnote/infnotechain/protocol"
	"github.com/Infnote/infnotechain/utils"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
)

var SharedServer *network.Server

func handleMessages(peer *network.Peer) {
	defer func() {
		// recover when any error occurred in message processing
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
		broadcast := <-protocol.BroadcastChannel
		utils.L.Debugf("broadcast a block")
		for _, peer := range SharedServer.Peers {
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
			server.Peers[peer.Addr] = peer
			peer.Send <- protocol.NewMessage(protocol.NewInfo()).Serialize()
			go handleMessages(peer)
		case peer := <-server.Out:
			utils.L.Infof("outcoming peer: %v", peer.Addr)
			delete(server.Peers, peer.Addr)
		}
	}
}

func addHook() {
	hook := viper.GetString("hooks.block")
	if len(hook) > 0 {
		addr, err := url.Parse(hook)
		if err != nil {
			utils.L.Warning("%v: %v", addr)
			return
		}
		blockchain.BlockSavedHook = func(block *blockchain.Block) {
			_, err := http.Post(
				addr.String(),
				"application/json",
				bytes.NewBuffer(block.Serialize()))
			if err != nil {
				utils.L.Warning("%v", err)
			}
		}
	}

}

func PeerService() {
	if SharedServer != nil {
		return
	}

	SharedServer = network.NewServer()
	go handlePeers(SharedServer)
	go handleBroadcast()

	addHook()

	utils.L.Info("network service start")
	SharedServer.Serve()
}
