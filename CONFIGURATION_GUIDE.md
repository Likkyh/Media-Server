# Guide de Configuration - Media Server

Guide pas à pas pour configurer chaque service de la stack. Suivez l'ordre indiqué pour une installation réussie.

---

## Prérequis

Avant de commencer, vérifiez que tous les containers sont en cours d'exécution :

```bash
docker compose ps
```

Tous les containers doivent être `Up` et `gluetun` doit être `(healthy)`.

**Règle UFW requise** (permet la communication entre containers) :
```bash
sudo ufw allow from 172.16.0.0/12 to any
sudo ufw reload
```

---

## 1. qBittorrent

**URL:** `http://<server-ip>:8080`

### Connexion initiale

```bash
# Récupérer le mot de passe temporaire
docker logs qbittorrent 2>&1 | grep -i password
```
- User: `admin`
- Pass: (voir logs, commence par `The WebUI administrator password was not set...`)

### Paramètres essentiels

**⚙️ Options → Downloads**

| Paramètre | Valeur |
|-----------|--------|
| Default Save Path | `/downloads` |
| Keep incomplete torrents in | ✅ Activé, `/downloads/incomplete` |

**⚙️ Options → Connection**

| Paramètre | Valeur |
|-----------|--------|
| Listening Port | `6881` |
| Use UPnP / NAT-PMP | ❌ Désactivé |

**⚙️ Options → BitTorrent**

| Paramètre | Valeur |
|-----------|--------|
| Enable DHT | ✅ |
| Enable PeX | ✅ |
| Seeding Limits → Ratio | `2.0` |
| Seeding Limits → Minutes | `1440` (24h) |
| When ratio reached | `Pause torrent` |

**⚙️ Options → Web UI**

| Paramètre | Valeur |
|-----------|--------|
| Username | `admin` |
| Password | *(Changez-le!)* |
| Bypass auth for localhost | ❌ Désactivé |

**⚙️ Options → Advanced**

| Paramètre | Valeur |
|-----------|--------|
| Network interface | `tun0` |
| Optional IP to bind | *(laisser vide)* |

> [!CAUTION]
> **Le paramètre `Network interface = tun0` est OBLIGATOIRE** - il force tout le trafic torrent à passer par le VPN. Sans cela, votre vraie IP sera exposée !

**Après configuration, cliquez "Save" en bas de page.**

---

## 2. SABnzbd (Usenet)

**URL:** `http://<server-ip>:8085`

SABnzbd télécharge les fichiers depuis Usenet (alternative aux torrents).

### Configuration initiale

1. Ouvrez l'interface web
2. Suivez l'assistant de configuration

### Configurer le serveur Usenet

**Config → Servers → Add Server**

| Champ | Valeur |
|-------|--------|
| Host | *(fourni par votre provider, ex: news.abnzb.com)* |
| Port | `563` (SSL) |
| SSL | ✅ |
| Username | *(votre username)* |
| Password | *(votre password)* |
| Connections | `10-20` |

Cliquez **Test Server** puis **Add Server**.

### Récupérer la clé API

**Config → General → Security**

Copiez la **API Key** - vous en aurez besoin pour Radarr/Sonarr.

### Dossiers

**Config → Folders**

| Champ | Valeur |
|-------|--------|
| Temporary Download Folder | `/incomplete-downloads` |
| Completed Download Folder | `/downloads` |

---

## 3. Jellyfin (Streaming)

**URL:** `http://<server-ip>:8096`

### Setup initial (wizard)

1. Langue: **Français**
2. Créer un compte admin (notez le mot de passe !)
3. **NE PAS ajouter de bibliothèques maintenant** → "Suivant"
4. Terminer l'assistant

### Ajouter les bibliothèques

**Tableau de bord → Bibliothèques → Ajouter une bibliothèque**

#### Films

**Onglet principal :**
| Champ | Valeur |
|-------|--------|
| Type de contenu | `Films` |
| Nom d'affichage | `Films` |
| Dossiers | Cliquer ➕ → entrer `/data/movies` → OK |
| Langue de métadonnée préférée | `French` |
| Pays | `France` |

**Récupérateurs de métadonnées (cocher) :**
- ✅ TheMovieDb
- ✅ The Open Movie Database

**Récupérateurs d'images (cocher) :**
- ✅ TheMovieDb

