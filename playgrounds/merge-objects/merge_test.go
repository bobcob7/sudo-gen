package mergeobjects

import (
	"testing"
	"time"
)

func TestMergeImplementations(t *testing.T) {
	// Setup test inputs
	input1 := &InputConfig{
		Name:       Ptr("service-a"),
		Port:       Ptr(8080),
		MaxRetries: Ptr(int32(3)),
		Enabled:    Ptr(true),
		Hosts:      []string{"host1.example.com", "host2.example.com"},
		Labels: map[string]string{
			"env":     "production",
			"version": "1.0",
		},
		Database: &InputDatabaseConfig{
			Host:     Ptr("db1.example.com"),
			Port:     Ptr(5432),
			Username: Ptr("admin"),
		},
		CreatedAt: Ptr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
	}

	input2 := &InputConfig{
		Port:        Ptr(9090), // Override port
		Rate:        Ptr(1.5),  // Add new field
		Description: Ptr("Updated service description"),
		Labels: map[string]string{
			"version": "2.0", // Override version
			"team":    "platform",
		},
		Database: &InputDatabaseConfig{
			Password: Ptr("secret123"), // Add password
			SSLMode:  Ptr("require"),
		},
		Tags: []InputTag{
			{Key: Ptr("priority"), Value: Ptr("high")},
		},
	}

	// Test all merge implementations
	implementations := []struct {
		name  string
		merge func(*InputConfig, *InputConfig) Config
	}{
		{"MergeManual", MergeManual},
		{"MergeReflection", MergeReflection},
		{"MergeReflectionGeneric", MergeReflectionGeneric},
		{"MergeJSON", MergeJSON},
		{"MergeJSONWithMap", MergeJSONWithMap},
	}

	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			result := impl.merge(input1, input2)

			// Check input1 values preserved
			if result.Name != "service-a" {
				t.Errorf("Name: expected 'service-a', got '%s'", result.Name)
			}
			if result.MaxRetries != 3 {
				t.Errorf("MaxRetries: expected 3, got %d", result.MaxRetries)
			}
			if !result.Enabled {
				t.Errorf("Enabled: expected true, got false")
			}

			// Check input2 values override
			if result.Port != 9090 {
				t.Errorf("Port: expected 9090, got %d", result.Port)
			}
			if result.Rate != 1.5 {
				t.Errorf("Rate: expected 1.5, got %f", result.Rate)
			}
			if result.Description == nil || *result.Description != "Updated service description" {
				t.Errorf("Description: expected 'Updated service description', got %v", result.Description)
			}

			// Check map merging
			if result.Labels["env"] != "production" {
				t.Errorf("Labels[env]: expected 'production', got '%s'", result.Labels["env"])
			}
			if result.Labels["version"] != "2.0" {
				t.Errorf("Labels[version]: expected '2.0', got '%s'", result.Labels["version"])
			}
			if result.Labels["team"] != "platform" {
				t.Errorf("Labels[team]: expected 'platform', got '%s'", result.Labels["team"])
			}

			// Check nested struct merging
			if result.Database.Host != "db1.example.com" {
				t.Errorf("Database.Host: expected 'db1.example.com', got '%s'", result.Database.Host)
			}
			if result.Database.Password != "secret123" {
				t.Errorf("Database.Password: expected 'secret123', got '%s'", result.Database.Password)
			}

			// Check slices (input2 replaces)
			if len(result.Tags) != 1 || result.Tags[0].Key != "priority" {
				t.Errorf("Tags: expected 1 tag with key 'priority', got %v", result.Tags)
			}

			t.Logf("Result: Name=%s, Port=%d, Rate=%f, Labels=%v",
				result.Name, result.Port, result.Rate, result.Labels)
		})
	}
}

func TestMergeWithNilInputs(t *testing.T) {
	input := &InputConfig{
		Name: Ptr("test-service"),
		Port: Ptr(8080),
	}

	// Test with nil first input
	result := MergeManual(nil, input)
	if result.Name != "test-service" {
		t.Errorf("Expected 'test-service', got '%s'", result.Name)
	}

	// Test with nil second input
	result = MergeManual(input, nil)
	if result.Name != "test-service" {
		t.Errorf("Expected 'test-service', got '%s'", result.Name)
	}

	// Test with both nil
	result = MergeManual(nil, nil)
	if result.Name != "" {
		t.Errorf("Expected empty string, got '%s'", result.Name)
	}
}

func BenchmarkMergeImplementations(b *testing.B) {
	input1 := &InputConfig{
		Name:       Ptr("service-a"),
		Port:       Ptr(8080),
		MaxRetries: Ptr(int32(3)),
		Hosts:      []string{"host1", "host2", "host3"},
		Labels:     map[string]string{"env": "prod", "version": "1.0"},
		Database: &InputDatabaseConfig{
			Host: Ptr("db.example.com"),
			Port: Ptr(5432),
		},
	}

	input2 := &InputConfig{
		Port:   Ptr(9090),
		Rate:   Ptr(1.5),
		Labels: map[string]string{"version": "2.0", "team": "platform"},
	}

	b.Run("MergeManual", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			MergeManual(input1, input2)
		}
	})

	b.Run("MergeReflection", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			MergeReflection(input1, input2)
		}
	})

	b.Run("MergeReflectionGeneric", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			MergeReflectionGeneric(input1, input2)
		}
	})

	b.Run("MergeJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			MergeJSON(input1, input2)
		}
	})

	b.Run("MergeJSONWithMap", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			MergeJSONWithMap(input1, input2)
		}
	})
}
