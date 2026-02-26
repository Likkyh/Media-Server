package models

import "time"

// DashboardData holds the complete dashboard state.
type DashboardData struct {
	Host       HostMetrics      `json:"host"`
	Services   []ServiceStatus  `json:"services"`
	Streams    []StreamSession  `json:"streams"`
	Torrents   TorrentData      `json:"torrents"`
	Downloads  []DownloadItem   `json:"downloads"`
	Requests   []MediaRequest   `json:"requests"`
	Transcodes TranscodeData    `json:"transcodes"`
	Library    LibraryCounts    `json:"library"`
	Health     []HealthWarning  `json:"health"`
	SSHSecurity SSHSecurityData `json:"sshSecurity"`
	UpdatedAt  time.Time        `json:"updatedAt"`
}

// HostMetrics represents host-level system metrics.
type HostMetrics struct {
	CPUPercent  float64     `json:"cpuPercent"`
	CPUCores    []CPUCore   `json:"cpuCores"`
	MemTotal    uint64      `json:"memTotal"`
	MemUsed     uint64      `json:"memUsed"`
	MemPercent  float64     `json:"memPercent"`
	SwapTotal   uint64      `json:"swapTotal"`
	SwapUsed    uint64      `json:"swapUsed"`
	SwapPercent float64     `json:"swapPercent"`
	Disks       []DiskInfo  `json:"disks"`
	GPU         GPUMetrics  `json:"gpu"`
	SSHSessions int         `json:"sshSessions"`
	Uptime      string      `json:"uptime"`
}

// CPUCore represents a single CPU core's usage.
type CPUCore struct {
	ID      int     `json:"id"`
	Percent float64 `json:"percent"`
}

// DiskInfo represents a single mount point's usage.
type DiskInfo struct {
	Mount   string  `json:"mount"`
	Label   string  `json:"label"`
	Total   uint64  `json:"total"`
	Used    uint64  `json:"used"`
	Percent float64 `json:"percent"`
}

// GPUMetrics represents NVIDIA GPU metrics.
type GPUMetrics struct {
	Name        string  `json:"name"`
	TempC       int     `json:"tempC"`
	UtilPercent int     `json:"utilPercent"`
	MemUsed     uint64  `json:"memUsed"`
	MemTotal    uint64  `json:"memTotal"`
	MemPercent  float64 `json:"memPercent"`
	Available   bool    `json:"available"`
}

// ServiceStatus represents the health of a Docker container.
type ServiceStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Health    string `json:"health"`
	Image     string `json:"image"`
	IP        string `json:"ip"`
	Uptime    string `json:"uptime"`
	ExternalURL string `json:"externalUrl,omitempty"`
}

// StreamSession represents a Jellyfin playback session.
type StreamSession struct {
	User        string  `json:"user"`
	Title       string  `json:"title"`
	Series      string  `json:"series,omitempty"`
	Episode     string  `json:"episode,omitempty"`
	Client      string  `json:"client"`
	Device      string  `json:"device"`
	Progress    float64 `json:"progress"`
	Transcoding bool    `json:"transcoding"`
	PlayMethod  string  `json:"playMethod"`
}

// TorrentData holds qBittorrent aggregate data.
type TorrentData struct {
	DLSpeed      uint64          `json:"dlSpeed"`
	UPSpeed      uint64          `json:"upSpeed"`
	Ratio        float64         `json:"ratio"`
	ActiveCount  int             `json:"activeCount"`
	SeedingCount int             `json:"seedingCount"`
	TotalCount   int             `json:"totalCount"`
	TopTorrents  []TorrentBrief  `json:"topTorrents"`
}

// TorrentBrief is a summary of a single torrent.
type TorrentBrief struct {
	Name     string  `json:"name"`
	State    string  `json:"state"`
	Progress float64 `json:"progress"`
	DLSpeed  uint64  `json:"dlSpeed"`
	UPSpeed  uint64  `json:"upSpeed"`
	Size     uint64  `json:"size"`
	Ratio    float64 `json:"ratio"`
}

// DownloadItem represents a queued download from Radarr/Sonarr/SABnzbd.
type DownloadItem struct {
	Title    string  `json:"title"`
	Source   string  `json:"source"`
	Status   string  `json:"status"`
	Progress float64 `json:"progress"`
	Size     uint64  `json:"size"`
	Timeleft string  `json:"timeleft"`
}

// MediaRequest represents a Seerr request.
type MediaRequest struct {
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	User      string    `json:"user"`
	RequestedAt time.Time `json:"requestedAt"`
}

// TranscodeData holds Unmanic transcoding status.
type TranscodeData struct {
	Workers []TranscodeWorker `json:"workers"`
	Pending int               `json:"pending"`
}

// TranscodeWorker represents a single Unmanic worker.
type TranscodeWorker struct {
	ID       string  `json:"id"`
	FileName string  `json:"fileName"`
	Progress float64 `json:"progress"`
	FPS      float64 `json:"fps"`
	Speed    string  `json:"speed"`
	Status   string  `json:"status"`
}

// HealthWarning represents a health issue from the arr suite.
type HealthWarning struct {
	Source  string `json:"source"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// LibraryCounts holds Jellyfin library item counts.
type LibraryCounts struct {
	Movies   int `json:"movies"`
	Series   int `json:"series"`
	Episodes int `json:"episodes"`
	Music    int `json:"music"`
}

// SSHSecurityData holds SSH authentication log analysis.
type SSHSecurityData struct {
	Failed24h      int              `json:"failed24h"`
	Failed7d       int              `json:"failed7d"`
	Failed30d      int              `json:"failed30d"`
	Accepted24h    int              `json:"accepted24h"`
	Accepted7d     int              `json:"accepted7d"`
	Accepted30d    int              `json:"accepted30d"`
	TopOffenders   []SSHOffender    `json:"topOffenders"`
	RecentFailed   []SSHAuthEvent   `json:"recentFailed"`
	RecentAccepted []SSHAuthEvent   `json:"recentAccepted"`
}

// SSHOffender is an IP with its failed attempt count.
type SSHOffender struct {
	IP       string `json:"ip"`
	Attempts int    `json:"attempts"`
	LastSeen string `json:"lastSeen"`
	Country  string `json:"country,omitempty"`
}

// SSHAuthEvent is a single SSH authentication event.
type SSHAuthEvent struct {
	Time     string `json:"time"`
	User     string `json:"user"`
	IP       string `json:"ip"`
	Method   string `json:"method"`
	Success  bool   `json:"success"`
}

// ArrStats holds combined stats for Radarr/Sonarr.
type ArrStats struct {
	Movies     int    `json:"movies"`
	Series     int    `json:"series"`
	Episodes   int    `json:"episodes"`
	MoviesMon  int    `json:"moviesMonitored"`
	SeriesMon  int    `json:"seriesMonitored"`
	DiskSpace  []DiskSpace `json:"diskSpace"`
}

// DiskSpace represents disk usage from an arr service.
type DiskSpace struct {
	Path       string `json:"path"`
	Total      uint64 `json:"total"`
	Free       uint64 `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}
