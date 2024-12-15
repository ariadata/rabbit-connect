# VPN Project Technical Definition

## Project Overview
Custom VPN implementation using Go and TUN interfaces (github.com/songgao/water) for secure network tunneling between sites.

## Architecture
- TUN-based VPN using UDP protocol
- RC4 encryption for packet security
- Client-Server model with full tunnel capability
- Uses Linux network stack (iptables, routing)

## Network Topology

### Home Site
```
VPN IP: 172.16.0.2/24 (TUN/rabbit-connect)
LAN Network: 192.168.2.0/23
LAN Client: 192.168.2.4
Target Access Required: 192.168.100.101 (Office LAN)
```

### Office Site
```
VPN IP: 172.16.0.100/24 (TUN/rabbit-connect)
LAN Network: 192.168.100.0/23
LAN Client: 192.168.100.102
Target Access Required: 192.168.2.202 (Home LAN)
```

## Current Status
- VPN tunnel is functional
- Direct VPN IPs (172.16.0.x) can communicate
- Server (172.16.0.1) is pingable
- Cross-LAN traffic not working

## Applied Configuration
```bash
# IP Forwarding
sysctl -w net.ipv4.ip_forward=1

# IPTables Rules Applied
iptables -t nat -A POSTROUTING -i rabbit-connect -j MASQUERADE
iptables -t nat -A POSTROUTING -o rabbit-connect -j MASQUERADE
iptables -t nat -A POSTROUTING -j MASQUERADE
iptables -A FORWARD -i rabbit-connect -j ACCEPT
iptables -A FORWARD -d rabbit-connect -j ACCEPT
iptables -A FORWARD -i go-vpn -j ACCEPT
iptables -A FORWARD -o go-vpn -j ACCEPT
iptables -t nat -A POSTROUTING -o go-vpn -j MASQUERADE

# Routes Added
# Home:
ip route add 192.168.100.0/24 via 172.16.0.100 dev rabbit-connect
# Office:
ip route add 192.168.2.0/23 via 172.16.0.2 dev rabbit-connect
```

## Core Requirements
1. Full tunnel capability between sites
2. Access to remote LAN resources
3. Proper packet forwarding between VPN and LAN interfaces

## Current Issue
While VPN tunnel works for direct VPN IPs (172.16.0.x), cross-LAN communication fails. Need to enable proper packet forwarding for:
- Home -> Office LAN (192.168.100.0/24)
- Office -> Home LAN (192.168.2.0/23)

## Technologies
- Go 1.22+
- Linux with TUN support
- github.com/songgao/water
- github.com/patrickmn/go-cache
- UDP/Raw IP packets