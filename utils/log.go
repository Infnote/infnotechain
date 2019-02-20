package utils

import (
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"log"
	"os"
)

const (
	NONE   int = 0
	STDOUT int = 1
	FILE   int = 1 << 1
)

var level = map[string]logging.Level{
	"debug":    logging.DEBUG,
	"info":     logging.INFO,
	"notice":   logging.NOTICE,
	"warning":  logging.WARNING,
	"error":    logging.ERROR,
	"critical": logging.CRITICAL,
}

var L = logging.MustGetLogger("ifc")
var stdout logging.LeveledBackend
var fout logging.LeveledBackend

func init() {
	formatter := logging.MustStringFormatter(
		`%{color}[%{time:15:04:05.000}][%{shortfunc}][%{level:.4s}]%{color:reset} %{message}`)
	stdout = logging.AddModuleLevel(
		logging.NewBackendFormatter(
			logging.NewLogBackend(os.Stdout, "", 0), formatter))
	stdout.SetLevel(logging.INFO, "")

	logging.SetBackend(stdout)
}

func SetLoggingMode(mode int) {
	if mode == NONE {
		logging.SetBackend()
	}

	l, ok := level[viper.GetString("log.level")]
	if !ok {
		log.Fatalf("log level '%v' is not supported\n", viper.GetString("log.level"))
	}

	file, err := os.OpenFile(viper.GetString("log.file"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	fileFormatter := logging.MustStringFormatter(
		`[%{time:2006-01-02 15:04:05.000}][%{longfunc}][%{level:.4s}] %{message}`)
	fout = logging.AddModuleLevel(
		logging.NewBackendFormatter(
			logging.NewLogBackend(file, "", 0), fileFormatter))

	stdout.SetLevel(l, "")
	fout.SetLevel(l, "")

	var backends []logging.Backend
	if mode&STDOUT > 0 {
		backends = append(backends, stdout)
	}

	if mode&FILE == FILE {
		backends = append(backends, fout)
	}

	logging.SetBackend(backends...)
}
