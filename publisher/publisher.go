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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"maps"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"goMqttDnp3/config"
	"goMqttDnp3/dnp3"
	"goMqttDnp3/mapping"
	"goMqttDnp3/modbus"
	"goMqttDnp3/modbusrtu"
	"goMqttDnp3/source"
	"goMqttDnp3/sparkplug"
	"goMqttDnp3/sysmon"
)

// Status snapshots the publisher's runtime state for the API.
type Status struct {
	Running       bool                     `json:"running"`
	MQTTConnected bool                     `json:"mqttConnected"`
	BdSeq         uint64                   `json:"bdSeq"`
	Outstations   map[string]source.Status `json:"outstations"`
	PublishCount  int64                    `json:"publishCount"`
	ErrorCount    int64                    `json:"errorCount"`
	DroppedCount  int64                    `json:"droppedCount"` // samples shed when the ingest queue was full
	Uptime        string                   `json:"uptime"`
	LastReadings  map[string]float64       `json:"lastReadings"`
	// ServerTime is the gateway clock at snapshot time (UTC RFC3339). The UI
	// anchors source-freshness ages to this instead of the browser clock, so an
	// embedded device whose clock disagrees with the browser still shows correct,
	// live-ticking "last read" labels.
	ServerTime time.Time `json:"serverTime"`
}

