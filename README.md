# Media Server

Serveur multimédia personnel avec streaming, téléchargement automatique et transcodage GPU.

## Architecture

```
Internet → Cloudflare → Nginx Proxy Manager
                                ↓
                ┌───────────────┼───────────────┐
                ↓               ↓               ↓
            Jellyfin         Seerr            (admin)
                ↑               ↓
            Unmanic      Radarr / Sonarr / Bazarr
            (GPU)               ↓
                ↓           Prowlarr
            /mnt/media          ↓
                        ┌───────┴────────┐
                        ↓               ↓
                   qBittorrent      SABnzbd
                        ↑
                   Gluetun VPN
                   + FlareSolverr
```

## Stack

| Service | Port | Description |
|---------|------|-------------|
| Jellyfin | 8096 | Streaming (GPU) |
| Seerr | 5055 | Requêtes utilisateurs |
| Radarr | 7878 | Automatisation films |
| Sonarr | 8989 | Automatisation séries |
| Bazarr | 6767 | Sous-titres |
| Prowlarr | 9696 | Gestionnaire indexeurs |
| FlareSolverr | 8191 | Bypass Cloudflare (via VPN) |
| qBittorrent | 8080 | Torrent (via VPN) |
| SABnzbd | 8085 | Usenet |
| Unmanic | 8888 | Transcodage automatique (GPU) |
| Nginx Proxy Manager | 80/443/81 | Reverse proxy |

> **Note réseau :** qBittorrent et FlareSolverr partagent le namespace réseau de Gluetun.
> Les ports 8080 et 8191 ne sont pas exposés sur l'hôte — accessibles uniquement via NPM.

## Sécurité

- **Accès local** : tous les services internes sont derrière `*.local.example.com` avec certificats Let's Encrypt (wildcard DNS-01 via Cloudflare) et access list NPM restreinte au LAN
- **Accès public** : seuls Jellyfin (`stream.example.com`) et Seerr (`request.example.com`) sont exposés, derrière Cloudflare
- **VPN** : tout le trafic torrent passe par Gluetun (ProtonVPN WireGuard) — qBittorrent ne peut pas fuiter l'IP réelle
- **SSH Proxmox** : authentification par clé uniquement, password auth désactivé
- **Firewall Proxmox** : seuls les ports 22, 443 et 8006 sont ouverts depuis le LAN
- **DNS local** : Pi-hole résout `*.local.example.com` vers le serveur NPM ou nginx Proxmox
- **Certificats Proxmox/Pi-hole** : gérés directement sur l'hôte Proxmox via acme.sh + Cloudflare DNS-01, renouvellement automatique avec déploiement dans le LXC

## Prérequis

- Ubuntu 22.04 (VM ou bare metal)
- GPU NVIDIA (pour le transcodage Jellyfin/Unmanic)
- Un disque dédié pour le stockage média (monté sur `/mnt/media`)
- Un compte ProtonVPN avec accès WireGuard

## Installation

### 1. Cloner le dépôt sur le serveur

```bash
ssh user@<server-ip>
git clone <repo-url> ~/mediaserver
cd ~/mediaserver
```

### 2. Préparer le système

```bash
cd ~/mediaserver/scripts
chmod +x *.sh
sudo ./01-setup-system.sh
```

Installe les paquets système, les drivers NVIDIA, Docker, le NVIDIA Container Toolkit, configure le disque média (`/mnt/media`) et le firewall.

Redémarrez après cette étape :

```bash
sudo reboot
```

### 3. Configurer WireGuard (accès distant)

```bash
sudo ./02-setup-wireguard.sh
```

Génère les clés serveur/client et active le tunnel WireGuard pour l'administration à distance.

### 4. Créer le fichier `.env`

```bash
cd ~/mediaserver
cp .env.example .env   # ou créer manuellement
nano .env
```

Variables requises :

```
WIREGUARD_PRIVATE_KEY=<votre clé privée ProtonVPN WireGuard>
```

Variables optionnelles (pour le dashboard Arcticmon, configurables après le premier démarrage) :

```
JELLYFIN_API_KEY=
SEERR_API_KEY=
RADARR_API_KEY=
SONARR_API_KEY=
PROWLARR_API_KEY=
BAZARR_API_KEY=
QBIT_USERNAME=
QBIT_PASSWORD=
SABNZBD_API_KEY=
DASHBOARD_USER=
DASHBOARD_PASS=
```

### 5. Démarrer la stack

```bash
./03-setup-docker.sh
```

Crée l'arborescence `config/`, pull les images et lance tous les containers.

### 6. Vérifier

```bash
./04-verify.sh
```

Vérifie que tous les services répondent, que le VPN fonctionne et que le GPU est accessible.

### 7. Configurer les services

Suivez le [Guide de Configuration](CONFIGURATION_GUIDE.md) pour paramétrer chaque service (Prowlarr, Radarr, Sonarr, etc.).

## Mise à jour

Le dépôt contient la configuration de la stack (`compose.yml`, `scripts/`, `dashboard/`). Les données des services sont dans `config/` et `.env`, qui ne sont pas versionnés.

Pour mettre à jour la stack depuis le dépôt :

```bash
cd ~/mediaserver

# Récupérer les dernières modifications
git pull

# Reconstruire le dashboard si modifié
docker compose build arcticmon

# Mettre à jour les images et redémarrer
docker compose pull
docker compose up -d
```

`docker compose up -d` ne recrée que les containers dont l'image ou la configuration a changé. Les volumes `config/` et `.env` sont préservés — aucune donnée de service n'est perdue.

> **Note :** si `compose.yml` ajoute un nouveau service avec un dossier `config/` manquant, Docker le créera automatiquement au démarrage.

## Structure

```
mediaserver/
├── compose.yml             # Stack Docker
├── .env.example            # Template des variables d'environnement
├── .env                    # Credentials (non versionné)
├── .gitignore
├── dashboard/              # Source du dashboard Arcticmon
├── config/                 # Données des services (non versionné)
│   ├── gluetun/
│   ├── qbittorrent/
│   ├── sabnzbd/
│   ├── prowlarr/
│   ├── radarr/
│   ├── sonarr/
│   ├── bazarr/
│   ├── jellyfin/
│   ├── jellyseerr/
│   ├── unmanic/
│   └── npm/
├── scripts/                # Scripts d'installation
└── wireguard/              # Templates WireGuard admin
```
