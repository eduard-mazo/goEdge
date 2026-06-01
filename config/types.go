package config

// AppConfig is the top-level persisted configuration.
//
// Sources are split by protocol: Outstations (DNP3) and ModbusDevices. Each
// SignalMapping carries a Protocol discriminator and references its source by
// SourceID; the publisher runs one source.Source per protocol and routes
// samples to mappings by source ID.
type AppConfig struct {
	MQTT          MQTTConfig       `json:"mqtt"`
	Sparkplug     SparkplugConfig  `json:"sparkplug"`
	Outstations   []DNP3Outstation `json:"outstations"`
	ModbusDevices []ModbusDevice   `json:"modbusDevices"`
	Mappings      []SignalMapping  `json:"mappings"`
	System        SystemConfig     `json:"system"`
}

// SystemConfig controls host-telemetry collection (CPU, memory, disk, network,
// temperature, uptime). When Enabled, the publisher samples the host every
// IntervalMs and publishes the result as Sparkplug node metrics under
// MetricPrefix. These are zero-config — no SignalMapping is required.
type SystemConfig struct {
	Enabled       bool          `json:"enabled"`
	IntervalMs    int           `json:"intervalMs"`    // poll cadence; default 5000
	MetricPrefix  string        `json:"metricPrefix"`  // folder prefix; default "System/"
	Mounts        []string      `json:"mounts"`        // filesystems to report; default ["/"]
	Interfaces    []string      `json:"interfaces"`    // NICs to report; empty = all non-loopback
	TempSensorKey string        `json:"tempSensorKey"` // sensor-key substring to prefer; empty = auto
	Metrics       SystemMetrics `json:"metrics"`       // which metric groups to publish

	// DisabledMetrics lists individual metric suffixes to exclude even when their
	// group is enabled (e.g. "Memory/Used_MB"). Disk/network entries use the
	// "<mount>"/"<iface>" placeholders, e.g. "Disk/<mount>/Free_MB", which match
	// every concrete mount/interface. Suffixes are relative to MetricPrefix.
	DisabledMetrics []string `json:"disabledMetrics"`
}

// SystemMetrics selects which host-telemetry groups the collector publishes.
// As a safety net, if every flag is false the collector treats it as "all on"
// (see Effective) — so older configs without this block, or a sparse PUT,
// still publish everything rather than silently nothing.
type SystemMetrics struct {
	CPU          bool `json:"cpu"`          // CPU/Usage_pct
	Load         bool `json:"load"`         // CPU/Load1, Load5, Load15
	Memory       bool `json:"memory"`       // Memory/Used_pct, Used_MB, Available_MB, Total_MB
	Swap         bool `json:"swap"`         // Memory/Swap_Used_pct
	Disk         bool `json:"disk"`         // Disk/<mount>/Used_pct, Free_MB
	Network      bool `json:"network"`      // Network/<iface>/Rx_MB, Tx_MB (totals)
	NetworkRates bool `json:"networkRates"` // Network/<iface>/RxRate_kbps, TxRate_kbps
	Temperature  bool `json:"temperature"`  // Temperature/CPU_C
	Uptime       bool `json:"uptime"`       // Uptime_h
	Processes    bool `json:"processes"`    // Process/Count
}

// Any reports whether at least one metric group is selected.
func (m SystemMetrics) Any() bool {
	return m.CPU || m.Load || m.Memory || m.Swap || m.Disk ||
		m.Network || m.NetworkRates || m.Temperature || m.Uptime || m.Processes
}

// Effective returns the selection to actually use: the configured one, or — when
// nothing is selected — all groups enabled, so the gateway never silently
// publishes zero system metrics while monitoring is on.
func (m SystemMetrics) Effective() SystemMetrics {
	if m.Any() {
		return m
	}
	return SystemMetrics{
		CPU: true, Load: true, Memory: true, Swap: true, Disk: true,
		Network: true, NetworkRates: true, Temperature: true, Uptime: true, Processes: true,
	}
}

