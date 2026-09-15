package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"transfer.agent/api"
	receiver "transfer.agent/service/receiver"
	"transfer.agent/service/sender"
	"transfer.agent/storage"
)

func main() {
	mode := flag.String("mode", "web", "Mode: 'web', 'sender', or 'receiver'")
	port := flag.String("port", "8080", "Port for web server or receiver")
	server := flag.String("server", "localhost:6789", "Server address for sender")
	file := flag.String("file", "./source_file/test_file.txt", "File to send")
	saveDir := flag.String("savedir", "./generated_file", "Directory to save received files")

	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	os.MkdirAll(*saveDir, 0755)

	switch *mode {
	case "receiver":
		log.Println("[RECEIVER] Starting in receiver mode")
		rvc := receiver.Init(*port, *saveDir)
		go func() {
			<-ctx.Done()
			rvc.Shutdown()
		}()
		if err := rvc.Start(); err != nil {
			log.Fatalf("[RECEIVER] Failed to start: %v", err)
		}

	case "sender":
		log.Println("[SENDER] Starting in sender mode")
		sndr := sender.Init(*server)
		if err := sndr.Send(*file); err != nil {
			log.Fatalf("[SENDER] Transfer failed: %v", err)
		}

	default:
		store := storage.New()
		hub := api.NewHub()
		h := api.NewHandler(store, *saveDir, hub)

		mux := http.NewServeMux()

		mux.HandleFunc("GET /api/health", h.Health)
		mux.HandleFunc("GET /api/transfers", h.ListTransfers)
		mux.HandleFunc("GET /api/transfers/", h.GetTransfer)
		mux.HandleFunc("DELETE /api/transfers/", h.DeleteTransfer)
		mux.HandleFunc("GET /api/files", h.ListFiles)
		mux.HandleFunc("GET /api/files/", h.DownloadFile)
		mux.HandleFunc("POST /api/upload", h.Upload)
		mux.HandleFunc("POST /api/send", h.SendToPeer)
		mux.HandleFunc("GET /ws", hub.HandleWebSocket)
		mux.HandleFunc("GET /d/", h.DownloadPage)

		corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			mux.ServeHTTP(w, r)
		})

		if p := os.Getenv("PORT"); p != "" {
			*port = p
		}

		log.Printf("[WEB] Transfer Agent API on :%s (saveDir=%s)", *port, *saveDir)
		go func() {
			<-ctx.Done()
			fmt.Println("\n[WEB] Shutting down...")
		}()
		if err := http.ListenAndServe(":"+*port, corsHandler); err != nil {
			log.Fatalf("[WEB] Server failed: %v", err)
		}
	}
}
