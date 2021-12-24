package config

import "os"

type Config struct {
	// Addr is a server address.
	Addr string
	// RPCAddr is an address of Ethereum RPC server.
	RPCAddr string

	DBPath string
}

var userHomeDir, _ = os.UserHomeDir()

var DefaultConfig = &Config{
	Addr:    "0.0.0.0:443",
	RPCAddr: "ws://127.0.0.1:8545",
	DBPath:  userHomeDir + "/.mettu/",
}