// ModbusDevice represents a Modbus/TCP slave reachable over the network.
// One TCP connection per device; reads are serialized and polled on a ticker.
type ModbusDevice struct {
	ID           string `json:"id"`           // unique slug (user-defined)
	Label        string `json:"label"`        // human-readable name
	Host         string `json:"host"`
	Port         int    `json:"port"`         // default 502
	UnitID       uint8  `json:"unitId"`       // Modbus slave/unit id (typical 1)
	ScanRateMs   int    `json:"scanRateMs"`   // poll cadence; default 1000
	TimeoutMs    int    `json:"timeoutMs"`    // per-request timeout; default 3000
	Retries      int    `json:"retries"`      // transient-error retries; default 2
	RetryDelayMs int    `json:"retryDelayMs"` // delay between retries; default 500
	Enabled      bool   `json:"enabled"`
}

// Addr returns "host:port" for the Modbus device.
func (d ModbusDevice) Addr() string {
	port := d.Port
	if port == 0 {
		port = 502
	}
	return d.Host + ":" + itoa(port)
}

// MQTTConfig holds broker connection parameters.
type MQTTConfig struct {
	Broker    string    `json:"broker"`    // tcp://host:port or ssl://host:port
	ClientID  string    `json:"clientId"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	QoS       byte      `json:"qos"`       // 0, 1, or 2
	Keepalive int       `json:"keepalive"` // seconds; 0 = library default
	TLS       TLSConfig `json:"tls"`
}

// TLSConfig holds optional TLS parameters for the MQTT connection.
type TLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CAFile   string `json:"caFile"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
	Insecure bool   `json:"insecure"` // skip server cert verification (dev only)
}

// SparkplugConfig holds Sparkplug B namespace parameters.
type SparkplugConfig struct {
	GroupID             string `json:"groupId"`
	NodeID              string `json:"nodeId"`
	BirthOnConfigChange bool   `json:"birthOnConfigChange"`
}

// DNP3Outstation represents a DNP3 outstation reachable via TCP.
// One TCP channel per (host:port); one association per (master, outstation) link address pair.
type DNP3Outstation struct {
	ID                 string `json:"id"`                 // unique slug (user-defined)
	Label              string `json:"label"`              // human-readable name
	Host               string `json:"host"`
	Port               int    `json:"port"`               // DNP3/IP default 20000
	MasterAddress      uint16 `json:"masterAddress"`      // local link-layer address (typical 1)
	OutstationAddress  uint16 `json:"outstationAddress"`  // remote link-layer address (typical 1024+)
	ResponseTimeoutMs  int    `json:"responseTimeoutMs"`  // app-layer response timeout; default 5000
	KeepAliveMs        int    `json:"keepAliveMs"`        // app-layer keep-alive interval; default 60000

	// Polling cadence (0 = disabled).
	IntegrityScanMs int `json:"integrityScanMs"` // periodic integrity poll (class 0+1+2+3); default 3600000
	Class1ScanMs    int `json:"class1ScanMs"`    // event class 1 poll; default 1000
	Class2ScanMs    int `json:"class2ScanMs"`    // event class 2 poll; default 5000
	Class3ScanMs    int `json:"class3ScanMs"`    // event class 3 poll; default 30000

	// Unsolicited responses.
	UnsolicitedEnabled bool `json:"unsolicitedEnabled"` // send ENABLE_UNSOLICITED on startup
	UnsolicitedClass1  bool `json:"unsolicitedClass1"`
	UnsolicitedClass2  bool `json:"unsolicitedClass2"`
	UnsolicitedClass3  bool `json:"unsolicitedClass3"`

	// Startup behavior.
	DisableUnsolOnStartup bool `json:"disableUnsolOnStartup"` // DISABLE_UNSOLICITED before initial integrity poll
	StartupIntegrity      bool `json:"startupIntegrity"`      // perform integrity poll on connect

	// Startup-integrity class selection. If StartupIntegrity is true but every
	// flag below is false, the lib treats it as "all classes" (DNP3 conformant default).
	// Set IntegrityClass0=false to work around outstations that emit malformed
	// objects (e.g. g50v4 with wrong qualifier) in class-0 responses.
	IntegrityClass0 bool `json:"integrityClass0"`
	IntegrityClass1 bool `json:"integrityClass1"`
	IntegrityClass2 bool `json:"integrityClass2"`
	IntegrityClass3 bool `json:"integrityClass3"`

	// StaticPollMs enables periodic group-specific static reads (one per
	// supported point type, "all objects" qualifier) at this period in ms.
	// 0 = disabled. Use this with outstations that don't flag events on update
	// (the data lives in static groups; class polls return empty).
	StaticPollMs int `json:"staticPollMs"`

	Enabled bool `json:"enabled"`
}

