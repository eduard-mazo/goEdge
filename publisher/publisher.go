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
	"os"
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"goMqttDnp3/config"
	"goMqttDnp3/dnp3"
	"goMqttDnp3/mapping"
	"goMqttDnp3/source"
	"goMqttDnp3/sparkplug"
)

// Status snapshots the publisher's runtime state for the API.
type Status struct {
	Running       bool                             `json:"running"`
	MQTTConnected bool                             `json:"mqttConnected"`
	BdSeq         uint64                           `json:"bdSeq"`
	Outstations   map[string]source.Status         `json:"outstations"`
	PublishCount  int64                            `json:"publishCount"`
	ErrorCount    int64                            `json:"errorCount"`
	DroppedCount  int64                            `json:"droppedCount"` // samples shed when the ingest queue was full
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
	droppedCount atomic.Int64

	// ingest decouples protocol callback threads (opendnp3 strands / Modbus
	// pollers) from the MQTT publish path: a slow or blocked broker must never
	// back-pressure the protocol stack. OnSample does a non-blocking enqueue;
	// ingestLoop drains it on its own goroutine.
	ingest chan source.Sample

	// last published value per metric (for deadband)
	lastMu  sync.RWMutex
	lastVal map[string]float64

	// store-and-forward offline buffer
	bufMu  sync.Mutex
	msgBuf []bufferedMsg
	bufMax int

	// observed outstation status (mirrored from master)
	statusMu sync.RWMutex
	osStatus map[string]source.Status

	// live status push (optional; set before Start)
	OnStatus func(Status)
	LogSink  func(level, msg string)
}

type bufferedMsg struct {
	deviceID string
	metrics  []*sparkplug.Metric
}

// ingestQueueSize bounds the protocol→publish handoff. Sized generously so it
// only sheds under sustained overload (which then shows up as Status.DroppedCount).
const ingestQueueSize = 4096

// New creates a Publisher from the current AppConfig.
func New(cfg config.AppConfig) *Publisher {
	p := &Publisher{
		store:    cfg,
		lastVal:  make(map[string]float64),
		mapIdx:   make(map[string][]config.SignalMapping),
		osStatus: make(map[string]source.Status),
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

	// MQTT clientId must be unique on the broker — duplicates get kicked
	// (mosquitto disconnects the older session). Append the PID so multiple
	// gateways on the same broker coexist; PID is stable within a process
	// lifetime, which Sparkplug needs for bdSeq continuity across reconnects.
	clientID := fmt.Sprintf("%s-%d", mq.ClientID, os.Getpid())

	p.node = sparkplug.NewNode(sp.GroupID, sp.NodeID)
	opts := p.node.NewClientOptions(mq.Broker, clientID, mq.Username, mq.Password)
	// Auto-reconnect on broker drop. Without this, an MQTT blip strands the
	// gateway with a live DNP3 master but no publish path; the offline buffer
	// drains once MQTT comes back.
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetMaxReconnectInterval(30 * time.Second)
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		p.logInfo("MQTT connected to " + mq.Broker + " (clientId=" + clientID + ")")
		p.drainBuffer()
	})
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		p.logWarn("MQTT connection lost: " + err.Error() + " (auto-reconnecting)")
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

	// Start the ingest worker BEFORE the master: startup integrity responses are
	// delivered synchronously during master.Start, so the queue must already be
	// draining or those callbacks would block a protocol thread.
	p.ingest = make(chan source.Sample, ingestQueueSize)
	p.wg.Add(1)
	go p.ingestLoop()

	if err := p.master.Start(pCtx); err != nil {
		// Tear down anything Start brought up (thread pool, partially-added
		// outstations) so a failed start doesn't orphan native resources.
		p.master.Stop()
		cancel()
		close(p.ingest)
		p.wg.Wait()
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
	// master.Stop() returns only once protocol callbacks have ceased, so no
	// goroutine can send on ingest after this — safe to close and drain.
	p.master.Stop()
	close(p.ingest)
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
		DroppedCount: p.droppedCount.Load(),
		Outstations:  make(map[string]source.Status),
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

	// Pull live per-outstation counters straight from the master so
	// MeasurementsRx / LastReadAt reflect every callback, not only the ones
	// that came alongside an OnStatusChange.
	if p.master != nil {
		for _, st := range p.master.Status() {
			s.Outstations[st.ID] = st
		}
	}
	// Overlay any extra fields the publisher cached via OnStatusChange
	// (e.g. LastError that the master doesn't re-emit on every measurement).
	p.statusMu.RLock()
	for k, v := range p.osStatus {
		if existing, ok := s.Outstations[k]; ok {
			if v.LastError != "" && existing.LastError == "" {
				existing.LastError = v.LastError
				s.Outstations[k] = existing
			}
		} else {
			s.Outstations[k] = v
		}
	}
	p.statusMu.RUnlock()

	p.lastMu.RLock()
	for k, v := range p.lastVal {
		s.LastReadings[k] = v
	}
	p.lastMu.RUnlock()
	return s
}

// --- source.Handler ---

// OnSample is called by a Source (DNP3 master, Modbus poller, …) for every point
// update from a protocol thread. It must not block: it does a non-blocking
// handoff to the ingest queue and returns immediately. If the queue is full
// (publish path wedged), the sample is shed and counted rather than stalling the
// protocol stack.
func (p *Publisher) OnSample(m source.Sample) {
	select {
	case p.ingest <- m:
	default:
		p.droppedCount.Add(1)
	}
}

// ingestLoop drains the queue on its own goroutine, doing mapping + deadband +
// publish off the protocol threads. Exits when ingest is closed (Stop).
func (p *Publisher) ingestLoop() {
	defer p.wg.Done()
	for m := range p.ingest {
		p.handleSample(m)
	}
}

// handleSample maps one measurement to its Sparkplug metric(s) and publishes
// (or buffers) it. Runs only on the ingest goroutine.
func (p *Publisher) handleSample(m source.Sample) {
	sigs := p.mappingsFor(m.SourceID)
	if len(sigs) == 0 {
		return
	}
	for _, sig := range sigs {
		if sig.PointType != string(m.PointType) || sig.Index != m.Index {
			continue
		}
		// By default publish only change events; static/poll responses (integrity
		// and static polls) publish only when the mapping opts in via PublishOnPoll.
		if !m.IsEvent && !sig.PublishOnPoll {
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
func (p *Publisher) OnStatusChange(s source.Status) {
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
