package mergeobjects

import (
	"encoding/json"
)

// MergeJSON merges two InputConfig pointers into a Config using JSON marshaling.
// This approach leverages JSON's omitempty behavior - fields with zero values
// in the input won't overwrite existing values when unmarshaling.
// Values from input2 take precedence over input1 when both are set.
func MergeJSON(input1, input2 *InputConfig) Config {
	result := Config{}

	// Marshal and unmarshal input1 first
	if input1 != nil {
		data, err := json.Marshal(input1)
		if err == nil {
			json.Unmarshal(data, &result)
		}
	}

	// Marshal and unmarshal input2, which will overwrite existing values
	if input2 != nil {
		data, err := json.Marshal(input2)
		if err == nil {
			json.Unmarshal(data, &result)
		}
	}

	return result
}

// MergeJSONWithMap is an alternative approach using map[string]any as intermediate.
// This gives more control over the merging process.
func MergeJSONWithMap(input1, input2 *InputConfig) Config {
	merged := make(map[string]any)

	// Convert input1 to map and add to merged
	if input1 != nil {
		data, err := json.Marshal(input1)
		if err == nil {
			var m map[string]any
			if json.Unmarshal(data, &m) == nil {
				deepMergeMap(merged, m)
			}
		}
	}

	// Convert input2 to map and merge (overwrites input1 values)
	if input2 != nil {
		data, err := json.Marshal(input2)
		if err == nil {
			var m map[string]any
			if json.Unmarshal(data, &m) == nil {
				deepMergeMap(merged, m)
			}
		}
	}

	// Convert merged map to Config
	result := Config{}
	data, err := json.Marshal(merged)
	if err == nil {
		json.Unmarshal(data, &result)
	}

	return result
}

// deepMergeMap recursively merges src into dst.
// Values in src overwrite values in dst, except for nested maps which are merged.
func deepMergeMap(dst, src map[string]any) {
	for key, srcVal := range src {
		if srcMap, ok := srcVal.(map[string]any); ok {
			// Source value is a map, try to merge with existing
			if dstVal, exists := dst[key]; exists {
				if dstMap, ok := dstVal.(map[string]any); ok {
					// Both are maps, merge recursively
					deepMergeMap(dstMap, srcMap)
					continue
				}
			}
			// Destination doesn't have this key or it's not a map
			// Make a copy of the source map
			newMap := make(map[string]any)
			deepMergeMap(newMap, srcMap)
			dst[key] = newMap
		} else {
			// Not a map, just overwrite
			dst[key] = srcVal
		}
	}
}

// MergeJSONStrict is a stricter version that handles errors properly.
func MergeJSONStrict(input1, input2 *InputConfig) (Config, error) {
	result := Config{}

	if input1 != nil {
		data, err := json.Marshal(input1)
		if err != nil {
			return result, err
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return result, err
		}
	}

	if input2 != nil {
		data, err := json.Marshal(input2)
		if err != nil {
			return result, err
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return result, err
		}
	}

	return result, nil
}
