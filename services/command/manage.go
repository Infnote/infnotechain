package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/network"
	"github.com/Infnote/infnotechain/protocol"
	"github.com/Infnote/infnotechain/services"
	"github.com/Infnote/infnotechain/services/codegen"
	"github.com/Infnote/infnotechain/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"
)

type ManageServer struct{}

type cachedChain struct {
	ID    string
	Ref   int64
	Count uint64
}

var IFCManageClient manage.IFCManageClient
var cachedChains = map[int64]*cachedChain{}
var chainContext *cachedChain
var cachedPeers = map[string]bool{}

func RunDaemon() {
	path := os.Args[0]
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = "/usr/local/bin/" + path
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		utils.L.Fatal("please put the executable file into /usr/local/bin/")
	}

	pid, err := syscall.ForkExec(path, []string{path, "run", "-fF"}, nil)
	if err != nil {
		utils.L.Fatal(err)
	}

	err = ioutil.WriteFile(viper.GetString("daemon.pid"), []byte(fmt.Sprintf("%d", pid)), 0655)
	if err != nil {
		utils.L.Fatal(err)
	}
	fmt.Printf("Infnote Chain service start in child process %v\n", pid)
}

func StopDaemon() {
	b, err := ioutil.ReadFile(viper.GetString("daemon.pid"))
	if err != nil {
		utils.L.Fatal(err)
	}

	pid, err := strconv.Atoi(string(b))
	if err != nil {
		utils.L.Fatal(err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		utils.L.Fatal(err)
	}

	err = process.Kill()
	if err != nil {
		utils.L.Fatal(err)
	}

	fmt.Printf("[PID: %v] service stopped", pid)
}

func RunManageServer() {
	conn, err := net.Listen(
		"tcp",
		fmt.Sprintf(
			"%v:%v",
			viper.GetString("manage.host"),
			viper.GetString("manage.port")))

	if err != nil {
		utils.L.Fatal(err)
	}

	server := grpc.NewServer()
	manage.RegisterIFCManageServer(server, &ManageServer{})

	utils.L.Info("manage service start")
	err = server.Serve(conn)
	if err != nil {
		utils.L.Fatal(err)
	}
}

func (*ManageServer) GetChains(request *manage.ChainRequest, stream manage.IFCManage_GetChainsServer) error {
	send := func(chain *blockchain.Chain) error {
		return stream.Send(&manage.ChainResponse{
			Id:    chain.ID,
			Ref:   chain.Ref,
			Count: chain.Count,
		})
	}
	if len(request.Id) <= 0 {
		for _, chain := range blockchain.LoadAllChains() {
			if err := send(chain); err != nil {
				return err
			}
		}
	} else if chain := blockchain.LoadChain(request.Id); chain != nil {
		if err := send(chain); err != nil {
			return err
		}
	}
	return nil
}

func (*ManageServer) GetBlocks(request *manage.BlockRequest, stream manage.IFCManage_GetBlocksServer) error {
	chain := blockchain.LoadChain(request.ChainID)
	if chain == nil {
		return nil
	}
	if request.From > request.To {
		return nil
	}
	for _, block := range chain.GetBlocks(request.From, request.To) {
		err := stream.Send(&manage.BlockResponse{
			Height:    block.Height,
			Time:      block.Time,
			PrevHash:  block.PrevHash,
			Hash:      block.Hash,
			Signature: block.Signature,
			Payload:   block.Payload,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (*ManageServer) CreateChain(ctx context.Context, request *manage.ChainCreationRequest) (*manage.ChainCreationResponse, error) {
	payload, _ := json.Marshal(map[string]string{
		"name":    request.Name,
		"author":  request.Author,
		"website": request.Website,
		"email":   request.Email,
		"desc":    request.Desc,
	})
	chain := blockchain.CreateChain(payload)
	return &manage.ChainCreationResponse{
		Ref: chain.Ref,
		Id:  chain.ID,
		Wif: chain.WIF(),
	}, nil
}

func (*ManageServer) CreateBlock(ctx context.Context, request *manage.BlockCreationRequest) (*manage.BlockCreationResponse, error) {
	utils.L.Debug("start creating a block")

	chain := blockchain.LoadChain(request.ChainID)
	if chain == nil {
		return nil, nil
	}
	if !chain.IsOwner() {
		return nil, nil
	}

	block := chain.CreateBlock(request.Payload)
	if chain.SaveBlock(block) {
		b := &protocol.BroadcastBlock{}
		b.SetBlock(block)
		defer func() { protocol.BroadcastChannel <- b }()
	}

	return &manage.BlockCreationResponse{
		Height:    block.Height,
		Time:      block.Time,

		PrevHash:  block.PrevHash,
		Hash:      block.Hash,
		Signature: block.Signature,
	}, nil
}

func (*ManageServer) AddChain(ctx context.Context, request *manage.ChainRequest) (*manage.CommonResponse, error) {
	blockchain.NewReadonlyChain(request.Id).Sync()
	return &manage.CommonResponse{Success: true}, nil
}

func (*ManageServer) DeleteChain(ctx context.Context, request *manage.ChainRequest) (*manage.CommonResponse, error) {
	if chain := blockchain.LoadChain(request.Id); chain != nil {
		blockchain.SharedStorage().CleanChain(chain)
		return &manage.CommonResponse{Success: true}, nil
	} else {
		return &manage.CommonResponse{Success: false, Error: "deleting chain is not exist"}, nil
	}
}

func (*ManageServer) GetPeers(request *manage.PeerListRequest, stream manage.IFCManage_GetPeersServer) error {
	peers := network.SharedStorage().GetPeers(int(request.Count))
	for _, peer := range peers {
		var response *manage.PeerResponse
		online := services.SharedServer.Peers[peer.Addr]
		if online == nil {
			response = &manage.PeerResponse{
				Addr:   peer.Addr,
				Rank:   int32(peer.Rank),
				Last:   peer.Last.Unix(),
				Server: peer.IsServer,
				Online: false}
		} else {
			response = &manage.PeerResponse{
				Addr:   online.Addr,
				Rank:   int32(online.Rank),
				Last:   online.Last.Unix(),
				Server: online.IsServer,
				Online: true}
		}
		if err := stream.Send(response); err != nil {
			return err
		}
	}
	return nil
}

func (*ManageServer) AddPeer(ctx context.Context, request *manage.PeerRequest) (*manage.CommonResponse, error) {
	peer := &network.Peer{Addr: request.Addr, Rank: 100}
	peer.Save()
	return &manage.CommonResponse{Success: true}, nil
}

func (*ManageServer) ConnectPeer(ctx context.Context, request *manage.PeerRequest) (*manage.CommonResponse, error) {
	if peer := services.SharedServer.Peers[request.Addr]; peer != nil {
		return &manage.CommonResponse{Success: false, Error: "already connected"}, nil
	}
	if peer := network.SharedStorage().GetPeer(request.Addr); peer != nil {
		services.SharedServer.Connect(peer)
		return &manage.CommonResponse{Success: true}, nil
	}
	peer := network.NewPeer(request.Addr, 100)
	services.SharedServer.Connect(peer)
	return &manage.CommonResponse{Success: true}, nil
}

func (*ManageServer) DisconnPeer(ctx context.Context, request *manage.PeerRequest) (*manage.CommonResponse, error) {
	if peer := services.SharedServer.Peers[request.Addr]; peer != nil {
		close(peer.Send)
		return &manage.CommonResponse{Success: true}, nil
	}
	return &manage.CommonResponse{Success: false, Error: "peer is not connected"}, nil
}

func (*ManageServer) DeletePeer(ctx context.Context, request *manage.PeerRequest) (*manage.CommonResponse, error) {
	if peer := services.SharedServer.Peers[request.Addr]; peer != nil {
		close(peer.Send)
		network.SharedStorage().DeletePeer(peer)
		return &manage.CommonResponse{Success: true}, nil
	}
	if peer := network.SharedStorage().GetPeer(request.Addr); peer != nil {
		network.SharedStorage().DeletePeer(peer)
		return &manage.CommonResponse{Success: true}, nil
	}
	return &manage.CommonResponse{Success: false, Error: "peer is not exist"}, nil
}

func GetPeers(count int32) {
	stream, err := IFCManageClient.GetPeers(context.Background(), &manage.PeerListRequest{Count: count})
	if err != nil {
		fmt.Println(err)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Address", "Rank", "Last/Duration", "Type", "Online?"})
	cachedPeers = map[string]bool{}
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		cachedPeers[in.Addr] = in.Online

		duration := "never"
		if in.Last > 0 {
			if in.Online {
				duration = time.Since(time.Unix(in.Last, 0)).Round(time.Second).String()
			} else {
				duration = time.Unix(in.Last, 0).String()
			}
		}

		t := "client"
		if in.Server {
			t = "server"
		}

		online := ""
		if in.Online {
			online = "✓"
		}

		table.Append([]string{in.Addr, strconv.Itoa(int(in.Rank)), duration, t, online})
	}
	table.Render()
}

func GetChains(id string) {
	stream, err := IFCManageClient.GetChains(context.Background(), &manage.ChainRequest{Id: id})
	if err != nil {
		fmt.Println(err)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Ref", "Chain ID", "Block Count"})
	cachedChains = map[int64]*cachedChain{}
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println()
			return
		}
		cachedChains[in.Ref] = &cachedChain{
			in.Id,
			in.Ref,
			in.Count,
		}
		table.Append([]string{strconv.FormatInt(in.Ref, 10), in.Id, strconv.FormatUint(in.Count, 10)})
	}
	table.Render()
}

func GetBlocks(id string, from uint64, to uint64) {
	stream, err := IFCManageClient.GetBlocks(context.Background(), &manage.BlockRequest{ChainID: id, From: from, To: to})
	if err != nil {
		fmt.Println(err)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Height", "Time", "Prev Hash", "Hash", "Signature"})
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		prev := in.PrevHash
		if len(prev) > 0 {
			prev = prev[:6]
		}
		table.Append([]string{
			strconv.FormatUint(in.Height, 10),
			time.Unix(int64(in.Time), 0).String(),
			prev,
			in.Hash[:6],
			in.Signature[:6],
		})
	}

	table.Render()
}

func DumpBlock(id string, height uint64) {
	stream, err := IFCManageClient.GetBlocks(context.Background(), &manage.BlockRequest{ChainID: id, From: height, To: height})
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("[Height   ] %v\n", in.Height)
		fmt.Printf("[Time     ] %v\n", in.Time)
		fmt.Printf("[PrevHash ] %v\n", in.PrevHash)
		fmt.Printf("[Hash     ] %v\n", in.Hash)
		fmt.Printf("[Signature] %v\n", in.Signature)

		buffer := &bytes.Buffer{}
		if err := json.Indent(buffer, in.Payload, "", "\t"); err != nil {
			fmt.Printf("[Payload  ] %v\n", string(in.Payload))
		} else {
			fmt.Println("[Payload  ]")
			_, err := buffer.WriteTo(os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println()
		}
	}
}

func CreateChain(name string, author string, website string, email string, desc string) {
	response, err := IFCManageClient.CreateChain(context.Background(), &manage.ChainCreationRequest{
		Name:    name,
		Author:  author,
		Website: website,
		Email:   email,
		Desc:    desc,
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	cachedChains[response.Ref] = &cachedChain{
		response.Id,
		response.Ref,
		1,
	}

	fmt.Printf("[Ref     ] %v\n", response.Ref)
	fmt.Printf("[Chain ID] %v\n", response.Id)
	fmt.Printf("[WIF     ] %v\n", response.Wif)
}

func CreateBlock(id string, payload []byte) {
	response, err := IFCManageClient.CreateBlock(context.Background(), &manage.BlockCreationRequest{ChainID: id, Payload: payload})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("[Height   ] %v\n", response.Height)
	fmt.Printf("[Time     ] %v\n", response.Time)
	fmt.Printf("[PrevHash ] %v\n", response.PrevHash)
	fmt.Printf("[Hash     ] %v\n", response.Hash)
	fmt.Printf("[Signature] %v\n", response.Signature)
}

func AddChain(id string) {
	response, err := IFCManageClient.AddChain(context.Background(), &manage.ChainRequest{Id: id})
	if err != nil {
		fmt.Println(err)
		return
	}

	if response.Success {
		fmt.Println("Added")
		GetChains(id)
	} else {
		fmt.Printf("%v\n", response.Error)
	}
}

func DeleteChain(id string) {
	response, err := IFCManageClient.DeleteChain(context.Background(), &manage.ChainRequest{Id: id})
	if err != nil {
		fmt.Println(err)
		return
	}

	if response.Success {
		fmt.Println("Deleted")
	} else {
		fmt.Printf("%v\n", response.Error)
	}
}

func AddPeer(addr string) {
	response, err := IFCManageClient.AddPeer(context.Background(), &manage.PeerRequest{Addr: addr})
	if err != nil {
		fmt.Println(err)
		return
	}

	if response.Success {
		fmt.Println("Added")
		cachedPeers[addr] = false
	} else {
		fmt.Printf("%v\n", response.Error)
	}
}

func ConnectPeer(addr string) {
	response, err := IFCManageClient.ConnectPeer(context.Background(), &manage.PeerRequest{Addr: addr})
	if err != nil {
		fmt.Println(err)
		return
	}

	if response.Success {
		fmt.Println("Connected")
		cachedPeers[addr] = true
	} else {
		fmt.Printf("%v\n", response.Error)
	}
}

func DisconnPeer(addr string) {
	response, err := IFCManageClient.DisconnPeer(context.Background(), &manage.PeerRequest{Addr: addr})
	if err != nil {
		fmt.Println(err)
		return
	}

	if response.Success {
		fmt.Println("Disconnected")
		cachedPeers[addr] = false
	} else {
		fmt.Printf("%v\n", response.Error)
	}
}

func DeletePeer(addr string) {
	response, err := IFCManageClient.DeletePeer(context.Background(), &manage.PeerRequest{Addr: addr})
	if err != nil {
		fmt.Println(err)
		return
	}

	if response.Success {
		fmt.Println("Deleted")
		delete(cachedPeers, addr)
	} else {
		fmt.Printf("%v\n", response.Error)
	}
}