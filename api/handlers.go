package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"goMqttDnp3/config"
	"goMqttDnp3/publisher"
	"goMqttDnp3/sysmon"
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
	mux *http.ServeMux
	gw  *gatewayImpl
}

// NewServer constructs the API server.
func NewServer(store *config.Store, hub *Hub, staticFS http.Handler) *Server {
	gw := &gatewayImpl{store: store, hub: hub}
	s := &Server{mux: http.NewServeMux(), gw: gw}
	s.routes(staticFS)
	return s
}

// Close stops the running publisher (sending NDEATH and disconnecting MQTT) if
// one is active. Call during graceful shutdown so the broker sees a clean death.
func (s *Server) Close() {
	if s.gw.pub != nil {
		s.gw.pub.Stop()
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
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
	s.mux.HandleFunc("/api/config/system", s.handleSystem)
	s.mux.HandleFunc("/api/system/telemetry", s.handleSystemTelemetry)

	s.mux.HandleFunc("/api/outstations", s.handleOutstations)
	s.mux.HandleFunc("/api/outstations/", s.handleOutstation)

	s.mux.HandleFunc("/api/modbusDevices", s.handleModbusDevices)
	s.mux.HandleFunc("/api/modbusDevices/", s.handleModbusDevice)

	s.mux.HandleFunc("/api/serialDevices", s.handleSerialDevices)
	s.mux.HandleFunc("/api/serialDevices/", s.handleSerialDevice)

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

// --- Health / Status ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeOK(w, map[string]string{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
}

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

func (s *Server) handleSystem(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeOK(w, s.gw.store.Get().System)
	case http.MethodPut:
		var cfg config.SystemConfig
		if err := decode(r.Body, &cfg); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := validateSystem(&cfg); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.gw.store.UpdateSystem(cfg); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		// Hot-apply to the running gateway so changes take effect without a
		// restart; when stopped, New() picks it up on the next Start.
		if s.gw.pub != nil {
			s.gw.pub.ApplySystemConfig(cfg)
		}
		s.gw.logEvent("info", "System monitoring config updated")
		writeOK(w, cfg)
	default:
		writeFail(w, http.StatusMethodNotAllowed, "GET or PUT")
	}
}

// handleSystemTelemetry returns host telemetry as a flat metric-name → value
// map. When the publisher is running with system monitoring enabled it returns
// the publisher's own live sample; otherwise it samples on demand via the
// collector (reading /proc and /sys directly), so it still works when the
// gateway is stopped or the MQTT broker is unreachable.
//
// The two paths are deliberate: gopsutil's cpu.Percent(0) keeps its previous
// sample in a package-global, so sampling a second collector here while the
// publisher is also sampling corrupts the shared CPU baseline and pegs the
// publisher's CPU% (and thus the UI) at 100. Reusing its readings avoids a
// competing caller entirely.
func (s *Server) handleSystemTelemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeFail(w, http.StatusMethodNotAllowed, "GET only")
		return
	}
	cfg := s.gw.store.Get().System
	prefix := cfg.MetricPrefix
	if prefix == "" {
		prefix = "System/"
	}

	var st publisher.Status
	if s.gw.pub != nil {
		st = s.gw.pub.Status()
	}

	values := make(map[string]float64)
	if st.Running && cfg.Enabled {
		// Publisher is sampling the host already — reuse its readings.
		for k, v := range st.LastReadings {
			if strings.HasPrefix(k, prefix) {
				values[k] = v
			}
		}
	} else {
		// Nothing else is sampling; safe to read on demand.
		for _, m := range sysmon.New(cfg).Collect(uint64(time.Now().UnixMilli())) {
			if m.DoubleValue != nil {
				values[m.Name] = *m.DoubleValue
			}
		}
	}

	writeOK(w, map[string]any{
		"time":    time.Now().UTC().Format(time.RFC3339),
		"metrics": values,
	})
}

// --- Outstations ---

func (s *Server) handleOutstations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeOK(w, s.gw.store.Get().Outstations)
	case http.MethodPost:
		var o config.DNP3Outstation
		if err := decode(r.Body, &o); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := validateOutstation(o); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.gw.store.UpsertOutstation(o); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Outstation added: "+o.ID)
		writeOK(w, o)
	default:
		writeFail(w, http.StatusMethodNotAllowed, "GET or POST")
	}
}

