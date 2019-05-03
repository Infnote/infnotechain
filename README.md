# Infnote Chain ![Go Version](https://img.shields.io/badge/go-1.11.4-green.svg) ![Go Report](https://goreportcard.com/badge/github.com/Infnote/infnotechain)



## Installation

Get the repo and build:

```bash
go get -u github.com/Infnote/infnotechain
cd $GOPATH/src/github.com/Infnote/infnotechain
go build -o /usr/local/bin/ifc ./

ifc run
```

Don't forget to set `GOPATH` correctly and add `/usr/local/bin` to your `PATH` variable

## How to Use

- `ifc run -f` run the program at foreground
- `ifc run -d` run the program with debug level log
- `ifc stop` stop the background process
- `ifc eject` eject config file to `/usr/local/etc/infnote/config.yaml`
- `ifc cli` run an interactive command line tool

## TODO:

- [ ] Communication starts with "Sync" message which will be responded an "Info"
- [ ] Dangled block error trigger "Sync"
- [ ] Writing test
- [ ] Check if peer connection is still alive by send a info 
- [x] ~~Respond 'Error' when cannot respond correctly~~
- [ ] Blocks request strategy
- [ ] Refresh connections strategy
- [ ] Peers updating strategy
    - [ ] Validate peer address
    - [ ] Ranking peers
    - [ ] Filter invalid peers received from outside
    - [ ] Check address equivalence then repleace the old one
- [ ] Clean boardcast id regularly
- [ ] Need some kind of authorization before boardcast
- [ ] Ranking strategy
    - [ ] Bad chain detection
    - [ ] Bad peer detection
    - [ ] Peer responding speed
    - [ ] Peer stability

### Maybe TODO:

- [ ] Adjust connections graph automatically