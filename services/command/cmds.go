package command

import (
	"errors"
	"fmt"
	"github.com/Infnote/infnotechain/database"
	"github.com/Infnote/infnotechain/services"
	"github.com/mr-tron/base58"
	"github.com/spf13/cobra"
	"strconv"
)

// - Direct Commands

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Infnote Chain",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Infnote Chain v0.1")
		fmt.Println("Protocol v1.1")
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create file and set environment for running",
	Run: func(cmd *cobra.Command, args []string) {
		database.Migrate()
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start Infnote Chain service",
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("foreground").Value.String() == "true" {
			go RunManageServer()
			services.PeerService()
		} else {
			RunDaemon()
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Infnote Chain service",
	Run: func(cmd *cobra.Command, args []string) {
		StopDaemon()
	},
}

var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Text interface for control Infnote Chain service",
	Run: func(cmd *cobra.Command, args []string) {
		UI()
	},
}

// - CLI Commands

var peersCmd = &cobra.Command{
	Use:   "peers",
	Short: "Print online peers",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			_, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		count := 0
		if len(args) > 0 {
			count, _ = strconv.Atoi(args[0])
		}
		GetPeers(int32(count))
	},
}

var chainsCmd = &cobra.Command{
	Use:   "chains",
	Short: "Print accepted chains detail",
	Run: func(cmd *cobra.Command, args []string) {
		id := ""
		if len(args) > 0 {
			id = args[1]
		}
		GetChains(id)
	},
}

var blocksCmd = &cobra.Command{
	Use:   "blocks",
	Short: "Print blocks detail",
	Args: func(cmd *cobra.Command, args []string) error {
		if chainContext == nil {
			return errors.New("chain context not set")
		}

		if len(args) < 2 {
			return errors.New("usage: blocks [chain id] [from] [to]")
		}

		from, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		to, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}

		if from > to {
			return errors.New("'from' should not bigger than 'to'")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		from, _ := strconv.Atoi(args[0])
		to, _ := strconv.Atoi(args[1])
		GetBlocks(chainContext.ID, uint64(from), uint64(to))
	},
}

var dumpCmd = &cobra.Command{
	Use: "dump",
	Short: "Print a all details of a block with height",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if chainContext == nil {
			fmt.Println("chain context not set")
			return
		}

		height, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		DumpBlock(chainContext.ID, uint64(height))
	},
}

var useChainCmd = &cobra.Command{
	Use: "use",
	Short: "Select a chain as current context",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ref, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		chain, ok := cachedChains[int64(ref)]
		if !ok {
			fmt.Println("unknown ref of chain, run 'chains' first")
			return
		}
		chainContext = chain
		fmt.Printf("switch to chain '%v'\n", chain.ID)
	},
}

var createChainCmd = &cobra.Command{
	Use:   "createchain",
	Short: "create a new chain with genesis block payload as chain info",
	Run: func(cmd *cobra.Command, args []string) {
		CreateChain(
			cmd.Flag("name").Value.String(),
			cmd.Flag("author").Value.String(),
			cmd.Flag("website").Value.String(),
			cmd.Flag("email").Value.String(),
			cmd.Flag("desc").Value.String())
	},
}

var createBlockCmd = &cobra.Command{
	Use:   "createblock",
	Short: "create a new block with payload",
	Args: func(cmd *cobra.Command, args []string) error {
		if chainContext == nil {
			return errors.New("chain context not set")
		}
		if len(args) < 2 {
			return errors.New("usage: createblock [string|base58] [payload]")
		}
		if args[0] != "string" && args[0] != "base58" {
			return errors.New("only support string or base58")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var payload []byte
		if args[0] == "string" {
			payload = []byte(args[1])
		} else if args[0] == "base58" {
			b, err := base58.Decode(args[1])
			if err != nil {
				fmt.Println(err)
				return
			}
			payload = b
		}
		CreateBlock(chainContext.ID, payload)
	},
}

func initDirectCommands() {
	runCmd.Flags().BoolP(
		"foreground",
		"f",
		false,
		"Running service and logging in foreground")

	directCmd.AddCommand(versionCmd)
	directCmd.AddCommand(initCmd)
	directCmd.AddCommand(runCmd)
	directCmd.AddCommand(stopCmd)
	directCmd.AddCommand(cliCmd)
}

func initCLICommands() {
	createChainCmd.Flags().StringP("name", "n", "", "name of the chain")
	createChainCmd.Flags().StringP("author", "a", "", "author of the chain")
	createChainCmd.Flags().StringP("website", "w", "", "website of the chain")
	createChainCmd.Flags().StringP("email", "e", "", "email of the chain")
	createChainCmd.Flags().StringP("desc", "d", "", "description of the chain")

	cliRootCmd.AddCommand(peersCmd)
	cliRootCmd.AddCommand(chainsCmd)
	cliRootCmd.AddCommand(useChainCmd)
	cliRootCmd.AddCommand(blocksCmd)
	cliRootCmd.AddCommand(dumpCmd)
	cliRootCmd.AddCommand(createChainCmd)
	cliRootCmd.AddCommand(createBlockCmd)
}