func (s *Server) handleOutstation(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/outstations/")
	if id == "" {
		writeFail(w, http.StatusBadRequest, "missing outstation id")
		return
	}
	switch r.Method {
	case http.MethodPut:
		var o config.DNP3Outstation
		if err := decode(r.Body, &o); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		o.ID = id
		if err := validateOutstation(o); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.gw.store.UpsertOutstation(o); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Outstation updated: "+id)
		writeOK(w, o)
	case http.MethodDelete:
		if err := s.gw.store.DeleteOutstation(id); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Outstation deleted: "+id)
		writeOK(w, nil)
	default:
		writeFail(w, http.StatusMethodNotAllowed, "PUT or DELETE")
	}
}

// --- Modbus devices ---

func (s *Server) handleModbusDevices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeOK(w, s.gw.store.Get().ModbusDevices)
	case http.MethodPost:
		var d config.ModbusDevice
		if err := decode(r.Body, &d); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := validateModbusDevice(d); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.gw.store.UpsertModbusDevice(d); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Modbus device added: "+d.ID)
		writeOK(w, d)
	default:
		writeFail(w, http.StatusMethodNotAllowed, "GET or POST")
	}
}

func (s *Server) handleModbusDevice(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/modbusDevices/")
	if id == "" {
		writeFail(w, http.StatusBadRequest, "missing device id")
		return
	}
	switch r.Method {
	case http.MethodPut:
		var d config.ModbusDevice
		if err := decode(r.Body, &d); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		d.ID = id
		if err := validateModbusDevice(d); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.gw.store.UpsertModbusDevice(d); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Modbus device updated: "+id)
		writeOK(w, d)
	case http.MethodDelete:
		if err := s.gw.store.DeleteModbusDevice(id); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Modbus device deleted: "+id)
		writeOK(w, nil)
	default:
		writeFail(w, http.StatusMethodNotAllowed, "PUT or DELETE")
	}
}

// --- Serial (Modbus RTU / RS-485) devices ---

func (s *Server) handleSerialDevices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeOK(w, s.gw.store.Get().SerialDevices)
	case http.MethodPost:
		var d config.SerialDevice
		if err := decode(r.Body, &d); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := validateSerialDevice(d); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.gw.store.UpsertSerialDevice(d); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Serial device added: "+d.ID)
		writeOK(w, d)
	default:
		writeFail(w, http.StatusMethodNotAllowed, "GET or POST")
	}
}

