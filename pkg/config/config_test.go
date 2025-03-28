package config

import (
	"os"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *PluginConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &PluginConfig{
				MaxEnclavesPerNode:      2,
				EnclaveCPUAdvertisement: true,
			},
			wantErr: false,
		},
		{
			name: "max enclaves too high",
			config: &PluginConfig{
				MaxEnclavesPerNode:      5,
				EnclaveCPUAdvertisement: true,
			},
			wantErr: true,
		},
		{
			name: "max enclaves zero",
			config: &PluginConfig{
				MaxEnclavesPerNode:      0,
				EnclaveCPUAdvertisement: true,
			},
			wantErr: true,
		},
		{
			name: "max enclaves negative",
			config: &PluginConfig{
				MaxEnclavesPerNode:      -1,
				EnclaveCPUAdvertisement: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.config.MaxEnclavesPerNode != maxEnclavesPerInstance {
				t.Errorf("Validate() did not set MaxEnclavesPerNode to max value when error occurred")
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name                 string
		envMaxEnclaves       string
		envCPUAdvertisement  string
		wantMaxEnclaves      int
		wantCPUAdvertisement bool
		shouldUnsetEnvVars   bool
	}{
		{
			name:                 "valid values",
			envMaxEnclaves:       "2",
			envCPUAdvertisement:  "true",
			wantMaxEnclaves:      2,
			wantCPUAdvertisement: true,
		},
		{
			name:                 "invalid max enclaves",
			envMaxEnclaves:       "invalid",
			envCPUAdvertisement:  "true",
			wantMaxEnclaves:      maxEnclavesPerInstance,
			wantCPUAdvertisement: true,
		},
		{
			name:                 "invalid cpu advertisement",
			envMaxEnclaves:       "2",
			envCPUAdvertisement:  "invalid",
			wantMaxEnclaves:      2,
			wantCPUAdvertisement: false,
		},
		{
			name:                 "unset environment variables",
			shouldUnsetEnvVars:   true,
			wantMaxEnclaves:      maxEnclavesPerInstance,
			wantCPUAdvertisement: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldUnsetEnvVars {
				os.Unsetenv("MAX_ENCLAVES_PER_NODE")
				os.Unsetenv("ENCLAVE_CPU_ADVERTISEMENT")
			} else {
				os.Setenv("MAX_ENCLAVES_PER_NODE", tt.envMaxEnclaves)
				os.Setenv("ENCLAVE_CPU_ADVERTISEMENT", tt.envCPUAdvertisement)
			}

			defer func() {
				os.Unsetenv("MAX_ENCLAVES_PER_NODE")
				os.Unsetenv("ENCLAVE_CPU_ADVERTISEMENT")
			}()

			config := LoadConfig()

			if config.MaxEnclavesPerNode != tt.wantMaxEnclaves {
				t.Errorf("LoadConfig() MaxEnclavesPerNode = %v, want %v",
					config.MaxEnclavesPerNode, tt.wantMaxEnclaves)
			}

			if config.EnclaveCPUAdvertisement != tt.wantCPUAdvertisement {
				t.Errorf("LoadConfig() EnclaveCPUAdvertisement = %v, want %v",
					config.EnclaveCPUAdvertisement, tt.wantCPUAdvertisement)
			}
		})
	}
}
