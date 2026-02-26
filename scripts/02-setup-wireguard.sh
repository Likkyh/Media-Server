#!/bin/bash
# =============================================================================
# Media Server Setup Script - Phase 2: Wireguard VPN
# Run this script after 01-setup-system.sh
# =============================================================================

set -e

echo "=========================================="
echo "  Media Server Setup - Phase 2"
echo "  Wireguard VPN Configuration"
echo "=========================================="

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root or with sudo${NC}"
    exit 1
fi

WG_DIR="/etc/wireguard"
CLIENT_DIR="$HOME/wireguard-clients"

echo -e "${YELLOW}[1/4] Generating server keys...${NC}"
mkdir -p "$WG_DIR"
cd "$WG_DIR"

# Generate server keys
wg genkey | tee server_private.key | wg pubkey > server_public.key
chmod 600 server_private.key

SERVER_PRIVATE=$(cat server_private.key)
SERVER_PUBLIC=$(cat server_public.key)

echo -e "${YELLOW}[2/4] Generating client keys...${NC}"
mkdir -p "$CLIENT_DIR"
cd "$CLIENT_DIR"

# Generate client 1 keys (Admin PC)
wg genkey | tee client1_private.key | wg pubkey > client1_public.key
CLIENT1_PRIVATE=$(cat client1_private.key)
CLIENT1_PUBLIC=$(cat client1_public.key)

# Generate client 2 keys (Mobile)
wg genkey | tee client2_private.key | wg pubkey > client2_public.key
CLIENT2_PRIVATE=$(cat client2_private.key)
CLIENT2_PUBLIC=$(cat client2_public.key)

echo -e "${YELLOW}[3/4] Creating server configuration...${NC}"

# Detect main network interface
INTERFACE=$(ip route | grep default | awk '{print $5}' | head -n1)

cat > "$WG_DIR/wg0.conf" << EOF
[Interface]
PrivateKey = $SERVER_PRIVATE
Address = 10.0.0.1/24
ListenPort = 51820
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o $INTERFACE -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o $INTERFACE -j MASQUERADE

# Client 1 - Admin PC
[Peer]
PublicKey = $CLIENT1_PUBLIC
AllowedIPs = 10.0.0.2/32

# Client 2 - Mobile
[Peer]
PublicKey = $CLIENT2_PUBLIC
AllowedIPs = 10.0.0.3/32
EOF

chmod 600 "$WG_DIR/wg0.conf"

echo -e "${YELLOW}[4/4] Creating client configurations...${NC}"

# Get public IP
PUBLIC_IP=$(curl -s ifconfig.me || echo "YOUR_PUBLIC_IP")

# Client 1 config
cat > "$CLIENT_DIR/client1.conf" << EOF
[Interface]
PrivateKey = $CLIENT1_PRIVATE
Address = 10.0.0.2/24
DNS = 1.1.1.1

[Peer]
PublicKey = $SERVER_PUBLIC
Endpoint = $PUBLIC_IP:51820
AllowedIPs = 10.0.0.0/24, 192.168.1.0/24
PersistentKeepalive = 25
EOF

# Client 2 config
cat > "$CLIENT_DIR/client2.conf" << EOF
[Interface]
PrivateKey = $CLIENT2_PRIVATE
Address = 10.0.0.3/24
DNS = 1.1.1.1

[Peer]
PublicKey = $SERVER_PUBLIC
Endpoint = $PUBLIC_IP:51820
AllowedIPs = 10.0.0.0/24, 192.168.1.0/24
PersistentKeepalive = 25
EOF

# Enable and start Wireguard
systemctl enable wg-quick@wg0
systemctl start wg-quick@wg0

echo ""
echo -e "${GREEN}=========================================="
echo "  Wireguard Configuration Complete!"
echo "==========================================${NC}"
echo ""
echo "Server Public Key: $SERVER_PUBLIC"
echo "Server Public IP: $PUBLIC_IP"
echo ""
echo "Client configurations saved in: $CLIENT_DIR"
echo "  - client1.conf (Admin PC)"
echo "  - client2.conf (Mobile)"
echo ""
echo "To verify: wg show"
echo ""
echo -e "${YELLOW}IMPORTANT: Configure port forwarding on your router:${NC}"
echo "  UDP 51820 -> <server-ip>:51820"
echo ""
echo "Next step: Run ./03-setup-docker.sh"
echo ""
