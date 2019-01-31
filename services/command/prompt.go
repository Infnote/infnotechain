package command

import (
	"fmt"
	"github.com/Infnote/infnotechain/services/codegen"
	"github.com/Infnote/infnotechain/utils"
	"github.com/c-bata/go-prompt"
	"google.golang.org/grpc"
	"os"
	"strings"
)

func cacheChainSuggest() []prompt.Suggest {
	var s []prompt.Suggest
	for _, c := range cachedChains {
		s = append(s, prompt.Suggest{
			Text:        fmt.Sprintf("%v", c.Ref),
			Description: fmt.Sprintf("ID: %v, Count: %v", c.ID, c.Count)})
	}
	return s
}

func Executer(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	} else if input == "quit" || input == "exit" {
		fmt.Println("Bye!")
		os.Exit(0)
		return
	}

	CLIExecute(strings.Split(input, " "))
}

func Completer(doc prompt.Document) []prompt.Suggest {
	if doc.TextBeforeCursor() == "" {
		return nil
	}

	args := strings.Split(doc.TextBeforeCursor(), " ")

	if len(args) == 1 {
		s := []prompt.Suggest{
			{Text: "peers", Description: "Print online peers"},
			{Text: "chains", Description: "Print accepted chains"},
			{Text: "blocks", Description: "Print blocks"},
			{Text: "dump", Description: "Print a block with detail"},
			{Text: "use", Description: "Set chain as current context"},
			{Text: "createblock", Description: "Create a new block"},
			{Text: "createchain", Description: "Create a new chain"},
		}
		return prompt.FilterContains(s, doc.GetWordBeforeCursor(), true)
	}

	if args[0] == "use" && len(args) == 2 {
		return cacheChainSuggest()
	}

	if args[0] == "createblock" && len(args) == 2 {
		return []prompt.Suggest{
			{Text: "string"},
			{Text: "base58"},
		}
	}

	if args[0] == "createchain" {
		return []prompt.Suggest{
			{Text: "--name"},
			{Text: "--author"},
			{Text: "--website"},
			{Text: "--email"},
			{Text: "--desc"},
		}
	}

	return nil
}

func connect() {
	conn, err := grpc.Dial("localhost:32700", grpc.WithInsecure())
	if err != nil {
		utils.L.Fatal(err)
	}

	IFCManageClient = manage.NewIFCManageClient(conn)
}

func UI() {
	initCLICommands()
	connect()
	fmt.Println("IFC service connected")
	ui := prompt.New(
		Executer,
		Completer,
		prompt.OptionTitle("IFC command line interface"),
		prompt.OptionPrefix("> "),
		prompt.OptionInputTextColor(prompt.Green))
	ui.Run()
}
