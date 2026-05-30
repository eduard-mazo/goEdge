package config

// AppConfig is the top-level persisted configuration.
type AppConfig struct {
	MQTT        MQTTConfig       `json:"mqtt"`
	Sparkplug   SparkplugConfig  `json:"sparkplug"`
	Outstations []DNP3Outstation `json:"outstations"`
	Mappings    []SignalMapping  `json:"mappings"`
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
	ID            string  `json:"id"`            // uuid
	MetricName    string  `json:"metricName"`    // unique metric name
	DeviceID      string  `json:"deviceId"`      // Sparkplug device ID; empty = node metric
	OutstationID  string  `json:"outstationId"`  // ref to DNP3Outstation.ID

	// DNP3 point identity.
	PointType string `json:"pointType"` // binary|double_bit_binary|binary_output_status|counter|frozen_counter|analog|analog_output_status|octet_string
	Index     uint16 `json:"index"`     // point index within its type
	EventClass uint8 `json:"eventClass"` // 0=static-only, 1|2|3 = event class assignment (informational/UI; outstation configures this)

	// Engineering value transform (kept for analog scaling at gateway side if outstation reports raw counts).
	Scale           float64 `json:"scale"`           // multiplier; default 1.0
	Offset          float64 `json:"offset"`          // addend after scale; default 0.0
	EngineeringUnit string  `json:"engineeringUnit"`

	// Publish-time controls.
	Deadband      float64 `json:"deadband"`      // min change to publish (post-scale); 0 = always publish on event
	PublishOnPoll bool    `json:"publishOnPoll"` // also publish static reads (not only events)

	Enabled bool `json:"enabled"`
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
		Outstations: []DNP3Outstation{},
		Mappings:    []SignalMapping{},
	}
}
