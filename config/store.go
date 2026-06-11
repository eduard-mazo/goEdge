package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Store persists AppConfig to a JSON file with atomic writes.
type Store struct {
	mu   sync.RWMutex
	path string
	cfg  AppConfig
}

// NewStore loads config from path, creating it with defaults if absent.
func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// Get returns a deep copy of the current config. The slices are cloned so a
// caller iterating the result can't race with a concurrent mutator (Upsert/
// Delete reuse the backing arrays under the write lock); the element structs
// hold only scalars/strings, so a shallow element copy is a full copy.
//
// Clones are always non-nil (cloneSlice): an empty source list marshals to a
// JSON [] rather than null, so the web UI's list panels never receive null and
// crash on `.length`/`.filter`.
func (s *Store) Get() AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c := s.cfg
	c.Outstations = cloneSlice(c.Outstations)
	c.ModbusDevices = cloneSlice(c.ModbusDevices)
	c.SerialDevices = cloneSlice(c.SerialDevices)
	c.Mappings = cloneSlice(c.Mappings)
	c.System.Mounts = cloneSlice(c.System.Mounts)
	c.System.Interfaces = cloneSlice(c.System.Interfaces)
	c.System.DisabledMetrics = cloneSlice(c.System.DisabledMetrics)
	return c
}

// cloneSlice returns a non-nil copy of s. A nil/empty input yields a non-nil
// empty slice so JSON encodes it as [] (not null) — REST consumers (the web UI)
// can rely on every list field being an array.
func cloneSlice[T any](s []T) []T {
	out := make([]T, len(s))
	copy(out, s)
	return out
}

// Set replaces the config and persists it.
func (s *Store) Set(cfg AppConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg
	return s.save()
}

// UpdateMQTT replaces only the MQTT section.
func (s *Store) UpdateMQTT(mqtt MQTTConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.MQTT = mqtt
	return s.save()
}

// UpdateSparkplug replaces only the Sparkplug section.
func (s *Store) UpdateSparkplug(sp SparkplugConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.Sparkplug = sp
	return s.save()
}

// UpdateSystem replaces only the System (host telemetry) section.
func (s *Store) UpdateSystem(sys SystemConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.System = sys
	return s.save()
}

// UpsertOutstation adds or replaces a DNP3Outstation by ID.
func (s *Store) UpsertOutstation(o DNP3Outstation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, ex := range s.cfg.Outstations {
		if ex.ID == o.ID {
			s.cfg.Outstations[i] = o
			return s.save()
		}
	}
	s.cfg.Outstations = append(s.cfg.Outstations, o)
	return s.save()
}

// DeleteOutstation removes an outstation and any mappings that reference it.
func (s *Store) DeleteOutstation(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	filtered := s.cfg.Outstations[:0]
	for _, o := range s.cfg.Outstations {
		if o.ID != id {
			filtered = append(filtered, o)
		}
	}
	s.cfg.Outstations = filtered
	mappings := s.cfg.Mappings[:0]
	for _, m := range s.cfg.Mappings {
		if m.OutstationID != id {
			mappings = append(mappings, m)
		}
	}
	s.cfg.Mappings = mappings
	return s.save()
}

// UpsertModbusDevice adds or replaces a ModbusDevice by ID.
func (s *Store) UpsertModbusDevice(d ModbusDevice) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, ex := range s.cfg.ModbusDevices {
		if ex.ID == d.ID {
			s.cfg.ModbusDevices[i] = d
			return s.save()
		}
	}
	s.cfg.ModbusDevices = append(s.cfg.ModbusDevices, d)
	return s.save()
}

// DeleteModbusDevice removes a device and any Modbus mappings that reference it.
func (s *Store) DeleteModbusDevice(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	filtered := s.cfg.ModbusDevices[:0]
	for _, d := range s.cfg.ModbusDevices {
		if d.ID != id {
			filtered = append(filtered, d)
		}
	}
	s.cfg.ModbusDevices = filtered
	mappings := s.cfg.Mappings[:0]
	for _, m := range s.cfg.Mappings {
		if !(m.IsModbus() && m.Src() == id) {
			mappings = append(mappings, m)
		}
	}
	s.cfg.Mappings = mappings
	return s.save()
}

// UpsertSerialDevice adds or replaces a SerialDevice (Modbus RTU slave) by ID.
func (s *Store) UpsertSerialDevice(d SerialDevice) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, ex := range s.cfg.SerialDevices {
		if ex.ID == d.ID {
			s.cfg.SerialDevices[i] = d
			return s.save()
		}
	}
	s.cfg.SerialDevices = append(s.cfg.SerialDevices, d)
	return s.save()
}

// DeleteSerialDevice removes a serial device and any Modbus RTU mappings that
// reference it.
func (s *Store) DeleteSerialDevice(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	filtered := s.cfg.SerialDevices[:0]
	for _, d := range s.cfg.SerialDevices {
		if d.ID != id {
			filtered = append(filtered, d)
		}
	}
	s.cfg.SerialDevices = filtered
	mappings := s.cfg.Mappings[:0]
	for _, m := range s.cfg.Mappings {
		if !(m.IsModbusRTU() && m.Src() == id) {
			mappings = append(mappings, m)
		}
	}
	s.cfg.Mappings = mappings
	return s.save()
}

// UpsertMapping adds or replaces a SignalMapping by ID.
func (s *Store) UpsertMapping(m SignalMapping) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sig := range s.cfg.Mappings {
		if sig.ID == m.ID {
			s.cfg.Mappings[i] = m
			return s.save()
		}
	}
	s.cfg.Mappings = append(s.cfg.Mappings, m)
	return s.save()
}

// DeleteMapping removes a SignalMapping by ID.
func (s *Store) DeleteMapping(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	filtered := s.cfg.Mappings[:0]
	for _, m := range s.cfg.Mappings {
		if m.ID != id {
			filtered = append(filtered, m)
		}
	}
	s.cfg.Mappings = filtered
	return s.save()
}

// ReplaceAllMappings atomically replaces the full mapping table.
func (s *Store) ReplaceAllMappings(mappings []SignalMapping) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.Mappings = mappings
	return s.save()
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		s.cfg = DefaultAppConfig()
		return s.save()
	}
	if err != nil {
		return fmt.Errorf("config load %q: %w", s.path, err)
	}
	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("config parse %q: %w", s.path, err)
	}
	s.cfg = cfg
	return nil
}

// save writes atomically via a temp file + rename.
func (s *Store) save() error {
	data, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("config marshal: %w", err)
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("config dir %q: %w", dir, err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o640); err != nil {
		return fmt.Errorf("config write tmp %q: %w", tmp, err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("config rename %q → %q: %w", tmp, s.path, err)
	}
	return nil
}