// Publisher manages the full lifecycle: MQTT connect → NBIRTH/DBIRTH → react to
// measurements → NDATA/DDATA → NDEATH.
type Publisher struct {
	store config.AppConfig

	// Field-protocol sources, all feeding OnSample. master is kept typed for
	// DNP3-specific ops (AddOutstation/IntegrityPoll); sources is the uniform
	// list used for Start/Stop/Status.
	master    dnp3.Master
	modbus    *modbus.Poller
	modbusRTU *modbusrtu.Poller
	sources   []source.Source

	// outstation is the northbound DNP3 outstation server (nil when disabled): a
	// sink that re-exposes ServeDNP3 mappings so a SCADA master can poll the
	// gateway's aggregated data. Fed from handleSample, not a source.Source.
	outstation dnp3.Outstation

	node   *sparkplug.Node
	client mqtt.Client

	// sysCollector samples host telemetry (CPU/mem/disk/net/temp) when enabled;
	// nil when disabled. Guarded by sysMu so the API can swap it in live
	// (ApplySystemConfig) without restarting the gateway. sysInterval is the
	// current poll cadence; sysReload nudges sysLoop to pick up a new interval.
	sysMu        sync.RWMutex
	sysCollector *sysmon.Collector
	sysInterval  time.Duration
	sysReload    chan struct{}

	// fast index: outstationID → []SignalMapping for measurement routing.
	mapIdx   map[string][]config.SignalMapping
	mapIdxMu sync.RWMutex

	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running atomic.Bool
	startAt time.Time

	// mqttUp is our own view of the live MQTT session, flipped by the connect /
	// connection-lost handlers. We can't use paho's client.IsConnected() for the
	// offline-buffer/status decision: with ConnectRetry/AutoReconnect it returns
	// true while merely *retrying* a down broker, which would defeat
	// store-and-forward. This is true only between OnConnect and OnConnectionLost.
	mqttUp atomic.Bool

	publishCount atomic.Int64
	errorCount   atomic.Int64
	droppedCount atomic.Int64

	// ingest decouples protocol callback threads (opendnp3 strands / Modbus
	// pollers) from the MQTT publish path: a slow or blocked broker must never
	// back-pressure the protocol stack. OnSample does a non-blocking enqueue;
	// ingestLoop drains it on its own goroutine.
	ingest chan source.Sample

	// batchWindow > 0 coalesces metrics arriving within the window into one
	// DDATA/NDATA per device (see ingestLoop); 0 publishes each sample
	// immediately. pending accumulates per-device (deviceID "" = node) metrics
	// between flushes; batchMax caps a device's batch so a single payload can't
	// grow unbounded. pending/batchMax are touched only by the ingest goroutine
	// (and synchronously in tests), so they need no lock.
	batchWindow time.Duration
	batchMax    int
	pending     map[string][]*sparkplug.Metric

	// last published value per metric (for deadband) and per-metric bad-quality
	// state (so a comm-loss marker publishes once and recovery republishes even
	// when the value is unchanged). Both guarded by lastMu.
	lastMu  sync.RWMutex
	lastVal map[string]float64
	qualBad map[string]bool

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

// batchMaxMetrics caps how many metrics accumulate for one device before its
// batch is flushed early, bounding Sparkplug payload size when coalescing is on.
const batchMaxMetrics = 100

// defaultBatchWindow is the coalescing window adopted when PublishBatchMs is
// unset (0). Small enough to be imperceptible at SCADA scan rates (~1s) yet wide
// enough to collapse a whole scan's points into one DDATA/NDATA per device.
const defaultBatchWindow = 200 * time.Millisecond

// New creates a Publisher from the current AppConfig.
func New(cfg config.AppConfig) *Publisher {
	p := &Publisher{
		store:    cfg,
		lastVal:  make(map[string]float64),
		qualBad:  make(map[string]bool),
		mapIdx:   make(map[string][]config.SignalMapping),
		osStatus: make(map[string]source.Status),
		pending:  make(map[string][]*sparkplug.Metric),
		batchMax: batchMaxMetrics,
		bufMax:   500,
	}
	// Coalescing is on by default: a positive value sets the window explicitly, 0
	// (unset) adopts defaultBatchWindow, and a negative value is the explicit
	// opt-out (legacy per-sample publish, batchWindow == 0).
	switch {
	case cfg.MQTT.PublishBatchMs > 0:
		p.batchWindow = time.Duration(cfg.MQTT.PublishBatchMs) * time.Millisecond
	case cfg.MQTT.PublishBatchMs == 0:
		p.batchWindow = defaultBatchWindow
	}
	for _, sig := range cfg.Mappings {
		if !sig.Enabled {
			continue
		}
		sig = normalizeMapping(sig)
		src := sig.Src()
		p.mapIdx[src] = append(p.mapIdx[src], sig)
	}
	p.master = dnp3.New(p)
	p.modbus = modbus.New(p)
	p.modbusRTU = modbusrtu.New(p)
	p.sources = []source.Source{p.master, p.modbus, p.modbusRTU}
	if cfg.DNP3Server.Enabled {
		p.outstation = dnp3.NewOutstation(cfg.DNP3Server, outstationSizes(cfg.Mappings), p)
	}
	if cfg.System.Enabled {
		p.sysCollector = sysmon.New(cfg.System)
	}
	iv := cfg.System.IntervalMs
	if iv <= 0 {
		iv = 5000
	}
	p.sysInterval = time.Duration(iv) * time.Millisecond
	return p
}

// normalizeMapping fills the routing fields (PointType/Index) for Modbus mappings
// (TCP or RTU) from their function/address so the publisher routes and maps them
// uniformly with DNP3. Modbus reads are polls, so they always publish-on-poll
// (deadband still applies); there is no event/static distinction.
func normalizeMapping(sig config.SignalMapping) config.SignalMapping {
	if !sig.UsesModbusFraming() {
		return sig
	}
	// Route by (function, address): the function is the point type so that the
	// same address under different functions (holding/input, coil/discrete) does
	// not collide. mapping.Apply formats by function.
	sig.PointType = sig.Function
	sig.Index = sig.Address
	sig.PublishOnPoll = true
	return sig
}

// outstationSizes computes the DNP3 outstation database dimensions from the
// ServeDNP3 mappings: the highest served index per OutType, plus one. An Update
// to an index beyond these bounds is dropped by the shim, so the database must
// be sized to cover every served point. Frozen-counter outputs are served as
// plain counters (opendnp3 has no direct frozen-counter setter), so they size
// the counter space.
func outstationSizes(mappings []config.SignalMapping) dnp3.OutstationDBSizes {
	var s dnp3.OutstationDBSizes
	bump := func(p *uint16, idx uint16) {
		n := min(uint32(idx)+1, 65535)
		if uint16(n) > *p {
			*p = uint16(n)
		}
	}
	for _, sig := range mappings {
		if !sig.Enabled || !sig.ServeDNP3 {
			continue
		}
		switch source.PointType(sig.OutType) {
		case source.PointBinary:
			bump(&s.Binary, sig.OutIndex)
		case source.PointDoubleBitBinary:
			bump(&s.DoubleBit, sig.OutIndex)
		case source.PointBinaryOutputStatus:
			bump(&s.BinaryOutputStatus, sig.OutIndex)
		case source.PointCounter, source.PointFrozenCounter:
			bump(&s.Counter, sig.OutIndex)
		case source.PointAnalog:
			bump(&s.Analog, sig.OutIndex)
		case source.PointAnalogOutputStatus:
			bump(&s.AnalogOutputStatus, sig.OutIndex)
		case source.PointOctetString:
			bump(&s.OctetString, sig.OutIndex)
		}
	}
	return s
}

// serveOutstation pushes one mapped point onto the DNP3 outstation database so a
// SCADA master sees the gateway's aggregated value. It serves the
// engineering-scaled value (result.Value, matching the MQTT side) at the
// configured (OutType, OutIndex), carrying the source quality flags through so a
// bad/comm-lost reading is served bad. Called for every matching sample
// (independent of the MQTT publish gating) so integrity polls stay fresh;
// opendnp3 does its own change/event detection.
func (p *Publisher) serveOutstation(sig config.SignalMapping, m source.Sample, result mapping.Result) {
	out := source.Sample{
		PointType: source.PointType(sig.OutType),
		Index:     sig.OutIndex,
		Time:      m.Time,
		Quality:   m.Quality,
	}
	switch out.PointType {
	case source.PointBinary, source.PointBinaryOutputStatus:
		out.BoolValue = result.Value != 0
	case source.PointDoubleBitBinary:
		out.DBBValue = source.DoubleBitState(clampDBB(result.Value))
	case source.PointCounter, source.PointFrozenCounter:
		out.UintValue = toUint32(result.Value)
	case source.PointAnalog, source.PointAnalogOutputStatus:
		out.FloatValue = result.Value
	case source.PointOctetString:
		out.BytesValue = m.BytesValue
	default:
		return // unknown OutType: skip rather than serve a wrong-typed point
	}
	p.outstation.Update(out)
}

// toUint32 clamps a (possibly scaled/negative) engineering value to the uint32
// range a DNP3 counter carries.
func toUint32(v float64) uint32 {
	if math.IsNaN(v) || v <= 0 {
		return 0
	}
	if v >= math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v)
}

