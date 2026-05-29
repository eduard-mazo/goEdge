// Package publisher orchestrates DNP3 measurement events → Sparkplug B publishing.
//
// Architecture (event-driven, not polled):
//
//	dnp3.Master ── OnMeasurement ──► Publisher.OnMeasurement
//	                                    │
//	                                    ▼
//	                          mapping.Apply (scale + quality)
//	                                    │
//	                                    ▼
//	                          deadband filter (per metric)
//	                                    │
//	                                    ▼
//	                  MQTT publish (NDATA/DDATA)  or  offline buffer
//
// The DNP3 master is responsible for connection lifecycle and polling cadence;
// this package only reacts to measurements.
package publisher

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"goMqttDnp3/config"
	"goMqttDnp3/dnp3"
	"goMqttDnp3/mapping"
	"goMqttDnp3/sparkplug"
)

// Status snapshots the publisher's runtime state for the API.
type Status struct {
	Running       bool                             `json:"running"`
	MQTTConnected bool                             `json:"mqttConnected"`
	BdSeq         uint64                           `json:"bdSeq"`
	Outstations   map[string]dnp3.OutstationStatus `json:"outstations"`
	PublishCount  int64                            `json:"publishCount"`
	ErrorCount    int64                            `json:"errorCount"`
	Uptime        string                           `json:"uptime"`
	LastReadings  map[string]float64               `json:"lastReadings"`
}

// Publisher manages the full lifecycle: MQTT connect → NBIRTH/DBIRTH → react to
// measurements → NDATA/DDATA → NDEATH.
type Publisher struct {
	store  config.AppConfig
	master dnp3.Master
	node   *sparkplug.Node
	client mqtt.Client

	// fast index: outstationID → []SignalMapping for measurement routing.
	mapIdx   map[string][]config.SignalMapping
	mapIdxMu sync.RWMutex

	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running atomic.Bool
	startAt time.Time

	publishCount atomic.Int64
	errorCount   atomic.Int64

	// last published value per metric (for deadband)
	lastMu  sync.RWMutex
	lastVal map[string]float64

	// store-and-forward offline buffer
	bufMu  sync.Mutex
	msgBuf []bufferedMsg
	bufMax int

	// observed outstation status (mirrored from master)
	statusMu sync.RWMutex
	osStatus map[string]dnp3.OutstationStatus

	// live status push (optional; set before Start)
	OnStatus func(Status)
	LogSink  func(level, msg string)
}

type bufferedMsg struct {
	deviceID string
	metrics  []*sparkplug.Metric
}

// New creates a Publisher from the current AppConfig.
func New(cfg config.AppConfig) *Publisher {
	p := &Publisher{
		store:    cfg,
		lastVal:  make(map[string]float64),
		mapIdx:   make(map[string][]config.SignalMapping),
		osStatus: make(map[string]dnp3.OutstationStatus),
		bufMax:   500,
	}
	for _, sig := range cfg.Mappings {
		if !sig.Enabled {
			continue
		}
		p.mapIdx[sig.OutstationID] = append(p.mapIdx[sig.OutstationID], sig)
	}
	p.master = dnp3.New(p)
	return p
}