**Options supplémentaires :**
| Option | Valeur |
|--------|--------|
| Enregistrer les illustrations dans les dossiers des médias | ❌ Non recommandé |
| Enregistrer les métadonnées NFO | ❌ Non (sauf si vous utilisez Kodi) |

#### Séries

**Onglet principal :**
| Champ | Valeur |
|-------|--------|
| Type de contenu | `Émissions de TV` |
| Nom d'affichage | `Séries` |
| Dossiers | Cliquer ➕ → entrer `/data/tv` → OK |
| Langue de métadonnée préférée | `French` |
| Pays | `France` |

#### Anime

**Onglet principal :**
| Champ | Valeur |
|-------|--------|
| Type de contenu | `Émissions de TV` |
| Nom d'affichage | `Anime` |
| Dossiers | Cliquer ➕ → entrer `/data/anime` → OK |
| Langue de métadonnée préférée | `French` |
| Pays | `France` |

**Récupérateurs de métadonnées :**
- ✅ TheMovieDb

### Activer le transcodage GPU (NVIDIA)

**Tableau de bord → Lecture → Transcodage**

| Paramètre | Valeur |
|-----------|--------|
| Accélération matérielle | `NVIDIA NVENC` |
| Activer le décodage matériel | ✅ |
| Activer l'encodage matériel | ✅ |

**Décodeurs matériels à activer (RTX 3060) :**

| Codec | Activer | Notes |
|-------|---------|-------|
| H264 | ✅ | Le plus courant |
| HEVC | ✅ | H.265, très répandu |
| HEVC 10bit | ✅ | HDR, haute qualité |
| VP9 | ✅ | YouTube, contenus web |
| VP9 10bit | ✅ | YouTube HDR |
| AV1 | ✅ | Netflix, YouTube récent |
| MPEG2 | ✅ | DVD, TV enregistrée |
| MPEG4 | ✅ | Anciens fichiers AVI |
| VC1 | ✅ | Blu-ray anciens |
| VP8 | ✅ | Optionnel mais supporté |
| HEVC RExt 8/10bit | ✅ | Rare mais supporté |
| HEVC RExt 12bit | ❌ | **Non supporté RTX 3060** |

> [!TIP]
> **Cochez tous les codecs sauf HEVC RExt 12bit**. Si un codec n'est pas coché, le CPU sera utilisé (plus lent, plus de consommation).

**Vérifier l'accès GPU :**
```bash
docker exec jellyfin nvidia-smi
```

### Créer des utilisateurs

**Tableau de bord → Utilisateurs → Ajouter utilisateur**

Pour chaque membre de la famille :
- Nom d'utilisateur
- Mot de passe
- Cocher les bibliothèques (Films, Séries)
- ❌ Autoriser la suppression de médias (sauf admin)

---

## 4. Prowlarr (Indexeurs)

**URL:** `http://<server-ip>:9696`

### Configuration initiale

1. **Settings → General → Security**
   - Authentication: `Forms (Login Page)`
   - Username: `admin`
   - Password: *(choisir un mot de passe)*
   - Cliquer **Save**

2. **Settings → UI**
   - ❌ Enable Analytics
   - **Save**

### Récupérer la clé API

**Settings → General → API Key**
- Copiez cette clé, vous en aurez besoin pour Radarr/Sonarr

### Configurer FlareSolverr (Anti-Cloudflare)

FlareSolverr passe par le VPN pour contourner les protections Cloudflare.

**Settings → Indexers (dans le menu gauche) → + → FlareSolverr**

| Champ | Valeur |
|-------|--------|
| Name | `FlareSolverr` |
| Tags | `flaresolverr` |
| Host | `http://<server-ip>:8191` |
| Request Timeout | `60` |

Cliquez **Test** puis **Save**.

> [!WARNING]
> Si le test échoue avec "Connection refused", exécutez sur le serveur :
> ```bash
> sudo ufw allow from 172.16.0.0/12 to any
> ```

### Ajouter des indexeurs

**Indexers → + → Rechercher l'indexeur**

Indexeurs recommandés :
| Indexeur | Type | Notes |
|----------|------|-------|
| `YTS` | Films | Fonctionne bien, petits fichiers |
| `EZTV` | Séries | Fiable |
| `LimeTorrents` | Général | Bon fallback |
| `Nyaa` | Anime | Si vous regardez des animes |
| `1337x` | Général | Nécessite tag `flaresolverr` |
| `YggTorrent` | FR | Nécessite un compte |

