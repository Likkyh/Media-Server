// SVG icon helpers (inline, 14x14)
const icons = {
    service: {
        jellyfin:    '<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>',
        radarr:      '<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M14.31 8l5.74 9.94M9.69 8h11.48M7.38 12l5.74-9.94M9.69 16L3.95 6.06M14.31 16H2.83M16.62 12l-5.74 9.94"/></svg>',
        sonarr:      '<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="7" width="20" height="15" rx="2" ry="2"/><polyline points="17 2 12 7 7 2"/></svg>',
        seerr:       '<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>',
        prowlarr:    '<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><line x1="11" y1="8" x2="11" y2="14"/><line x1="8" y1="11" x2="14" y2="11"/></svg>',
        bazarr:      '<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/></svg>',
        qbittorrent: '<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg>',
        sabnzbd:     '<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v14M5 10l7 7 7-7"/><path d="M21 17v2a2 2 0 01-2 2H5a2 2 0 01-2-2v-2"/></svg>',
        gluetun:     '<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>',
        flaresolverr:'<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>',
        unmanic:     '<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14.7 6.3a1 1 0 000 1.4l1.6 1.6a1 1 0 001.4 0l3.77-3.77a6 6 0 01-7.94 7.94l-6.91 6.91a2.12 2.12 0 01-3-3l6.91-6.91a6 6 0 017.94-7.94l-3.76 3.76z"/></svg>',
        npm:         '<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>',
        arcticmon:   '<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>',
        autoheal:    '<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20.84 4.61a5.5 5.5 0 00-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 00-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 000-7.78z"/></svg>',
    },
    _default: '<svg class="svc-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="2" width="20" height="20" rx="5" ry="5"/></svg>'
};

function getSvcIcon(name) {
    return icons.service[name] || icons._default;
}

// DOM render functions for each dashboard section

function renderHost(host) {
    const ramBar = document.getElementById('ram-bar');
    const gpuBar = document.getElementById('gpu-bar');
    const swapBar = document.getElementById('swap-bar');

    // Per-core CPU bars with avg as first bar
    const coresContainer = document.getElementById('cpu-cores');

    if (host.cpuCores && host.cpuCores.length > 0) {
        const avgPct = host.cpuPercent.toFixed(0);
        const avgBar = `<div class="cpu-core cpu-core-avg">
                <div class="cpu-core-bar">
                    <div class="cpu-core-fill" style="width:${host.cpuPercent.toFixed(1)}%"></div>
                    <span class="cpu-core-id">AVG</span>
                    <span class="cpu-core-pct">${avgPct}%</span>
                </div>
            </div>`;
        coresContainer.innerHTML = avgBar + host.cpuCores.map(c => {
            const pct = c.percent.toFixed(0);
            return `<div class="cpu-core">
                <div class="cpu-core-bar">
                    <div class="cpu-core-fill" style="width:${c.percent.toFixed(1)}%"></div>
                    <span class="cpu-core-id">${c.id}</span>
                    <span class="cpu-core-pct">${pct}%</span>
                </div>
            </div>`;
        }).join('');
    }

    ramBar.style.width = host.memPercent.toFixed(1) + '%';
    document.getElementById('ram-value').textContent =
        formatPercent(host.memPercent) + ' (' + formatBytes(host.memUsed) + '/' + formatBytes(host.memTotal) + ')';

    if (host.gpu && host.gpu.available) {
        gpuBar.style.width = host.gpu.utilPercent + '%';
        document.getElementById('gpu-value').textContent = host.gpu.utilPercent + '%';
        document.getElementById('gpu-info').textContent = 'GPU: ' + esc(host.gpu.name);
        document.getElementById('gpu-temp').textContent = 'GPU Temp: ' + host.gpu.tempC + '\u00B0C';
        document.getElementById('gpu-mem').textContent =
            'VRAM: ' + formatBytes(host.gpu.memUsed) + '/' + formatBytes(host.gpu.memTotal);
    } else {
        gpuBar.style.width = '0%';
        document.getElementById('gpu-value').textContent = 'N/A';
        document.getElementById('gpu-info').textContent = 'GPU: unavailable';
        document.getElementById('gpu-temp').textContent = '';
        document.getElementById('gpu-mem').textContent = '';
    }

    swapBar.style.width = (host.swapPercent || 0).toFixed(1) + '%';
    if (host.swapTotal > 0) {
        document.getElementById('swap-value').textContent =
            formatPercent(host.swapPercent) + ' (' + formatBytes(host.swapUsed) + '/' + formatBytes(host.swapTotal) + ')';
    } else {
        document.getElementById('swap-value').textContent = 'None';
    }

    // Disks
    const diskContainer = document.getElementById('disk-metrics');
    if (host.disks && host.disks.length > 0) {
        diskContainer.innerHTML = host.disks.map((d, i) => {
            const barClass = i === 0 ? 'bar-disk-nvme' : 'bar-disk-hdd';
            return `<div class="metric">
                <div class="metric-header"><span>${esc(d.label)}</span><span>${formatPercent(d.percent)} (${formatBytes(d.used)}/${formatBytes(d.total)})</span></div>
                <div class="bar"><div class="bar-fill ${barClass}" style="width:${d.percent.toFixed(1)}%"></div></div>
            </div>`;
        }).join('');
    } else {
        diskContainer.innerHTML = '';
    }

    document.getElementById('ssh-count').textContent = 'SSH: ' + (host.sshSessions || 0);
    document.getElementById('server-uptime').textContent = host.uptime || '--';
}

