package utils

import (
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
)

func init() {
	if err := os.MkdirAll("/usr/local/var/infnote/payloads/", 0755); err != nil {
		L.Fatal(err)
	}

	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 32767)
	viper.SetDefault("manage.host", "127.0.0.1")
	viper.SetDefault("manage.port", 32700)
	viper.SetDefault("data.file", "/usr/local/var/infnote/data.db")
	viper.SetDefault("data.root", "/usr/local/var/infnote/payloads/")
	viper.SetDefault("peers.sync", false)
	viper.SetDefault("peers.retry", 5)
	viper.SetDefault("hooks.block", nil)
	viper.SetDefault("daemon.pid", "/tmp/ifc.pid")
	viper.SetDefault("message.division", true)
	viper.SetDefault("message.maxsize", 1)

	// debug, info, notice, warning, error, critical
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", "/usr/local/var/infnote/daemon.log")

	viper.SetConfigType("yaml")
	viper.AddConfigPath("/usr/local/etc/infnote/")
}

func Migrate() {
	err := os.MkdirAll("/usr/local/etc/infnote/", 0755)
	if err != nil {
		L.Fatal(err)
	}

	err = ioutil.WriteFile("/usr/local/etc/infnote/config.yaml", []byte(
`daemon:
	# ifc service process
	pid: /tmp/ifc.pid
data:
	# all chains and blocks are saved here
	file: /usr/local/var/infnote/data.db
	root: /usr/local/var/infnote/payloads/
log:
	# avaliable: debug, info, notice, warning, error, critical
	level: info
	file: /usr/local/var/infnote/daemon.log
manage:
	# rpc management listen on
	host: 127.0.0.1
	port: 32700
peers:
	retry: 5
	# ifc will automatically sync peer list with any connected peer when set true
	sync: false
server:
	# ifc service listen on
	host: 0.0.0.0
	port: 32767
message:
	# message for transmit blocks will be divided to several messages
	division: true
	
	# max block payload size (MB) of one message can contain
	# only effective when division is true
	maxsize: 1
# hooks:
	# ifc service will call this by POST every time when receive a new block
	# blocks: "http://localhost/hooks/new_block"`), 0655)

	if err != nil {
		L.Fatal(err)
	}

	L.Info("create config file at /usr/local/etc/infnote/config.yaml")
}
