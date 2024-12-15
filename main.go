package main

import (
	"flag"

	"github.com/ariadata/rabbit-connect/client"
	"github.com/ariadata/rabbit-connect/common/config"
	"github.com/ariadata/rabbit-connect/server"
)

func main() {
	cfg := config.Config{}
	flag.StringVar(&cfg.CIDR, "c", "172.16.0.1/24", "tun interface CIDR")
	flag.StringVar(&cfg.LocalAddr, "l", "0.0.0.0:3000", "local address")
	flag.StringVar(&cfg.ServerAddr, "s", "0.0.0.0:3001", "server address")
	flag.StringVar(&cfg.Key, "k", "6w9z$C&F)J@NcRfWjXn3r4u7x!A%D*G-", "encryption key")
	flag.BoolVar(&cfg.ServerMode, "S", false, "Server mode")
	flag.Parse()

	if cfg.ServerMode {
		server.StartUDPServer(cfg)
	} else {
		client.StartUDPClient(cfg)
	}
}
