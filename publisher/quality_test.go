package publisher

import (
	"testing"

	"goMqttDnp3/config"
	"goMqttDnp3/source"
)

// modbusMapping builds a minimal enabled Modbus/TCP holding-register mapping on
// device dev. New() runs normalizeMapping, so PointType/Index become the
// function/address and PublishOnPoll is forced on.
func modbusMapping(dev, metric string, addr uint16) config.SignalMapping {
	return config.SignalMapping{
		MetricName: metric,
		Protocol:   "modbus",
		SourceID:   dev,
		Function:   "holding_register",
		Address:    addr,
		DataType:   "int16",
		Enabled:    true,
	}
}

// goodSample / badSample for a holding-register point at addr.
func regSample(dev string, addr uint16, q source.Quality, v float64) source.Sample {
	return source.Sample{
		SourceID:   dev,
		PointType:  source.PointType("holding_register"),
		Index:      addr,
		Quality:    q,
		FloatValue: v,
		IsEvent:    true,
	}
}

// bufferedMetrics drains the offline buffer (publishOrBuffer enqueues here while
// the MQTT client is unset) and returns it for assertions.
func (p *Publisher) drainTestBuffer() []bufferedMsg {
	p.bufMu.Lock()
	defer p.bufMu.Unlock()
	out := p.msgBuf
	p.msgBuf = nil
	return out
}

func TestHandleSample_BadQualitySetsIsNull(t *testing.T) {
	cfg := config.AppConfig{Mappings: []config.SignalMapping{modbusMapping("MOD1", "T1", 10)}}
	p := New(cfg)

	// Good read → published, value carried, not null.
	p.handleSample(regSample("MOD1", 10, source.QualityOnline, 42))
	buf := p.drainTestBuffer()
	if len(buf) != 1 || len(buf[0].metrics) != 1 {
		t.Fatalf("good sample: want 1 buffered metric, got %+v", buf)
	}
	if buf[0].metrics[0].IsNull {
		t.Fatalf("good sample published with is_null set")
	}

	// Comm-lost read (online bit clear) → published with is_null set (calidad=Mala).
	p.handleSample(regSample("MOD1", 10, 0, 0))
	buf = p.drainTestBuffer()
	if len(buf) != 1 || len(buf[0].metrics) != 1 {
		t.Fatalf("bad sample: want 1 buffered metric, got %+v", buf)
	}
	if !buf[0].metrics[0].IsNull {
		t.Fatalf("bad-quality sample published WITHOUT is_null set")
	}
}

func TestHandleSample_RecoveryRepublishesThroughDeadband(t *testing.T) {
	m := modbusMapping("MOD1", "T1", 10)
	m.Deadband = 100 // huge deadband: an unchanged value would normally be suppressed
	cfg := config.AppConfig{Mappings: []config.SignalMapping{m}}
	p := New(cfg)

	p.handleSample(regSample("MOD1", 10, source.QualityOnline, 42)) // good, baseline=42
	p.drainTestBuffer()
	p.handleSample(regSample("MOD1", 10, 0, 0)) // comm lost → bad marker
	if buf := p.drainTestBuffer(); len(buf) != 1 || !buf[0].metrics[0].IsNull {
		t.Fatalf("expected one is_null marker on comm loss, got %+v", buf)
	}
	// Recovery: same value as before the loss. Deadband alone would suppress it,
	// but a bad→good transition must republish so calidad returns to Buena.
	p.handleSample(regSample("MOD1", 10, source.QualityOnline, 42))
	buf := p.drainTestBuffer()
	if len(buf) != 1 {
		t.Fatalf("recovery within deadband not republished: %+v", buf)
	}
	if buf[0].metrics[0].IsNull {
		t.Fatalf("recovery sample still marked is_null")
	}
}

func TestOnStatusChange_FallingEdgeEmitsBadQuality(t *testing.T) {
	cfg := config.AppConfig{Mappings: []config.SignalMapping{
		modbusMapping("MOD1", "T1", 10),
		modbusMapping("MOD1", "T2", 11),
		modbusMapping("MOD2", "X1", 20), // different device — must not be touched
	}}
	p := New(cfg)
	p.ingest = make(chan source.Sample, 16) // emitBadQuality routes through OnSample→ingest

	// Bring MOD1 up, then drop it: a strict falling edge.
	p.OnStatusChange(source.Status{ID: "MOD1", Connected: true})
	if len(p.ingest) != 0 {
		t.Fatalf("rising edge unexpectedly emitted samples")
	}
	p.OnStatusChange(source.Status{ID: "MOD1", Connected: false})

	got := len(p.ingest)
	if got != 2 {
		t.Fatalf("falling edge: want 2 bad samples (T1,T2), got %d", got)
	}
	for range got {
		s := <-p.ingest
		if s.SourceID != "MOD1" {
			t.Fatalf("bad sample for wrong source %q", s.SourceID)
		}
		if s.Quality.Good() {
			t.Fatalf("comm-loss sample has Good() quality: %v", s.Quality)
		}
	}
}

func TestOnStatusChange_NoEmitWithoutPriorConnect(t *testing.T) {
	cfg := config.AppConfig{Mappings: []config.SignalMapping{modbusMapping("MOD1", "T1", 10)}}
	p := New(cfg)
	p.ingest = make(chan source.Sample, 16)

	// First-ever status is "down" (e.g. DNP3 OPENING/CLOSED before first connect).
	// No prior Connected==true, so we must NOT emit — avoids startup false positives.
	p.OnStatusChange(source.Status{ID: "MOD1", Connected: false})
	if len(p.ingest) != 0 {
		t.Fatalf("emitted bad quality without a prior successful connect")
	}
}
