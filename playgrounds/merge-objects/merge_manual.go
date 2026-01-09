package mergeobjects

// MergeManual merges two InputConfig pointers into a Config using explicit field assignments.
// This is the most straightforward approach - explicit, type-safe, but requires maintenance.
// Values from input2 take precedence over input1 when both are set.
func MergeManual(input1, input2 *InputConfig) Config {
	result := Config{}

	// Apply input1 first
	if input1 != nil {
		applyInputManual(&result, input1)
	}

	// Apply input2 second (overwrites input1 values)
	if input2 != nil {
		applyInputManual(&result, input2)
	}

	return result
}

func applyInputManual(dst *Config, src *InputConfig) {
	// Basic pointer fields
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
		v := *src.EnabledPtr
		dst.EnabledPtr = &v
	}
	if src.Description != nil {
		v := *src.Description
		dst.Description = &v
	}

	// Slices - replace entirely if present
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
			dst.Tags[i] = convertInputTag(t)
		}
	}

	// Maps - merge keys
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

	// Nested structs
	if src.Database != nil {
		applyInputDatabaseManual(&dst.Database, src.Database)
	}
	if src.DatabasePtr != nil {
		if dst.DatabasePtr == nil {
			dst.DatabasePtr = &DatabaseConfig{}
		}
		applyInputDatabaseManual(dst.DatabasePtr, src.DatabasePtr)
	}

	// Time fields
	if src.CreatedAt != nil {
		dst.CreatedAt = *src.CreatedAt
	}
	if src.UpdatedAt != nil {
		t := *src.UpdatedAt
		dst.UpdatedAt = &t
	}
}

func applyInputDatabaseManual(dst *DatabaseConfig, src *InputDatabaseConfig) {
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

func convertInputTag(t InputTag) Tag {
	tag := Tag{}
	if t.Key != nil {
		tag.Key = *t.Key
	}
	if t.Value != nil {
		tag.Value = *t.Value
	}
	return tag
}

// MergeManualWithDefaults is a variant that allows specifying default values.
func MergeManualWithDefaults(defaults *Config, input1, input2 *InputConfig) Config {
	result := Config{}

	// Apply defaults first
	if defaults != nil {
		result = *defaults
	}

	// Apply inputs
	if input1 != nil {
		applyInputManual(&result, input1)
	}
	if input2 != nil {
		applyInputManual(&result, input2)
	}

	return result
}
