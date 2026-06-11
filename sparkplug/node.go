package sparkplug

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Node represents a Sparkplug B Edge of Network (EoN) node.
// It is safe to use from multiple goroutines.
type Node struct {
	GroupID string
	NodeID  string

	mu             sync.Mutex
	bdSeq          uint64 // next session's bdSeq (advanced by NewClientOptions)
	sessionBdSeq   uint64 // bdSeq registered in the current session's NDEATH will
	lastBirthBdSeq uint64
	seq            uint64
	nodeProps      *PropertySet

	// pubMu serializes seq assignment WITH the publish enqueue. Sparkplug B has
	// ONE seq counter per EoN node across NDATA/DBIRTH/DDATA/DDEATH; if one
	// goroutine takes seq N, another takes N+1 and reaches the socket first,
	// every consumer sees an out-of-order seq and requests a rebirth (storm).
	pubMu sync.Mutex

	// global alias counter (thread-safe); shared across node + device metrics
	// so alias values are globally unique within the session.
	aliasCounter atomic.Uint64

	nodeRegistry    *Registry            // name → alias for node metrics
	deviceRegistries map[string]*Registry // deviceID → registry
}

// NewNode constructs an EoN node.
func NewNode(groupID, nodeID string) *Node {
	return &Node{
		GroupID:          groupID,
		NodeID:           nodeID,
		nodeRegistry:     NewRegistry(),
		deviceRegistries: make(map[string]*Registry),
	}
}

// NowMs returns current time as Unix epoch milliseconds.
func NowMs() uint64 {
	return uint64(time.Now().UnixMilli())
}

// SetNodeProperties attaches node-level metadata to every NBIRTH payload.
func (n *Node) SetNodeProperties(ps *PropertySet) {
	n.mu.Lock()
	n.nodeProps = ps
	n.mu.Unlock()
}

// NewClientOptions builds paho ClientOptions with NDEATH LWT.
//
// The bdSeq written into the will here DEFINES the session's bdSeq: every
// NBIRTH published within this client's lifetime repeats it (Sparkplug B
// §6.4.5 — hosts correlate an NDEATH with the birth via matching bdSeq).
// bdSeq advances per client session, never per NBIRTH; an in-session rebirth
// must NOT change it or the registered will becomes uncorrelatable.
func (n *Node) NewClientOptions(broker, clientID, username, password string) *mqtt.ClientOptions {
	n.mu.Lock()
	currentBdSeq := n.bdSeq
	n.sessionBdSeq = currentBdSeq
	n.bdSeq = (n.bdSeq + 1) % 256
	n.mu.Unlock()

	deathPayload := n.buildNDEATHPayloadWith(currentBdSeq)
	deathTopic := NodeTopic(n.GroupID, NDEATH, n.NodeID)

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetConnectTimeout(5 * time.Second).
		SetAutoReconnect(false)

	if username != "" {
		opts.SetUsername(username)
		opts.SetPassword(password)
	}
	opts.SetBinaryWill(deathTopic, deathPayload, 1, false)
	return opts
}

// PublishNBirth publishes NBIRTH and resets sequence + alias registries.
// The bdSeq metric repeats the session bdSeq registered in the NDEATH will.
func (n *Node) PublishNBirth(client mqtt.Client, metrics []*Metric) error {
	n.pubMu.Lock()
	defer n.pubMu.Unlock()

	n.mu.Lock()
	currentBdSeq := n.sessionBdSeq
	n.lastBirthBdSeq = currentBdSeq
	n.seq = 0
	n.nodeRegistry.Clear()
	for k := range n.deviceRegistries {
		delete(n.deviceRegistries, k)
	}
	n.aliasCounter.Store(0)
	n.mu.Unlock()

	ts := NowMs()
	bdSeqMetric := MetricUInt64("bdSeq", ts, currentBdSeq)

	all := make([]*Metric, 0, 1+len(metrics))
	all = append(all, bdSeqMetric)
	all = append(all, metrics...)

	for _, m := range all {
		alias := n.aliasCounter.Add(1) - 1
		m.Alias = alias
		n.nodeRegistry.Register(m.Name, alias)
	}

	n.mu.Lock()
	nodeProps := n.nodeProps
	n.mu.Unlock()

	p := &Payload{Timestamp: ts, Seq: 0, Metrics: all, Properties: nodeProps}
	topic := NodeTopic(n.GroupID, NBIRTH, n.NodeID)
	return n.publishBinary(client, topic, p.Marshal(), 0)
}

// PublishNData publishes NDATA using node-metric aliases.
func (n *Node) PublishNData(client mqtt.Client, metrics []*Metric) error {
	n.pubMu.Lock()
	defer n.pubMu.Unlock()
	seq := n.nextSeq()
	ts := NowMs()
	for _, m := range metrics {
		if alias, ok := n.nodeRegistry.GetAlias(m.Name); ok {
			m.Alias = alias
			m.Name = ""
		}
	}
	p := &Payload{Timestamp: ts, Seq: seq, Metrics: metrics}
	topic := NodeTopic(n.GroupID, NDATA, n.NodeID)
	return n.publishBinary(client, topic, p.Marshal(), 0)
}