function renderServices(services) {
    const grid = document.getElementById('service-grid');
    if (!services || services.length === 0) {
        grid.innerHTML = '<p class="empty-state">No services detected</p>';
        return;
    }

    const healthLabels = {healthy: 'Healthy', unhealthy: 'Unhealthy', starting: 'Starting', running: 'Running', stopped: 'Stopped', none: 'No check'};
    grid.innerHTML = services.map(svc => {
        const tag = svc.externalUrl ? 'a' : 'div';
        const href = svc.externalUrl ? ` href="${esc(svc.externalUrl)}" target="_blank" rel="noopener"` : '';
        const healthClass = svc.health || 'stopped';
        const healthLabel = healthLabels[healthClass] || healthLabels.none;
        const icon = getSvcIcon(svc.name);
        const linkIcon = svc.externalUrl ? '<svg class="link-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>' : '';
        return `<${tag} class="service-card"${href}>
            <span class="service-dot ${healthClass}"></span>
            ${icon}
            <div class="service-info">
                <div class="service-name">${esc(svc.name)}${linkIcon}</div>
                <div class="service-meta"><span class="service-health-text ${healthClass}">${healthLabel}</span> \u00B7 ${esc(svc.uptime)}</div>
            </div>
        </${tag}>`;
    }).join('');
}

function renderStreams(streams) {
    const wrap = document.getElementById('streams-table');
    if (!streams || streams.length === 0) {
        wrap.innerHTML = '<p class="empty-state">No active streams</p>';
        return;
    }

    wrap.innerHTML = `<table class="streams-table">
        <thead><tr>
            <th>User</th><th>Title</th><th>Progress</th><th>Client</th><th>Method</th>
        </tr></thead>
        <tbody>${streams.map(s => {
            const title = s.series
                ? `${esc(s.series)} ${esc(s.episode)} - ${esc(s.title)}`
                : esc(s.title);
            const methodIcon = s.transcoding
                ? '<svg class="inline-icon warn" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14.7 6.3a1 1 0 000 1.4l1.6 1.6a1 1 0 001.4 0l3.77-3.77a6 6 0 01-7.94 7.94l-6.91 6.91a2.12 2.12 0 01-3-3l6.91-6.91a6 6 0 017.94-7.94l-3.76 3.76z"/></svg>'
                : '<svg class="inline-icon ok" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>';
            return `<tr>
                <td>${esc(s.user)}</td>
                <td>${title}</td>
                <td>
                    <span class="progress-mini"><span class="progress-mini-fill" style="width:${s.progress.toFixed(1)}%"></span></span>
                    ${s.progress.toFixed(0)}%
                </td>
                <td>${esc(s.client)}</td>
                <td>${methodIcon} ${esc(s.playMethod)}</td>
            </tr>`;
        }).join('')}</tbody>
    </table>`;
}