// clampDBB clamps a numeric value to the DNP3 double-bit state range [0,3].
func clampDBB(v float64) uint8 {
	if math.IsNaN(v) || v < 0 {
		return 0
	}
	if v > 3 {
		return 3
	}
	return uint8(v)
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

	// paho selects its transport from the broker URL scheme, so when TLS is
	// enabled coerce tcp:// → ssl://. crypto/tls is pure Go and cross-compiles
	// to the static ICR-3232 (linux/arm/v7) build with no native dependency.
	broker := mq.Broker
	if mq.TLS.Enabled {
		broker = forceTLSScheme(broker)
	}

	p.node = sparkplug.NewNode(sp.GroupID, sp.NodeID)
	// Declare the Planta UI alias as an NBIRTH payload-level property (contract
	// v3 §1.1/§5) so the consumer can stage the Planta on first contact.
	if sp.PlantaAlias != "" {
		nodeProps := &sparkplug.PropertySet{}
		nodeProps.AddString("uns/planta", sp.PlantaAlias)
		p.node.SetNodeProperties(nodeProps)
	}
	opts := p.node.NewClientOptions(broker, clientID, mq.Username, mq.Password)
	if mq.TLS.Enabled {
		tlsCfg, err := mqttTLSConfig(mq.TLS)
		if err != nil {
			return fmt.Errorf("mqtt tls: %w", err)
		}
		opts.SetTLSConfig(tlsCfg)
	}
	// Auto-reconnect on broker drop. Without this, an MQTT blip strands the
	// gateway with a live DNP3 master but no publish path; the offline buffer
	// drains once MQTT comes back.
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetMaxReconnectInterval(30 * time.Second)
	// All MQTT session setup happens here so it runs on the initial connect AND on
	// every auto-reconnect. Sparkplug requires a fresh NBIRTH after each session is
	// (re)established; we also (re)subscribe to NCMD and flush anything the field
	// sources buffered while the broker was unreachable. (paho fires this on its
	// own goroutine once the connection is up, so the client is connected here and
	// births publish directly rather than re-buffering.)
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		p.mqttUp.Store(true)
		p.logInfo("MQTT connected to " + mq.Broker + " (clientId=" + clientID + ")")
		if err := p.node.PublishNBirth(c, p.buildBirthMetrics()); err != nil {
			p.logWarn("NBIRTH failed: " + err.Error())
		}
		p.node.SubscribeNCMD(c, func() {
			p.logInfo("Rebirth requested, re-publishing NBIRTH")
			p.node.PublishNBirth(c, p.buildBirthMetrics())
			p.publishDBirths()
		})
		p.publishDBirths()
		p.drainBuffer()
	})
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		p.mqttUp.Store(false)
		p.logWarn("MQTT connection lost: " + err.Error() + " (auto-reconnecting)")
	})

	client := mqtt.NewClient(opts)
	// Set the client before connecting so the field sources start buffering
	// immediately (publishOrBuffer sees IsConnected()==false → offline buffer)
	// instead of dropping samples during the initial connect.
	p.client = client
	// Connect in the background with retry — never block or fail startup on an
	// unreachable broker. An edge gateway must start polling the field at once and
	// hold data in the store-and-forward buffer even when the broker is down
	// (cloud broker, link not up, broker not yet provisioned). Births + buffer
	// drain happen in the OnConnect handler once the broker is reachable.
	client.Connect()

	for _, o := range cfg.Outstations {
		if !o.Enabled {
			continue
		}
		if err := p.master.AddOutstation(o); err != nil {
			p.logWarn(fmt.Sprintf("AddOutstation %s: %v", o.ID, err))
		}
	}
	for _, d := range cfg.ModbusDevices {
		if !d.Enabled {
			continue
		}
		if err := p.modbus.AddDevice(d, cfg.Mappings); err != nil {
			p.logWarn(fmt.Sprintf("AddDevice %s: %v", d.ID, err))
		}
	}
	for _, d := range cfg.SerialDevices {
		if !d.Enabled {
			continue
		}
		if err := p.modbusRTU.AddDevice(d, cfg.Mappings); err != nil {
			p.logWarn(fmt.Sprintf("AddDevice (rtu) %s: %v", d.ID, err))
		}
	}

	pCtx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.startAt = time.Now()

	// Start the ingest worker BEFORE the sources: DNP3 startup integrity responses
	// are delivered synchronously during Start, so the queue must already be
	// draining or those callbacks would block a protocol thread.
	p.ingest = make(chan source.Sample, ingestQueueSize)
	p.wg.Add(1)
	go p.ingestLoop()

	// Bring up the DNP3 outstation server (if enabled) before the field sources,
	// so its served points exist before the first sample is pushed to them.
	if p.outstation != nil {
		if err := p.outstation.Start(pCtx); err != nil {
			cancel()
			close(p.ingest)
			p.wg.Wait()
			p.running.Store(false)
			return fmt.Errorf("outstation start: %w", err)
		}
	}

	for _, s := range p.sources {
		if err := s.Start(pCtx); err != nil {
			// Tear down everything already brought up so a failed start doesn't
			// orphan native resources or goroutines. Stop is a no-op on sources
			// that never started.
			for _, other := range p.sources {
				other.Stop()
			}
			if p.outstation != nil {
				p.outstation.Stop()
			}
			cancel()
			close(p.ingest)
			p.wg.Wait()
			p.running.Store(false)
			return fmt.Errorf("source start: %w", err)
		}
	}

	// The system-telemetry loop always runs; it no-ops while no collector is set,
	// so monitoring can be toggled live (ApplySystemConfig) without a restart.
	p.sysReload = make(chan struct{}, 1)
	p.wg.Add(1)
	go p.sysLoop(pCtx)
	if p.collector() != nil {
		p.logInfo(fmt.Sprintf("System monitoring enabled (every %s)", p.currentInterval()))
	}

	dnp3Server := "off"
	if p.outstation != nil {
		dnp3Server = "on @ " + cfg.DNP3Server.Addr()
	}
	p.logInfo(fmt.Sprintf("Publisher started: %d outstations, %d modbus/tcp devices, %d modbus/rtu devices, %d mappings, dnp3 server %s",
		len(cfg.Outstations), len(cfg.ModbusDevices), len(cfg.SerialDevices), len(cfg.Mappings), dnp3Server))
	return nil
}

