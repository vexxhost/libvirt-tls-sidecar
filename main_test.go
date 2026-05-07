// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateQMPResponse(t *testing.T) {
	err := validateQMPResponse(`{"return":{},"id":"libvirt-1"}`)
	require.NoError(t, err)
}

func TestValidateQMPResponseWithQMPError(t *testing.T) {
	err := validateQMPResponse(`{"id":"libvirt-1","error":{"class":"GenericError","desc":"Unable to access credentials /etc/pki/libvirt-vnc/ca-cert.pem: No such file or directory"}}`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GenericError")
	assert.Contains(t, err.Error(), "Unable to access credentials")
}

func TestValidateQMPResponseWithInvalidJSON(t *testing.T) {
	err := validateQMPResponse(`not-json`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse qmp response")
}
