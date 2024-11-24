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
