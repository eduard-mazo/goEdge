package publisher

import (
	"testing"

	"goMqttDnp3/config"
	"goMqttDnp3/source"
)

// devMapping is a modbusMapping whose Sparkplug DeviceID (the batch key) is set,
// so samples route by SourceID (src) but publish under the child device dev.
func devMapping(src, dev, metric string, addr uint16) config.SignalMapping {
	m := modbusMapping(src, metric, addr)
	m.DeviceID = dev
	return m
}

// flushedByDevice drains the offline buffer and indexes batches by deviceID,
// asserting at most one batch per device (the point of coalescing).
func (p *Publisher) flushedByDevice(t *testing.T) map[string][]string {
	t.Helper()
	out := make(map[string][]string)
	for _, msg := range p.drainTestBuffer() {
		if _, dup := out[msg.deviceID]; dup {
			t.Fatalf("device %q produced more than one batch", msg.deviceID)
		}
		names := make([]string, len(msg.metrics))
		for i, m := range msg.metrics {
			names[i] = m.Name
		}
		out[msg.deviceID] = names
	}
	return out
}

// TestBatch_GroupsPerDevice: with batching on, metrics for the same device that
// arrive within a window are flushed as ONE message; distinct devices each get
// their own message.
func TestBatch_GroupsPerDevice(t *testing.T) {
	cfg := config.AppConfig{
		MQTT: config.MQTTConfig{PublishBatchMs: 200},
		Mappings: []config.SignalMapping{
			devMapping("MOD1", "DEVA", "T1", 10),
			devMapping("MOD1", "DEVA", "T2", 11),
			devMapping("MOD1", "DEVB", "X1", 12),
		},
	}
	p := New(cfg)
	if p.batchWindow <= 0 {
		t.Fatalf("PublishBatchMs>0 should enable batching")
	}

	// Three good reads accumulate; nothing is published until the flush.
	p.handleSample(regSample("MOD1", 10, source.QualityOnline, 1))
	p.handleSample(regSample("MOD1", 11, source.QualityOnline, 2))
	p.handleSample(regSample("MOD1", 12, source.QualityOnline, 3))
	if buf := p.drainTestBuffer(); len(buf) != 0 {
		t.Fatalf("metrics published before flush: %+v", buf)
	}

	p.flushPending()
	got := p.flushedByDevice(t)
	if len(got) != 2 {
		t.Fatalf("want 2 device batches, got %d (%v)", len(got), got)
	}
	if len(got["DEVA"]) != 2 {
		t.Fatalf("DEVA: want 2 metrics in one batch, got %v", got["DEVA"])
	}
	if len(got["DEVB"]) != 1 {
		t.Fatalf("DEVB: want 1 metric, got %v", got["DEVB"])
	}
}

// TestBatch_SizeCapFlushesEarly: a device whose pending batch reaches batchMax is
// flushed immediately, without waiting for the timer.
func TestBatch_SizeCapFlushesEarly(t *testing.T) {
	cfg := config.AppConfig{
		MQTT: config.MQTTConfig{PublishBatchMs: 200},
		Mappings: []config.SignalMapping{
			devMapping("MOD1", "DEVA", "T1", 10),
			devMapping("MOD1", "DEVA", "T2", 11),
			devMapping("MOD1", "DEVA", "T3", 12),
		},
	}
	p := New(cfg)
	p.batchMax = 2 // flush after every 2 accumulated metrics

	p.handleSample(regSample("MOD1", 10, source.QualityOnline, 1))
	p.handleSample(regSample("MOD1", 11, source.QualityOnline, 2)) // hits cap → flush

	buf := p.drainTestBuffer()
	if len(buf) != 1 || len(buf[0].metrics) != 2 {
		t.Fatalf("size cap: want one batch of 2, got %+v", buf)
	}

	// The third sample starts a fresh batch, still pending until the next flush.
	p.handleSample(regSample("MOD1", 12, source.QualityOnline, 3))
	if buf := p.drainTestBuffer(); len(buf) != 0 {
		t.Fatalf("third sample should still be pending, got %+v", buf)
	}
	p.flushPending()
	if buf := p.drainTestBuffer(); len(buf) != 1 || len(buf[0].metrics) != 1 {
		t.Fatalf("flush: want one batch of 1, got %+v", buf)
	}
}

// TestBatch_DisabledPublishesPerSample: with a negative PublishBatchMs (explicit
// opt-out) behavior is the legacy path — one message per signal, no buffering.
func TestBatch_DisabledPublishesPerSample(t *testing.T) {
	cfg := config.AppConfig{
		MQTT: config.MQTTConfig{PublishBatchMs: -1},
		Mappings: []config.SignalMapping{
			devMapping("MOD1", "DEVA", "T1", 10),
			devMapping("MOD1", "DEVA", "T2", 11),
		},
	}
	p := New(cfg)
	if p.batchWindow != 0 {
		t.Fatalf("negative PublishBatchMs should disable batching")
	}

	p.handleSample(regSample("MOD1", 10, source.QualityOnline, 1))
	p.handleSample(regSample("MOD1", 11, source.QualityOnline, 2))

	buf := p.drainTestBuffer()
	if len(buf) != 2 {
		t.Fatalf("batching off: want 2 separate messages, got %d", len(buf))
	}
	for _, msg := range buf {
		if len(msg.metrics) != 1 {
			t.Fatalf("batching off: each message should carry 1 metric, got %+v", msg)
		}
	}
}

// TestBatch_DefaultOnWhenUnset: an unset (0) PublishBatchMs adopts the default
// coalescing window, so batching is on out of the box.
func TestBatch_DefaultOnWhenUnset(t *testing.T) {
	cfg := config.AppConfig{Mappings: []config.SignalMapping{devMapping("MOD1", "DEVA", "T1", 10)}}
	p := New(cfg)
	if p.batchWindow != defaultBatchWindow {
		t.Fatalf("unset PublishBatchMs: want default window %v, got %v", defaultBatchWindow, p.batchWindow)
	}
}

// TestBatch_FlushEmptyIsNoop guards the timer path: flushing with nothing pending
// must not publish an empty message.
func TestBatch_FlushEmptyIsNoop(t *testing.T) {
	cfg := config.AppConfig{
		MQTT:     config.MQTTConfig{PublishBatchMs: 200},
		Mappings: []config.SignalMapping{devMapping("MOD1", "DEVA", "T1", 10)},
	}
	p := New(cfg)
	p.flushPending()
	if buf := p.drainTestBuffer(); len(buf) != 0 {
		t.Fatalf("empty flush published: %+v", buf)
	}
}
