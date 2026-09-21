package web

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"go.bug.st/serial"

	"monitor-esp32-host/pkg/models"
	embedded "monitor-esp32-host/web"
)

// Hub gestiona el estado compartido entre el motor serial y los clientes web (SSE).
type Hub struct {
	mu           sync.RWMutex
	Connected    bool
	Port         string
	Baud         int
	IntervalMs   int
	LastPayload  models.TelemetryPayload
	clients      map[chan []byte]struct{}
	PortChangeCh chan string
}

func NewHub(baud, intervalMs int) *Hub {
	return &Hub{
		Baud:         baud,
		IntervalMs:   intervalMs,
		clients:      make(map[chan []byte]struct{}),
		PortChangeCh: make(chan string, 10),
	}
}

// UpdateState actualiza la información de conexión con el ESP32.
func (h *Hub) UpdateState(connected bool, port string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Connected = connected
	h.Port = port
}

// Broadcast envía el nuevo paquete de telemetría a todos los clientes web conectados por SSE.
func (h *Hub) Broadcast(payload models.TelemetryPayload) {
	h.mu.Lock()
	h.LastPayload = payload

	// Estructura combinada para la web
	wrapper := struct {
		models.TelemetryPayload
		Device struct {
			Connected bool   `json:"connected"`
			Port      string `json:"port"`
		} `json:"device"`
	}{
		TelemetryPayload: payload,
	}
	wrapper.Device.Connected = h.Connected
	wrapper.Device.Port = h.Port

	data, err := json.Marshal(wrapper)
	if err != nil {
		h.mu.Unlock()
		return
	}

	for ch := range h.clients {
		select {
		case ch <- data:
		default:
			// Si el cliente está lento, no bloquear a los demás
		}
	}
	h.mu.Unlock()
}

func (h *Hub) register(ch chan []byte) {
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) unregister(ch chan []byte) {
	h.mu.Lock()
	delete(h.clients, ch)
	close(ch)
	h.mu.Unlock()
}

// Server encapsula el servidor HTTP y los manejadores de API.
type Server struct {
	port int
	hub  *Hub
}

func NewServer(port int, hub *Hub) *Server {
	return &Server{
		port: port,
		hub:  hub,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// 1. Servir archivos estáticos embebidos (HTML, CSS, JS)
	staticFS, err := fs.Sub(embedded.Assets, ".")
	if err == nil {
		mux.Handle("/", http.FileServer(http.FS(staticFS)))
	}

	// 2. API Endpoints
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/telemetry", s.handleTelemetry)
	mux.HandleFunc("/api/ports", s.handlePorts)
	mux.HandleFunc("/api/settings", s.handleSettings)
	mux.HandleFunc("/api/events", s.handleEvents)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("[WEB] Servidor de control iniciado en http://localhost:%d\n", s.port)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.hub.mu.RLock()
	defer s.hub.mu.RUnlock()

	res := map[string]interface{}{
		"device": map[string]interface{}{
			"connected": s.hub.Connected,
			"port":      s.hub.Port,
			"baud":      s.hub.Baud,
		},
		"interval_ms": s.hub.IntervalMs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	s.hub.mu.RLock()
	defer s.hub.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.hub.LastPayload)
}

func (s *Server) handlePorts(w http.ResponseWriter, r *http.Request) {
	ports, _ := serial.GetPortsList()
	s.hub.mu.RLock()
	currPort := s.hub.Port
	s.hub.mu.RUnlock()

	res := map[string]interface{}{
		"ports":   ports,
		"current": currPort,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Port string `json:"port"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Port != "" {
		s.hub.PortChangeCh <- req.Port
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Puerto actualizado"})
		return
	}

	http.Error(w, "Petición inválida", http.StatusBadRequest)
}

// handleEvents transmite telemetría en tiempo real mediante Server-Sent Events (SSE).
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming no soportado", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	messageChan := make(chan []byte, 10)
	s.hub.register(messageChan)
	defer s.hub.unregister(messageChan)

	// Enviar estado inicial inmediatamente
	s.hub.mu.RLock()
	initialPayload := s.hub.LastPayload
	connected := s.hub.Connected
	port := s.hub.Port
	s.hub.mu.RUnlock()

	initialWrapper := map[string]interface{}{
		"cpu": initialPayload.CPU,
		"gpu": initialPayload.GPU,
		"ram": initialPayload.RAM,
		"device": map[string]interface{}{
			"connected": connected,
			"port":      port,
		},
	}
	if initBytes, err := json.Marshal(initialWrapper); err == nil {
		fmt.Fprintf(w, "data: %s\n\n", initBytes)
		flusher.Flush()
	}

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case msg, open := <-messageChan:
			if !open {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-time.After(15 * time.Second):
			// Keep-alive heartbeat para evitar desconexiones de proxies
			fmt.Fprintf(w, ": keep-alive\n\n")
			flusher.Flush()
		}
	}
}
