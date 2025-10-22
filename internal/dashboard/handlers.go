package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// handleRequests serves the request list API
func (s *Server) handleRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	requests := s.proxy.GetRequests()
	json.NewEncoder(w).Encode(requests)
}

// handleRequestsStream serves SSE stream of new requests
func (s *Server) handleRequestsStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	
	// Subscribe to new requests
	ch := s.proxy.Subscribe()
	defer s.proxy.Unsubscribe(ch)
	
	// Send existing requests first
	for _, req := range s.proxy.GetRequests() {
		data, _ := json.Marshal(req)
		fmt.Fprintf(w, "data: %s\n\n", data)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
	
	// Stream new requests
	for req := range ch {
		data, _ := json.Marshal(req)
		fmt.Fprintf(w, "data: %s\n\n", data)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

// Enhanced handleIndex with request inspector
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	stats := s.stats.GetStats()
	uptime := time.Since(stats.StartTime).Round(time.Second)
	
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>lrok - %s</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #0a0a0a; color: #f0f0f0; line-height: 1.6; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        
        /* Header */
        .header { text-align: center; padding: 30px 0; border-bottom: 1px solid #333; }
        .logo { font-size: 36px; font-weight: 700; background: linear-gradient(135deg, #FFD700 0%%, #FF8000 50%%, #E94055 100%%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
        .status { display: inline-block; padding: 6px 12px; background: rgba(16, 185, 129, 0.2); color: #10b981; border-radius: 20px; font-size: 13px; margin-top: 8px; }
        .status::before { content: "●"; margin-right: 6px; animation: pulse 2s infinite; }
        @keyframes pulse { 0%%, 100%% { opacity: 1; } 50%% { opacity: 0.5; } }
        
        /* Cards */
        .card { background: #1a1a1a; border: 1px solid #333; border-radius: 8px; padding: 20px; margin: 16px 0; }
        .url { font-size: 16px; font-weight: 600; color: #FF8000; word-break: break-all; padding: 12px; background: rgba(255, 128, 0, 0.1); border-radius: 6px; margin: 12px 0; font-family: monospace; }
        
        /* Stats Grid */
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 12px; margin-top: 16px; }
        .stat { background: rgba(255, 255, 255, 0.05); padding: 12px; border-radius: 6px; text-align: center; }
        .stat-label { font-size: 11px; color: #888; text-transform: uppercase; margin-bottom: 4px; }
        .stat-value { font-size: 18px; font-weight: 600; color: #FF8000; }
        
        /* Request List */
        .requests-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
        .requests-header h2 { font-size: 18px; color: #f0f0f0; }
        .btn { padding: 6px 12px; background: rgba(255, 128, 0, 0.2); color: #FF8000; border: 1px solid #FF8000; border-radius: 4px; cursor: pointer; font-size: 12px; }
        .btn:hover { background: rgba(255, 128, 0, 0.3); }
        
        .request-list { max-height: 400px; overflow-y: auto; }
        .request-item { background: rgba(255, 255, 255, 0.03); border: 1px solid #2a2a2a; padding: 12px; margin-bottom: 8px; border-radius: 6px; cursor: pointer; transition: all 0.2s; display: grid; grid-template-columns: 80px 60px 80px 1fr 80px 120px; gap: 12px; align-items: center; font-size: 13px; }
        .request-item:hover { background: rgba(255, 255, 255, 0.08); border-color: #FF8000; }
        
        .req-time { color: #888; font-size: 12px; }
        .req-status { padding: 4px 8px; border-radius: 4px; font-weight: 600; text-align: center; font-size: 12px; }
        .status-2xx { background: rgba(16, 185, 129, 0.2); color: #10b981; }
        .status-3xx { background: rgba(251, 191, 36, 0.2); color: #fbbf24; }
        .status-4xx { background: rgba(239, 68, 68, 0.2); color: #ef4444; }
        .status-5xx { background: rgba(220, 38, 38, 0.3); color: #dc2626; }
        .req-method { color: #FF8000; font-weight: 600; font-size: 12px; }
        .req-path { color: #f0f0f0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .req-duration { color: #888; font-size: 12px; }
        .req-size { color: #888; font-size: 12px; }
        
        .empty { text-align: center; padding: 40px; color: #666; font-size: 14px; }
        
        /* Scrollbar */
        ::-webkit-scrollbar { width: 8px; }
        ::-webkit-scrollbar-track { background: #1a1a1a; }
        ::-webkit-scrollbar-thumb { background: #333; border-radius: 4px; }
        ::-webkit-scrollbar-thumb:hover { background: #FF8000; }
        
        .info { font-size: 12px; color: #888; margin-top: 8px; }
        a { color: #FF8000; text-decoration: none; }
        a:hover { color: #E94055; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">lrok</div>
            <div class="status">%s</div>
        </div>
        
        <div class="card">
            <h2 style="margin-bottom: 12px; font-size: 16px; color: #f0f0f0;">🌐 Public URL</h2>
            <div class="url">%s</div>
            <div class="info">
                📍 Forwarding to: <code style="color: #10b981;">localhost:%d</code>
            </div>
        </div>
        
        <div class="card">
            <h2 style="margin-bottom: 12px; font-size: 16px; color: #f0f0f0;">📊 Statistics</h2>
            <div class="stats">
                <div class="stat">
                    <div class="stat-label">↓ Received</div>
                    <div class="stat-value" id="bytes-in">0 B</div>
                </div>
                <div class="stat">
                    <div class="stat-label">↑ Sent</div>
                    <div class="stat-value" id="bytes-out">0 B</div>
                </div>
                <div class="stat">
                    <div class="stat-label">Connections</div>
                    <div class="stat-value" id="connections">0</div>
                </div>
                <div class="stat">
                    <div class="stat-label">Uptime</div>
                    <div class="stat-value" id="uptime">%s</div>
                </div>
            </div>
        </div>
        
        <div class="card">
            <div class="requests-header">
                <h2>🔍 Request Inspector</h2>
                <div>
                    <button class="btn" onclick="clearRequests()">Clear</button>
                    <button class="btn" onclick="togglePause()" id="pauseBtn">Pause</button>
                </div>
            </div>
            <div class="request-list" id="requestList">
                <div class="empty">No requests yet. Send a request to your public URL to see it here!</div>
            </div>
        </div>
        
        <div class="card">
            <h2 style="margin-bottom: 12px; font-size: 16px; color: #f0f0f0;">💡 Tips</h2>
            <ul style="list-style: none; padding: 0; color: #888; font-size: 13px; line-height: 1.8;">
                <li>• Click any request above to see full headers and body</li>
                <li>• Requests update in real-time (auto-refresh)</li>
                <li>• Last 100 requests are kept in memory</li>
                <li>• View full stats at <a href="https://platform.lum.tools/tunnels" target="_blank">platform.lum.tools/tunnels</a></li>
            </ul>
        </div>
    </div>
    
    <script>
        let paused = false;
        
        // Update stats
        async function updateStats() {
            try {
                const response = await fetch('/api/stats');
                const data = await response.json();
                document.getElementById('bytes-in').textContent = formatBytes(data.bytes_in);
                document.getElementById('bytes-out').textContent = formatBytes(data.bytes_out);
                document.getElementById('connections').textContent = data.connections;
                document.getElementById('uptime').textContent = formatDuration(data.start_time);
            } catch (e) {}
        }
        
        // Load requests via SSE
        const eventSource = new EventSource('/api/requests/stream');
        const requests = [];
        
        eventSource.onmessage = function(event) {
            if (paused) return;
            
            const req = JSON.parse(event.data);
            requests.unshift(req);
            if (requests.length > 100) requests.pop();
            
            renderRequests();
        };
        
        function renderRequests() {
            const container = document.getElementById('requestList');
            
            if (requests.length === 0) {
                container.innerHTML = '<div class="empty">No requests yet. Send a request to your public URL!</div>';
                return;
            }
            
            container.innerHTML = requests.map(req => {
                const statusClass = 'status-' + Math.floor(req.status_code / 100) + 'xx';
                const time = new Date(req.timestamp).toLocaleTimeString();
                const duration = Math.round(req.duration / 1000000) + 'ms';
                
                return ` + "`" + `
                    <div class="request-item" onclick="showRequest('${req.id}')">
                        <div class="req-time">${time}</div>
                        <div class="req-status ${statusClass}">${req.status_code}</div>
                        <div class="req-method">${req.method}</div>
                        <div class="req-path">${req.path}</div>
                        <div class="req-duration">${duration}</div>
                        <div class="req-size">↓${formatBytes(req.bytes_in)} ↑${formatBytes(req.bytes_out)}</div>
                    </div>
                ` + "`" + `;
            }).join('');
        }
        
        function showRequest(id) {
            const req = requests.find(r => r.id === id);
            if (!req) return;
            
            // Format JSON if content-type is JSON
            let reqBody = req.request_body;
            let resBody = req.response_body;
            
            const reqCT = req.request_headers['Content-Type'] || '';
            const resCT = req.response_headers['Content-Type'] || '';
            
            if (reqCT.includes('json') && reqBody) {
                try { reqBody = JSON.stringify(JSON.parse(reqBody), null, 2); } catch(e) {}
            }
            if (resCT.includes('json') && resBody) {
                try { resBody = JSON.stringify(JSON.parse(resBody), null, 2); } catch(e) {}
            }
            
            const modal = ` + "`" + `
                <div style="position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.8); z-index: 1000; overflow-y: auto; padding: 20px;" onclick="this.remove()">
                    <div class="card" style="max-width: 900px; margin: 40px auto;" onclick="event.stopPropagation()">
                        <div style="display: flex; justify-content: space-between; margin-bottom: 20px;">
                            <button class="btn" onclick="this.closest('[style*=fixed]').remove()">◀ Back</button>
                            <button class="btn" onclick="copyCurl('${req.id}')">Copy cURL</button>
                        </div>
                        
                        <h2 style="font-size: 20px; margin-bottom: 8px; color: #FF8000;">${req.method} ${req.path}</h2>
                        <div style="font-size: 13px; color: #888; margin-bottom: 20px;">
                            <span class="req-status status-${Math.floor(req.status_code / 100)}xx">${req.status_code}</span>
                            • ${Math.round(req.duration / 1000000)}ms
                            • ↓ ${formatBytes(req.bytes_in)}
                            • ↑ ${formatBytes(req.bytes_out)}
                        </div>
                        
                        <h3 style="font-size: 14px; color: #f0f0f0; margin: 16px 0 8px;">📤 Request Headers</h3>
                        <pre style="background: #0a0a0a; padding: 12px; border-radius: 4px; font-size: 12px; overflow-x: auto; color: #888;">${formatHeaders(req.request_headers)}</pre>
                        
                        ${reqBody ? ` + "`" + `
                        <h3 style="font-size: 14px; color: #f0f0f0; margin: 16px 0 8px;">📥 Request Body</h3>
                        <pre style="background: #0a0a0a; padding: 12px; border-radius: 4px; font-size: 12px; overflow-x: auto; color: #10b981;">${escapeHtml(reqBody)}</pre>
                        ` + "`" + ` : ''}
                        
                        <h3 style="font-size: 14px; color: #f0f0f0; margin: 16px 0 8px;">📤 Response Headers</h3>
                        <pre style="background: #0a0a0a; padding: 12px; border-radius: 4px; font-size: 12px; overflow-x: auto; color: #888;">${formatHeaders(req.response_headers)}</pre>
                        
                        ${resBody ? ` + "`" + `
                        <h3 style="font-size: 14px; color: #f0f0f0; margin: 16px 0 8px;">📥 Response Body</h3>
                        <pre style="background: #0a0a0a; padding: 12px; border-radius: 4px; font-size: 12px; overflow-x: auto; color: #E94055;">${escapeHtml(resBody)}</pre>
                        ` + "`" + ` : ''}
                    </div>
                </div>
            ` + "`" + `;
            
            document.body.insertAdjacentHTML('beforeend', modal);
        }
        
        function formatHeaders(headers) {
            return Object.entries(headers).map(([k, v]) => k + ': ' + v).join('\n');
        }
        
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
        
        function clearRequests() {
            requests.length = 0;
            renderRequests();
        }
        
        function togglePause() {
            paused = !paused;
            document.getElementById('pauseBtn').textContent = paused ? 'Resume' : 'Pause';
        }
        
        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
        }
        
        function formatDuration(startTime) {
            const start = new Date(startTime);
            const now = new Date();
            const diff = Math.floor((now - start) / 1000);
            const hours = Math.floor(diff / 3600);
            const minutes = Math.floor((diff %% 3600) / 60);
            const seconds = diff %% 60;
            return hours + 'h ' + minutes + 'm ' + seconds + 's';
        }
        
        setInterval(updateStats, 1000);
        updateStats();
    </script>
</body>
</html>`,
		stats.TunnelName,
		stats.Status,
		stats.PublicURL,
		stats.LocalPort,
		uptime.String(),
	)
	
	fmt.Fprint(w, html)
}