// Addr returns "host:port" for the outstation.
func (d DNP3Outstation) Addr() string {
	port := d.Port
	if port == 0 {
		port = 20000
	}
	return d.Host + ":" + itoa(port)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [10]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}

// SignalMapping maps a DNP3 point to a Sparkplug B metric.
// DNP3 points are addressed by (Group, Variation, Index).
// The library delivers typed measurements; no byte-order/scaling decoding is needed.
type SignalMapping struct {
	ID         string `json:"id"`         // uuid
	MetricName string `json:"metricName"` // unique metric name
	DeviceID   string `json:"deviceId"`   // Sparkplug device ID; empty = node metric

	// Protocol selects how the source point is addressed: "dnp3" (default when
	// empty) or "modbus".
	Protocol string `json:"protocol"`

	// SourceID references the source this point belongs to (DNP3Outstation.ID or
	// ModbusDevice.ID). OutstationID is the legacy DNP3-only alias; Src() resolves
	// SourceID first, falling back to OutstationID for older configs.
	SourceID     string `json:"sourceId"`
	OutstationID string `json:"outstationId"`

	// DNP3 point identity (Protocol=="dnp3").
	PointType  string `json:"pointType"`  // binary|double_bit_binary|binary_output_status|counter|frozen_counter|analog|analog_output_status|octet_string
	Index      uint16 `json:"index"`      // point index within its type
	EventClass uint8  `json:"eventClass"` // 0=static-only, 1|2|3 = event class (informational/UI)

	// Modbus point identity (Protocol=="modbus").
	Function  string `json:"function"`  // coil|discrete_input|input_register|holding_register
	Address   uint16 `json:"address"`   // register/coil start address
	Quantity  uint16 `json:"quantity"`  // registers to read; 0 = derived from DataType
	DataType  string `json:"dataType"`  // bool|int16|uint16|int32|uint32|float32|float64
	ByteOrder string `json:"byteOrder"` // ABCD|DCBA|BADC|CDAB (word/byte order for 32/64-bit)

	// Engineering value transform.
	Scale           float64 `json:"scale"`  // multiplier; default 1.0
	Offset          float64 `json:"offset"` // addend after scale; default 0.0
	EngineeringUnit string  `json:"engineeringUnit"`

	// Publish-time controls.
	Deadband      float64 `json:"deadband"`      // min change to publish (post-scale); 0 = always
	PublishOnPoll bool    `json:"publishOnPoll"` // also publish static/poll reads (not only events)

	Enabled bool `json:"enabled"`
}

// Src returns the source ID this mapping belongs to, preferring the neutral
// SourceID and falling back to the legacy OutstationID.
func (m SignalMapping) Src() string {
	if m.SourceID != "" {
		return m.SourceID
	}
	return m.OutstationID
}

// IsModbus reports whether this mapping addresses a Modbus point.
func (m SignalMapping) IsModbus() bool {
	return m.Protocol == "modbus"
}

// DefaultAppConfig returns a config with sane defaults.
func DefaultAppConfig() AppConfig {
	return AppConfig{
		MQTT: MQTTConfig{
			Broker:    "tcp://localhost:1883",
			ClientID:  "goMqttDnp3",
			QoS:       1,
			Keepalive: 60,
		},
		Sparkplug: SparkplugConfig{
			GroupID: "plant-floor",
			NodeID:  "dnp3-gw",
		},
		Outstations:   []DNP3Outstation{},
		ModbusDevices: []ModbusDevice{},
		Mappings:      []SignalMapping{},
		System: SystemConfig{
			Enabled:      false,
			IntervalMs:   5000,
			MetricPrefix: "System/",
			Mounts:       []string{"/"},
			Metrics: SystemMetrics{
				CPU: true, Load: true, Memory: true, Swap: true, Disk: true,
				Network: true, NetworkRates: true, Temperature: true, Uptime: true, Processes: true,
			},
		},
	}
}
