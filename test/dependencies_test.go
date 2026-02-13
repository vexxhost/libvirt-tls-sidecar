// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package dependencies_test validates that all key dependencies are properly integrated
package dependencies_test

import (
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/vexxhost/pod-tls-sidecar/pkg/podinfo"
	"github.com/vexxhost/pod-tls-sidecar/pkg/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// TestDependencyCertManagerImport validates cert-manager can be imported
func TestDependencyCertManagerImport(t *testing.T) {
	// Test that cert-manager types are available
	cert := &cmv1.Certificate{}
	assert.NotNil(t, cert)

	// Test Certificate struct has expected fields
	cert.Name = "test-cert"
	cert.Namespace = "test-ns"
	cert.Spec.CommonName = "example.com"

	assert.Equal(t, "test-cert", cert.Name)
	assert.Equal(t, "test-ns", cert.Namespace)
	assert.Equal(t, "example.com", cert.Spec.CommonName)
}

// TestDependencyCertManagerUsages validates KeyUsage types
func TestDependencyCertManagerUsages(t *testing.T) {
	usages := []cmv1.KeyUsage{
		cmv1.UsageClientAuth,
		cmv1.UsageServerAuth,
		cmv1.UsageDigitalSignature,
		cmv1.UsageKeyEncipherment,
	}

	assert.Len(t, usages, 4)
	assert.Contains(t, usages, cmv1.UsageClientAuth)
	assert.Contains(t, usages, cmv1.UsageServerAuth)
}

// TestDependencyCertManagerIssuerRef validates IssuerRef
func TestDependencyCertManagerIssuerRef(t *testing.T) {
	cert := &cmv1.Certificate{
		Spec: cmv1.CertificateSpec{
			IssuerRef: cmmeta.IssuerReference{
				Name: "test-issuer",
				Kind: "ClusterIssuer",
			},
		},
	}

	assert.Equal(t, "test-issuer", cert.Spec.IssuerRef.Name)
	assert.Equal(t, "ClusterIssuer", cert.Spec.IssuerRef.Kind)
}

// TestDependencyPodTLSSidecarTemplate validates pod-tls-sidecar template package
func TestDependencyPodTLSSidecarTemplate(t *testing.T) {
	// Test that we can create a template
	tmpl, err := template.New(`
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{ .PodInfo.Name }}
  namespace: {{ .PodInfo.Namespace }}
spec:
  commonName: "{{ .FQDN }}"
`)
	require.NoError(t, err)
	assert.NotNil(t, tmpl)
}

// TestDependencyPodTLSSidecarTemplateExecution validates template execution
func TestDependencyPodTLSSidecarTemplateExecution(t *testing.T) {
	tmpl, err := template.New(`
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{ .PodInfo.Name }}-cert
  namespace: {{ .PodInfo.Namespace }}
spec:
  commonName: "{{ .FQDN }}"
  dnsNames:
    - "{{ .Hostname }}"
  ipAddresses:
    - "{{ .PodInfo.IP }}"
`)
	require.NoError(t, err)

	values := &template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "my-pod",
			Namespace: "my-namespace",
			IP:        "10.0.0.1",
		},
		Hostname: "host",
		FQDN:     "host.example.com",
	}

	cert, err := tmpl.Execute(values)
	require.NoError(t, err)

	assert.Equal(t, "my-pod-cert", cert.Name)
	assert.Equal(t, "my-namespace", cert.Namespace)
	assert.Equal(t, "host.example.com", cert.Spec.CommonName)
	assert.Contains(t, cert.Spec.DNSNames, "host")
	assert.Contains(t, cert.Spec.IPAddresses, "10.0.0.1")
}

// TestDependencyPodTLSSidecarPodInfo validates podinfo package
func TestDependencyPodTLSSidecarPodInfo(t *testing.T) {
	pod := podinfo.PodInfo{
		Name:      "test-pod",
		Namespace: "test-namespace",
		IP:        "192.168.1.1",
	}

	assert.Equal(t, "test-pod", pod.Name)
	assert.Equal(t, "test-namespace", pod.Namespace)
	assert.Equal(t, "192.168.1.1", pod.IP)
}

