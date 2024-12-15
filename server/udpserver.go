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
	connCache *cache.Cache
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
	log.Printf("rabbit-connect udp server started on %v, CIDR is %v", config.LocalAddr, config.CIDR)

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

		// Write packet to TUN interface
		iface.Write(b)

		// Cache the client's address based on source IP for return traffic
		srcIP := waterutil.IPv4Source(b).String()
		forwarder.connCache.Set(srcIP, cliAddr, cache.DefaultExpiration)
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

		// Get the destination IP to determine which client to send to
		dstIP := waterutil.IPv4Destination(b).String()

		// Find the client that owns this destination IP range
		if clientAddr, found := f.connCache.Get(dstIP); found {
			b = cipher.Encrypt(b)
			_, _ = f.localConn.WriteToUDP(b, clientAddr.(*net.UDPAddr))
		}
	}
}