function renderTorrents(torrents) {
    document.getElementById('torrent-dl').textContent = formatSpeed(torrents.dlSpeed);
    document.getElementById('torrent-ul').textContent = formatSpeed(torrents.upSpeed);
    document.getElementById('torrent-ratio').textContent = torrents.ratio ? torrents.ratio.toFixed(2) : '--';
    document.getElementById('torrent-counts').textContent =
        `${torrents.activeCount} / ${torrents.seedingCount} / ${torrents.totalCount}`;

    const list = document.getElementById('torrent-list');
    if (!torrents.topTorrents || torrents.topTorrents.length === 0) {
        list.innerHTML = '<p class="empty-state">No active torrents</p>';
        return;
    }

    list.innerHTML = torrents.topTorrents.slice(0, 5).map(t => `
        <div class="torrent-item">
            <span class="torrent-item-name" title="${esc(t.name)}">${esc(truncate(t.name, 50))}</span>
            <span class="torrent-item-stats">
                <span>\u2193 ${formatSpeed(t.dlSpeed)}</span>
                <span>\u2191 ${formatSpeed(t.upSpeed)}</span>
                <span>${t.progress.toFixed(0)}%</span>
            </span>
        </div>
    `).join('');
}

function renderDownloads(downloads) {
    const list = document.getElementById('download-list');
    if (!downloads || downloads.length === 0) {
        list.innerHTML = '<p class="empty-state">Queue empty</p>';
        return;
    }

    list.innerHTML = downloads.slice(0, 5).map(d => `
        <div class="download-item">
            <span class="download-item-name" title="${esc(d.title)}">
                <span class="status-badge ${d.source.toLowerCase()}">${esc(d.source)}</span>
                ${esc(truncate(d.title, 45))}
            </span>
            <span class="download-item-stats">
                <span class="download-progress"><span class="download-progress-fill" style="width:${d.progress.toFixed(1)}%"></span></span>
                <span>${d.progress.toFixed(0)}%</span>
                <span>${d.timeleft || '--'}</span>
            </span>
        </div>
    `).join('');
}

function renderRequests(requests) {
    const list = document.getElementById('request-list');
    if (!requests || requests.length === 0) {
        list.innerHTML = '<p class="empty-state">No recent requests</p>';
        return;
    }

    list.innerHTML = requests.map(r => `
        <div class="request-item">
            <span class="request-item-name">
                <span class="status-badge ${r.status.toLowerCase()}">${esc(r.status)}</span>
                ${esc(r.title)}
            </span>
            <span class="request-item-meta">
                <span>${esc(r.user)}</span>
                <span>${timeAgo(r.requestedAt)}</span>
            </span>
        </div>
    `).join('');
}

function renderTranscoding(transcodes) {
    document.getElementById('transcode-pending').textContent = 'Pending: ' + (transcodes.pending || 0);

    const list = document.getElementById('transcode-workers');
    if (!transcodes.workers || transcodes.workers.length === 0) {
        list.innerHTML = '<p class="empty-state">No active workers</p>';
        return;
    }

    list.innerHTML = transcodes.workers.map(w => {
        const isIdle = w.status === 'idle';
        return `<div class="worker-card">
            <div class="worker-header">
                <span class="worker-file">${isIdle ? 'Idle' : esc(truncate(w.fileName, 50))}</span>
                <span class="worker-stats">${isIdle ? '' : `${w.progress.toFixed(1)}% \u00B7 ${w.speed || '--'}`}</span>
            </div>
            ${isIdle ? '' : `<div class="bar"><div class="bar-fill bar-gpu" style="width:${w.progress.toFixed(1)}%"></div></div>`}
        </div>`;
    }).join('');
}

function renderLibrary(lib) {
    if (!lib) return;
    document.getElementById('lib-movies').textContent = (lib.movies || 0).toLocaleString();
    document.getElementById('lib-series').textContent = (lib.series || 0).toLocaleString();
    document.getElementById('lib-episodes').textContent = (lib.episodes || 0).toLocaleString();
    document.getElementById('lib-music').textContent = (lib.songs || 0).toLocaleString();
}