// TestDependencyEnvconfig validates envconfig package
func TestDependencyEnvconfig(t *testing.T) {
	type Config struct {
		Host string `envconfig:"HOST" required:"true"`
		Port int    `envconfig:"PORT" default:"8080"`
	}

	t.Setenv("TEST_HOST", "localhost")

	var cfg Config
	err := envconfig.Process("TEST", &cfg)
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
}

// TestDependencyEnvconfigRequired validates required field handling
func TestDependencyEnvconfigRequired(t *testing.T) {
	type Config struct {
		RequiredField string `envconfig:"REQUIRED_FIELD" required:"true"`
	}

	var cfg Config
	err := envconfig.Process("TEST", &cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required key")
	assert.Contains(t, err.Error(), "REQUIRED_FIELD")
}

// TestDependencyK8sClientGoRest validates k8s.io/client-go/rest package
func TestDependencyK8sClientGoRest(t *testing.T) {
	// Test that we can create a rest config structure
	config := &rest.Config{
		Host: "https://kubernetes.default.svc",
	}

	assert.Equal(t, "https://kubernetes.default.svc", config.Host)
	assert.NotNil(t, config)
}

// TestDependencyK8sApimachineryMetav1 validates k8s.io/apimachinery/pkg/apis/meta/v1
func TestDependencyK8sApimachineryMetav1(t *testing.T) {
	// Test ObjectMeta
	meta := metav1.ObjectMeta{
		Name:      "test-object",
		Namespace: "test-namespace",
		Labels: map[string]string{
			"app": "test",
		},
	}

	assert.Equal(t, "test-object", meta.Name)
	assert.Equal(t, "test-namespace", meta.Namespace)
	assert.Equal(t, "test", meta.Labels["app"])
}

// TestDependencyK8sApimachineryTypeMeta validates TypeMeta
func TestDependencyK8sApimachineryTypeMeta(t *testing.T) {
	typeMeta := metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "ConfigMap",
	}

	assert.Equal(t, "v1", typeMeta.APIVersion)
	assert.Equal(t, "ConfigMap", typeMeta.Kind)
}

// TestDependencyCertManagerCertificateConditions validates Certificate conditions
func TestDependencyCertManagerCertificateConditions(t *testing.T) {
	cert := &cmv1.Certificate{
		Status: cmv1.CertificateStatus{
			Conditions: []cmv1.CertificateCondition{
				{
					Type:   cmv1.CertificateConditionReady,
					Status: cmmeta.ConditionTrue,
				},
			},
		},
	}

	assert.Len(t, cert.Status.Conditions, 1)
	assert.Equal(t, cmv1.CertificateConditionReady, cert.Status.Conditions[0].Type)
	assert.Equal(t, cmmeta.ConditionTrue, cert.Status.Conditions[0].Status)
}

// TestDependencyCertManagerCertificateSpec validates full CertificateSpec
func TestDependencyCertManagerCertificateSpec(t *testing.T) {
	spec := cmv1.CertificateSpec{
		CommonName: "example.com",
		DNSNames:   []string{"example.com", "www.example.com"},
		IPAddresses: []string{"192.168.1.1"},
		Usages: []cmv1.KeyUsage{
			cmv1.UsageServerAuth,
			cmv1.UsageClientAuth,
		},
		IssuerRef: cmmeta.IssuerReference{
			Name: "ca-issuer",
			Kind: "ClusterIssuer",
		},
		SecretName: "example-tls",
	}

	assert.Equal(t, "example.com", spec.CommonName)
	assert.Len(t, spec.DNSNames, 2)
	assert.Len(t, spec.IPAddresses, 1)
	assert.Len(t, spec.Usages, 2)
	assert.Equal(t, "ca-issuer", spec.IssuerRef.Name)
	assert.Equal(t, "example-tls", spec.SecretName)
}