// PublishDBirth publishes DBIRTH for a child device, establishing per-device aliases.
func (n *Node) PublishDBirth(client mqtt.Client, deviceID string, metrics []*Metric) error {
	n.pubMu.Lock()
	defer n.pubMu.Unlock()
	seq := n.nextSeq()
	ts := NowMs()

	n.mu.Lock()
	if _, ok := n.deviceRegistries[deviceID]; !ok {
		n.deviceRegistries[deviceID] = NewRegistry()
	}
	devReg := n.deviceRegistries[deviceID]
	devReg.Clear()
	n.mu.Unlock()

	for _, m := range metrics {
		alias := n.aliasCounter.Add(1) - 1
		m.Alias = alias
		devReg.Register(m.Name, alias)
	}

	p := &Payload{Timestamp: ts, Seq: seq, Metrics: metrics}
	topic := DeviceTopic(n.GroupID, DBIRTH, n.NodeID, deviceID)
	return n.publishBinary(client, topic, p.Marshal(), 0)
}

// PublishDData publishes DDATA for a child device using per-device aliases.
func (n *Node) PublishDData(client mqtt.Client, deviceID string, metrics []*Metric) error {
	n.pubMu.Lock()
	defer n.pubMu.Unlock()
	seq := n.nextSeq()
	ts := NowMs()

	n.mu.Lock()
	devReg := n.deviceRegistries[deviceID]
	n.mu.Unlock()

	var selected []*Metric
	for _, m := range metrics {
		if devReg != nil {
			if alias, ok := devReg.GetAlias(m.Name); ok {
				m.Alias = alias
				m.Name = ""
				selected = append(selected, m)
				continue
			}
		}
		selected = append(selected, m)
	}
	if len(selected) == 0 {
		return nil
	}
	p := &Payload{Timestamp: ts, Seq: seq, Metrics: selected}
	topic := DeviceTopic(n.GroupID, DDATA, n.NodeID, deviceID)
	return n.publishBinary(client, topic, p.Marshal(), 0)
}

// PublishDDeath publishes DDEATH for a child device.
func (n *Node) PublishDDeath(client mqtt.Client, deviceID string) error {
	n.pubMu.Lock()
	defer n.pubMu.Unlock()
	seq := n.nextSeq()
	ts := NowMs()
	p := &Payload{Timestamp: ts, Seq: seq}
	topic := DeviceTopic(n.GroupID, DDEATH, n.NodeID, deviceID)
	return n.publishBinary(client, topic, p.Marshal(), 0)
}

// PublishNDeath publishes an explicit graceful NDEATH.
func (n *Node) PublishNDeath(client mqtt.Client) error {
	n.mu.Lock()
	birthBdSeq := n.lastBirthBdSeq
	n.mu.Unlock()
	payload := n.buildNDEATHPayloadWith(birthBdSeq)
	topic := NodeTopic(n.GroupID, NDEATH, n.NodeID)
	return n.publishBinary(client, topic, payload, 1)
}

// SubscribeNCMD subscribes to NCMD and calls onRebirth when requested.
func (n *Node) SubscribeNCMD(client mqtt.Client, onRebirth func()) {
	topic := NodeTopic(n.GroupID, NCMD, n.NodeID)
	token := client.Subscribe(topic, 1, func(_ mqtt.Client, msg mqtt.Message) {
		p := &Payload{}
		if err := p.Unmarshal(msg.Payload()); err != nil {
			slog.Warn("NCMD unmarshal failed", "err", err)
			return
		}
		for _, m := range p.Metrics {
			name := m.Name
			if name == "" {
				name, _ = n.nodeRegistry.GetName(m.Alias)
			}
			if name == "Node Control/Rebirth" && m.BoolValue != nil && *m.BoolValue {
				slog.Info("NCMD Rebirth received")
				if onRebirth != nil {
					onRebirth()
				}
				return
			}
		}
		if onRebirth != nil {
			onRebirth()
		}
	})
	token.Wait()
}

// Bdseq returns the current session's bdSeq (for status reporting). Constant
// for the lifetime of the MQTT client; it no longer counts rebirths.
func (n *Node) Bdseq() uint64 {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.sessionBdSeq
}

// nextSeq advances the node sequence number per Sparkplug B: NBIRTH carries 0,
// each subsequent message increments by one, wrapping 255 → 0. (The previous
// `(seq % 255) + 1` wrapped 255 → 1, skipping 0 — a spec-compliant consumer
// detects that as a gap once per 255 messages and requests a rebirth, causing
// a periodic birth storm.) Callers must hold pubMu.
func (n *Node) nextSeq() uint64 {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.seq = (n.seq + 1) % 256
	return n.seq
}

func (n *Node) buildNDEATHPayloadWith(bdSeq uint64) []byte {
	ts := NowMs()
	p := &Payload{
		Timestamp: ts,
		Metrics:   []*Metric{MetricUInt64("bdSeq", ts, bdSeq)},
	}
	return p.Marshal()
}

func (n *Node) publishBinary(client mqtt.Client, topic string, payload []byte, qos byte) error {
	if client == nil {
		return nil
	}
	token := client.Publish(topic, qos, false, payload)
	token.Wait()
	if err := token.Error(); err != nil {
		slog.Error("publish failed", "topic", topic, "err", err)
		return err
	}
	return nil
}
