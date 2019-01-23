package command

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var directCmd = &cobra.Command{
	Use:   "ifc",
	Short: "Infnote Chain is a blockchain implementation with peer-to-peer network.",
	Long: `Infnote is a decentralized information sharing system based on blockchain and peer-to-peer network, 
				aiming to provide an easy-to-use medium for users to share their thoughts, 
				insights and views freely without worrying about anonymity, data tampering and data loss.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			fmt.Println(err)
		}
	},
}

var cliRootCmd = &cobra.Command{}

func DirectExecute() {
	initDirectCommands()
	if err := directCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func CLIExecute(args []string) {
	cliRootCmd.SetArgs(args)
	if err := cliRootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
