package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Infnote/infnotechain/blockchain"
	"github.com/Infnote/infnotechain/protocol"
	"github.com/Infnote/infnotechain/services"
	"github.com/Infnote/infnotechain/services/codegen"
	"github.com/Infnote/infnotechain/utils"
	"github.com/olekukonko/tablewriter"
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

func RunDaemon() {
	pid, err := syscall.ForkExec(os.Args[0], []string{os.Args[0], "run", "-f"}, nil)
	if err != nil {
		utils.L.Fatal(err)
	}

	err = ioutil.WriteFile("/tmp/ifc.pid", []byte(fmt.Sprintf("%d", pid)), 0655)
	if err != nil {
		utils.L.Fatal(err)
	}
	fmt.Printf("Infnote Chain service start in child process %v\n", pid)
}

func StopDaemon() {
	b, err := ioutil.ReadFile("/tmp/ifc.pid")
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
	conn, err := net.Listen("tcp", "localhost:32700")
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

func (*ManageServer) GetPeers(request *manage.PeerRequest, stream manage.IFCManage_GetPeersServer) error {
	var count int32
	for peer := range services.ConnectedPeers {
		count += 1
		if err := stream.Send(&manage.PeerResponse{
			Addr: peer.Addr,
			Rank: int32(peer.Rank),
			Last: peer.Last.Unix()}); err != nil {

			return err
		}
		if count != 0 && count >= request.Count {
			break
		}
	}
	return nil
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

func GetPeers(count int32) {
	stream, err := IFCManageClient.GetPeers(context.Background(), &manage.PeerRequest{Count: count})
	if err != nil {
		fmt.Println(err)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Address", "Rank", "Duration"})
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		duration := "never"
		if in.Last > 0 {
			duration = time.Since(time.Unix(in.Last, 0)).String()
		}
		table.Append([]string{in.Addr, strconv.Itoa(int(in.Rank)), duration})
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
