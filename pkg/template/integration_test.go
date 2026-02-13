// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vexxhost/pod-tls-sidecar/pkg/podinfo"
	"github.com/vexxhost/pod-tls-sidecar/pkg/template"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestIntegrationCertManagerTypes validates that cert-manager types are compatible
func TestIntegrationCertManagerTypes(t *testing.T) {
	tmpl, err := New("api", &IssuerInfo{
		Kind: "ClusterIssuer",
		Name: "test-issuer",
	})
	require.NoError(t, err)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "test-pod",
			Namespace: "test-ns",
			IP:        "10.0.0.1",
		},
		Hostname: "test",
		FQDN:     "test.example.com",
	})
	require.NoError(t, err)

	// Validate cert-manager Certificate fields are correctly populated
	assert.IsType(t, &cmv1.Certificate{}, certificate)
	assert.NotNil(t, certificate.TypeMeta)
	assert.NotNil(t, certificate.ObjectMeta)
	assert.NotNil(t, certificate.Spec)
}

// TestIntegrationCertManagerAPIVersion validates API version compatibility
func TestIntegrationCertManagerAPIVersion(t *testing.T) {
	tmpl, err := New("api", &IssuerInfo{
		Kind: "Issuer",
		Name: "test-issuer",
	})
	require.NoError(t, err)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "test-pod",
			Namespace: "test-ns",
			IP:        "10.0.0.1",
		},
		Hostname: "test",
		FQDN:     "test.example.com",
	})
	require.NoError(t, err)

	// Ensure the certificate uses the correct API version
	assert.Equal(t, "cert-manager.io/v1", certificate.APIVersion)
	assert.Equal(t, "Certificate", certificate.Kind)
}

// TestIntegrationPodTLSSidecarTemplate validates pod-tls-sidecar template integration
func TestIntegrationPodTLSSidecarTemplate(t *testing.T) {
	// Test that the template package from pod-tls-sidecar works correctly
	tmpl, err := New("vnc", &IssuerInfo{
		Kind: "ClusterIssuer",
		Name: "ca-issuer",
	})
	require.NoError(t, err)
	require.NotNil(t, tmpl)

	// Execute the template with various inputs
	tests := []struct {
		name   string
		values *template.Values
	}{
		{
			name: "Basic values",
			values: &template.Values{
				PodInfo: podinfo.PodInfo{
					Name:      "pod1",
					Namespace: "ns1",
					IP:        "10.0.0.1",
				},
				Hostname: "host1",
				FQDN:     "host1.example.com",
			},
		},
		{
			name: "Complex namespace name",
			values: &template.Values{
				PodInfo: podinfo.PodInfo{
					Name:      "my-pod",
					Namespace: "openstack-nova",
					IP:        "172.16.0.10",
				},
				Hostname: "compute-01",
				FQDN:     "compute-01.nova.svc.cluster.local",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			certificate, err := tmpl.Execute(tt.values)
			require.NoError(t, err)
			assert.NotNil(t, certificate)
			assert.NotEmpty(t, certificate.Name)
			assert.NotEmpty(t, certificate.Namespace)
		})
	}
}

// TestIntegrationCertificateUsages validates cert-manager usage types
func TestIntegrationCertificateUsages(t *testing.T) {
	tmpl, err := New("api", &IssuerInfo{
		Kind: "Issuer",
		Name: "test-issuer",
	})
	require.NoError(t, err)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "test-pod",
			Namespace: "test-ns",
			IP:        "10.0.0.1",
		},
		Hostname: "test",
		FQDN:     "test.example.com",
	})
	require.NoError(t, err)

	// Validate that cert-manager usage types are correctly set
	expectedUsages := []cmv1.KeyUsage{
		cmv1.UsageClientAuth,
		cmv1.UsageServerAuth,
	}

	assert.ElementsMatch(t, expectedUsages, certificate.Spec.Usages)

	// Ensure the usage types are from the correct package
	for _, usage := range certificate.Spec.Usages {
		assert.IsType(t, cmv1.KeyUsage(""), usage)
	}
}

// TestIntegrationIssuerRef validates IssuerRef types
func TestIntegrationIssuerRef(t *testing.T) {
	tests := []struct {
		name       string
		issuerKind string
		issuerName string
	}{
		{
			name:       "ClusterIssuer",
			issuerKind: "ClusterIssuer",
			issuerName: "ca-issuer",
		},
		{
			name:       "Issuer",
			issuerKind: "Issuer",
			issuerName: "namespace-issuer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := New("api", &IssuerInfo{
				Kind: tt.issuerKind,
				Name: tt.issuerName,
			})
			require.NoError(t, err)

			certificate, err := tmpl.Execute(&template.Values{
				PodInfo: podinfo.PodInfo{
					Name:      "test-pod",
					Namespace: "test-ns",
					IP:        "10.0.0.1",
				},
				Hostname: "test",
				FQDN:     "test.example.com",
			})
			require.NoError(t, err)

			assert.Equal(t, tt.issuerKind, certificate.Spec.IssuerRef.Kind)
			assert.Equal(t, tt.issuerName, certificate.Spec.IssuerRef.Name)
		})
	}
}

