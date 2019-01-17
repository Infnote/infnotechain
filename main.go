package main

import "github.com/Infnote/infnotechain/network"

func main() {
	s := network.NewServer()
	s.Run()
}
