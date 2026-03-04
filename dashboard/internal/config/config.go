package config

import "os"

type Config struct {
	ListenAddr string

	JellyfinURL    string
	JellyfinAPIKey string

	QbitURL      string
	QbitUsername  string
	QbitPassword string

	RadarrURL    string
	RadarrAPIKey string

	SonarrURL    string
	SonarrAPIKey string

	SeerrURL    string
	SeerrAPIKey string

	ProwlarrURL    string
	ProwlarrAPIKey string

	BazarrURL    string
	BazarrAPIKey string

	SabnzbdURL    string
	SabnzbdAPIKey string

	UnmanicURL string

	DockerSocket string
	HostProcPath string
	HostUtmpPath string

	DashboardUser string
	DashboardPass string

	PiholeURL      string
	PiholePassword string
}

func Load() *Config {
	return &Config{
		ListenAddr: envOr("LISTEN_ADDR", ":3000"),

		JellyfinURL:    envOr("JELLYFIN_URL", "http://jellyfin:8096"),
		JellyfinAPIKey: os.Getenv("JELLYFIN_API_KEY"),

		QbitURL:      envOr("QBIT_URL", "http://gluetun:8080"),
		QbitUsername:  envOr("QBIT_USERNAME", "admin"),
		QbitPassword:  os.Getenv("QBIT_PASSWORD"),

		RadarrURL:    envOr("RADARR_URL", "http://radarr:7878"),
		RadarrAPIKey: os.Getenv("RADARR_API_KEY"),

		SonarrURL:    envOr("SONARR_URL", "http://sonarr:8989"),
		SonarrAPIKey: os.Getenv("SONARR_API_KEY"),

		SeerrURL:    envOr("SEERR_URL", "http://seerr:5055"),
		SeerrAPIKey: os.Getenv("SEERR_API_KEY"),

		ProwlarrURL:    envOr("PROWLARR_URL", "http://prowlarr:9696"),
		ProwlarrAPIKey: os.Getenv("PROWLARR_API_KEY"),

		BazarrURL:    envOr("BAZARR_URL", "http://bazarr:6767"),
		BazarrAPIKey: os.Getenv("BAZARR_API_KEY"),

		SabnzbdURL:    envOr("SABNZBD_URL", "http://sabnzbd:8080"),
		SabnzbdAPIKey: os.Getenv("SABNZBD_API_KEY"),

		UnmanicURL: envOr("UNMANIC_URL", "http://unmanic:8888"),

		DockerSocket: envOr("DOCKER_SOCKET", "/var/run/docker.sock"),
		HostProcPath: envOr("HOST_PROC", "/host/proc"),
		HostUtmpPath: envOr("HOST_UTMP", "/host/run/utmp"),

		DashboardUser: os.Getenv("DASHBOARD_USER"),
		DashboardPass: os.Getenv("DASHBOARD_PASS"),

		PiholeURL:      envOr("PIHOLE_URL", "http://192.168.1.254"),
		PiholePassword: os.Getenv("PIHOLE_PASSWORD"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