> [!NOTE]
> Pour les indexeurs avec erreur Cloudflare, ajoutez le tag `flaresolverr` dans leurs paramètres.

---

## 5. Radarr (Films)

**URL:** `http://<server-ip>:7878`

### Configuration initiale

**Settings → General → Security**
- Authentication: `Forms (Login Page)`
- Username/Password
- **Save**

### Récupérer la clé API

**Settings → General → API Key** → Copiez-la

### Configurer le dossier racine

**Settings → Media Management**

| Paramètre | Valeur |
|-----------|--------|
| Rename Movies | ✅ |
| Replace Illegal Characters | ✅ |
| Standard Movie Format | `{Movie Title} ({Release Year})` |

**Root Folders → Add Root Folder**
- Path: `/movies`
- Cliquez **OK**

### Client de téléchargement

**Settings → Download Clients → + → qBittorrent**

| Champ | Valeur |
|-------|--------|
| Name | `qBittorrent` |
| Host | `gluetun` |
| Port | `8080` |
| Username | `admin` |
| Password | *(votre mot de passe qBittorrent)* |
| Category | `radarr` |
| Remove Completed | ✅ |

Cliquez **Test** puis **Save**.

> [!IMPORTANT]
> Le host est `gluetun` et non l'IP car qBittorrent utilise le réseau VPN.

**Settings → Download Clients → + → SABnzbd**

| Champ | Valeur |
|-------|--------|
| Name | `SABnzbd` |
| Host | `sabnzbd` |
| Port | `8080` |
| API Key | *(clé API de SABnzbd)* |
| Category | `radarr` |

Cliquez **Test** puis **Save**.

### Profils de qualité

**Settings → Profiles**

Éditez le profil `HD-1080p` ou créez-en un :

| Qualité | Activé |
|---------|--------|
| Bluray-1080p | ✅ |
| WEBDL-1080p | ✅ |
| WEBRip-1080p | ✅ |
| Bluray-720p | ✅ (fallback) |
| Remux-1080p | ❌ (trop gros) |
| Raw-HD | ❌ |

---

## 6. Sonarr (Séries)

**URL:** `http://<server-ip>:8989`

### Configuration initiale

**Settings → General → Security**
- Authentication: `Forms (Login Page)`
- Username/Password
- **Save**

### Récupérer la clé API

**Settings → General → API Key** → Copiez-la

### Configurer le dossier racine

**Settings → Media Management**

| Paramètre | Valeur |
|-----------|--------|
| Rename Episodes | ✅ |
| Season Folder Format | `Season {season:00}` |
| Standard Episode Format | `{Series Title} - S{season:00}E{episode:00} - {Episode Title}` |

**Root Folders → Add Root Folder**
- `/tv` (pour les séries)
- `/anime` (pour les animes)

### Client de téléchargement

**Settings → Download Clients → + → qBittorrent**

| Champ | Valeur |
|-------|--------|
| Name | `qBittorrent` |
| Host | `gluetun` |
| Port | `8080` |
| Username | `admin` |
| Password | *(votre mot de passe qBittorrent)* |
| Category | `sonarr` |
| Remove Completed | ✅ |

**Settings → Download Clients → + → SABnzbd**

| Champ | Valeur |
|-------|--------|
| Name | `SABnzbd` |
| Host | `sabnzbd` |
| Port | `8080` |
| API Key | *(clé API de SABnzbd)* |
| Category | `sonarr` |

Cliquez **Test** puis **Save**.

---

## 7. Prowlarr → Connexion Apps

**Retournez dans Prowlarr : `http://<server-ip>:9696`**

### Connecter Radarr

**Settings → Apps → + → Radarr**

| Champ | Valeur |
|-------|--------|
| Name | `Radarr` |
| Sync Level | `Full Sync` |
| Prowlarr Server | `http://prowlarr:9696` |
| Radarr Server | `http://radarr:7878` |
| API Key | *(clé API de Radarr)* |

Cliquez **Test** puis **Save**.

### Connecter Sonarr

**Settings → Apps → + → Sonarr**

| Champ | Valeur |
|-------|--------|
| Name | `Sonarr` |
| Sync Level | `Full Sync` |
| Prowlarr Server | `http://prowlarr:9696` |
| Sonarr Server | `http://sonarr:8989` |
| API Key | *(clé API de Sonarr)* |

