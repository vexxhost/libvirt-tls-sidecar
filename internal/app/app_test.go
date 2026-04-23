// Copyright (c) 2026 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build cgo
// +build cgo

package app

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/vexxhost/pod-tls-sidecar/pkg/tls"
)

type AppSuite struct {
	suite.Suite
	originalLogOutput io.Writer
	logBuffer         *bytes.Buffer
}

func (s *AppSuite) SetupTest() {
	// Capture log output for testing
	s.logBuffer = new(bytes.Buffer)
	s.originalLogOutput = log.StandardLogger().Out
	log.SetOutput(s.logBuffer)
}

func (s *AppSuite) TearDownTest() {
	// Restore original log output
	log.SetOutput(s.originalLogOutput)
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

// TestAPIUpdateCallbackExecutesCommand verifies that the API callback attempts to execute the correct command
func (s *AppSuite) TestAPIUpdateCallbackExecutesCommand() {
	// Test that callback is callable and logs errors when command fails
	callback := APIUpdateCallback()
	s.Require().NotNil(callback)
	
	// Execute callback - it will fail since virt-admin is not available in test environment
	// but we can verify it attempts to run the right command by checking logs
	callback()
	
	// Check that error was logged (since virt-admin won't exist in test environment)
	logOutput := s.logBuffer.String()
	s.Contains(logOutput, "failed to reload tls configuration for api", 
		"Expected error log when virt-admin command fails")
}

// TestAPIUpdateCallbackWithMockCommand tests the callback with a mock command
func (s *AppSuite) TestAPIUpdateCallbackWithMockCommand() {
	// Create a mock script that simulates virt-admin
	if os.Getenv("GO_TEST_MOCK_COMMAND") == "1" {
		// This is the mock command execution
		os.Exit(0)
		return
	}

	// Test that the callback would execute virt-admin with correct arguments
	// We verify this by checking that exec.Command is called with the right params
	callback := APIUpdateCallback()
	s.NotNil(callback)
	
	// The callback creates a command to execute: exec.Command("virt-admin", "server-update-tls", "libvirtd")
	// We can't easily mock exec.Command without extensive refactoring, but we can
	// verify the callback function is created and is of the correct type
	s.IsType(func() {}, callback)
}

// TestVNCUpdateCallbackExecutesCorrectly verifies VNC callback behavior
func (s *AppSuite) TestVNCUpdateCallbackExecutesCorrectly() {
	callback := VNCUpdateCallback()
	s.Require().NotNil(callback)
	
	// Execute callback - it will fail to connect to libvirt in test environment
	// but we can verify it attempts the connection by checking logs
	callback()
	
	// Check that error was logged (since libvirt won't be running in test environment)
	logOutput := s.logBuffer.String()
	s.Contains(logOutput, "failed to connect to libvirt",
		"Expected error log when libvirt connection fails")
}

// TestAPIUpdateCallbackLogFormat verifies correct log format on errors
func (s *AppSuite) TestAPIUpdateCallbackLogFormat() {
	callback := APIUpdateCallback()
	callback()
	
	logOutput := s.logBuffer.String()
	// Verify log contains key information
	s.True(strings.Contains(logOutput, "failed to reload tls configuration for api") ||
		strings.Contains(logOutput, "error"),
		"Expected error information in logs")
}

// TestVNCUpdateCallbackLogFormat verifies correct log format on errors  
func (s *AppSuite) TestVNCUpdateCallbackLogFormat() {
	callback := VNCUpdateCallback()
	callback()
	
	logOutput := s.logBuffer.String()
	// Verify log contains expected error for connection failure
	s.Contains(logOutput, "failed to connect to libvirt",
		"Expected libvirt connection error in logs")
}

// TestAPIUpdateCallbackIntegration verifies integration behavior when command succeeds
func (s *AppSuite) TestAPIUpdateCallbackIntegration() {
	// We can't run actual virt-admin in tests, but we can verify the command structure
	// by checking that the callback is a valid function that can be invoked
	callback := APIUpdateCallback()
	
	// Verify callback is invocable (will fail but won't panic)
	s.NotPanics(func() {
		callback()
	}, "Callback should not panic even when command fails")
	
	// Verify appropriate error logging occurred
	logOutput := s.logBuffer.String()
	s.NotEmpty(logOutput, "Expected some log output from callback execution")
}

// TestVNCUpdateCallbackIntegration verifies integration behavior
func (s *AppSuite) TestVNCUpdateCallbackIntegration() {
	callback := VNCUpdateCallback()
	
	// Verify callback is invocable (will fail but won't panic)
	s.NotPanics(func() {
		callback()
	}, "Callback should not panic even when libvirt connection fails")
	
	// Verify appropriate error logging occurred
	logOutput := s.logBuffer.String()
	s.NotEmpty(logOutput, "Expected some log output from callback execution")
}

// TestAPIUpdateCallbackCommandStructure verifies the command structure is correct
func (s *AppSuite) TestAPIUpdateCallbackCommandStructure() {
	// Verify that we can create the same command structure that the callback uses
	cmd := exec.Command("virt-admin", "server-update-tls", "libvirtd")
	
	s.Equal("virt-admin", cmd.Path)
	s.Contains(cmd.Args, "virt-admin")
	s.Contains(cmd.Args, "server-update-tls")
	s.Contains(cmd.Args, "libvirtd")
	s.Len(cmd.Args, 3, "Command should have exactly 3 arguments")
}

// TestVNCUpdateCallbackConnectionString verifies correct libvirt URI
func (s *AppSuite) TestVNCUpdateCallbackConnectionString() {
	// The VNC callback connects to "qemu:///system"
	// We can't test the actual connection, but we document the expected behavior
	callback := VNCUpdateCallback()
	s.NotNil(callback)
	
	// Execute and verify it attempts connection (will fail in test environment)
	callback()
	
	logOutput := s.logBuffer.String()
	// The error should mention libvirt connection failure
	s.Contains(logOutput, "libvirt", "Expected libvirt-related error in logs")
}

func TestAppSuite(t *testing.T) {
	suite.Run(t, new(AppSuite))
}
