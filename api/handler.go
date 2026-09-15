package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"transfer.agent/service/sender"
	"transfer.agent/storage"
	"transfer.agent/utils"
)

type Handler struct {
	store   *storage.Store
	saveDir string
	hub     *Hub
}

func NewHandler(store *storage.Store, saveDir string, hub *Hub) *Handler {
	return &Handler{store: store, saveDir: saveDir, hub: hub}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *Handler) ListTransfers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"transfers": h.store.ListTransfers()})
}

func (h *Handler) GetTransfer(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/transfers/")
	t := h.store.GetTransfer(id)
	if t == nil {
		writeError(w, "transfer not found", http.StatusNotFound)
		return
	}
	writeJSON(w, t)
}

func (h *Handler) DeleteTransfer(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/transfers/")
	t := h.store.GetTransfer(id)
	if t == nil {
		writeError(w, "transfer not found", http.StatusNotFound)
		return
	}
	if t.FileName != "" {
		os.Remove(filepath.Join(h.saveDir, t.FileName))
		h.store.RemoveFile(t.FileName)
	}
	h.store.DeleteTransfer(id)
	writeJSON(w, map[string]string{"status": "deleted"})
}

func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"files": h.store.ListFiles()})
}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/files/")
	if name == "" {
		writeError(w, "filename required", http.StatusBadRequest)
		return
	}
	clean := filepath.Base(name)
	path := filepath.Join(h.saveDir, clean)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		writeError(w, "file not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", clean))
	http.ServeFile(w, r, path)
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, "invalid form data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, "file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	cleanName := filepath.Base(header.Filename)
	destPath := filepath.Join(h.saveDir, cleanName)
	dest, err := os.Create(destPath)
	if err != nil {
		writeError(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	defer dest.Close()

	written, err := io.Copy(dest, file)
	if err != nil {
		writeError(w, "failed to write file", http.StatusInternalServerError)
		return
	}

	checkSum, _ := utils.CalculateChecksum(destPath)

	t := h.store.CreateTransfer(cleanName, written, "upload")
	h.store.UpdateTransfer(t.ID, func(tr *storage.Transfer) {
		tr.Status = "completed"
		tr.Progress = 100
		tr.CheckSum = checkSum
	})
	h.store.AddFile(cleanName, written, checkSum)

	h.hub.Broadcast(Event{
		Type: "complete",
		ID:   t.ID,
		Data: map[string]any{
			"file_name": cleanName,
			"file_size": written,
			"link":      t.Link,
		},
	})

	writeJSON(w, t)
}

func (h *Handler) SendToPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FileName string `json:"file_name"`
		Addr     string `json:"addr"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.FileName == "" || req.Addr == "" {
		writeError(w, "file_name and addr required", http.StatusBadRequest)
		return
	}

	cleanName := filepath.Base(req.FileName)
	srcPath := filepath.Join(h.saveDir, cleanName)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		writeError(w, "file not found locally. Upload first.", http.StatusBadRequest)
		return
	}

	t := h.store.CreateTransfer(cleanName, 0, "send")
	h.store.UpdateTransfer(t.ID, func(tr *storage.Transfer) {
		tr.RemoteAddr = req.Addr
		tr.Status = "sending"
	})

	go func() {
		sndr := sender.Init(req.Addr)
		sndr.OnProgress(func(sent, total int64) {
			pct := float64(sent) / float64(total) * 100
			h.store.UpdateTransfer(t.ID, func(tr *storage.Transfer) {
				tr.Progress = pct
				tr.FileSize = total
			})
			h.hub.Broadcast(Event{
				Type: "progress",
				ID:   t.ID,
				Data: map[string]any{"progress": pct, "sent": sent, "total": total},
			})
		})

		if err := sndr.Send(srcPath); err != nil {
			h.store.UpdateTransfer(t.ID, func(tr *storage.Transfer) {
				tr.Status = "failed"
			})
			h.hub.Broadcast(Event{Type: "error", ID: t.ID, Data: map[string]string{"message": err.Error()}})
			return
		}

		h.store.UpdateTransfer(t.ID, func(tr *storage.Transfer) {
			tr.Status = "completed"
			tr.Progress = 100
		})
		h.hub.Broadcast(Event{Type: "complete", ID: t.ID, Data: map[string]any{"link": t.Link}})
	}()

	writeJSON(w, t)
}

func (h *Handler) DownloadPage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/d/")
	t := h.store.GetTransfer(id)
	if t == nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Download - %s</title>
<style>
body{font-family:system-ui;display:flex;justify-content:center;align-items:center;min-height:100vh;margin:0;background:#0f172a;color:#e2e8f0}
.card{background:#1e293b;border-radius:16px;padding:40px;max-width:420px;width:90%%;text-align:center}
h1{font-size:1.5rem;margin-bottom:8px}
p{color:#94a3b8;margin:4px 0}
.btn{display:inline-block;margin-top:20px;padding:12px 32px;background:#3b82f6;color:#fff;border:none;border-radius:8px;font-size:1rem;cursor:pointer;text-decoration:none}
.btn:hover{background:#2563eb}
.icon{font-size:3rem;margin-bottom:16px}
</style></head><body>
<div class="card">
<div class="icon">&#128194;</div>
<h1>%s</h1>
<p>%s</p>
<p style="color:#64748b;font-size:0.85rem">SHA-256: %s</p>
<a class="btn" href="/api/files/%s">Download</a>
</div></body></html>`,
		t.FileName, t.FileName,
		formatSize(t.FileSize),
		t.CheckSum[:16]+"...",
		filepath.Base(t.FileName))
}

func formatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

