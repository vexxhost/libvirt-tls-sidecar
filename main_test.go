// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build cgo
// +build cgo

package main

import (
	"os"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssuerInfoEnvconfig(t *testing.T) {
	tests := []struct {
		name        string
		prefix      string
		envVars     map[string]string
		wantErr     bool
		expectedErr string
	}{
		{
			name:   "Valid API issuer configuration",
			prefix: "API",
			envVars: map[string]string{
				"API_ISSUER_KIND": "ClusterIssuer",
				"API_ISSUER_NAME": "ca-issuer",
			},
			wantErr: false,
		},
		{
			name:   "Valid VNC issuer configuration",
			prefix: "VNC",
			envVars: map[string]string{
				"VNC_ISSUER_KIND": "Issuer",
				"VNC_ISSUER_NAME": "vnc-issuer",
			},
			wantErr: false,
		},
		{
			name:   "Missing ISSUER_KIND",
			prefix: "API",
			envVars: map[string]string{
				"API_ISSUER_NAME": "ca-issuer",
			},
			wantErr:     true,
			expectedErr: "required key ISSUER_KIND missing value",
		},
		{
			name:   "Missing ISSUER_NAME",
			prefix: "API",
			envVars: map[string]string{
				"API_ISSUER_KIND": "ClusterIssuer",
			},
			wantErr:     true,
			expectedErr: "required key ISSUER_NAME missing value",
		},
		{
			name:        "Missing both required fields",
			prefix:      "API",
			envVars:     map[string]string{},
			wantErr:     true,
			expectedErr: "required key ISSUER_KIND missing value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			var issuer IssuerInfo
			err := envconfig.Process(tt.prefix, &issuer)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.envVars[tt.prefix+"_ISSUER_KIND"], issuer.Kind)
				assert.Equal(t, tt.envVars[tt.prefix+"_ISSUER_NAME"], issuer.Name)
			}
		})
	}
}

func TestIssuerInfoStructure(t *testing.T) {
	t.Run("IssuerInfo has correct fields", func(t *testing.T) {
		issuer := IssuerInfo{
			Kind: "ClusterIssuer",
			Name: "test-issuer",
		}

		assert.Equal(t, "ClusterIssuer", issuer.Kind)
		assert.Equal(t, "test-issuer", issuer.Name)
	})

	t.Run("IssuerInfo can be used with envconfig", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_ISSUER_KIND", "Issuer")
		os.Setenv("TEST_ISSUER_NAME", "my-issuer")

		var issuer IssuerInfo
		err := envconfig.Process("TEST", &issuer)

		require.NoError(t, err)
		assert.Equal(t, "Issuer", issuer.Kind)
		assert.Equal(t, "my-issuer", issuer.Name)
	})
}

func TestIssuerInfoWithDifferentTypes(t *testing.T) {
	tests := []struct {
		name       string
		issuerKind string
		issuerName string
	}{
		{
			name:       "ClusterIssuer type",
			issuerKind: "ClusterIssuer",
			issuerName: "cluster-ca",
		},
		{
			name:       "Issuer type",
			issuerKind: "Issuer",
			issuerName: "namespace-ca",
		},
		{
			name:       "Complex issuer name",
			issuerKind: "ClusterIssuer",
			issuerName: "prod-ca-issuer-v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("TEST_ISSUER_KIND", tt.issuerKind)
			os.Setenv("TEST_ISSUER_NAME", tt.issuerName)

			var issuer IssuerInfo
			err := envconfig.Process("TEST", &issuer)

			require.NoError(t, err)
			assert.Equal(t, tt.issuerKind, issuer.Kind)
			assert.Equal(t, tt.issuerName, issuer.Name)
		})
	}
}

func TestIssuerInfoEnvconfigTags(t *testing.T) {
	t.Run("Verify envconfig tags are correct", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("PREFIX_ISSUER_KIND", "TestKind")
		os.Setenv("PREFIX_ISSUER_NAME", "TestName")

		var issuer IssuerInfo
		err := envconfig.Process("PREFIX", &issuer)

		require.NoError(t, err)
		assert.Equal(t, "TestKind", issuer.Kind)
		assert.Equal(t, "TestName", issuer.Name)
	})
}

func TestIssuerInfoEmptyValues(t *testing.T) {
	t.Run("Empty values are accepted by envconfig", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_ISSUER_KIND", "")
		os.Setenv("TEST_ISSUER_NAME", "")

		var issuer IssuerInfo
		err := envconfig.Process("TEST", &issuer)

		// envconfig accepts empty strings for required fields
		require.NoError(t, err)
		assert.Equal(t, "", issuer.Kind)
		assert.Equal(t, "", issuer.Name)
	})
}

func TestIssuerInfoCaseSensitivity(t *testing.T) {
	t.Run("Environment variable names are case sensitive", func(t *testing.T) {
		os.Clearenv()
		// envconfig converts field names to uppercase by default
		os.Setenv("TEST_ISSUER_KIND", "ClusterIssuer")
		os.Setenv("TEST_ISSUER_NAME", "ca-issuer")

		var issuer IssuerInfo
		err := envconfig.Process("TEST", &issuer)

		require.NoError(t, err)
		assert.Equal(t, "ClusterIssuer", issuer.Kind)
		assert.Equal(t, "ca-issuer", issuer.Name)
	})
}

// Benchmark tests for environment configuration parsing
func BenchmarkEnvconfigProcess(b *testing.B) {
	os.Clearenv()
	os.Setenv("BENCH_ISSUER_KIND", "ClusterIssuer")
	os.Setenv("BENCH_ISSUER_NAME", "bench-issuer")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var issuer IssuerInfo
		if err := envconfig.Process("BENCH", &issuer); err != nil {
			b.Fatal(err)
		}
	}
}
