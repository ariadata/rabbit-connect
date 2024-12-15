package register

import (
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

type ClientInfo struct {
	IP       string
	Subnets  []string
	Count    int
	LastSeen time.Time
}

var _register *cache.Cache
var _routeMap sync.Map // maps subnet to client IP

func init() {
	_register = cache.New(30*time.Minute, 3*time.Minute)
}

// AddClientIP adds a client IP to the registry
func AddClientIP(ip string) {
	info := &ClientInfo{
		IP:       ip,
		Subnets:  []string{},
		Count:    0,
		LastSeen: time.Now(),
	}
	_register.Add(ip, info, cache.DefaultExpiration)
}

// DeleteClientIP removes a client IP from the registry
func DeleteClientIP(ip string) {
	if info, ok := getClientInfo(ip); ok {
		// Remove subnet mappings
		for _, subnet := range info.Subnets {
			_routeMap.Delete(subnet)
		}
	}
	_register.Delete(ip)
}

// ExistClientIP checks if a client IP exists
func ExistClientIP(ip string) bool {
	_, ok := _register.Get(ip)
	return ok
}

// AddClientSubnet adds a subnet that a client can route to
func AddClientSubnet(clientIP string, subnet string) bool {
	if info, ok := getClientInfo(clientIP); ok {
		// Check if subnet is valid
		_, _, err := net.ParseCIDR(subnet)
		if err != nil {
			return false
		}

		// Add subnet to client info
		for _, existing := range info.Subnets {
			if existing == subnet {
				return true // Already exists
			}
		}
		info.Subnets = append(info.Subnets, subnet)
		info.LastSeen = time.Now()
		_register.Set(clientIP, info, cache.DefaultExpiration)

		// Add to route map
		_routeMap.Store(subnet, clientIP)
		return true
	}
	return false
}

// GetRouteForSubnet finds which client can route to a given subnet
func GetRouteForSubnet(subnet string) (string, bool) {
	if clientIP, ok := _routeMap.Load(subnet); ok {
		return clientIP.(string), true
	}
	return "", false
}

// KeepAliveClientIP updates the last seen time for a client
func KeepAliveClientIP(ip string) {
	if info, ok := getClientInfo(ip); ok {
		info.Count++
		info.LastSeen = time.Now()
		_register.Set(ip, info, cache.DefaultExpiration)
	} else {
		AddClientIP(ip)
	}
}

// PickClientIP allocates a new client IP from the CIDR range
func PickClientIP(cidr string) (clientIP string, prefixLength string) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Panicf("error cidr %v", cidr)
	}
	total := addressCount(ipNet) - 3
	index := uint64(0)
	//skip first ip
	ip = incr(ipNet.IP.To4())
	for {
		ip = incr(ip)
		index++
		if index >= total {
			break
		}
		if !ExistClientIP(ip.String()) {
			AddClientIP(ip.String())
			return ip.String(), strings.Split(cidr, "/")[1]
		}
	}
	return "", ""
}

// ListClientIP returns all registered client IPs
func ListClientIP() []string {
	var result []string
	for k, v := range _register.Items() {
		if info, ok := v.Object.(*ClientInfo); ok {
			result = append(result, k+" (subnets: "+strings.Join(info.Subnets, ", ")+")")
		} else {
			result = append(result, k)
		}
	}
	return result
}

// ListClientSubnets returns all subnets registered for a client
func ListClientSubnets(clientIP string) []string {
	if info, ok := getClientInfo(clientIP); ok {
		return info.Subnets
	}
	return nil
}

// Helper functions

func getClientInfo(ip string) (*ClientInfo, bool) {
	if val, ok := _register.Get(ip); ok {
		if info, ok := val.(*ClientInfo); ok {
			return info, true
		}
	}
	return nil, false
}

func addressCount(network *net.IPNet) uint64 {
	prefixLen, bits := network.Mask.Size()
	return 1 << (uint64(bits) - uint64(prefixLen))
}

func incr(IP net.IP) net.IP {
	IP = checkIPv4(IP)
	incIP := make([]byte, len(IP))
	copy(incIP, IP)
	for j := len(incIP) - 1; j >= 0; j-- {
		incIP[j]++
		if incIP[j] > 0 {
			break
		}
	}
	return incIP
}

func checkIPv4(ip net.IP) net.IP {
	if v4 := ip.To4(); v4 != nil {
		return v4
	}
	return ip
}
