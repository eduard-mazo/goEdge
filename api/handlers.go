package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"goMqttModbus/config"
	"goMqttModbus/publisher"
)

// Gateway is the interface that api handlers call into.
type Gateway interface {
	Start() error
	Stop()
	Status() publisher.Status
	Reload(config.AppConfig)
}

// gatewayImpl wraps a *publisher.Publisher and the config store.
type gatewayImpl struct {
	store *config.Store
	pub   *publisher.Publisher
	hub   *Hub
}

func (g *gatewayImpl) logEvent(level, msg string) {
	g.hub.Broadcast(WSEvent{
		Type: EventLog,
		Payload: LogEntry{
			Level:   level,
			Message: msg,
			Time:    time.Now().Format("15:04:05.000"),
		},
	})
}

// Server is the HTTP API + WebSocket server.
type Server struct {
	mux   *http.ServeMux
	gw    *gatewayImpl
}

// NewServer constructs the API server.
func NewServer(store *config.Store, hub *Hub, staticFS http.Handler) *Server {
	gw := &gatewayImpl{store: store, hub: hub}
	s := &Server{mux: http.NewServeMux(), gw: gw}
	s.routes(staticFS)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	// CORS for local dev (Vite dev server proxy)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes(staticFS http.Handler) {
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/status", s.handleStatus)

	s.mux.HandleFunc("/api/config", s.handleConfig)
	s.mux.HandleFunc("/api/config/mqtt", s.handleMQTT)
	s.mux.HandleFunc("/api/config/sparkplug", s.handleSparkplug)

	s.mux.HandleFunc("/api/devices", s.handleDevices)
	s.mux.HandleFunc("/api/devices/", s.handleDevice)

	s.mux.HandleFunc("/api/mappings", s.handleMappings)
	s.mux.HandleFunc("/api/mappings/export", s.handleMappingsExport)
	s.mux.HandleFunc("/api/mappings/import", s.handleMappingsImport)
	s.mux.HandleFunc("/api/mappings/", s.handleMapping)

	s.mux.HandleFunc("/api/gateway/start", s.handleGatewayStart)
	s.mux.HandleFunc("/api/gateway/stop", s.handleGatewayStop)

	s.mux.HandleFunc("/ws", s.gw.hub.ServeWS)

	if staticFS != nil {
		s.mux.Handle("/", staticFS)
	}
}

// --- Health ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeOK(w, map[string]string{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
}

// --- Status ---

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	st := publisher.Status{}
	if s.gw.pub != nil {
		st = s.gw.pub.Status()
	}
	writeOK(w, st)
}

// --- Config ---

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeFail(w, http.StatusMethodNotAllowed, "GET only")
		return
	}
	writeOK(w, s.gw.store.Get())
}

func (s *Server) handleMQTT(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeOK(w, s.gw.store.Get().MQTT)
	case http.MethodPut:
		var cfg config.MQTTConfig
		if err := decode(r.Body, &cfg); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := validateMQTT(cfg); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.gw.store.UpdateMQTT(cfg); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "MQTT config updated")
		writeOK(w, cfg)
	default:
		writeFail(w, http.StatusMethodNotAllowed, "GET or PUT")
	}
}

func (s *Server) handleSparkplug(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeOK(w, s.gw.store.Get().Sparkplug)
	case http.MethodPut:
		var cfg config.SparkplugConfig
		if err := decode(r.Body, &cfg); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := validateSparkplug(cfg); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.gw.store.UpdateSparkplug(cfg); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Sparkplug config updated")
		writeOK(w, cfg)
	default:
		writeFail(w, http.StatusMethodNotAllowed, "GET or PUT")
	}
}

// --- Devices ---

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeOK(w, s.gw.store.Get().Devices)
	case http.MethodPost:
		var dev config.ModbusDevice
		if err := decode(r.Body, &dev); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := validateDevice(dev); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.gw.store.UpsertDevice(dev); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Device added: "+dev.ID)
		writeOK(w, dev)
	default:
		writeFail(w, http.StatusMethodNotAllowed, "GET or POST")
	}
}

func (s *Server) handleDevice(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/devices/")
	if id == "" {
		writeFail(w, http.StatusBadRequest, "missing device id")
		return
	}
	switch r.Method {
	case http.MethodPut:
		var dev config.ModbusDevice
		if err := decode(r.Body, &dev); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		dev.ID = id
		if err := validateDevice(dev); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.gw.store.UpsertDevice(dev); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Device updated: "+id)
		writeOK(w, dev)
	case http.MethodDelete:
		if err := s.gw.store.DeleteDevice(id); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Device deleted: "+id)
		writeOK(w, nil)
	default:
		writeFail(w, http.StatusMethodNotAllowed, "PUT or DELETE")
	}
}

// --- Mappings ---

func (s *Server) handleMappings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeOK(w, s.gw.store.Get().Mappings)
	case http.MethodPost:
		var sig config.SignalMapping
		if err := decode(r.Body, &sig); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if sig.ID == "" {
			sig.ID = newID()
		}
		if err := validateMapping(sig, s.gw.store.Get()); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.gw.store.UpsertMapping(sig); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Mapping added: "+sig.MetricName)
		writeOK(w, sig)
	default:
		writeFail(w, http.StatusMethodNotAllowed, "GET or POST")
	}
}

