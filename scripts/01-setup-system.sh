#!/bin/bash
# =============================================================================
# Media Server Setup Script - Phase 1: System Configuration
# Run this script on the freshly installed Ubuntu 22.04 VM
# =============================================================================

set -e

echo "=========================================="
echo "  Media Server Setup - Phase 1"
echo "  System Configuration"
echo "=========================================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root or with sudo${NC}"
    exit 1
fi

# Configuration
MEDIA_DISK="/dev/sdb"  # Change if your HDD is on a different device
MEDIA_MOUNT="/mnt/media"

echo -e "${YELLOW}[1/8] Updating system...${NC}"
apt update && apt upgrade -y

echo -e "${YELLOW}[2/8] Installing essential packages...${NC}"
apt install -y \
    curl \
    wget \
    git \
    htop \
    net-tools \
    ufw \
    fail2ban \
    wireguard \
    ca-certificates \
    gnupg

echo -e "${YELLOW}[3/8] Installing NVIDIA drivers...${NC}"
apt install -y nvidia-driver-535 nvidia-utils-535

echo -e "${YELLOW}[4/8] Installing Docker...${NC}"
# Remove old versions
apt remove -y docker docker-engine docker.io containerd runc 2>/dev/null || true

# Add Docker repo
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null

apt update
apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

echo -e "${YELLOW}[5/8] Installing NVIDIA Container Toolkit...${NC}"
# New repository format (2024+)
curl -fsSL https://nvidia.github.io/libnvidia-container/gpgkey | gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg

curl -s -L https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list | \
  sed 's#deb https://#deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://#g' | \
  tee /etc/apt/sources.list.d/nvidia-container-toolkit.list

apt update
apt install -y nvidia-container-toolkit
nvidia-ctk runtime configure --runtime=docker
systemctl restart docker

echo -e "${YELLOW}[6/8] Configuring media storage...${NC}"
if [ -b "$MEDIA_DISK" ]; then
    # Check if already formatted
    if ! blkid "$MEDIA_DISK" | grep -q ext4; then
        echo -e "${YELLOW}Formatting $MEDIA_DISK...${NC}"
        mkfs.ext4 "$MEDIA_DISK"
    fi
    
    mkdir -p "$MEDIA_MOUNT"
    
    # Add to fstab if not already present
    if ! grep -q "$MEDIA_MOUNT" /etc/fstab; then
        UUID=$(blkid -s UUID -o value "$MEDIA_DISK")
        echo "UUID=$UUID $MEDIA_MOUNT ext4 defaults 0 2" >> /etc/fstab
    fi
    
    mount -a
    
    # Create directory structure
    mkdir -p "$MEDIA_MOUNT"/{downloads/{incomplete,complete},movies,tv,music,transcode}
    chown -R 1000:1000 "$MEDIA_MOUNT"
    chmod -R 755 "$MEDIA_MOUNT"
    
    echo -e "${GREEN}Media storage configured at $MEDIA_MOUNT${NC}"
else
    echo -e "${RED}Warning: $MEDIA_DISK not found. Skipping media disk setup.${NC}"
fi

echo -e "${YELLOW}[7/8] Configuring firewall...${NC}"
ufw default deny incoming
ufw default allow outgoing

# SSH from local network
ufw allow from 192.168.1.0/24 to any port 22

# Wireguard
ufw allow 51820/udp

# HTTP/HTTPS for reverse proxy
ufw allow 80/tcp
ufw allow 443/tcp

# Enable firewall
echo "y" | ufw enable

echo -e "${YELLOW}[8/8] Configuring IP forwarding...${NC}"
echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf
sysctl -p

# Add current user to docker group
SUDO_USER_NAME=${SUDO_USER:-$USER}
usermod -aG docker "$SUDO_USER_NAME"

echo ""
echo -e "${GREEN}=========================================="
echo "  Phase 1 Complete!"
echo "==========================================${NC}"
echo ""
echo "Next steps:"
echo "1. Reboot the system: sudo reboot"
echo "2. After reboot, verify GPU: nvidia-smi"
echo "3. Run: ./02-setup-wireguard.sh"
echo ""
