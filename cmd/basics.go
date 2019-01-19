package cmd

import (
	"fmt"
	"github.com/Infnote/infnotechain/database"
	"github.com/Infnote/infnotechain/services"
	"github.com/Infnote/infnotechain/utils"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
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
			pid, err := syscall.ForkExec(os.Args[0], []string{"run"}, nil)
			if err != nil {
				utils.L.Fatal(err)
			}
			if pid == 0 {
				services.PeerService()
			} else {
				err := ioutil.WriteFile("/tmp/ifc.pid", []byte(fmt.Sprintf("%d", pid)), 0655)
				if err != nil {
					utils.L.Fatal(err)
				}
				fmt.Printf("Infnote Chain service start in child process %v\n", pid)
			}
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Infnote Chain service",
	Run: func(cmd *cobra.Command, args []string) {
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

		utils.L.Infof("[PID: %v] service stopped", pid)
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