function renderHealth(health) {
    const list = document.getElementById('health-list');
    if (!health || health.length === 0) {
        list.innerHTML = '<p class="empty-state">All systems healthy</p>';
        return;
    }

    list.innerHTML = health.map(h => {
        const icon = h.type === 'error'
            ? '<svg class="inline-icon err" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>'
            : '<svg class="inline-icon warn" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>';
        return `<div class="health-item ${h.type === 'error' ? 'error' : ''}">
            ${icon}
            <span class="health-source">${esc(h.source)}</span>
            <span class="health-message">${esc(h.message)}</span>
        </div>`;
    }).join('');
}

// Store latest SSH data for period switching
var _sshData = null;
var _sshPeriod = '24h';

function renderSSHSecurity(data) {
    if (!data) return;
    _sshData = data;

    // Update counters based on active period
    updateSSHCounters();

    // Top offenders (always 30d window)
    const offList = document.getElementById('ssh-offenders');
    if (!data.topOffenders || data.topOffenders.length === 0) {
        offList.innerHTML = '<p class="empty-state">No failed attempts</p>';
    } else {
        offList.innerHTML = data.topOffenders.map(o =>
            `<div class="ssh-entry ssh-entry-offender">
                <span class="ssh-entry-ip">${esc(o.ip)}</span>
                <span class="ssh-entry-meta">
                    <span class="ssh-entry-time">${timeAgo(o.lastSeen)}</span>
                    <span class="ssh-entry-count">${o.attempts}</span>
                </span>
            </div>`
        ).join('');
    }

    // Recent failed
    const failList = document.getElementById('ssh-recent-failed');
    if (!data.recentFailed || data.recentFailed.length === 0) {
        failList.innerHTML = '<p class="empty-state">No failed attempts</p>';
    } else {
        failList.innerHTML = data.recentFailed.map(e =>
            `<div class="ssh-entry ssh-entry-fail">
                <span class="ssh-entry-user">${esc(e.user)}</span>
                <span class="ssh-entry-ip">${esc(e.ip)}</span>
                <span class="ssh-entry-meta">
                    <span class="ssh-entry-method">${esc(e.method)}</span>
                    <span class="ssh-entry-time">${timeAgo(e.time)}</span>
                </span>
            </div>`
        ).join('');
    }

    // Recent accepted
    const okList = document.getElementById('ssh-recent-accepted');
    if (!data.recentAccepted || data.recentAccepted.length === 0) {
        okList.innerHTML = '<p class="empty-state">No logins</p>';
    } else {
        okList.innerHTML = data.recentAccepted.map(e =>
            `<div class="ssh-entry ssh-entry-ok">
                <span class="ssh-entry-user">${esc(e.user)}</span>
                <span class="ssh-entry-ip">${esc(e.ip)}</span>
                <span class="ssh-entry-meta">
                    <span class="ssh-entry-method">${esc(e.method)}</span>
                    <span class="ssh-entry-time">${timeAgo(e.time)}</span>
                </span>
            </div>`
        ).join('');
    }
}

function updateSSHCounters() {
    if (!_sshData) return;
    var d = _sshData;
    var accepted, failed;
    if (_sshPeriod === '7d') {
        accepted = d.accepted7d || 0;
        failed = d.failed7d || 0;
    } else if (_sshPeriod === '30d') {
        accepted = d.accepted30d || 0;
        failed = d.failed30d || 0;
    } else {
        accepted = d.accepted24h || 0;
        failed = d.failed24h || 0;
    }
    document.getElementById('ssh-accepted').textContent = accepted.toLocaleString();
    document.getElementById('ssh-failed').textContent = failed.toLocaleString();
    document.getElementById('ssh-unique-ips').textContent =
        (d.topOffenders ? d.topOffenders.length : 0).toLocaleString();
}

// Period toggle click handler
document.getElementById('ssh-period-toggle').addEventListener('click', function(e) {
    var btn = e.target.closest('.period-btn');
    if (!btn) return;
    _sshPeriod = btn.dataset.period;
    this.querySelectorAll('.period-btn').forEach(function(b) { b.classList.remove('active'); });
    btn.classList.add('active');
    updateSSHCounters();
});