// Start connects MQTT, sends NBIRTH/DBIRTHs, and starts the DNP3 master.
func (p *Publisher) Start(ctx context.Context) error {
	if p.running.Swap(true) {
		return fmt.Errorf("publisher already running")
	}

	cfg := p.store
	sp := cfg.Sparkplug
	mq := cfg.MQTT

	p.node = sparkplug.NewNode(sp.GroupID, sp.NodeID)
	opts := p.node.NewClientOptions(mq.Broker, mq.ClientID, mq.Username, mq.Password)
	opts.SetAutoReconnect(false)
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		p.logInfo("MQTT connected to " + mq.Broker)
		p.drainBuffer()
	})
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		p.logWarn("MQTT connection lost: " + err.Error())
	})

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if err := token.Error(); err != nil {
		p.running.Store(false)
		return fmt.Errorf("MQTT connect %s: %w", mq.Broker, err)
	}
	p.client = client

	birthMetrics := p.buildBirthMetrics()
	if err := p.node.PublishNBirth(client, birthMetrics); err != nil {
		p.logWarn("NBIRTH failed: " + err.Error())
	}

	p.node.SubscribeNCMD(client, func() {
		p.logInfo("Rebirth requested, re-publishing NBIRTH")
		p.node.PublishNBirth(client, p.buildBirthMetrics())
		p.publishDBirths()
	})

	p.publishDBirths()

	for _, o := range cfg.Outstations {
		if !o.Enabled {
			continue
		}
		if err := p.master.AddOutstation(o); err != nil {
			p.logWarn(fmt.Sprintf("AddOutstation %s: %v", o.ID, err))
		}
	}

	pCtx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.startAt = time.Now()

	if err := p.master.Start(pCtx); err != nil {
		p.running.Store(false)
		return fmt.Errorf("DNP3 master start: %w", err)
	}

	p.logInfo(fmt.Sprintf("Publisher started: %d outstations, %d mappings", len(cfg.Outstations), len(cfg.Mappings)))
	return nil
}

// Stop tears down the master and sends NDEATH.
func (p *Publisher) Stop() {
	if !p.running.Load() {
		return
	}
	if p.cancel != nil {
		p.cancel()
	}
	p.master.Stop()
	p.wg.Wait()
	if p.client != nil && p.client.IsConnected() {
		if err := p.node.PublishNDeath(p.client); err != nil {
			p.logWarn("NDEATH failed: " + err.Error())
		}
		p.client.Disconnect(500)
	}
	p.running.Store(false)
	p.logInfo("Publisher stopped")
}

// Status returns a snapshot of the publisher's current state.
func (p *Publisher) Status() Status {
	s := Status{
		Running:      p.running.Load(),
		PublishCount: p.publishCount.Load(),
		ErrorCount:   p.errorCount.Load(),
		Outstations:  make(map[string]dnp3.OutstationStatus),
		LastReadings: make(map[string]float64),
	}
	if p.client != nil {
		s.MQTTConnected = p.client.IsConnected()
	}
	if p.node != nil {
		s.BdSeq = p.node.Bdseq()
	}
	if !p.startAt.IsZero() {
		s.Uptime = time.Since(p.startAt).Round(time.Second).String()
	}

	p.statusMu.RLock()
	for k, v := range p.osStatus {
		s.Outstations[k] = v
	}
	p.statusMu.RUnlock()

	p.lastMu.RLock()
	for k, v := range p.lastVal {
		s.LastReadings[k] = v
	}
	p.lastMu.RUnlock()
	return s
}

// --- dnp3.Handler ---

// OnMeasurement is called by the DNP3 master for every point update (event or
// integrity response). Must return quickly; mapping + MQTT publish run inline.
func (p *Publisher) OnMeasurement(m dnp3.Measurement) {
	sigs := p.mappingsFor(m.OutstationID)
	if len(sigs) == 0 {
		return
	}
	for _, sig := range sigs {
		if sig.PointType != string(m.PointType) || sig.Index != m.Index {
			continue
		}
		result, err := mapping.Apply(sig, m)
		if err != nil {
			p.errorCount.Add(1)
			slog.Warn("mapping error", "metric", sig.MetricName, "err", err)
			continue
		}
		if !p.passDeadband(sig, result.Value) {
			continue
		}
		if err := p.publishOrBuffer(sig.DeviceID, []*sparkplug.Metric{result.Metric}); err != nil {
			p.errorCount.Add(1)
			slog.Warn("publish failed", "metric", sig.MetricName, "err", err)
		}
	}
}

