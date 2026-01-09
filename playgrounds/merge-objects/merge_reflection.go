package mergeobjects

import (
	"reflect"
	"time"
)

// MergeReflection merges two InputConfig pointers into a Config using reflection.
// Values from input2 take precedence over input1 when both are set.
func MergeReflection(input1, input2 *InputConfig) Config {
	result := Config{}

	// Merge input1 first, then input2 (so input2 takes precedence)
	if input1 != nil {
		mergeInputToConfig(&result, input1)
	}
	if input2 != nil {
		mergeInputToConfig(&result, input2)
	}

	return result
}

func mergeInputToConfig(dst *Config, src *InputConfig) {
	if src == nil {
		return
	}

	// Handle basic pointer fields
	if src.Name != nil {
		dst.Name = *src.Name
	}
	if src.Port != nil {
		dst.Port = *src.Port
	}
	if src.MaxRetries != nil {
		dst.MaxRetries = *src.MaxRetries
	}
	if src.Timeout != nil {
		dst.Timeout = *src.Timeout
	}
	if src.Rate != nil {
		dst.Rate = *src.Rate
	}
	if src.Enabled != nil {
		dst.Enabled = *src.Enabled
	}
	if src.EnabledPtr != nil {
		dst.EnabledPtr = src.EnabledPtr
	}
	if src.Description != nil {
		dst.Description = src.Description
	}

	// Handle slices (non-nil slice replaces)
	if src.Hosts != nil {
		dst.Hosts = make([]string, len(src.Hosts))
		copy(dst.Hosts, src.Hosts)
	}
	if src.Ports != nil {
		dst.Ports = make([]int, len(src.Ports))
		copy(dst.Ports, src.Ports)
	}
	if src.Tags != nil {
		dst.Tags = make([]Tag, len(src.Tags))
		for i, t := range src.Tags {
			tag := Tag{}
			if t.Key != nil {
				tag.Key = *t.Key
			}
			if t.Value != nil {
				tag.Value = *t.Value
			}
			dst.Tags[i] = tag
		}
	}

	// Handle maps (merge keys, input2 values take precedence)
	if src.Labels != nil {
		if dst.Labels == nil {
			dst.Labels = make(map[string]string)
		}
		for k, v := range src.Labels {
			dst.Labels[k] = v
		}
	}
	if src.Metadata != nil {
		if dst.Metadata == nil {
			dst.Metadata = make(map[string]any)
		}
		for k, v := range src.Metadata {
			dst.Metadata[k] = v
		}
	}

	// Handle nested struct
	if src.Database != nil {
		mergeInputDatabaseToDatabase(&dst.Database, src.Database)
	}
	if src.DatabasePtr != nil {
		if dst.DatabasePtr == nil {
			dst.DatabasePtr = &DatabaseConfig{}
		}
		mergeInputDatabaseToDatabase(dst.DatabasePtr, src.DatabasePtr)
	}

	// Handle time
	if src.CreatedAt != nil {
		dst.CreatedAt = *src.CreatedAt
	}
	if src.UpdatedAt != nil {
		dst.UpdatedAt = src.UpdatedAt
	}
}

func mergeInputDatabaseToDatabase(dst *DatabaseConfig, src *InputDatabaseConfig) {
	if src.Host != nil {
		dst.Host = *src.Host
	}
	if src.Port != nil {
		dst.Port = *src.Port
	}
	if src.Username != nil {
		dst.Username = *src.Username
	}
	if src.Password != nil {
		dst.Password = *src.Password
	}
	if src.SSLMode != nil {
		dst.SSLMode = *src.SSLMode
	}
}

// MergeReflectionGeneric is a more generic reflection-based approach.
// It attempts to merge any two structs with similar JSON tags into a target type.
func MergeReflectionGeneric(input1, input2 *InputConfig) Config {
	result := Config{}
	resultVal := reflect.ValueOf(&result).Elem()

	if input1 != nil {
		mergeStructByReflection(resultVal, reflect.ValueOf(input1).Elem())
	}
	if input2 != nil {
		mergeStructByReflection(resultVal, reflect.ValueOf(input2).Elem())
	}

	return result
}

