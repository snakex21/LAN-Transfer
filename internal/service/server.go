package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"lantransfer/internal/app"
	"lantransfer/internal/config"
	"lantransfer/internal/discovery"
	"lantransfer/internal/filebrowser"
	"lantransfer/internal/transfer"
)

func Start(state *app.State) (*http.Server, error) {
	listener, err := net.Listen("tcp4", address())
	if err != nil {
		return nil, err
	}
	server := newHTTPServer(state)
	go discovery.Responder{ID: state.ID, Name: state.Name}.Run()
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			_ = listener.Close()
		}
	}()
	return server, nil
}

func Run(state *app.State) error {
	listener, err := net.Listen("tcp4", address())
	if err != nil {
		return err
	}
	go discovery.Responder{ID: state.ID, Name: state.Name}.Run()
	return newHTTPServer(state).Serve(listener)
}

func newHTTPServer(state *app.State) *http.Server {
	mux := http.NewServeMux()
	receiver := transfer.Receiver{Directory: state.ReceiveDir}
	mux.HandleFunc("/api/info", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "GET only", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   state.ID,
			"name": state.Name,
			"port": config.HTTPPort,
		})
	})
	mux.HandleFunc("/api/receive", receiver.HandleHTTP)
	mux.HandleFunc("/api/browse", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "GET only", http.StatusMethodNotAllowed)
			return
		}
		listing, err := filebrowser.Browse(req.URL.Query().Get("path"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listing)
	})
	return &http.Server{Handler: mux}
}

func address() string {
	return net.JoinHostPort(config.BindHost(), stringPort(config.HTTPPort))
}

func stringPort(port int) string {
	return fmt.Sprintf("%d", port)
}
