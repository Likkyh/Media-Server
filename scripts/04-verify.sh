#!/bin/bash
# =============================================================================
# Media Server - Verification Script
# Run after setup to verify all components
# =============================================================================

echo "=========================================="
echo "  Media Server Verification"
echo "=========================================="

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

cd ~/mediaserver 2>/dev/null || cd /home/*/mediaserver 2>/dev/null

check_service() {
    local name=$1
    local port=$2
    local timeout=3
    
    if curl -s --max-time $timeout -o /dev/null -w "%{http_code}" "http://localhost:$port" 2>/dev/null | grep -qE "200|302|401|403"; then
        echo -e "${GREEN}✓${NC} $name (port $port)"
        return 0
    else
        echo -e "${RED}✗${NC} $name (port $port)"
        return 1
    fi
}

echo ""
echo -e "${YELLOW}[1/5] Docker containers status...${NC}"
echo ""
docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Health}}" 2>/dev/null || docker compose ps

echo ""
echo -e "${YELLOW}[2/5] VPN connection check...${NC}"
VPN_IP=$(docker exec gluetun wget -qO- https://api.ipify.org 2>/dev/null || echo "FAILED")
LOCAL_IP=$(wget -qO- https://api.ipify.org 2>/dev/null || curl -s https://api.ipify.org 2>/dev/null || echo "UNKNOWN")

if [ "$VPN_IP" != "FAILED" ] && [ "$VPN_IP" != "" ]; then
    if [ "$VPN_IP" != "$LOCAL_IP" ]; then
        echo -e "${GREEN}✓${NC} VPN is working"
        echo "  VPN IP:  $VPN_IP"
        echo "  Real IP: $LOCAL_IP"
    else
        echo -e "${RED}✗${NC} VPN IP matches real IP - VPN may not be working"
    fi
else
    echo -e "${RED}✗${NC} VPN check failed - Gluetun may be starting or misconfigured"
    echo ""
    echo "  Check logs: docker logs gluetun --tail 20"
fi

echo ""
echo -e "${YELLOW}[3/5] GPU access...${NC}"
if docker exec jellyfin nvidia-smi &>/dev/null; then
    echo -e "${GREEN}✓${NC} Jellyfin has GPU access"
else
    echo -e "${YELLOW}~${NC} Jellyfin - cannot verify GPU (container may be initializing)"
fi

if docker exec unmanic nvidia-smi &>/dev/null; then
    echo -e "${GREEN}✓${NC} Unmanic has GPU access"
else
    echo -e "${YELLOW}~${NC} Unmanic - cannot verify GPU (container may be initializing)"
fi

echo ""
echo -e "${YELLOW}[4/5] Service endpoints...${NC}"
check_service "qBittorrent" 8080
check_service "Prowlarr" 9696
check_service "Radarr" 7878
check_service "Sonarr" 8989
check_service "Bazarr" 6767
check_service "Jellyfin" 8096
check_service "Jellyseerr" 5055
check_service "Unmanic" 8888
check_service "Authelia" 9091
check_service "NPM" 81

echo ""
echo -e "${YELLOW}[5/5] Wireguard status...${NC}"
if command -v wg &> /dev/null; then
    sudo wg show 2>/dev/null || echo "Run with sudo to see Wireguard status"
else
    echo "Wireguard not installed on this machine"
fi

echo ""
echo "=========================================="
echo "  Verification Complete"
echo "=========================================="
echo ""
echo -e "${YELLOW}Troubleshooting:${NC}"
echo "  View logs:     docker logs <container_name>"
echo "  Restart stack: docker compose restart"
echo "  Recreate:      docker compose up -d --force-recreate"
echo ""