func mergeStructByReflection(dst, src reflect.Value) {
	srcType := src.Type()

	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		srcFieldType := srcType.Field(i)

		// Find matching field in destination by JSON tag or name
		jsonTag := srcFieldType.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = srcFieldType.Name
		}
		// Strip ",omitempty" etc.
		if idx := findCommaIndex(jsonTag); idx != -1 {
			jsonTag = jsonTag[:idx]
		}

		dstField := findFieldByJSONTag(dst, jsonTag)
		if !dstField.IsValid() || !dstField.CanSet() {
			continue
		}

		// Handle pointer fields in source
		if srcField.Kind() == reflect.Ptr {
			if srcField.IsNil() {
				continue
			}
			srcElem := srcField.Elem()
			setFieldValue(dstField, srcElem)
		} else if srcField.Kind() == reflect.Slice {
			if srcField.IsNil() {
				continue
			}
			setSliceField(dstField, srcField)
		} else if srcField.Kind() == reflect.Map {
			if srcField.IsNil() {
				continue
			}
			setMapField(dstField, srcField)
		} else if srcField.Kind() == reflect.Struct {
			// Handle time.Time specially
			if srcField.Type() == reflect.TypeOf(time.Time{}) {
				if !srcField.Interface().(time.Time).IsZero() {
					dstField.Set(srcField)
				}
			} else {
				mergeStructByReflection(dstField, srcField)
			}
		}
	}
}

func findCommaIndex(s string) int {
	for i, c := range s {
		if c == ',' {
			return i
		}
	}
	return -1
}

func findFieldByJSONTag(v reflect.Value, tag string) reflect.Value {
	vType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := vType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		}
		if idx := findCommaIndex(jsonTag); idx != -1 {
			jsonTag = jsonTag[:idx]
		}
		if jsonTag == tag {
			return v.Field(i)
		}
	}
	return reflect.Value{}
}

func setFieldValue(dst, src reflect.Value) {
	if dst.Kind() == reflect.Ptr {
		// Destination is pointer, source is struct - need to convert types
		if src.Kind() == reflect.Struct {
			newPtr := reflect.New(dst.Type().Elem())
			convertStructByReflection(newPtr.Elem(), src)
			dst.Set(newPtr)
		} else {
			newPtr := reflect.New(src.Type())
			newPtr.Elem().Set(src)
			dst.Set(newPtr)
		}
	} else if dst.Kind() == reflect.Struct && src.Kind() == reflect.Struct {
		// Both are structs but different types - need to convert field by field
		convertStructByReflection(dst, src)
	} else if dst.Kind() == src.Kind() {
		dst.Set(src)
	} else if src.Type().ConvertibleTo(dst.Type()) {
		dst.Set(src.Convert(dst.Type()))
	}
}

func setSliceField(dst, src reflect.Value) {
	if dst.Kind() != reflect.Slice {
		return
	}
	newSlice := reflect.MakeSlice(dst.Type(), src.Len(), src.Cap())
	for i := 0; i < src.Len(); i++ {
		srcElem := src.Index(i)
		dstElem := newSlice.Index(i)

		if srcElem.Kind() == reflect.Struct && dstElem.Kind() == reflect.Struct {
			// Convert struct elements (e.g., InputTag to Tag)
			convertStructByReflection(dstElem, srcElem)
		} else {
			dstElem.Set(srcElem)
		}
	}
	dst.Set(newSlice)
}

func setMapField(dst, src reflect.Value) {
	if dst.Kind() != reflect.Map {
		return
	}
	if dst.IsNil() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}
	for _, key := range src.MapKeys() {
		dst.SetMapIndex(key, src.MapIndex(key))
	}
}

func convertStructByReflection(dst, src reflect.Value) {
	srcType := src.Type()
	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		srcFieldType := srcType.Field(i)

		jsonTag := srcFieldType.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = srcFieldType.Name
		}
		if idx := findCommaIndex(jsonTag); idx != -1 {
			jsonTag = jsonTag[:idx]
		}

		dstField := findFieldByJSONTag(dst, jsonTag)
		if !dstField.IsValid() || !dstField.CanSet() {
			continue
		}

		if srcField.Kind() == reflect.Ptr && !srcField.IsNil() {
			setFieldValue(dstField, srcField.Elem())
		}
	}
}
