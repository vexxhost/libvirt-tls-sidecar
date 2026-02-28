// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/vexxhost/pod-tls-sidecar/pkg/podinfo"
	"github.com/vexxhost/pod-tls-sidecar/pkg/template"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IntegrationSuite struct {
	suite.Suite
}

func (s *IntegrationSuite) newTemplate(name string, issuerKind string, issuerName string) *template.Template {
	tmpl, err := New(name, &IssuerInfo{
		Kind: issuerKind,
		Name: issuerName,
	})
	s.Require().NoError(err)
	return tmpl
}

func (s *IntegrationSuite) defaultValues() *template.Values {
	return &template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "test-pod",
			Namespace: "test-ns",
			IP:        "10.0.0.1",
		},
		Hostname: "test",
		FQDN:     "test.example.com",
	}
}

func (s *IntegrationSuite) TestCertManagerTypes() {
	tmpl := s.newTemplate("api", "ClusterIssuer", "test-issuer")
	certificate, err := tmpl.Execute(s.defaultValues())
	s.Require().NoError(err)

	s.IsType(&cmv1.Certificate{}, certificate)
	s.NotNil(certificate.TypeMeta)
	s.NotNil(certificate.ObjectMeta)
	s.NotNil(certificate.Spec)
}

func (s *IntegrationSuite) TestCertManagerAPIVersion() {
	tmpl := s.newTemplate("api", "Issuer", "test-issuer")
	certificate, err := tmpl.Execute(s.defaultValues())
	s.Require().NoError(err)

	s.Equal("cert-manager.io/v1", certificate.APIVersion)
	s.Equal("Certificate", certificate.Kind)
}

func (s *IntegrationSuite) TestPodTLSSidecarTemplate() {
	tmpl := s.newTemplate("vnc", "ClusterIssuer", "ca-issuer")

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
		s.Run(tt.name, func() {
			certificate, err := tmpl.Execute(tt.values)
			s.Require().NoError(err)
			s.NotNil(certificate)
			s.NotEmpty(certificate.Name)
			s.NotEmpty(certificate.Namespace)
		})
	}
}

func (s *IntegrationSuite) TestCertificateUsages() {
	tmpl := s.newTemplate("api", "Issuer", "test-issuer")
	certificate, err := tmpl.Execute(s.defaultValues())
	s.Require().NoError(err)

	expectedUsages := []cmv1.KeyUsage{
		cmv1.UsageClientAuth,
		cmv1.UsageServerAuth,
	}

	s.ElementsMatch(expectedUsages, certificate.Spec.Usages)

	for _, usage := range certificate.Spec.Usages {
		s.IsType(cmv1.KeyUsage(""), usage)
	}
}

func (s *IntegrationSuite) TestIssuerRef() {
	tests := []struct {
		name       string
		issuerKind string
		issuerName string
	}{
		{"ClusterIssuer", "ClusterIssuer", "ca-issuer"},
		{"Issuer", "Issuer", "namespace-issuer"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tmpl := s.newTemplate("api", tt.issuerKind, tt.issuerName)
			certificate, err := tmpl.Execute(s.defaultValues())
			s.Require().NoError(err)

			s.Equal(tt.issuerKind, certificate.Spec.IssuerRef.Kind)
			s.Equal(tt.issuerName, certificate.Spec.IssuerRef.Name)
		})
	}
}

func (s *IntegrationSuite) TestMetadataFields() {
	tmpl := s.newTemplate("api", "Issuer", "test-issuer")
	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "my-pod",
			Namespace: "my-namespace",
			IP:        "10.0.0.1",
		},
		Hostname: "hostname",
		FQDN:     "hostname.example.com",
	})
	s.Require().NoError(err)

	s.Equal("my-pod-api", certificate.ObjectMeta.Name)
	s.Equal("my-namespace", certificate.ObjectMeta.Namespace)
	s.IsType(metav1.ObjectMeta{}, certificate.ObjectMeta)
}

func (s *IntegrationSuite) TestDNSNamesAndIPAddresses() {
	tmpl := s.newTemplate("api", "Issuer", "test-issuer")
	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "test-pod",
			Namespace: "test-ns",
			IP:        "192.168.1.100",
		},
		Hostname: "myhost",
		FQDN:     "myhost.example.com",
	})
	s.Require().NoError(err)

	s.IsType([]string{}, certificate.Spec.DNSNames)
	s.Len(certificate.Spec.DNSNames, 2)

	s.IsType([]string{}, certificate.Spec.IPAddresses)
	s.Len(certificate.Spec.IPAddresses, 1)
	s.Equal("192.168.1.100", certificate.Spec.IPAddresses[0])
}

func (s *IntegrationSuite) TestCommonName() {
	tmpl := s.newTemplate("api", "Issuer", "test-issuer")

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
	s.Require().NoError(err)

	s.Equal(fqdn, certificate.Spec.CommonName)
	s.IsType("", certificate.Spec.CommonName)
}

func (s *IntegrationSuite) TestSecretName() {
	tests := []struct {
		certName   string
		podName    string
		wantSecret string
	}{
		{"api", "libvirt-node-1", "libvirt-node-1-api"},
		{"vnc", "compute-host", "compute-host-vnc"},
	}

	for _, tt := range tests {
		s.Run(tt.certName, func() {
			tmpl := s.newTemplate(tt.certName, "Issuer", "test-issuer")
			certificate, err := tmpl.Execute(&template.Values{
				PodInfo: podinfo.PodInfo{
					Name:      tt.podName,
					Namespace: "test-ns",
					IP:        "10.0.0.1",
				},
				Hostname: "test",
				FQDN:     "test.example.com",
			})
			s.Require().NoError(err)

			s.Equal(tt.wantSecret, certificate.Spec.SecretName)
		})
	}
}

func (s *IntegrationSuite) TestMultipleTemplateInstances() {
	apiTmpl := s.newTemplate("api", "ClusterIssuer", "api-issuer")
	vncTmpl := s.newTemplate("vnc", "Issuer", "vnc-issuer")

	values := s.defaultValues()

	apiCert, err := apiTmpl.Execute(values)
	s.Require().NoError(err)

	vncCert, err := vncTmpl.Execute(values)
	s.Require().NoError(err)

	s.Equal("test-pod-api", apiCert.Name)
	s.Equal("test-pod-vnc", vncCert.Name)
	s.Equal("ClusterIssuer", apiCert.Spec.IssuerRef.Kind)
	s.Equal("Issuer", vncCert.Spec.IssuerRef.Kind)
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}