// sysLoop samples host telemetry on a ticker and publishes it as node metrics.
// It records each value into lastVal first (so the dashboard sees it), then
// publishes/buffers — publishOrBuffer's alias-compression mutates metric names,
// so the record must happen before the publish. The collector and interval can
// change at runtime (ApplySystemConfig): a nil collector makes the tick a no-op,
// and a sysReload signal resets the ticker. Exits on ctx cancellation.
func (p *Publisher) sysLoop(ctx context.Context) {
	defer p.wg.Done()
	interval := p.currentInterval()
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.sysReload:
			if iv := p.currentInterval(); iv != interval {
				interval = iv
				t.Reset(interval)
			}
		case <-t.C:
			c := p.collector()
			if c == nil {
				continue
			}
			ms := c.Collect(sparkplug.NowMs())
			if len(ms) == 0 {
				continue
			}
			p.recordReadings(ms)
			if err := p.publishOrBuffer("", ms); err != nil {
				p.errorCount.Add(1)
				slog.Warn("system metrics publish failed", "err", err)
			}
		}
	}
}

// collector returns the active system collector (nil when monitoring is off).
func (p *Publisher) collector() *sysmon.Collector {
	p.sysMu.RLock()
	defer p.sysMu.RUnlock()
	return p.sysCollector
}

