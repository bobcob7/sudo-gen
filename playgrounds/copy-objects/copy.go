package copyobjects

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"maps"
	"reflect"
	"time"
)

// CopyManual creates a deep copy using manual field-by-field copying.
// This is the most performant approach but requires maintenance when fields change.
func (c *Config) CopyManual() *Config {
	if c == nil {
		return nil
	}

	dst := &Config{
		// Basic types (copied by value)
		Name:       c.Name,
		Port:       c.Port,
		MaxRetries: c.MaxRetries,
		Timeout:    c.Timeout,
		Rate:       c.Rate,
		Enabled:    c.Enabled,

		// Nested struct (copied by value since it has no pointers/slices/maps)
		Database:  c.Database,
		CreatedAt: c.CreatedAt,
	}

	// Pointer types
	if c.EnabledPtr != nil {
		v := *c.EnabledPtr
		dst.EnabledPtr = &v
	}
	if c.Description != nil {
		v := *c.Description
		dst.Description = &v
	}
	if c.DatabasePtr != nil {
		v := *c.DatabasePtr
		dst.DatabasePtr = &v
	}
	if c.UpdatedAt != nil {
		v := *c.UpdatedAt
		dst.UpdatedAt = &v
	}

	// Slices
	if c.Hosts != nil {
		dst.Hosts = make([]string, len(c.Hosts))
		copy(dst.Hosts, c.Hosts)
	}
	if c.Ports != nil {
		dst.Ports = make([]int, len(c.Ports))
		copy(dst.Ports, c.Ports)
	}
	if c.Tags != nil {
		dst.Tags = make([]Tag, len(c.Tags))
		copy(dst.Tags, c.Tags)
	}

	// Maps
	if c.Labels != nil {
		dst.Labels = make(map[string]string, len(c.Labels))
		maps.Copy(dst.Labels, c.Labels)
	}
	if c.Metadata != nil {
		dst.Metadata = make(map[string]any, len(c.Metadata))
		for k, v := range c.Metadata {
			dst.Metadata[k] = deepCopyAny(v)
		}
	}

	return dst
}

// deepCopyAny performs a deep copy of arbitrary values (for map[string]any).
func deepCopyAny(v any) any {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(val))
		for k, v := range val {
			m[k] = deepCopyAny(v)
		}
		return m
	case []any:
		s := make([]any, len(val))
		for i, v := range val {
			s[i] = deepCopyAny(v)
		}
		return s
	case []string:
		s := make([]string, len(val))
		copy(s, val)
		return s
	case []int:
		s := make([]int, len(val))
		copy(s, val)
		return s
	default:
		// Primitives (string, int, float64, bool) are copied by value
		return val
	}
}

// CopyJSON creates a deep copy by marshaling to JSON and back.
// Simple to implement, works with any JSON-serializable struct.
// Downsides: slower, loses type info for some fields, time.Time formatting.
func (c *Config) CopyJSON() (*Config, error) {
	if c == nil {
		return nil, nil
	}

	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	var dst Config
	if err := json.Unmarshal(data, &dst); err != nil {
		return nil, err
	}

	return &dst, nil
}

// CopyGob creates a deep copy using gob encoding.
// Faster than JSON for Go-native types, preserves more type information.
// Requires types to be gob-encodable.
func (c *Config) CopyGob() (*Config, error) {
	if c == nil {
		return nil, nil
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	if err := enc.Encode(c); err != nil {
		return nil, err
	}

	var dst Config
	if err := dec.Decode(&dst); err != nil {
		return nil, err
	}

	return &dst, nil
}

// CopyReflect creates a deep copy using reflection.
// Works with any struct without manual field mapping.
// More flexible but slower than manual copying.
func (c *Config) CopyReflect() *Config {
	if c == nil {
		return nil
	}

	dst := new(Config)
	deepCopyReflect(reflect.ValueOf(dst).Elem(), reflect.ValueOf(c).Elem())
	return dst
}

// deepCopyReflect recursively deep copies using reflection.
func deepCopyReflect(dst, src reflect.Value) {
	switch src.Kind() {
	case reflect.Ptr:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.New(src.Elem().Type()))
		deepCopyReflect(dst.Elem(), src.Elem())

	case reflect.Struct:
		// Special handling for time.Time (it's a struct but should be copied by value)
		if src.Type() == reflect.TypeFor[time.Time]() {
			dst.Set(src)
			return
		}
		for i := 0; i < src.NumField(); i++ {
			deepCopyReflect(dst.Field(i), src.Field(i))
		}

	case reflect.Slice:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			deepCopyReflect(dst.Index(i), src.Index(i))
		}

	case reflect.Map:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeMapWithSize(src.Type(), src.Len()))
		for _, key := range src.MapKeys() {
			// Copy the key
			keyCopy := reflect.New(key.Type()).Elem()
			deepCopyReflect(keyCopy, key)
			// Copy the value
			valCopy := reflect.New(src.MapIndex(key).Type()).Elem()
			deepCopyReflect(valCopy, src.MapIndex(key))
			dst.SetMapIndex(keyCopy, valCopy)
		}

	case reflect.Interface:
		if src.IsNil() {
			return
		}
		elem := src.Elem()
		elemCopy := reflect.New(elem.Type()).Elem()
		deepCopyReflect(elemCopy, elem)
		dst.Set(elemCopy)

	default:
		// Basic types (string, int, float, bool, etc.) - copy by value
		dst.Set(src)
	}
}
