// Arctic Monitor - Main Application

// Action buttons
function doAction(action, btn) {
    if (btn.classList.contains('loading')) return;
    btn.classList.add('loading');
    fetch('/api/actions/' + action, { method: 'POST' })
        .then(function(r) { return r.json(); })
        .then(function(data) {
            if (data.error) {
                alert('Error: ' + data.error);
            }
        })
        .catch(function(err) {
            alert('Action failed: ' + err.message);
        })
        .finally(function() {
            btn.classList.remove('loading');
        });
}

var _pendingAction = null;

function confirmAction(action) {
    _pendingAction = action;
    var messages = {
        'restart-vm': 'Are you sure you want to reboot the server? All services will be temporarily unavailable.',
        'restart-stack': 'Are you sure you want to restart all containers?',
        'update-stack': 'Pull latest images, recreate containers, and prune unused images?',
        'update-system': 'Run apt full-upgrade on the host? This may take a while.'
    };
    document.getElementById('confirm-message').textContent = messages[action] || 'Are you sure?';
    document.getElementById('confirm-overlay').style.display = '';
    document.getElementById('confirm-btn').onclick = function() {
        closeConfirm();
        if (_pendingAction) {
            var btn = document.querySelector('.btn-action.danger');
            doAction(_pendingAction, btn || document.createElement('button'));
            _pendingAction = null;
        }
    };
}

function closeConfirm() {
    document.getElementById('confirm-overlay').style.display = 'none';
    _pendingAction = null;
}

(function() {
    'use strict';

    // Clock
    function updateClock() {
        var now = new Date();
        document.getElementById('clock').textContent = now.toLocaleTimeString('en-GB', {
            hour: '2-digit', minute: '2-digit', second: '2-digit'
        });
    }
    setInterval(updateClock, 1000);
    updateClock();

    // Initial data load
    async function loadOverview() {
        try {
            var resp = await fetch('/api/overview');
            if (!resp.ok) return;
            var data = await resp.json();

            if (data.host) renderHost(data.host);
            if (data.services) renderServices(data.services);
            if (data.streams) renderStreams(data.streams);
            if (data.torrents) renderTorrents(data.torrents);
            if (data.downloads) renderDownloads(data.downloads);
            if (data.requests) renderRequests(data.requests);
            if (data.transcodes) renderTranscoding(data.transcodes);
            if (data.health) renderHealth(data.health);
            if (data.library) renderLibrary(data.library);
            if (data.sshSecurity) renderSSHSecurity(data.sshSecurity);
        } catch (e) {
            console.error('Failed to load overview:', e);
        }
    }

    loadOverview();

    // SSE connection with auto-reconnect
    var sse = null;
    var reconnectTimer = null;
    var statusDot = document.getElementById('sse-status');

    function connectSSE() {
        if (sse) {
            sse.close();
        }

        sse = new EventSource('/api/events');

        sse.onopen = function() {
            statusDot.className = 'sse-status connected';
            statusDot.title = 'Connected';
            if (reconnectTimer) {
                clearTimeout(reconnectTimer);
                reconnectTimer = null;
            }
        };

        sse.onmessage = function(e) {
            try {
                var msg = JSON.parse(e.data);
                handleEvent(msg.event, msg.data);
            } catch (err) {
                console.error('SSE parse error:', err);
            }
        };

        sse.onerror = function() {
            statusDot.className = 'sse-status disconnected';
            statusDot.title = 'Disconnected - reconnecting...';
            sse.close();
            if (!reconnectTimer) {
                reconnectTimer = setTimeout(connectSSE, 3000);
            }
        };
    }

    function handleEvent(event, data) {
        switch (event) {
            case 'host':
                renderHost(data);
                break;
            case 'services':
                renderServices(data);
                break;
            case 'streams':
                renderStreams(data);
                break;
            case 'torrents':
                renderTorrents(data);
                break;
            case 'downloads':
                renderDownloads(data);
                break;
            case 'requests':
                renderRequests(data);
                break;
            case 'transcodes':
                renderTranscoding(data);
                break;
            case 'health':
                renderHealth(data);
                break;
            case 'library':
                renderLibrary(data);
                break;
            case 'sshSecurity':
                renderSSHSecurity(data);
                break;
        }
    }

    connectSSE();

})();
