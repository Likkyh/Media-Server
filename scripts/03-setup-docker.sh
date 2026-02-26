#!/bin/bash
# =============================================================================
# Media Server Setup Script - Phase 3: Docker Stack
# Run this script after 02-setup-wireguard.sh
# =============================================================================

set -e

echo "=========================================="
echo "  Media Server Setup - Phase 3"
echo "  Docker Stack Deployment"
echo "=========================================="

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

MEDIASERVER_DIR="$HOME/mediaserver"

echo -e "${YELLOW}[1/6] Creating directory structure...${NC}"
mkdir -p "$MEDIASERVER_DIR"
cd "$MEDIASERVER_DIR"

mkdir -p config/{gluetun,qbittorrent,prowlarr,radarr,sonarr,jellyfin,jellyseerr,bazarr,sabnzbd,unmanic,npm/data,npm/letsencrypt}

echo -e "${YELLOW}[2/5] Checking for .env file...${NC}"
if [ ! -f ".env" ]; then
    echo -e "${RED}ERROR: .env file not found!${NC}"
    echo ""
    echo "Create .env with at least:"
    echo "  WIREGUARD_PRIVATE_KEY=your_protonvpn_wireguard_key"
    echo ""
    echo "Get your WireGuard key from: https://account.protonvpn.com/downloads#wireguard-configuration"
    echo ""
    echo "Optional (for the dashboard):"
    echo "  JELLYFIN_API_KEY=..."
    echo "  SEERR_API_KEY=..."
    echo "  RADARR_API_KEY=..."
    echo "  SONARR_API_KEY=..."
    echo "  PROWLARR_API_KEY=..."
    echo "  BAZARR_API_KEY=..."
    echo "  QBIT_USERNAME=..."
    echo "  QBIT_PASSWORD=..."
    echo "  SABNZBD_API_KEY=..."
    echo "  DASHBOARD_USER=..."
    echo "  DASHBOARD_PASS=..."
    exit 1
fi

# Validate .env has real credentials
if grep -q "your_protonvpn_wireguard_key" .env; then
    echo -e "${RED}ERROR: .env still contains placeholder values!${NC}"
    echo "Edit .env and add your real WireGuard private key."
    exit 1
fi

echo -e "${YELLOW}[3/5] Checking for compose.yml...${NC}"
if [ ! -f "compose.yml" ]; then
    echo -e "${RED}ERROR: compose.yml not found!${NC}"
    echo "Please copy compose.yml to $MEDIASERVER_DIR"
    exit 1
fi

echo -e "${YELLOW}[4/5] Pulling Docker images...${NC}"
docker compose pull

echo -e "${YELLOW}[5/5] Starting Docker stack...${NC}"
docker compose up -d

echo ""
echo -e "${GREEN}=========================================="
echo "  Docker Stack Deployed!"
echo "==========================================${NC}"
echo ""
echo "Waiting for services to initialize..."
sleep 15

echo ""
echo "Service Status:"
docker compose ps --format "table {{.Name}}\t{{.Status}}"

echo ""
echo -e "${GREEN}Access your services:${NC}"
echo ""
echo "  qBittorrent:    http://$(hostname -I | awk '{print $1}'):8080"
echo "  Prowlarr:       http://$(hostname -I | awk '{print $1}'):9696"
echo "  Radarr:         http://$(hostname -I | awk '{print $1}'):7878"
echo "  Sonarr:         http://$(hostname -I | awk '{print $1}'):8989"
echo "  Bazarr:         http://$(hostname -I | awk '{print $1}'):6767"
echo "  Jellyfin:       http://$(hostname -I | awk '{print $1}'):8096"
echo "  Jellyseerr:     http://$(hostname -I | awk '{print $1}'):5055"
echo "  Unmanic:        http://$(hostname -I | awk '{print $1}'):8888"
echo "  NPM Admin:      http://$(hostname -I | awk '{print $1}'):81"
echo ""
echo -e "${YELLOW}Default credentials:${NC}"
echo "  qBittorrent: admin / (check: docker logs qbittorrent 2>&1 | grep password)"
echo "  NPM: admin@example.com / changeme"
echo ""
echo -e "${YELLOW}Run verification:${NC} ./scripts/04-verify.sh"
echo ""