// currentInterval returns the system poll cadence, defaulting to 5s.
func (p *Publisher) currentInterval() time.Duration {
	p.sysMu.RLock()
	defer p.sysMu.RUnlock()
	if p.sysInterval <= 0 {
		return 5 * time.Second
	}
	return p.sysInterval
}

// ApplySystemConfig swaps the system-telemetry collection live, with no gateway
// restart: it rebuilds (or clears) the collector, updates the interval, and
// re-issues NBIRTH so the new metric set is declared with fresh aliases. Safe to
// call from the API goroutine while the gateway runs; a no-op when stopped (the
// new config takes effect on the next Start via New).
func (p *Publisher) ApplySystemConfig(cfg config.SystemConfig) {
	if !p.running.Load() {
		return
	}
	p.sysMu.Lock()
	if cfg.Enabled {
		p.sysCollector = sysmon.New(cfg)
	} else {
		p.sysCollector = nil
	}
	iv := cfg.IntervalMs
	if iv <= 0 {
		iv = 5000
	}
	p.sysInterval = time.Duration(iv) * time.Millisecond
	p.sysMu.Unlock()

	// Rebirth so subscribers learn the new metric set (and aliases) immediately.
	// If the broker is down, skip — OnConnect rebuilds births from current config.
	if p.client != nil && p.mqttUp.Load() {
		if err := p.node.PublishNBirth(p.client, p.buildBirthMetrics()); err != nil {
			p.logWarn("rebirth after system-config change failed: " + err.Error())
		}
		p.publishDBirths()
	}
	// Nudge sysLoop to adopt the new interval (non-blocking; coalesced).
	select {
	case p.sysReload <- struct{}{}:
	default:
	}
	if cfg.Enabled {
		p.logInfo(fmt.Sprintf("System monitoring updated live (every %s)", p.currentInterval()))
	} else {
		p.logInfo("System monitoring disabled live")
	}
}