func (s *Server) handleSerialDevice(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/serialDevices/")
	if id == "" {
		writeFail(w, http.StatusBadRequest, "missing device id")
		return
	}
	switch r.Method {
	case http.MethodPut:
		var d config.SerialDevice
		if err := decode(r.Body, &d); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		d.ID = id
		if err := validateSerialDevice(d); err != nil {
			writeFail(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.gw.store.UpsertSerialDevice(d); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Serial device updated: "+id)
		writeOK(w, d)
	case http.MethodDelete:
		if err := s.gw.store.DeleteSerialDevice(id); err != nil {
			writeFail(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.gw.logEvent("info", "Serial device deleted: "+id)
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
	pub.LogSink = func(level, msg string) {
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

// validateSystem normalizes and bounds-checks the system-monitoring config,
// filling defaults in place so a sparse PUT (e.g. just {"enabled":true}) is valid.
func validateSystem(cfg *config.SystemConfig) error {
	if cfg.IntervalMs == 0 {
		cfg.IntervalMs = 5000
	}
	if cfg.IntervalMs < 500 {
		return fmt.Errorf("intervalMs must be >= 500")
	}
	if cfg.MetricPrefix == "" {
		cfg.MetricPrefix = "System/"
	}
	if len(cfg.Mounts) == 0 {
		cfg.Mounts = []string{"/"}
	}
	return nil
}

func validateOutstation(o config.DNP3Outstation) error {
	if o.ID == "" {
		return fmt.Errorf("id is required")
	}
	if o.Host == "" {
		return fmt.Errorf("host is required")
	}
	if o.Port == 0 {
		o.Port = 20000
	}
	if o.Port < 1 || o.Port > 65535 {
		return fmt.Errorf("port must be 1-65535")
	}
	if o.MasterAddress == 0 {
		return fmt.Errorf("masterAddress is required (typical: 1)")
	}
	if o.OutstationAddress == 0 {
		return fmt.Errorf("outstationAddress is required")
	}
	return nil
}

func validateModbusDevice(d config.ModbusDevice) error {
	if d.ID == "" {
		return fmt.Errorf("id is required")
	}
	if d.Host == "" {
		return fmt.Errorf("host is required")
	}
	if d.Port != 0 && (d.Port < 1 || d.Port > 65535) {
		return fmt.Errorf("port must be 1-65535")
	}
	switch d.Transport {
	case "", "tcp", "rtuovertcp":
	default:
		return fmt.Errorf("transport must be \"tcp\" or \"rtuovertcp\"")
	}
	return nil
}

var validSerialParity = map[string]bool{"": true, "N": true, "E": true, "O": true}

func validateSerialDevice(d config.SerialDevice) error {
	if d.ID == "" {
		return fmt.Errorf("id is required")
	}
	if d.Port == "" {
		return fmt.Errorf("port is required (e.g. /dev/ttyS1)")
	}
	if d.UnitID < 1 || d.UnitID > 247 {
		return fmt.Errorf("unitId must be 1-247")
	}
	if d.BaudRate != 0 && d.BaudRate < 1 {
		return fmt.Errorf("baudRate must be positive")
	}
	if d.DataBits != 0 && (d.DataBits < 5 || d.DataBits > 8) {
		return fmt.Errorf("dataBits must be 5-8")
	}
	if d.StopBits != 0 && d.StopBits != 1 && d.StopBits != 2 {
		return fmt.Errorf("stopBits must be 1 or 2")
	}
	if !validSerialParity[d.Parity] {
		return fmt.Errorf("parity must be N, E or O")
	}
	return nil
}

var validPointTypes = map[string]bool{
	"binary":               true,
	"double_bit_binary":    true,
	"binary_output_status": true,
	"counter":              true,
	"frozen_counter":       true,
	"analog":               true,
	"analog_output_status": true,
	"octet_string":         true,
}

var validModbusFunctions = map[string]bool{
	"coil":             true,
	"discrete_input":   true,
	"input_register":   true,
	"holding_register": true,
}

var validModbusDataTypes = map[string]bool{
	"bool": true, "int16": true, "uint16": true,
	"int32": true, "uint32": true, "float32": true, "float64": true,
}

func validateMapping(sig config.SignalMapping, cfg config.AppConfig) error {
	if sig.MetricName == "" {
		return fmt.Errorf("metricName is required")
	}
	// Contract v3 catalog limits: codigo_senal VARCHAR(20), nombre_instancia
	// VARCHAR(30) on the consumer side.
	if len([]rune(sig.SignalCode)) > 20 {
		return fmt.Errorf("signalCode must be at most 20 characters")
	}
	if len([]rune(sig.Instance)) > 30 {
		return fmt.Errorf("instance must be at most 30 characters")
	}

	if sig.UsesModbusFraming() {
		if !validModbusFunctions[sig.Function] {
			return fmt.Errorf("function must be one of coil|discrete_input|input_register|holding_register")
		}
		if sig.DataType != "" && !validModbusDataTypes[sig.DataType] {
			return fmt.Errorf("dataType must be one of bool|int16|uint16|int32|uint32|float32|float64")
		}
		if sig.Src() == "" {
			return fmt.Errorf("sourceId is required for modbus mappings")
		}
		// Modbus/TCP mappings reference a ModbusDevice; Modbus RTU mappings
		// reference a SerialDevice. The point identity is otherwise identical.
		found := false
		if sig.IsModbusRTU() {
			for _, d := range cfg.SerialDevices {
				if d.ID == sig.Src() {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("serial device %q not found", sig.Src())
			}
		} else {
			for _, d := range cfg.ModbusDevices {
				if d.ID == sig.Src() {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("modbus device %q not found", sig.Src())
			}
		}
		return nil
	}

	// DNP3 mapping
	if !validPointTypes[sig.PointType] {
		return fmt.Errorf("pointType must be one of binary|double_bit_binary|binary_output_status|counter|frozen_counter|analog|analog_output_status|octet_string")
	}
	if sig.EventClass > 3 {
		return fmt.Errorf("eventClass must be 0..3")
	}
	if sig.Src() != "" {
		found := false
		for _, o := range cfg.Outstations {
			if o.ID == sig.Src() {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("outstationId %q not found", sig.Src())
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
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
