package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vexxhost/pod-tls-sidecar/pkg/podinfo"
	"github.com/vexxhost/pod-tls-sidecar/pkg/template"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
)

func TestNew(t *testing.T) {
	tmpl, err := New("api", &IssuerInfo{
		Kind: "IssuerKind",
		Name: "IssuerName",
	})

	assert.NoError(t, err)
	assert.NotNil(t, tmpl)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "test-pod",
			Namespace: "test-namespace",
			IP:        "1.2.3.4",
		},
		Hostname: "test-hostname",
		FQDN:     "test-hostname.atmosphere.dev",
	})
	require.NoError(t, err)

	assert.Equal(t, "test-pod-api", certificate.Name)
	assert.Equal(t, "test-namespace", certificate.Namespace)

	assert.Equal(t, "test-hostname.atmosphere.dev", certificate.Spec.CommonName)

	assert.Len(t, certificate.Spec.DNSNames, 2)
	assert.Contains(t, certificate.Spec.DNSNames, "test-hostname")
	assert.Contains(t, certificate.Spec.DNSNames, "test-hostname.atmosphere.dev")

	assert.Len(t, certificate.Spec.IPAddresses, 1)
	assert.Contains(t, certificate.Spec.IPAddresses, "1.2.3.4")

	assert.Len(t, certificate.Spec.Usages, 2)
	assert.Contains(t, certificate.Spec.Usages, cmv1.UsageClientAuth)
	assert.Contains(t, certificate.Spec.Usages, cmv1.UsageServerAuth)

	assert.Equal(t, "IssuerKind", certificate.Spec.IssuerRef.Kind)
	assert.Equal(t, "IssuerName", certificate.Spec.IssuerRef.Name)

	assert.Equal(t, "test-pod-api", certificate.Spec.SecretName)
}

func TestNewWithVNCName(t *testing.T) {
	tmpl, err := New("vnc", &IssuerInfo{
		Kind: "ClusterIssuer",
		Name: "vnc-issuer",
	})

	assert.NoError(t, err)
	assert.NotNil(t, tmpl)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "compute-node",
			Namespace: "openstack",
			IP:        "192.168.1.100",
		},
		Hostname: "compute-01",
		FQDN:     "compute-01.openstack.svc.cluster.local",
	})
	require.NoError(t, err)

	assert.Equal(t, "compute-node-vnc", certificate.Name)
	assert.Equal(t, "openstack", certificate.Namespace)
	assert.Equal(t, "compute-01.openstack.svc.cluster.local", certificate.Spec.CommonName)
	assert.Equal(t, "ClusterIssuer", certificate.Spec.IssuerRef.Kind)
	assert.Equal(t, "vnc-issuer", certificate.Spec.IssuerRef.Name)
	assert.Equal(t, "compute-node-vnc", certificate.Spec.SecretName)
}

func TestNewWithDifferentIssuerTypes(t *testing.T) {
	tests := []struct {
		name       string
		certName   string
		issuerKind string
		issuerName string
	}{
		{
			name:       "Issuer type",
			certName:   "api",
			issuerKind: "Issuer",
			issuerName: "ca-issuer",
		},
		{
			name:       "ClusterIssuer type",
			certName:   "vnc",
			issuerKind: "ClusterIssuer",
			issuerName: "cluster-ca",
		},
		{
			name:       "Custom name with dashes",
			certName:   "custom-cert",
			issuerKind: "Issuer",
			issuerName: "my-custom-issuer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := New(tt.certName, &IssuerInfo{
				Kind: tt.issuerKind,
				Name: tt.issuerName,
			})

			require.NoError(t, err)
			require.NotNil(t, tmpl)

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
			assert.Equal(t, "test-pod-"+tt.certName, certificate.Name)
		})
	}
}

func TestNewWithIPv6Address(t *testing.T) {
	tmpl, err := New("api", &IssuerInfo{
		Kind: "Issuer",
		Name: "test-issuer",
	})
	require.NoError(t, err)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "ipv6-pod",
			Namespace: "default",
			IP:        "2001:db8::1",
		},
		Hostname: "ipv6-host",
		FQDN:     "ipv6-host.example.com",
	})
	require.NoError(t, err)

	assert.Contains(t, certificate.Spec.IPAddresses, "2001:db8::1")
}

func TestNewWithSpecialCharactersInNames(t *testing.T) {
	tmpl, err := New("api-v2", &IssuerInfo{
		Kind: "Issuer",
		Name: "special-issuer-name",
	})
	require.NoError(t, err)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "pod-with-special-chars",
			Namespace: "my-namespace",
			IP:        "172.16.0.1",
		},
		Hostname: "host-123",
		FQDN:     "host-123.my-domain.example.org",
	})
	require.NoError(t, err)

	assert.Equal(t, "pod-with-special-chars-api-v2", certificate.Name)
	assert.Equal(t, "my-namespace", certificate.Namespace)
}

func TestNewValidatesCertificateUsages(t *testing.T) {
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

	// Validate that both client auth and server auth are present
	assert.Len(t, certificate.Spec.Usages, 2)
	assert.Contains(t, certificate.Spec.Usages, cmv1.UsageClientAuth)
	assert.Contains(t, certificate.Spec.Usages, cmv1.UsageServerAuth)
}

func TestNewWithMultipleDNSNames(t *testing.T) {
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
		Hostname: "short-name",
		FQDN:     "short-name.long.domain.example.com",
	})
	require.NoError(t, err)

	assert.Len(t, certificate.Spec.DNSNames, 2)
	assert.Contains(t, certificate.Spec.DNSNames, "short-name")
	assert.Contains(t, certificate.Spec.DNSNames, "short-name.long.domain.example.com")
	assert.Equal(t, "short-name.long.domain.example.com", certificate.Spec.CommonName)
}

func TestNewGeneratesCorrectSecretName(t *testing.T) {
	tests := []struct {
		name           string
		certName       string
		podName        string
		expectedSecret string
	}{
		{
			name:           "API certificate",
			certName:       "api",
			podName:        "libvirt-node",
			expectedSecret: "libvirt-node-api",
		},
		{
			name:           "VNC certificate",
			certName:       "vnc",
			podName:        "compute-host",
			expectedSecret: "compute-host-vnc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			assert.Equal(t, tt.expectedSecret, certificate.Spec.SecretName)
			assert.Equal(t, tt.expectedSecret, certificate.Name)
		})
	}
}

// Benchmark tests
func BenchmarkNew(b *testing.B) {
	issuer := &IssuerInfo{
		Kind: "Issuer",
		Name: "test-issuer",
	}

	for i := 0; i < b.N; i++ {
		_, err := New("api", issuer)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTemplateExecute(b *testing.B) {
	tmpl, err := New("api", &IssuerInfo{
		Kind: "Issuer",
		Name: "test-issuer",
	})
	if err != nil {
		b.Fatal(err)
	}

	values := &template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "test-pod",
			Namespace: "test-ns",
			IP:        "10.0.0.1",
		},
		Hostname: "test",
		FQDN:     "test.example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tmpl.Execute(values)
		if err != nil {
			b.Fatal(err)
		}
	}
}
