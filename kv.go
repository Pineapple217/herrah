package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]json.RawMessage
	path string
}

// OpenStore loads the store from a JSON file if it exists.
func OpenStore(path string) *Store {
	s := &Store{data: make(map[string]json.RawMessage), path: path}
	f, err := os.Open(path)
	if err == nil {
		defer f.Close()
		json.NewDecoder(f).Decode(&s.data)
	} else if !os.IsNotExist(err) {
		panic(err)
	}
	return s
}

// Set stores any Go value by JSON-encoding it.
func (s *Store) Set(key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = b
	return s.persist()
}

// Get retrieves and decodes a value from JSON into generic Go types (map[string]any, []any, etc.).
func (s *Store) Get(key string) (any, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.data[key]
	if !ok {
		return nil, false, nil
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, true, err
	}
	return v, true, nil
}

// Delete removes a key from the store and persists the change.
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return s.persist()
}

// persist writes the entire store atomically as JSON.
func (s *Store) persist() error {
	dir := filepath.Dir(s.path)
	tmp := filepath.Join(dir, "data.tmp")
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s.data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// GetByPrefix returns all key-value pairs where the key starts with the given prefix.
func (s *Store) GetByPrefix(prefix string) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]any)
	for k, raw := range s.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			var v any
			if err := json.Unmarshal(raw, &v); err != nil {
				return nil, err
			}
			result[k] = v
		}
	}
	return result, nil
}
