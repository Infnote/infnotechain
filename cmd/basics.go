package cmd

import (
	"fmt"
	"github.com/Infnote/infnotechain/database"
	"github.com/Infnote/infnotechain/services"
	"github.com/spf13/cobra"
)

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

var foreground bool
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start Infnote Chain service",
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("foreground").Value.String() == "true" {
			services.PeerService()
		} else {
			services.RunDaemon()
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Infnote Chain service",
	Run: func(cmd *cobra.Command, args []string) {
		services.StopDaemon()
	},
}

func init() {
	runCmd.Flags().BoolVarP(
		&foreground,
		"foreground",
		"f",
		false,
		"Running service and logging in foreground")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(stopCmd)
}
