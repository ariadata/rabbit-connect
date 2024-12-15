package client

import (
	"log"
	"net"

	"github.com/ariadata/rabbit-connect/common/cipher"
	"github.com/ariadata/rabbit-connect/common/config"
	"github.com/ariadata/rabbit-connect/tun"
	"github.com/songgao/water/waterutil"
)

// StartUDPClient start udp client
func StartUDPClient(config config.Config) {
	config.Init()
	iface := tun.CreateTun(config.CIDR)
	serverAddr, err := net.ResolveUDPAddr("udp", config.ServerAddr)
	if err != nil {
		log.Fatalln("failed to resolve server addr:", err)
	}
	localAddr, err := net.ResolveUDPAddr("udp", config.LocalAddr)
	if err != nil {
		log.Fatalln("failed to get UDP socket:", err)
	}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		log.Fatalln("failed to listen on UDP socket:", err)
	}
	defer conn.Close()
	log.Printf("go-vpn udp client started on %v, CIDR is %v", config.LocalAddr, config.CIDR)

	// read data from server
	go func() {
		buf := make([]byte, 1500)
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil || n == 0 {
				continue
			}
			// decrypt data
			b := cipher.Decrypt(buf[:n])
			if !waterutil.IsIPv4(b) {
				continue
			}
			// Forward all IPv4 packets to the TUN interface
			iface.Write(b)
		}
	}()

	// read data from tun
	packet := make([]byte, 1500)
	for {
		n, err := iface.Read(packet)
		if err != nil || n == 0 {
			continue
		}
		b := packet[:n]
		if !waterutil.IsIPv4(b) {
			continue
		}
		// Forward all IPv4 packets to the server
		b = cipher.Encrypt(b)
		_, _ = conn.WriteToUDP(b, serverAddr)
	}
}