// recordReadings mirrors numeric metric values into lastVal so they surface in
// Status.LastReadings (and thus the UI) even though they bypass the mapping path.
func (p *Publisher) recordReadings(ms []*sparkplug.Metric) {
	p.lastMu.Lock()
	for _, m := range ms {
		switch {
		case m.DoubleValue != nil:
			p.lastVal[m.Name] = *m.DoubleValue
		case m.LongValue != nil:
			p.lastVal[m.Name] = float64(*m.LongValue)
		case m.IntValue != nil:
			p.lastVal[m.Name] = float64(*m.IntValue)
		}
	}
	p.lastMu.Unlock()
}

// Stop tears down the master and sends NDEATH.
func (p *Publisher) Stop() {
	if !p.running.Load() {
		return
	}
	if p.cancel != nil {
		p.cancel()
	}
	// Each source's Stop() returns only once its callbacks have ceased, so after
	// all sources stop no goroutine can send on ingest — safe to close and drain.
	for _, s := range p.sources {
		s.Stop()
	}
	close(p.ingest)
	p.wg.Wait()
	// Stop the outstation only after the ingest queue has fully drained, so no
	// in-flight serveOutstation → Update can race the teardown.
	if p.outstation != nil {
		p.outstation.Stop()
	}
	if p.client != nil && p.mqttUp.Load() {
		if err := p.node.PublishNDeath(p.client); err != nil {
			p.logWarn("NDEATH failed: " + err.Error())
		}
		p.client.Disconnect(500)
	}
	p.mqttUp.Store(false)
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
		ServerTime:   time.Now().UTC(),
	}
	s.MQTTConnected = p.mqttUp.Load()
	if p.node != nil {
		s.BdSeq = p.node.Bdseq()
	}
	if !p.startAt.IsZero() {
		s.Uptime = time.Since(p.startAt).Round(time.Second).String()
	}

	// Pull live per-source counters straight from each source so MeasurementsRx /
	// LastReadAt reflect every callback, not only the ones that came alongside an
	// OnStatusChange.
	for _, src := range p.sources {
		for _, st := range src.Status() {
			s.Outstations[st.ID] = st
		}
	}
	// The northbound DNP3 outstation server reports under its own ID; Connected
	// here means a SCADA master is currently connected.
	if p.outstation != nil {
		st := p.outstation.Status()
		s.Outstations[st.ID] = st
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
	maps.Copy(s.LastReadings, p.lastVal)
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
//
// With batching off it publishes each sample as it arrives. With batching on it
// accumulates per-device metrics (handleSample → emit) and flushes them as one
// message per device on the flush timer, plus once more when the channel closes
// so nothing buffered is lost at shutdown.
func (p *Publisher) ingestLoop() {
	defer p.wg.Done()
	if p.batchWindow <= 0 {
		for m := range p.ingest {
			p.handleSample(m)
		}
		return
	}
	t := time.NewTicker(p.batchWindow)
	defer t.Stop()
	for {
		select {
		case m, ok := <-p.ingest:
			if !ok {
				p.flushPending()
				return
			}
			p.handleSample(m)
		case <-t.C:
			p.flushPending()
		}
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
		serve := sig.ServeDNP3 && p.outstation != nil
		// By default only change events flow to MQTT; static/poll responses publish
		// only when the mapping opts in via PublishOnPoll.
		mqttGated := !m.IsEvent && !sig.PublishOnPoll
		// A non-served, MQTT-gated sample has nothing to do — skip before mapping
		// so the common MQTT-only path keeps its early-out.
		if !serve && mqttGated {
			continue
		}
		result, err := mapping.Apply(sig, m)
		if err != nil {
			p.errorCount.Add(1)
			slog.Warn("mapping error", "metric", sig.MetricName, "err", err)
			continue
		}
		// Feed the DNP3 outstation first, independent of the MQTT gating below, so
		// a SCADA master's integrity polls always see the latest value (opendnp3
		// does its own change/event detection on the served side).
		if serve {
			p.serveOutstation(sig, m, result)
		}
		if mqttGated {
			continue
		}
		// Carry the computed quality onto the wire: a not-good measurement
		// (comm-lost/restart/offline) publishes with the Sparkplug is_null bit
		// set, which the consumer maps to calidad=Mala. Without this the bad
		// quality is computed but silently dropped.
		result.Metric.IsNull = result.IsNull
		if !p.shouldPublish(sig, result.Value, result.IsNull) {
			continue
		}
		p.emit(sig.DeviceID, result.Metric)
	}
}

// emit routes one mapped metric to its device's outgoing message. With batching
// off (batchWindow==0) it publishes immediately (legacy: one message per
// signal); with batching on it appends to the device's pending batch, flushing
// early when that batch reaches batchMax. Ingest-goroutine only (or synchronous
// in tests), so pending needs no lock.
func (p *Publisher) emit(deviceID string, m *sparkplug.Metric) {
	if p.batchWindow <= 0 {
		if err := p.publishOrBuffer(deviceID, []*sparkplug.Metric{m}); err != nil {
			p.errorCount.Add(1)
			slog.Warn("publish failed", "device", deviceID, "err", err)
		}
		return
	}
	p.pending[deviceID] = append(p.pending[deviceID], m)
	if len(p.pending[deviceID]) >= p.batchMax {
		p.flushDevice(deviceID)
	}
}

// flushDevice publishes (or buffers) one device's pending batch and clears it;
// no-op when nothing is pending. Ingest-goroutine only.
func (p *Publisher) flushDevice(deviceID string) {
	metrics := p.pending[deviceID]
	if len(metrics) == 0 {
		return
	}
	delete(p.pending, deviceID)
	if err := p.publishOrBuffer(deviceID, metrics); err != nil {
		p.errorCount.Add(1)
		slog.Warn("batch publish failed", "device", deviceID, "err", err)
	}
}

// flushPending flushes every device's pending batch. Called on the flush timer
// and once when the ingest channel closes (Stop). Ingest-goroutine only.
// Deleting the current key during range is safe in Go.
func (p *Publisher) flushPending() {
	for deviceID := range p.pending {
		p.flushDevice(deviceID)
	}
}

// OnStatusChange mirrors the master's per-outstation status into our snapshot.
// On the falling edge of a device's connection (was up, now down) it marks every
// signal mapped to that device bad-quality once, so a comm loss is recorded as
// calidad=Mala in the historian instead of freezing the last good value.
func (p *Publisher) OnStatusChange(s source.Status) {
	p.statusMu.Lock()
	prev, had := p.osStatus[s.ID]
	p.osStatus[s.ID] = s
	p.statusMu.Unlock()
	// Strict falling edge only: requires a prior Connected==true so we don't fire
	// on the OPENING/CLOSED states a channel passes through before it first
	// connects (DNP3 in particular reports those as Connected==false).
	if had && prev.Connected && !s.Connected {
		p.emitBadQuality(s.ID)
	}
	if p.OnStatus != nil {
		p.OnStatus(p.Status())
	}
}

// emitBadQuality enqueues one comm-lost (not-online) sample per point mapped to
// sourceID. Quality 0 has the online bit clear, so Quality.Good() is false and
// mapping.Apply yields IsNull — published with is_null set (→ calidad=Mala).
// Routed through the ingest queue like any sample, so it never blocks the
// protocol thread that reported the status change.
func (p *Publisher) emitBadQuality(sourceID string) {
	now := time.Now()
	for _, sig := range p.mappingsFor(sourceID) {
		p.OnSample(source.Sample{
			SourceID:  sourceID,
			PointType: source.PointType(sig.PointType),
			Index:     sig.Index,
			Time:      now,
			Quality:   0,    // online bit clear → not Good() → is_null
			IsEvent:   true, // force handleSample past the PublishOnPoll gate
		})
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

// shouldPublish decides whether to emit this metric, accounting for both the
// value deadband and quality transitions. A bad-quality marker always publishes
// (and records the bad state); a good sample publishes when its value clears the
// deadband OR when it is the first good sample after a bad one (recovery) — so
// the historian's calidad returns to Buena even if the value never changed. The
// deadband baseline is refreshed only on a published good value.
func (p *Publisher) shouldPublish(sig config.SignalMapping, cur float64, bad bool) bool {
	p.lastMu.Lock()
	defer p.lastMu.Unlock()
	if bad {
		p.qualBad[sig.MetricName] = true
		return true
	}
	wasBad := p.qualBad[sig.MetricName]
	delete(p.qualBad, sig.MetricName)
	last, ok := p.lastVal[sig.MetricName]
	if !ok {
		last = math.NaN()
	}
	if wasBad || mapping.CheckDeadband(cur, last, sig.Deadband) {
		p.lastVal[sig.MetricName] = cur
		return true
	}
	return false
}

func (p *Publisher) publishOrBuffer(deviceID string, metrics []*sparkplug.Metric) error {
	if p.client == nil || !p.mqttUp.Load() {
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
		metrics = append(metrics, mapping.BirthMetric(sig, ts))
	}
	// Declare system metrics in NBIRTH so they alias-compress in NDATA. This also
	// primes the collector's network-rate baseline.
	if c := p.collector(); c != nil {
		metrics = append(metrics, c.Collect(ts)...)
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
		devMetrics[sig.DeviceID] = append(devMetrics[sig.DeviceID], mapping.BirthMetric(sig, ts))
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

// forceTLSScheme rewrites a broker URL to use the TLS transport (ssl://), since
// paho picks plain vs. TLS from the scheme. A bare host:port becomes ssl://host:port.
func forceTLSScheme(broker string) string {
	switch {
	case strings.HasPrefix(broker, "ssl://"), strings.HasPrefix(broker, "tls://"):
		return broker
	case strings.HasPrefix(broker, "tcp://"):
		return "ssl://" + strings.TrimPrefix(broker, "tcp://")
	case strings.Contains(broker, "://"):
		return broker // some other explicit scheme — leave as-is
	default:
		return "ssl://" + broker
	}
}

// mqttTLSConfig builds a *tls.Config from the MQTT TLS settings. Supports a
// custom CA (private/self-signed brokers), optional client certificate (mutual
// TLS), and an explicit insecure escape hatch. With no CA file it uses the
// system root pool. crypto/tls is pure Go, so this works in the fully-static
// ICR-3232 (linux/arm/v7) build.
func mqttTLSConfig(c config.TLSConfig) (*tls.Config, error) {
	t := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: c.Insecure, //nolint:gosec // operator opt-in for self-signed brokers
	}
	if c.CAFile != "" {
		pem, err := os.ReadFile(c.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read ca file: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("ca file %s: no certificates parsed", c.CAFile)
		}
		t.RootCAs = pool
	}
	if c.CertFile != "" && c.KeyFile != "" {
		crt, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load client cert/key: %w", err)
		}
		t.Certificates = []tls.Certificate{crt}
	}
	return t, nil
}
