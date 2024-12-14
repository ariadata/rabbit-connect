# Go VPN

A simple VPN server and client based on Go, featuring easy installation and systemd service management.

## Quick Install

### Server Side
```bash
bash <(curl -sSL https://raw.githubusercontent.com/ariadata/rabbit-connect/main/installer-server.sh)
```

The server installer will:
- Check system requirements (Linux x64 + root access)
- Install the latest version of VPN server
- Configure as a systemd service
- Set up IP forwarding and iptables rules
- Start the service automatically

### Client Side
```bash
apt install -y netfilter-persistent iptables
systemctl enable --now netfilter-persistent
bash <(curl -sSL https://raw.githubusercontent.com/ariadata/rabbit-connect/main/installer-client.sh)

# for whole system :
# Add net.ipv4.ip_forward=1 end of /etc/sysctl.conf
# Or :
sysctl -w net.ipv4.ip_forward=1
echo "net.ipv4.ip_forward=1" | tee /etc/sysctl.d/99-ip-forward.conf
sysctl -p /etc/sysctl.d/99-ip-forward.conf

ip route add 195.201.194.199 via 192.168.100.1 dev eth0
ip route del default
ip route add default via 172.18.0.1 dev rabbit-connect
iptables -t nat -A POSTROUTING -o rabbit-connect -j MASQUERADE
iptables-save > /etc/iptables/rules.v4


netfilter-persistent save
```

The client installer will:
- Check system requirements (Linux x64 + root access)
- Install the latest version of VPN client
- Configure as a systemd service
- Set up required iptables rules
- Start the service automatically
- Test the connection

## Service Management

### Server Commands
```bash
systemctl start rabbit-connect-server    # Start the server
systemctl stop rabbit-connect-server     # Stop the server
systemctl restart rabbit-connect-server  # Restart the server
systemctl status rabbit-connect-server   # Check server status
```

### Client Commands
```bash
systemctl start rabbit-connect-client    # Start the client
systemctl stop rabbit-connect-client     # Stop the client
systemctl restart rabbit-connect-client  # Restart the client
systemctl status rabbit-connect-client   # Check client status
```

## Testing VPN Connection

After installing the client, you can test the connection using these commands:

### Ping Test
```bash
# Ping the VPN gateway (if your VPN CIDR is 172.18.0.10/24, this pings 172.18.0.1)
ping 172.18.0.1

# Ping other hosts in your VPN network
ping 172.18.0.2
```

### Check Your IP
```bash
# Check your IP through the VPN interface
curl --interface rabbit-connect myip4.ir

# Compare with your regular IP
curl myip4.ir
```

## Requirements
- Linux x64 system
- Root access
- systemd
- iptables

## Features
- Automated installation
- systemd service management
- Persistent iptables rules
- Automatic IP forwarding
- Connection testing
- Service auto-restart on failure

## Uninstallation
To completely remove the VPN server or client:

### Server
```bash
systemctl stop rabbit-connect-server
systemctl disable rabbit-connect-server
rm /etc/systemd/system/rabbit-connect-server.service
rm /usr/local/bin/rabbit-connect-server
systemctl daemon-reload
```

### Client
```bash
systemctl stop rabbit-connect-client
systemctl disable rabbit-connect-client
rm /etc/systemd/system/rabbit-connect-client.service
rm /usr/local/bin/rabbit-connect-client
systemctl daemon-reload
```

## License
MIT