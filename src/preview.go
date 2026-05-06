package mango

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

type RenderPreview struct {
	buffer *ImageBuffer
	server *http.Server
	addr   string
	start  time.Time

	mu       sync.RWMutex
	mode     string
	status   string
	samples  int
	progress float64
}

type renderPreviewStatus struct {
	Mode     string  `json:"mode"`
	Status   string  `json:"status"`
	Samples  int     `json:"samples"`
	Progress float64 `json:"progress"`
	Elapsed  string  `json:"elapsed"`
}

func StartRenderPreview(buffer *ImageBuffer, mode string) (*RenderPreview, error) {
	preview := &RenderPreview{
		buffer:  buffer,
		mode:    mode,
		status:  "Starting",
		samples: 1,
		start:   time.Now(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", preview.handleIndex)
	mux.HandleFunc("/preview.png", preview.handlePreviewPNG)
	mux.HandleFunc("/status", preview.handleStatus)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	preview.addr = "http://" + listener.Addr().String()
	preview.server = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
	}

	go func() {
		if err := preview.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Render preview stopped: %v\n", err)
		}
	}()

	return preview, nil
}

func (preview *RenderPreview) URL() string {
	if preview == nil {
		return ""
	}
	return preview.addr
}

func (preview *RenderPreview) Open() error {
	if preview == nil || preview.addr == "" {
		return nil
	}
	return openBrowser(preview.addr)
}

func (preview *RenderPreview) Update(samples int, progress float64, status string) {
	if preview == nil {
		return
	}

	preview.mu.Lock()
	defer preview.mu.Unlock()

	preview.samples = max(samples, 1)
	preview.progress = Clamp(progress, 0, 1)
	preview.status = status
}

func (preview *RenderPreview) Close(ctx context.Context) error {
	if preview == nil || preview.server == nil {
		return nil
	}
	return preview.server.Shutdown(ctx)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func (preview *RenderPreview) sampleScale() float64 {
	preview.mu.RLock()
	defer preview.mu.RUnlock()

	return 1.0 / float64(max(preview.samples, 1))
}

func (preview *RenderPreview) snapshotStatus() renderPreviewStatus {
	preview.mu.RLock()
	defer preview.mu.RUnlock()

	return renderPreviewStatus{
		Mode:     preview.mode,
		Status:   preview.status,
		Samples:  preview.samples,
		Progress: preview.progress,
		Elapsed:  time.Since(preview.start).Round(time.Second).String(),
	}
}

func (preview *RenderPreview) handlePreviewPNG(w http.ResponseWriter, r *http.Request) {
	pngBytes, err := preview.buffer.PNGBytes(preview.sampleScale())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(pngBytes)
}

func (preview *RenderPreview) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(preview.snapshotStatus())
}

func (preview *RenderPreview) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Mango Render Preview</title>
  <style>
    html, body { margin: 0; min-height: 100%; background: #181818; color: #eee; font: 14px system-ui, sans-serif; }
    header { position: sticky; top: 0; display: flex; gap: 18px; align-items: center; padding: 10px 14px; background: #111; border-bottom: 1px solid #333; }
    .bar { flex: 1; height: 8px; background: #333; border-radius: 999px; overflow: hidden; }
    .fill { height: 100%; width: 0%; background: #f1c76b; }
    main { display: grid; place-items: start center; padding: 16px; }
    img { max-width: 100%; height: auto; image-rendering: auto; box-shadow: 0 8px 36px rgba(0,0,0,.45); }
  </style>
</head>
<body>
  <header>
    <strong id="mode">Render</strong>
    <span id="status">Starting</span>
    <span id="samples"></span>
    <span id="elapsed"></span>
    <div class="bar"><div class="fill" id="fill"></div></div>
  </header>
  <main><img id="preview" src="/preview.png"></main>
  <script>
    async function refresh() {
      const status = await fetch('/status', { cache: 'no-store' }).then(r => r.json());
      document.getElementById('mode').textContent = status.mode;
      document.getElementById('status').textContent = status.status;
      document.getElementById('samples').textContent = status.samples + ' sample(s)';
      document.getElementById('elapsed').textContent = status.elapsed;
      document.getElementById('fill').style.width = Math.round(status.progress * 100) + '%';
      document.getElementById('preview').src = '/preview.png?t=' + Date.now();
    }
    refresh();
    setInterval(refresh, 750);
  </script>
</body>
</html>`)
}
