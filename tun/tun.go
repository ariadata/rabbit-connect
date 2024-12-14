package tun

import (
	"log"
	"runtime"

	"github.com/ariadata/rabbit-connect/common/osutil"
	"github.com/songgao/water"
)

func CreateTun(cidr string) (iface *water.Interface) {
	c := water.Config{DeviceType: water.TUN}
	os := runtime.GOOS
	if os != "darwin" {
		c.Name = "rabbit-connect"
	}
	iface, err := water.New(c)
	if err != nil {
		log.Fatalln("failed to allocate TUN interface:", err)
	}
	log.Println("interface allocated:", iface.Name())
	osutil.ConfigTun(cidr, iface)
	return iface
}
