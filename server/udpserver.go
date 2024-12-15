// server/udpserver.go
package server

import (
	"log"
	"net"
	"time"

	"github.com/ariadata/rabbit-connect/common/cipher"
	"github.com/ariadata/rabbit-connect/common/config"
	"github.com/ariadata/rabbit-connect/tun"
	"github.com/patrickmn/go-cache"
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
)

type Forwarder struct {
	localConn *net.UDPConn
	connCache *cache.Cache // Maps IP to client UDP address
}

func StartUDPServer(config config.Config) {
	config.Init()
	iface := tun.CreateTun(config.CIDR)
	localAddr, err := net.ResolveUDPAddr("udp", config.LocalAddr)
	if err != nil {
		log.Fatalln("failed to get UDP socket:", err)
	}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		log.Fatalln("failed to listen on UDP socket:", err)
	}
	defer conn.Close()

	forwarder := &Forwarder{
		localConn: conn,
		connCache: cache.New(30*time.Minute, 10*time.Minute),
	}

	go forwarder.forward(iface)

	buf := make([]byte, 1500)
	for {
		n, cliAddr, err := conn.ReadFromUDP(buf)
		if err != nil || n == 0 {
			continue
		}

		b := cipher.Decrypt(buf[:n])
		if !waterutil.IsIPv4(b) {
			continue
		}

		// Get source and destination IPs
		srcIP := waterutil.IPv4Source(b)
		dstIP := waterutil.IPv4Destination(b)

		log.Printf("Received packet: src=%v dst=%v", srcIP, dstIP)

		// Update client address mapping
		forwarder.connCache.Set(srcIP.String(), cliAddr, cache.DefaultExpiration)

		// Write to TUN interface
		_, err = iface.Write(b)
		if err != nil {
			log.Printf("Error writing to TUN: %v", err)
		}
	}
}

func (f *Forwarder) forward(iface *water.Interface) {
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

		srcIP := waterutil.IPv4Source(b)
		dstIP := waterutil.IPv4Destination(b)

		log.Printf("Forwarding packet: src=%v dst=%v", srcIP, dstIP)

		// Find the right VPN client to forward to
		// This might be either the direct destination IP (for VPN IPs)
		// or the VPN IP that routes to this destination network
		var foundClient *net.UDPAddr

		// First try direct VPN IP match
		if client, exists := f.connCache.Get(dstIP.String()); exists {
			if addr, ok := client.(*net.UDPAddr); ok {
				foundClient = addr
				log.Printf("Found direct route to VPN IP %v", dstIP)
			}
		}

		// If no direct match, check all connected clients
		if foundClient == nil {
			for _, item := range f.connCache.Items() {
				if addr, ok := item.Object.(*net.UDPAddr); ok {
					foundClient = addr
					log.Printf("Using client %v for routing", addr)
					break
				}
			}
		}

		if foundClient != nil {
			encrypted := cipher.Encrypt(b)
			_, err = f.localConn.WriteToUDP(encrypted, foundClient)
			if err != nil {
				log.Printf("Error forwarding to %v: %v", foundClient, err)
			}
		} else {
			log.Printf("No route found for destination %v", dstIP)
		}
	}
}