// TestIntegrationMetadataFields validates k8s metadata integration
func TestIntegrationMetadataFields(t *testing.T) {
	tmpl, err := New("api", &IssuerInfo{
		Kind: "Issuer",
		Name: "test-issuer",
	})
	require.NoError(t, err)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "my-pod",
			Namespace: "my-namespace",
			IP:        "10.0.0.1",
		},
		Hostname: "hostname",
		FQDN:     "hostname.example.com",
	})
	require.NoError(t, err)

	// Validate ObjectMeta fields
	assert.Equal(t, "my-pod-api", certificate.ObjectMeta.Name)
	assert.Equal(t, "my-namespace", certificate.ObjectMeta.Namespace)
	assert.IsType(t, metav1.ObjectMeta{}, certificate.ObjectMeta)
}

// TestIntegrationDNSNamesAndIPAddresses validates DNS and IP field types
func TestIntegrationDNSNamesAndIPAddresses(t *testing.T) {
	tmpl, err := New("api", &IssuerInfo{
		Kind: "Issuer",
		Name: "test-issuer",
	})
	require.NoError(t, err)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "test-pod",
			Namespace: "test-ns",
			IP:        "192.168.1.100",
		},
		Hostname: "myhost",
		FQDN:     "myhost.example.com",
	})
	require.NoError(t, err)

	// Validate DNSNames is a string slice
	assert.IsType(t, []string{}, certificate.Spec.DNSNames)
	assert.Len(t, certificate.Spec.DNSNames, 2)

	// Validate IPAddresses is a string slice
	assert.IsType(t, []string{}, certificate.Spec.IPAddresses)
	assert.Len(t, certificate.Spec.IPAddresses, 1)
	assert.Equal(t, "192.168.1.100", certificate.Spec.IPAddresses[0])
}

// TestIntegrationCommonName validates CommonName field
func TestIntegrationCommonName(t *testing.T) {
	tmpl, err := New("api", &IssuerInfo{
		Kind: "Issuer",
		Name: "test-issuer",
	})
	require.NoError(t, err)

	fqdn := "test.example.com"
	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "test-pod",
			Namespace: "test-ns",
			IP:        "10.0.0.1",
		},
		Hostname: "test",
		FQDN:     fqdn,
	})
	require.NoError(t, err)

	// Validate CommonName matches FQDN
	assert.Equal(t, fqdn, certificate.Spec.CommonName)
	assert.IsType(t, "", certificate.Spec.CommonName)
}

// TestIntegrationSecretName validates SecretName field
func TestIntegrationSecretName(t *testing.T) {
	tests := []struct {
		certName   string
		podName    string
		wantSecret string
	}{
		{
			certName:   "api",
			podName:    "libvirt-node-1",
			wantSecret: "libvirt-node-1-api",
		},
		{
			certName:   "vnc",
			podName:    "compute-host",
			wantSecret: "compute-host-vnc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.certName, func(t *testing.T) {
			tmpl, err := New(tt.certName, &IssuerInfo{
				Kind: "Issuer",
				Name: "test-issuer",
			})
			require.NoError(t, err)

			certificate, err := tmpl.Execute(&template.Values{
				PodInfo: podinfo.PodInfo{
					Name:      tt.podName,
					Namespace: "test-ns",
					IP:        "10.0.0.1",
				},
				Hostname: "test",
				FQDN:     "test.example.com",
			})
			require.NoError(t, err)

			assert.Equal(t, tt.wantSecret, certificate.Spec.SecretName)
		})
	}
}

// TestIntegrationMultipleTemplateInstances validates creating multiple templates
func TestIntegrationMultipleTemplateInstances(t *testing.T) {
	// Create API template
	apiTmpl, err := New("api", &IssuerInfo{
		Kind: "ClusterIssuer",
		Name: "api-issuer",
	})
	require.NoError(t, err)

	// Create VNC template
	vncTmpl, err := New("vnc", &IssuerInfo{
		Kind: "Issuer",
		Name: "vnc-issuer",
	})
	require.NoError(t, err)

	values := &template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "test-pod",
			Namespace: "test-ns",
			IP:        "10.0.0.1",
		},
		Hostname: "test",
		FQDN:     "test.example.com",
	}

	// Execute both templates
	apiCert, err := apiTmpl.Execute(values)
	require.NoError(t, err)

	vncCert, err := vncTmpl.Execute(values)
	require.NoError(t, err)

	// Verify they have different names and issuers
	assert.Equal(t, "test-pod-api", apiCert.Name)
	assert.Equal(t, "test-pod-vnc", vncCert.Name)
	assert.Equal(t, "ClusterIssuer", apiCert.Spec.IssuerRef.Kind)
	assert.Equal(t, "Issuer", vncCert.Spec.IssuerRef.Kind)
}