Cliquez **Test** puis **Save**.

> [!NOTE]
> Après avoir ajouté les apps, les indexeurs de Prowlarr seront automatiquement synchronisés avec Radarr et Sonarr.

---

## 8. Bazarr (Sous-titres)

**URL:** `http://<server-ip>:6767`

### Connexion à Sonarr

**Settings → Sonarr**

| Champ | Valeur |
|-------|--------|
| Use Sonarr | ✅ |
| Hostname or IP | `sonarr` |
| Port | `8989` |
| API Key | *(clé API de Sonarr)* |
| Download Only Monitored | ✅ |

Cliquez **Test** puis **Save**.

### Connexion à Radarr

**Settings → Radarr**

| Champ | Valeur |
|-------|--------|
| Use Radarr | ✅ |
| Hostname or IP | `radarr` |
| Port | `7878` |
| API Key | *(clé API de Radarr)* |
| Download Only Monitored | ✅ |

### Configurer les langues

**Settings → Languages**

| Champ | Valeur |
|-------|--------|
| Languages Filter | `French, English` |
| Subtitles Languages | Cocher `French` et `English` |
| Default enabled | French en premier |

### Fournisseurs de sous-titres

**Settings → Providers**

Cliquez sur les providers à activer :

| Provider | Configuration |
|----------|---------------|
| OpenSubtitles.com | Créer un compte gratuit, entrer username/password |
| Subscene | Aucune config requise |
| Addic7ed | Aucune config requise |

> [!TIP]
> OpenSubtitles.com offre plus de résultats avec un compte (gratuit).

---

## 9. Seerr (Requêtes)

**URL:** `http://<server-ip>:5055`

### Setup initial

1. Cliquez **Sign in with Jellyfin**
2. Jellyfin URL: `http://jellyfin:8096`
3. Entrez vos identifiants admin Jellyfin
4. Cliquez **Sign In**

### Configuration Jellyfin

| Champ | Valeur |
|-------|--------|
| Internal URL | `http://jellyfin:8096` |
| External URL | `https://jellyfin.example.com` *(ou vide)* |

Cliquez **Sync Libraries** et sélectionnez Films + Séries.

### Connecter Radarr

**Settings → Services → Radarr → Add Radarr Server**

| Champ | Valeur |
|-------|--------|
| Default Server | ✅ |
| 4K Server | ❌ |
| Server Name | `Radarr` |
| Hostname or IP | `radarr` |
| Port | `7878` |
| Use SSL | ❌ |
| API Key | *(clé API de Radarr)* |

Cliquez **Test**, puis sélectionnez :
- Quality Profile: `HD-1080p`
- Root Folder: `/movies`
- Minimum Availability: `Released`

**Save Changes**

### Connecter Sonarr

**Settings → Services → Sonarr → Add Sonarr Server**

| Champ | Valeur |
|-------|--------|
| Default Server | ✅ |
| 4K Server | ❌ |
| Server Name | `Sonarr` |
| Hostname or IP | `sonarr` |
| Port | `8989` |
| API Key | *(clé API de Sonarr)* |

Sélectionnez :
- Quality Profile: `HD-1080p`
- Root Folder: `/tv`
- Language Profile: `Deprecated` ou le profil disponible

**Save Changes**

### Importer les utilisateurs

**Settings → Users → Import Jellyfin Users**

Les utilisateurs Jellyfin apparaîtront. Définissez leurs permissions de requête.

---

## 10. Unmanic (Transcodage automatique)

**URL:** `http://<server-ip>:8888`

Unmanic transcode automatiquement vos fichiers en H.265 (HEVC) pour économiser de l'espace.

> [!NOTE]
> La version gratuite est limitée à 2 bibliothèques. Solution : utiliser une seule bibliothèque `/library` qui contient movies, tv, anime.

### 1. Configurer la bibliothèque

**Settings → Library**

Modifiez la bibliothèque par défaut existante :

| Champ | Valeur |
|-------|--------|
| Path | `/library` |

> [!TIP]
> `/library` contient automatiquement movies, tv et anime grâce aux volumes Docker.

### 2. Installer le plugin de transcodage

**Settings → Plugins**

1. Cherchez **"Transcode video files"**
2. Cliquez pour l'installer

