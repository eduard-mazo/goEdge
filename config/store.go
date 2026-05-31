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

// Get returns a copy of the current config.
func (s *Store) Get() AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
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