// OnStatusChange mirrors the master's per-outstation status into our snapshot.
func (p *Publisher) OnStatusChange(s dnp3.OutstationStatus) {
	p.statusMu.Lock()
	p.osStatus[s.ID] = s
	p.statusMu.Unlock()
	if p.OnStatus != nil {
		p.OnStatus(p.Status())
	}
}

// OnLog forwards DNP3 master log messages.
func (p *Publisher) OnLog(level, msg string) {
	switch level {
	case "warn", "error":
		p.logWarn(msg)
	default:
		p.logInfo(msg)
	}
}

// --- internals ---

func (p *Publisher) mappingsFor(outstationID string) []config.SignalMapping {
	p.mapIdxMu.RLock()
	defer p.mapIdxMu.RUnlock()
	return p.mapIdx[outstationID]
}

func (p *Publisher) passDeadband(sig config.SignalMapping, cur float64) bool {
	p.lastMu.Lock()
	defer p.lastMu.Unlock()
	last, ok := p.lastVal[sig.MetricName]
	if !ok {
		last = math.NaN()
	}
	if mapping.CheckDeadband(cur, last, sig.Deadband) {
		p.lastVal[sig.MetricName] = cur
		return true
	}
	return false
}

func (p *Publisher) publishOrBuffer(deviceID string, metrics []*sparkplug.Metric) error {
	if p.client == nil || !p.client.IsConnected() {
		p.enqueue(deviceID, metrics)
		return nil
	}
	p.publishCount.Add(1)
	if deviceID == "" {
		return p.node.PublishNData(p.client, metrics)
	}
	return p.node.PublishDData(p.client, deviceID, metrics)
}

func (p *Publisher) enqueue(deviceID string, metrics []*sparkplug.Metric) {
	p.bufMu.Lock()
	defer p.bufMu.Unlock()
	if len(p.msgBuf) >= p.bufMax {
		p.msgBuf = p.msgBuf[1:]
	}
	cp := make([]*sparkplug.Metric, len(metrics))
	copy(cp, metrics)
	p.msgBuf = append(p.msgBuf, bufferedMsg{deviceID: deviceID, metrics: cp})
}

func (p *Publisher) drainBuffer() {
	p.bufMu.Lock()
	msgs := make([]bufferedMsg, len(p.msgBuf))
	copy(msgs, p.msgBuf)
	p.msgBuf = nil
	p.bufMu.Unlock()

	if len(msgs) == 0 {
		return
	}
	slog.Info("draining offline buffer", "count", len(msgs))
	for i := range msgs {
		_ = p.publishOrBuffer(msgs[i].deviceID, msgs[i].metrics)
	}
}

func (p *Publisher) buildBirthMetrics() []*sparkplug.Metric {
	ts := sparkplug.NowMs()
	var metrics []*sparkplug.Metric
	for _, sig := range p.store.Mappings {
		if !sig.Enabled || sig.DeviceID != "" {
			continue
		}
		metrics = append(metrics, sparkplug.MetricDouble(sig.MetricName, ts, 0))
	}
	return metrics
}

func (p *Publisher) publishDBirths() {
	ts := sparkplug.NowMs()
	devMetrics := make(map[string][]*sparkplug.Metric)
	for _, sig := range p.store.Mappings {
		if !sig.Enabled || sig.DeviceID == "" {
			continue
		}
		devMetrics[sig.DeviceID] = append(devMetrics[sig.DeviceID], sparkplug.MetricDouble(sig.MetricName, ts, 0))
	}
	for devID, metrics := range devMetrics {
		if err := p.node.PublishDBirth(p.client, devID, metrics); err != nil {
			p.logWarn(fmt.Sprintf("DBIRTH %s failed: %v", devID, err))
		}
	}
}

func (p *Publisher) logInfo(msg string) {
	slog.Info(msg)
	if p.LogSink != nil {
		p.LogSink("info", msg)
	}
}

func (p *Publisher) logWarn(msg string) {
	slog.Warn(msg)
	if p.LogSink != nil {
		p.LogSink("warn", msg)
	}
}