// TestDependencyIntegrationFullCertificate validates creating a complete certificate
func TestDependencyIntegrationFullCertificate(t *testing.T) {
	cert := &cmv1.Certificate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "Certificate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-certificate",
			Namespace: "default",
		},
		Spec: cmv1.CertificateSpec{
			CommonName: "test.example.com",
			DNSNames:   []string{"test.example.com", "test"},
			IPAddresses: []string{"10.0.0.1"},
			Usages: []cmv1.KeyUsage{
				cmv1.UsageClientAuth,
				cmv1.UsageServerAuth,
			},
			IssuerRef: cmmeta.IssuerReference{
				Kind: "ClusterIssuer",
				Name: "ca-issuer",
			},
			SecretName: "test-certificate-secret",
		},
	}

	// Validate the certificate structure
	assert.Equal(t, "cert-manager.io/v1", cert.APIVersion)
	assert.Equal(t, "Certificate", cert.Kind)
	assert.Equal(t, "test-certificate", cert.Name)
	assert.Equal(t, "default", cert.Namespace)
	assert.Equal(t, "test.example.com", cert.Spec.CommonName)
	assert.Equal(t, "test-certificate-secret", cert.Spec.SecretName)
}

// TestDependencyTemplateWithMultipleValues validates template with various input values
func TestDependencyTemplateWithMultipleValues(t *testing.T) {
	tests := []struct {
		name      string
		podName   string
		namespace string
		ip        string
		hostname  string
		fqdn      string
	}{
		{
			name:      "Simple values",
			podName:   "pod1",
			namespace: "ns1",
			ip:        "10.0.0.1",
			hostname:  "host1",
			fqdn:      "host1.example.com",
		},
		{
			name:      "Complex values with dashes",
			podName:   "my-pod-with-dashes",
			namespace: "openstack-nova",
			ip:        "192.168.1.100",
			hostname:  "compute-01",
			fqdn:      "compute-01.nova.svc.cluster.local",
		},
		{
			name:      "IPv6 address",
			podName:   "ipv6-pod",
			namespace: "default",
			ip:        "2001:db8::1",
			hostname:  "ipv6-host",
			fqdn:      "ipv6-host.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New(`
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{ .PodInfo.Name }}-cert
  namespace: {{ .PodInfo.Namespace }}
spec:
  commonName: "{{ .FQDN }}"
  ipAddresses:
    - "{{ .PodInfo.IP }}"
`)
			require.NoError(t, err)

			values := &template.Values{
				PodInfo: podinfo.PodInfo{
					Name:      tt.podName,
					Namespace: tt.namespace,
					IP:        tt.ip,
				},
				Hostname: tt.hostname,
				FQDN:     tt.fqdn,
			}

			cert, err := tmpl.Execute(values)
			require.NoError(t, err)

			assert.Equal(t, tt.podName+"-cert", cert.Name)
			assert.Equal(t, tt.namespace, cert.Namespace)
			assert.Equal(t, tt.fqdn, cert.Spec.CommonName)
			assert.Contains(t, cert.Spec.IPAddresses, tt.ip)
		})
	}
}

// BenchmarkCertManagerCertificateCreation benchmarks certificate creation
func BenchmarkCertManagerCertificateCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cert := &cmv1.Certificate{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "cert-manager.io/v1",
				Kind:       "Certificate",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cert",
				Namespace: "default",
			},
			Spec: cmv1.CertificateSpec{
				CommonName: "test.example.com",
				IssuerRef: cmmeta.IssuerReference{
					Kind: "Issuer",
					Name: "test-issuer",
				},
			},
		}
		_ = cert
	}
}

// BenchmarkTemplateExecution benchmarks template execution
func BenchmarkTemplateExecution(b *testing.B) {
	tmpl, err := template.New(`
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{ .PodInfo.Name }}-cert
  namespace: {{ .PodInfo.Namespace }}
spec:
  commonName: "{{ .FQDN }}"
`)
	if err != nil {
		b.Fatal(err)
	}

	values := &template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "bench-pod",
			Namespace: "bench-ns",
			IP:        "10.0.0.1",
		},
		Hostname: "bench",
		FQDN:     "bench.example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tmpl.Execute(values)
		if err != nil {
			b.Fatal(err)
		}
	}
}