func (s *Server) handleMapping(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/mappings/")
	if id == "" || id == "export" || id == "import" {
		return
	}
	switch r.Method {
	case http.MethodPut:
		var sig config.SignalMapping
		if err := decode(r.Body, &sig); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		sig.ID = id
		if err := validateMapping(sig, s.gw.store.Get()); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.gw.store.UpsertMapping(sig); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Mapping updated: "+sig.MetricName)
		writeOK(w, sig)
	case http.MethodDelete:
		if err := s.gw.store.DeleteMapping(id); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Mapping deleted: "+id)
		writeOK(w, nil)
	default:
		writeFail(w, http.StatusMethodNotAllowed, "PUT or DELETE")
	}
}

func (s *Server) handleMappingsExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeFail(w, http.StatusMethodNotAllowed, "GET only")
		return
	}
	mappings := s.gw.store.Get().Mappings
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="mappings.json"`)
	_ = json.NewEncoder(w).Encode(mappings)
}

func (s *Server) handleMappingsImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeFail(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	var mappings []config.SignalMapping
	if err := decode(r.Body, &mappings); err != nil {
		writeFail(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	cfg := s.gw.store.Get()
	for i := range mappings {
		if mappings[i].ID == "" {
			mappings[i].ID = newID()
		}
		if err := validateMapping(mappings[i], cfg); err != nil {
			writeFail(w, http.StatusBadRequest, fmt.Sprintf("mapping[%d] %q: %v", i, mappings[i].MetricName, err))
			return
		}
	}
	if err := s.gw.store.ReplaceAllMappings(mappings); err != nil {
		writeFail(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.gw.logEvent("info", fmt.Sprintf("Imported %d mappings", len(mappings)))
	writeOK(w, map[string]int{"imported": len(mappings)})
}

// --- Gateway start/stop ---

func (s *Server) handleGatewayStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeFail(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	if s.gw.pub != nil {
		s.gw.pub.Stop()
	}
	cfg := s.gw.store.Get()
	pub := publisher.New(cfg)
	pub.OnLog = func(level, msg string) {
		s.gw.logEvent(level, msg)
	}
	pub.OnStatus = func(st publisher.Status) {
		s.gw.hub.Broadcast(WSEvent{Type: EventStatus, Payload: st})
	}
	if err := pub.Start(context.Background()); err != nil {
		writeFail(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.gw.pub = pub
	s.gw.logEvent("info", "Gateway started")
	writeOK(w, pub.Status())
}

func (s *Server) handleGatewayStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeFail(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	if s.gw.pub != nil {
		s.gw.pub.Stop()
		s.gw.pub = nil
	}
	s.gw.logEvent("info", "Gateway stopped")
	writeOK(w, nil)
}

// --- Validation ---

func validateMQTT(cfg config.MQTTConfig) error {
	if cfg.Broker == "" {
		return fmt.Errorf("broker is required")
	}
	if cfg.ClientID == "" {
		return fmt.Errorf("clientId is required")
	}
	return nil
}

func validateSparkplug(cfg config.SparkplugConfig) error {
	if cfg.GroupID == "" {
		return fmt.Errorf("groupId is required")
	}
	if cfg.NodeID == "" {
		return fmt.Errorf("nodeId is required")
	}
	return nil
}

func validateDevice(dev config.ModbusDevice) error {
	if dev.ID == "" {
		return fmt.Errorf("id is required")
	}
	if dev.Host == "" {
		return fmt.Errorf("host is required")
	}
	if dev.Port == 0 {
		dev.Port = 502
	}
	if dev.Port < 1 || dev.Port > 65535 {
		return fmt.Errorf("port must be 1-65535")
	}
	return nil
}

var validFunctions = map[string]bool{
	"coil": true, "discrete_input": true,
	"input_register": true, "holding_register": true,
}

var validDataTypes = map[string]bool{
	"bool": true, "int16": true, "uint16": true,
	"int32": true, "uint32": true, "float32": true,
	"int64": true, "uint64": true, "float64": true,
}

var validByteOrders = map[string]bool{
	"": true, "ABCD": true, "DCBA": true, "BADC": true, "CDAB": true,
}

func validateMapping(sig config.SignalMapping, cfg config.AppConfig) error {
	if sig.MetricName == "" {
		return fmt.Errorf("metricName is required")
	}
	if !validFunctions[sig.Function] {
		return fmt.Errorf("function must be coil|discrete_input|input_register|holding_register")
	}
	if !validDataTypes[sig.DataType] {
		return fmt.Errorf("invalid dataType %q", sig.DataType)
	}
	if !validByteOrders[sig.ByteOrder] {
		return fmt.Errorf("byteOrder must be ABCD|DCBA|BADC|CDAB")
	}
	if sig.ScanRateMs != 0 && sig.ScanRateMs < 100 {
		return fmt.Errorf("scanRateMs must be >= 100 (got %d)", sig.ScanRateMs)
	}
	// Check referenced device exists
	if sig.ModbusDeviceID != "" {
		found := false
		for _, d := range cfg.Devices {
			if d.ID == sig.ModbusDeviceID {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("modbusDeviceId %q not found", sig.ModbusDeviceID)
		}
	}
	return nil
}

// --- Helpers ---

func decode(r io.Reader, v any) error {
	return json.NewDecoder(io.LimitReader(r, 1<<20)).Decode(v)
}

func writeOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(APIResponse{Success: true, Data: data})
}

func writeFail(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIResponse{Success: false, Message: msg})
}

func newID() string {
	// Simple UUID-like ID using current time + random suffix.
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
