package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
		_ = json.NewDecoder(f).Decode(&s.data)
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

// Get retrieves a key and unmarshals it into the provided pointer.
// Example: store.Get("user:123", &user)
func (s *Store) Get(key string, out any) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.data[key]
	if !ok {
		return false, nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return true, err
	}
	return true, nil
}

// GetByPrefix unmarshals all matching keys into the given map pointer.
// Example: store.GetByPrefix("user:", &map[string]User{})
func (s *Store) GetByPrefix(prefix string, out any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// out must be a pointer to a map[string]T
	outMap, ok := out.(*map[string]any)
	if !ok {
		// We need reflection to handle map[string]T properly
		return s.getByPrefixReflect(prefix, out)
	}

	result := make(map[string]any)
	for k, raw := range s.data {
		if strings.HasPrefix(k, prefix) {
			var v any
			if err := json.Unmarshal(raw, &v); err != nil {
				return err
			}
			result[k] = v
		}
	}
	*outMap = result
	return nil
}

func (s *Store) getByPrefixReflect(prefix string, out any) error {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Map {
		return &json.InvalidUnmarshalError{Type: reflect.TypeOf(out)}
	}
	m := reflect.MakeMap(rv.Elem().Type())
	elemType := rv.Elem().Type().Elem()

	for k, raw := range s.data {
		if strings.HasPrefix(k, prefix) {
			elemPtr := reflect.New(elemType)
			if err := json.Unmarshal(raw, elemPtr.Interface()); err != nil {
				return err
			}
			m.SetMapIndex(reflect.ValueOf(k), elemPtr.Elem())
		}
	}
	rv.Elem().Set(m)
	return nil
}

// Delete removes a key and persists the change.
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return s.persist()
}

// persist writes the entire store atomically.
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
