// Copyright (c) 2026 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build cgo
// +build cgo

package app

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/vexxhost/pod-tls-sidecar/pkg/tls"
)

type AppSuite struct {
	suite.Suite
}

func (s *AppSuite) TestAPIPaths() {
	paths := APIPaths()
	s.NotNil(paths)
	s.Len(paths.CertificateAuthorityPaths, 2)
	s.Len(paths.CertificatePaths, 4)
	s.Len(paths.CertificateKeyPaths, 4)
	s.Contains(paths.CertificateAuthorityPaths, "/etc/pki/CA/cacert.pem")
	s.Contains(paths.CertificatePaths, "/etc/pki/libvirt/servercert.pem")
}

func (s *AppSuite) TestVNCPaths() {
	paths := VNCPaths()
	s.NotNil(paths)
	s.Len(paths.CertificateAuthorityPaths, 1)
	s.Len(paths.CertificatePaths, 1)
	s.Len(paths.CertificateKeyPaths, 1)
	s.Equal("/etc/pki/libvirt-vnc/ca-cert.pem", paths.CertificateAuthorityPaths[0])
	s.Equal("/etc/pki/libvirt-vnc/server-cert.pem", paths.CertificatePaths[0])
}

func (s *AppSuite) TestAPIUpdateCallback() {
	callback := APIUpdateCallback()
	s.NotNil(callback)
}

func (s *AppSuite) TestVNCUpdateCallback() {
	callback := VNCUpdateCallback()
	s.NotNil(callback)
}

func (s *AppSuite) TestPathsReturnCorrectTypes() {
	apiPaths := APIPaths()
	vncPaths := VNCPaths()
	
	s.IsType(&tls.WritePathConfig{}, apiPaths)
	s.IsType(&tls.WritePathConfig{}, vncPaths)
}

func TestAppSuite(t *testing.T) {
	suite.Run(t, new(AppSuite))
}
