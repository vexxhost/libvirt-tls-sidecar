// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package dependencies_test

import (
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/suite"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/vexxhost/pod-tls-sidecar/pkg/podinfo"
	"github.com/vexxhost/pod-tls-sidecar/pkg/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type CertManagerSuite struct {
	suite.Suite
}

func (s *CertManagerSuite) TestImport() {
	cert := &cmv1.Certificate{}
	s.NotNil(cert)

	cert.Name = "test-cert"
	cert.Namespace = "test-ns"
	cert.Spec.CommonName = "example.com"

	s.Equal("test-cert", cert.Name)
	s.Equal("test-ns", cert.Namespace)
	s.Equal("example.com", cert.Spec.CommonName)
}

func (s *CertManagerSuite) TestUsages() {
	usages := []cmv1.KeyUsage{
		cmv1.UsageClientAuth,
		cmv1.UsageServerAuth,
		cmv1.UsageDigitalSignature,
		cmv1.UsageKeyEncipherment,
	}

	s.Len(usages, 4)
	s.Contains(usages, cmv1.UsageClientAuth)
	s.Contains(usages, cmv1.UsageServerAuth)
}

func (s *CertManagerSuite) TestIssuerRef() {
	cert := &cmv1.Certificate{
		Spec: cmv1.CertificateSpec{
			IssuerRef: cmmeta.IssuerReference{
				Name: "test-issuer",
				Kind: "ClusterIssuer",
			},
		},
	}

	s.Equal("test-issuer", cert.Spec.IssuerRef.Name)
	s.Equal("ClusterIssuer", cert.Spec.IssuerRef.Kind)
}

func (s *CertManagerSuite) TestCertificateConditions() {
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

	s.Len(cert.Status.Conditions, 1)
	s.Equal(cmv1.CertificateConditionReady, cert.Status.Conditions[0].Type)
	s.Equal(cmmeta.ConditionTrue, cert.Status.Conditions[0].Status)
}

func (s *CertManagerSuite) TestCertificateSpec() {
	spec := cmv1.CertificateSpec{
		CommonName:  "example.com",
		DNSNames:    []string{"example.com", "www.example.com"},
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

	s.Equal("example.com", spec.CommonName)
	s.Len(spec.DNSNames, 2)
	s.Len(spec.IPAddresses, 1)
	s.Len(spec.Usages, 2)
	s.Equal("ca-issuer", spec.IssuerRef.Name)
	s.Equal("example-tls", spec.SecretName)
}

func (s *CertManagerSuite) TestFullCertificate() {
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
			CommonName:  "test.example.com",
			DNSNames:    []string{"test.example.com", "test"},
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

	s.Equal("cert-manager.io/v1", cert.APIVersion)
	s.Equal("Certificate", cert.Kind)
	s.Equal("test-certificate", cert.Name)
	s.Equal("default", cert.Namespace)
	s.Equal("test.example.com", cert.Spec.CommonName)
	s.Equal("test-certificate-secret", cert.Spec.SecretName)
}

func TestCertManagerSuite(t *testing.T) {
	suite.Run(t, new(CertManagerSuite))
}

type PodTLSSidecarSuite struct {
	suite.Suite
}

func (s *PodTLSSidecarSuite) TestTemplate() {
	tmpl, err := template.New(`
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{ .PodInfo.Name }}
  namespace: {{ .PodInfo.Namespace }}
spec:
  commonName: "{{ .FQDN }}"
`)
	s.Require().NoError(err)
	s.NotNil(tmpl)
}

func (s *PodTLSSidecarSuite) TestTemplateExecution() {
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
	s.Require().NoError(err)

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
	s.Require().NoError(err)

	s.Equal("my-pod-cert", cert.Name)
	s.Equal("my-namespace", cert.Namespace)
	s.Equal("host.example.com", cert.Spec.CommonName)
	s.Contains(cert.Spec.DNSNames, "host")
	s.Contains(cert.Spec.IPAddresses, "10.0.0.1")
}

func (s *PodTLSSidecarSuite) TestPodInfo() {
	pod := podinfo.PodInfo{
		Name:      "test-pod",
		Namespace: "test-namespace",
		IP:        "192.168.1.1",
	}

	s.Equal("test-pod", pod.Name)
	s.Equal("test-namespace", pod.Namespace)
	s.Equal("192.168.1.1", pod.IP)
}

func (s *PodTLSSidecarSuite) TestTemplateWithMultipleValues() {
	tests := []struct {
		name      string
		podName   string
		namespace string
		ip        string
		hostname  string
		fqdn      string
	}{
		{"Simple values", "pod1", "ns1", "10.0.0.1", "host1", "host1.example.com"},
		{"Complex values with dashes", "my-pod-with-dashes", "openstack-nova", "192.168.1.100", "compute-01", "compute-01.nova.svc.cluster.local"},
		{"IPv6 address", "ipv6-pod", "default", "2001:db8::1", "ipv6-host", "ipv6-host.example.com"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
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
			s.Require().NoError(err)

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
			s.Require().NoError(err)

			s.Equal(tt.podName+"-cert", cert.Name)
			s.Equal(tt.namespace, cert.Namespace)
			s.Equal(tt.fqdn, cert.Spec.CommonName)
			s.Contains(cert.Spec.IPAddresses, tt.ip)
		})
	}
}

func TestPodTLSSidecarSuite(t *testing.T) {
	suite.Run(t, new(PodTLSSidecarSuite))
}

type KubernetesSuite struct {
	suite.Suite
}

func (s *KubernetesSuite) TestEnvconfig() {
	type Config struct {
		Host string `envconfig:"HOST" required:"true"`
		Port int    `envconfig:"PORT" default:"8080"`
	}

	s.T().Setenv("TEST_HOST", "localhost")

	var cfg Config
	err := envconfig.Process("TEST", &cfg)
	s.Require().NoError(err)

	s.Equal("localhost", cfg.Host)
	s.Equal(8080, cfg.Port)
}

func (s *KubernetesSuite) TestEnvconfigRequired() {
	type Config struct {
		RequiredField string `envconfig:"REQUIRED_FIELD" required:"true"`
	}

	var cfg Config
	err := envconfig.Process("TEST", &cfg)
	s.Error(err)
	s.Contains(err.Error(), "required key")
	s.Contains(err.Error(), "REQUIRED_FIELD")
}

func (s *KubernetesSuite) TestClientGoRest() {
	config := &rest.Config{
		Host: "https://kubernetes.default.svc",
	}

	s.Equal("https://kubernetes.default.svc", config.Host)
	s.NotNil(config)
}

func (s *KubernetesSuite) TestApimachineryMetav1() {
	meta := metav1.ObjectMeta{
		Name:      "test-object",
		Namespace: "test-namespace",
		Labels: map[string]string{
			"app": "test",
		},
	}

	s.Equal("test-object", meta.Name)
	s.Equal("test-namespace", meta.Namespace)
	s.Equal("test", meta.Labels["app"])
}

func (s *KubernetesSuite) TestApimachineryTypeMeta() {
	typeMeta := metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "ConfigMap",
	}

	s.Equal("v1", typeMeta.APIVersion)
	s.Equal("ConfigMap", typeMeta.Kind)
}

func TestKubernetesSuite(t *testing.T) {
	suite.Run(t, new(KubernetesSuite))
}

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
