// Copyright (c) 2026 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build cgo
// +build cgo

package main

import (
	"os"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/suite"

	"github.com/vexxhost/libvirt-tls-sidecar/pkg/template"
)

type IssuerInfoSuite struct {
	suite.Suite
}

func (s *IssuerInfoSuite) SetupTest() {
	os.Clearenv()
}

func (s *IssuerInfoSuite) TestValidAPIIssuerConfiguration() {
	os.Setenv("API_ISSUER_KIND", "ClusterIssuer")
	os.Setenv("API_ISSUER_NAME", "ca-issuer")

	var issuer template.IssuerInfo
	err := envconfig.Process("API", &issuer)

	s.Require().NoError(err)
	s.Equal("ClusterIssuer", issuer.Kind)
	s.Equal("ca-issuer", issuer.Name)
}

func (s *IssuerInfoSuite) TestValidVNCIssuerConfiguration() {
	os.Setenv("VNC_ISSUER_KIND", "Issuer")
	os.Setenv("VNC_ISSUER_NAME", "vnc-issuer")

	var issuer template.IssuerInfo
	err := envconfig.Process("VNC", &issuer)

	s.Require().NoError(err)
	s.Equal("Issuer", issuer.Kind)
	s.Equal("vnc-issuer", issuer.Name)
}

func (s *IssuerInfoSuite) TestMissingIssuerKind() {
	os.Setenv("API_ISSUER_NAME", "ca-issuer")

	var issuer template.IssuerInfo
	err := envconfig.Process("API", &issuer)

	s.Require().Error(err)
	s.Contains(err.Error(), "required key ISSUER_KIND missing value")
}

func (s *IssuerInfoSuite) TestMissingIssuerName() {
	os.Setenv("API_ISSUER_KIND", "ClusterIssuer")

	var issuer template.IssuerInfo
	err := envconfig.Process("API", &issuer)

	s.Require().Error(err)
	s.Contains(err.Error(), "required key ISSUER_NAME missing value")
}

func (s *IssuerInfoSuite) TestMissingBothRequiredFields() {
	var issuer template.IssuerInfo
	err := envconfig.Process("API", &issuer)

	s.Require().Error(err)
	s.Contains(err.Error(), "required key ISSUER_KIND missing value")
}

func (s *IssuerInfoSuite) TestStructureFields() {
	issuer := template.IssuerInfo{
		Kind: "ClusterIssuer",
		Name: "test-issuer",
	}

	s.Equal("ClusterIssuer", issuer.Kind)
	s.Equal("test-issuer", issuer.Name)
}

func (s *IssuerInfoSuite) TestWithDifferentTypes() {
	tests := []struct {
		name       string
		issuerKind string
		issuerName string
	}{
		{"ClusterIssuer type", "ClusterIssuer", "cluster-ca"},
		{"Issuer type", "Issuer", "namespace-ca"},
		{"Complex issuer name", "ClusterIssuer", "prod-ca-issuer-v2"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			os.Clearenv()
			os.Setenv("TEST_ISSUER_KIND", tt.issuerKind)
			os.Setenv("TEST_ISSUER_NAME", tt.issuerName)

			var issuer template.IssuerInfo
			err := envconfig.Process("TEST", &issuer)

			s.Require().NoError(err)
			s.Equal(tt.issuerKind, issuer.Kind)
			s.Equal(tt.issuerName, issuer.Name)
		})
	}
}

func (s *IssuerInfoSuite) TestEmptyValuesAccepted() {
	os.Setenv("TEST_ISSUER_KIND", "")
	os.Setenv("TEST_ISSUER_NAME", "")

	var issuer template.IssuerInfo
	err := envconfig.Process("TEST", &issuer)

	s.Require().NoError(err)
	s.Equal("", issuer.Kind)
	s.Equal("", issuer.Name)
}

func TestIssuerInfoSuite(t *testing.T) {
	suite.Run(t, new(IssuerInfoSuite))
}

func BenchmarkEnvconfigProcess(b *testing.B) {
	os.Clearenv()
	os.Setenv("BENCH_ISSUER_KIND", "ClusterIssuer")
	os.Setenv("BENCH_ISSUER_NAME", "bench-issuer")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var issuer template.IssuerInfo
		if err := envconfig.Process("BENCH", &issuer); err != nil {
			b.Fatal(err)
		}
	}
}
