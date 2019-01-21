package services

import (
	"fmt"
	"github.com/Infnote/infnotechain/utils"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
)

func RunDaemon() {
	pid, err := syscall.ForkExec(os.Args[0], []string{os.Args[0], "run", "-f"}, nil)
	if err != nil {
		utils.L.Fatal(err)
	}
	if pid == 0 {
		PeerService()
	} else {
		err := ioutil.WriteFile("/tmp/ifc.pid", []byte(fmt.Sprintf("%d", pid)), 0655)
		if err != nil {
			utils.L.Fatal(err)
		}
		fmt.Printf("Infnote Chain service start in child process %v\n", pid)
	}
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