### 3. Configurer le plugin

Après installation, cliquez sur le plugin pour le configurer :

| Paramètre | Valeur |
|-----------|--------|
| Video Codec | `HEVC/H265` |
| Video Encoder | `NVENC - hevc_nvenc` |
| NVIDIA Device | `NVIDIA GeForce RTX 3060...` |
| Enable HW Decoding | `NVDEC/CUDA - Use the GPUs HW decoding...` |

> [!TIP]
> NVDEC/CUDA permet un transcodage 100% GPU (décodage + encodage).

### 4. Ajouter au Plugin Flow

**Settings → Plugin Flow**

Vérifiez que **"Transcode video files"** apparaît dans la liste. Sinon, ajoutez-le avec **+**.

### 5. Configurer les Workers

**Settings → Workers**

| Paramètre | Valeur |
|-----------|--------|
| Number of Workers | `1` |

> [!NOTE]
> 1 seul worker car le GPU encode efficacement un fichier à la fois.

### 6. Planifier les heures de travail (optionnel)

**Settings → Schedule**

Pour ne transcoder qu'en heures creuses :

| Champ | Valeur |
|-------|--------|
| Heures actives | `22:00 → 06:00` |

Unmanic mettra les workers en pause en dehors de ces heures.

---

## 11. Nginx Proxy Manager (Accès externe)

**URL:** `http://<server-ip>:81`

### Connexion initiale

- Email: `admin@example.com`
- Password: `changeme`

→ **Changez immédiatement le mot de passe !**

### Configuration DNS Cloudflare

Dans **Cloudflare Dashboard** → votre domaine → DNS :

| Type | Nom | Contenu | Proxy |
|------|-----|---------|-------|
| A | `jellyfin` | *Votre IP publique* | ✅ Proxied |
| A | `request` | *Votre IP publique* | ✅ Proxied |

### Redirection des ports (routeur)

Configurez la redirection de ports sur votre box/routeur :

| Port externe | → | Destination |
|--------------|---|-------------|
| 80/TCP | → | <server-ip>:80 |
| 443/TCP | → | <server-ip>:443 |

### Créer Proxy Host - Jellyfin

**Hosts → Proxy Hosts → Add Proxy Host**

**Details :**
| Champ | Valeur |
|-------|--------|
| Domain Names | `jellyfin.example.com` |
| Scheme | `http` |
| Forward Hostname/IP | `jellyfin` |
| Forward Port | `8096` |
| Websockets Support | ✅ |

**SSL :**
- SSL Certificate: `Request a new SSL Certificate`
- ✅ Force SSL
- ✅ HTTP/2 Support
- Email: *(votre email)*
- ✅ I Agree

### Créer Proxy Host - Seerr

**Hosts → Proxy Hosts → Add Proxy Host**

| Champ | Valeur |
|-------|--------|
| Domain Names | `request.example.com` |
| Scheme | `http` |
| Forward Hostname/IP | `seerr` |
| Forward Port | `5055` |
| Websockets Support | ✅ |

SSL comme précédemment.

---

## Test Final

### 1. Vérifier le VPN
```bash
# IP VPN (doit être différente de votre vraie IP)
docker exec gluetun wget -qO- https://api.ipify.org && echo
```

### 2. Tester une requête complète

1. Ouvrez **Seerr** (`http://<server-ip>:5055`)
2. Recherchez un film populaire (ex: "Inception")
3. Cliquez **Request**
4. Vérifiez dans **Radarr** que le film est ajouté
5. Vérifiez dans **qBittorrent** que le téléchargement démarre
6. Une fois terminé, vérifiez dans **Jellyfin** que le film apparaît
7. **Bazarr** téléchargera les sous-titres automatiquement

### 3. Tester l'accès externe

1. Désactivez le WiFi sur votre téléphone (utilisez 4G)
2. Accédez à `https://jellyfin.example.com`
3. Connectez-vous avec vos identifiants

---

## Dépannage

### Container qui ne démarre pas
```bash
docker logs <nom_container> --tail 50
```

### Redémarrer un service
```bash
docker compose restart <nom_service>
```

### Recreer tous les containers
```bash
docker compose down
docker compose up -d
```

### Vérifier l'espace disque
```bash
df -h
```

### Voir l'utilisation des ressources
```bash
docker stats
```

---

🎉 **Votre media server est maintenant opérationnel !**
