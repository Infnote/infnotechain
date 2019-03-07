package utils

import (
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
)

func CheckProcessAlive() bool {
	filename := viper.GetString("daemon.pid")
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(string(content))
	if err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return false
	}

	return true
}
