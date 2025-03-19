// Copyright 2025 Amazon.com, Inc. or its affiliates. All Rights Reserved.

package nitro_enclaves_cpu_plugin

import (
	"k8s-ne-device-plugin/pkg/config"
	"testing"
)

func TestDetermineAdvisableCPUs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:    "empty content",
			input:   "",
			want:    0,
			wantErr: false,
		},
		{
			name:    "single CPU",
			input:   "1",
			want:    1,
			wantErr: false,
		},
		{
			name:    "multiple single CPUs",
			input:   "1,2,3,20",
			want:    4,
			wantErr: false,
		},
		{
			name:    "corrupt file",
			input:   "1,2,",
			want:    2,
			wantErr: true,
		},
		{
			name:    "CPU range",
			input:   "1-3",
			want:    3,
			wantErr: false,
		},
		{
			name:    "CPU ranges",
			input:   "1-3,5-7",
			want:    6,
			wantErr: false,
		},
		{
			name:    "CPU range with single",
			input:   "1-3,9",
			want:    4,
			wantErr: false,
		},
		{
			name:    "multiple ranges",
			input:   "1-3,5,7-8",
			want:    6,
			wantErr: false,
		},
		{
			name:    "invalid range",
			input:   "1-3-4",
			wantErr: true,
		},
		{
			name:    "invalid number",
			input:   "a-b",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := determineAdvisableCPUs(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("determineAdvisableCPUs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("determineAdvisableCPUs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIncrementalDeviceIdGenerationSuccess(t *testing.T) {
	deviceName := "dummy_cpu"
	expected := "dummy_cpu_0"

	id := generateEnclaveCPUID(deviceName)

	if expected != id {
		t.Fatalf("Expected %s but got invalid id: %s!", expected, id)
		return
	}

	cpuIdCounter = 99
	_ = generateEnclaveCPUID(deviceName)
	deviceName = "dummy_cpu2"
	expected = "dummy_cpu2_100"
	id = generateEnclaveCPUID(deviceName)

	if expected != id {
		t.Fatalf("Expected %s but got invalid id: %s!", expected, id)
		return
	}
}

func TestValidateDeviceNameSuccess(t *testing.T) {
	p := NewNitroEnclavesCPUDevicePlugin(&config.PluginConfig{MaxEnclavesPerNode: 4, EnclaveCPUAdvertisement: false})

	// enclave cpu advertisement is disabled
	if len(p.devices) != 0 {
		t.Fatalf("Expected %v but got invalid id: %v!", 0, len(p.devices))
		return
	}
}
