package utils

import (
	"github.com/op/go-logging"
	"log"
	"os"
)

var L = logging.MustGetLogger("ifc")

func init() {
	formatter := logging.MustStringFormatter(
		`%{color}[%{time:15:04:05.000}][%{shortfunc}][%{level:.4s}]%{color:reset} %{message}`)
	stdout := logging.AddModuleLevel(
		logging.NewBackendFormatter(
			logging.NewLogBackend(os.Stdout, "", 0), formatter))
	stdout.SetLevel(logging.DEBUG, "")

	file, err := os.OpenFile(
		"/usr/local/var/infnote/daemon.log",
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0666)
	if err != nil {
		log.Fatal(err)
	}

	fileFormatter := logging.MustStringFormatter(
		`[%{time:15:04:05.000}][%{longfunc}][%{level:.4s}] %{message}`)
	fout := logging.AddModuleLevel(
		logging.NewBackendFormatter(
			logging.NewLogBackend(file, "", 0), fileFormatter))
	fout.SetLevel(logging.INFO, "")

	logging.SetBackend(stdout, fout)
}
