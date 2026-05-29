package config

// AppConfig is the top-level persisted configuration.
type AppConfig struct {
	MQTT      MQTTConfig      `json:"mqtt"`
	Sparkplug SparkplugConfig `json:"sparkplug"`
	Devices   []ModbusDevice  `json:"devices"`
	Mappings  []SignalMapping  `json:"mappings"`
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

// ModbusDevice represents a Modbus TCP target device.
type ModbusDevice struct {
	ID         string `json:"id"`         // unique slug (user-defined)
	Label      string `json:"label"`      // human-readable name
	Host       string `json:"host"`
	Port       int    `json:"port"`       // default 502
	TimeoutMs  int    `json:"timeoutMs"`  // connect + read timeout; default 3000
	Retries    int    `json:"retries"`    // per-read retry count; default 2
	RetryDelayMs int  `json:"retryDelayMs"` // delay between retries; default 500
	Enabled    bool   `json:"enabled"`
}

// Addr returns "host:port" for the device.
func (d ModbusDevice) Addr() string {
	port := d.Port
	if port == 0 {
		port = 502
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

// SignalMapping maps a Modbus register read to a Sparkplug B metric.
type SignalMapping struct {
	ID              string  `json:"id"`              // uuid
	MetricName      string  `json:"metricName"`      // unique metric name
	DeviceID        string  `json:"deviceId"`        // Sparkplug device ID; empty = node metric
	ModbusDeviceID  string  `json:"modbusDeviceId"`  // ref to ModbusDevice.ID
	UnitID          uint8   `json:"unitId"`
	Function        string  `json:"function"`        // coil|discrete_input|input_register|holding_register
	Address         uint16  `json:"address"`
	Quantity        uint16  `json:"quantity"`        // registers/coils; must match dataType
	DataType        string  `json:"dataType"`        // bool|int16|uint16|int32|uint32|int64|uint64|float32|float64
	ByteOrder       string  `json:"byteOrder"`       // ABCD|DCBA|BADC|CDAB
	Scale           float64 `json:"scale"`           // multiplier; default 1.0
	Offset          float64 `json:"offset"`          // addend after scale; default 0.0
	EngineeringUnit string  `json:"engineeringUnit"`
	ScanRateMs      int     `json:"scanRateMs"`      // >= 100
	Deadband        float64 `json:"deadband"`        // min change to publish; 0 = always publish
	QualityPolicy   string  `json:"qualityPolicy"`   // good|bad_on_error|last_known
	Enabled         bool    `json:"enabled"`
}

// DefaultAppConfig returns a config with sane defaults.
func DefaultAppConfig() AppConfig {
	return AppConfig{
		MQTT: MQTTConfig{
			Broker:    "tcp://localhost:1883",
			ClientID:  "goMqttModbus",
			QoS:       1,
			Keepalive: 60,
		},
		Sparkplug: SparkplugConfig{
			GroupID: "plant-floor",
			NodeID:  "modbus-gw",
		},
		Devices:  []ModbusDevice{},
		Mappings: []SignalMapping{},
	}
}
