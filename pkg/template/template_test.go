package template

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/vexxhost/pod-tls-sidecar/pkg/podinfo"
	"github.com/vexxhost/pod-tls-sidecar/pkg/template"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
)

type TemplateSuite struct {
	suite.Suite
}

func (s *TemplateSuite) TestNew() {
	tmpl, err := New("api", &IssuerInfo{
		Kind: "IssuerKind",
		Name: "IssuerName",
	})

	s.NoError(err)
	s.NotNil(tmpl)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "test-pod",
			Namespace: "test-namespace",
			IP:        "1.2.3.4",
		},
		Hostname: "test-hostname",
		FQDN:     "test-hostname.atmosphere.dev",
	})
	s.Require().NoError(err)

	s.Equal("test-pod-api", certificate.Name)
	s.Equal("test-namespace", certificate.Namespace)

	s.Equal("test-hostname.atmosphere.dev", certificate.Spec.CommonName)

	s.Len(certificate.Spec.DNSNames, 2)
	s.Contains(certificate.Spec.DNSNames, "test-hostname")
	s.Contains(certificate.Spec.DNSNames, "test-hostname.atmosphere.dev")

	s.Len(certificate.Spec.IPAddresses, 1)
	s.Contains(certificate.Spec.IPAddresses, "1.2.3.4")

	s.Len(certificate.Spec.Usages, 2)
	s.Contains(certificate.Spec.Usages, cmv1.UsageClientAuth)
	s.Contains(certificate.Spec.Usages, cmv1.UsageServerAuth)

	s.Equal("IssuerKind", certificate.Spec.IssuerRef.Kind)
	s.Equal("IssuerName", certificate.Spec.IssuerRef.Name)

	s.Equal("test-pod-api", certificate.Spec.SecretName)
}

func (s *TemplateSuite) TestNewWithVNCName() {
	tmpl, err := New("vnc", &IssuerInfo{
		Kind: "ClusterIssuer",
		Name: "vnc-issuer",
	})

	s.NoError(err)
	s.NotNil(tmpl)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "compute-node",
			Namespace: "openstack",
			IP:        "192.168.1.100",
		},
		Hostname: "compute-01",
		FQDN:     "compute-01.openstack.svc.cluster.local",
	})
	s.Require().NoError(err)

	s.Equal("compute-node-vnc", certificate.Name)
	s.Equal("openstack", certificate.Namespace)
	s.Equal("compute-01.openstack.svc.cluster.local", certificate.Spec.CommonName)
	s.Equal("ClusterIssuer", certificate.Spec.IssuerRef.Kind)
	s.Equal("vnc-issuer", certificate.Spec.IssuerRef.Name)
	s.Equal("compute-node-vnc", certificate.Spec.SecretName)
}

func (s *TemplateSuite) TestNewWithDifferentIssuerTypes() {
	tests := []struct {
		name       string
		certName   string
		issuerKind string
		issuerName string
	}{
		{"Issuer type", "api", "Issuer", "ca-issuer"},
		{"ClusterIssuer type", "vnc", "ClusterIssuer", "cluster-ca"},
		{"Custom name with dashes", "custom-cert", "Issuer", "my-custom-issuer"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tmpl, err := New(tt.certName, &IssuerInfo{
				Kind: tt.issuerKind,
				Name: tt.issuerName,
			})

			s.Require().NoError(err)
			s.Require().NotNil(tmpl)

			certificate, err := tmpl.Execute(&template.Values{
				PodInfo: podinfo.PodInfo{
					Name:      "test-pod",
					Namespace: "test-ns",
					IP:        "10.0.0.1",
				},
				Hostname: "test",
				FQDN:     "test.example.com",
			})
			s.Require().NoError(err)

			s.Equal(tt.issuerKind, certificate.Spec.IssuerRef.Kind)
			s.Equal(tt.issuerName, certificate.Spec.IssuerRef.Name)
			s.Equal("test-pod-"+tt.certName, certificate.Name)
		})
	}
}

func (s *TemplateSuite) TestNewWithIPv6Address() {
	tmpl, err := New("api", &IssuerInfo{
		Kind: "Issuer",
		Name: "test-issuer",
	})
	s.Require().NoError(err)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "ipv6-pod",
			Namespace: "default",
			IP:        "2001:db8::1",
		},
		Hostname: "ipv6-host",
		FQDN:     "ipv6-host.example.com",
	})
	s.Require().NoError(err)

	s.Contains(certificate.Spec.IPAddresses, "2001:db8::1")
}

func (s *TemplateSuite) TestNewWithSpecialCharactersInNames() {
	tmpl, err := New("api-v2", &IssuerInfo{
		Kind: "Issuer",
		Name: "special-issuer-name",
	})
	s.Require().NoError(err)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "pod-with-special-chars",
			Namespace: "my-namespace",
			IP:        "172.16.0.1",
		},
		Hostname: "host-123",
		FQDN:     "host-123.my-domain.example.org",
	})
	s.Require().NoError(err)

	s.Equal("pod-with-special-chars-api-v2", certificate.Name)
	s.Equal("my-namespace", certificate.Namespace)
}

func (s *TemplateSuite) TestNewValidatesCertificateUsages() {
	tmpl, err := New("api", &IssuerInfo{
		Kind: "Issuer",
		Name: "test-issuer",
	})
	s.Require().NoError(err)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "test-pod",
			Namespace: "test-ns",
			IP:        "10.0.0.1",
		},
		Hostname: "test",
		FQDN:     "test.example.com",
	})
	s.Require().NoError(err)

	s.Len(certificate.Spec.Usages, 2)
	s.Contains(certificate.Spec.Usages, cmv1.UsageClientAuth)
	s.Contains(certificate.Spec.Usages, cmv1.UsageServerAuth)
}

func (s *TemplateSuite) TestNewWithMultipleDNSNames() {
	tmpl, err := New("api", &IssuerInfo{
		Kind: "Issuer",
		Name: "test-issuer",
	})
	s.Require().NoError(err)

	certificate, err := tmpl.Execute(&template.Values{
		PodInfo: podinfo.PodInfo{
			Name:      "test-pod",
			Namespace: "test-ns",
			IP:        "10.0.0.1",
		},
		Hostname: "short-name",
		FQDN:     "short-name.long.domain.example.com",
	})
	s.Require().NoError(err)

	s.Len(certificate.Spec.DNSNames, 2)
	s.Contains(certificate.Spec.DNSNames, "short-name")
	s.Contains(certificate.Spec.DNSNames, "short-name.long.domain.example.com")
	s.Equal("short-name.long.domain.example.com", certificate.Spec.CommonName)
}

func (s *TemplateSuite) TestNewGeneratesCorrectSecretName() {
	tests := []struct {
		name           string
		certName       string
		podName        string
		expectedSecret string
	}{
		{"API certificate", "api", "libvirt-node", "libvirt-node-api"},
		{"VNC certificate", "vnc", "compute-host", "compute-host-vnc"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tmpl, err := New(tt.certName, &IssuerInfo{
				Kind: "Issuer",
				Name: "test-issuer",
			})
			s.Require().NoError(err)

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

			s.Equal(tt.expectedSecret, certificate.Spec.SecretName)
			s.Equal(tt.expectedSecret, certificate.Name)
		})
	}
}

func TestTemplateSuite(t *testing.T) {
	suite.Run(t, new(TemplateSuite))
}

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
