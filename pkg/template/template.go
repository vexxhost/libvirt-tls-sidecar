package template

import (
	"fmt"

	"github.com/vexxhost/pod-tls-sidecar/pkg/template"
)

type IssuerInfo struct {
	Kind string `envconfig:"ISSUER_KIND" required:"true"`
	Name string `envconfig:"ISSUER_NAME" required:"true"`
}

func New(name string, issuer *IssuerInfo) (*template.Template, error) {
	return template.New(fmt.Sprintf(`
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{ .PodInfo.Name }}-%s
  namespace: {{ .PodInfo.Namespace }}
spec:
  commonName: "{{ .FQDN }}"
  dnsNames:
    - "{{ .Hostname }}"
    - "{{ .FQDN }}"
  ipAddresses:
    - "{{ .PodInfo.IP }}"
  usages:
    - client auth
    - server auth
  issuerRef:
    kind: %s
    name: %s
  secretName: {{ .PodInfo.Name }}-%s`,
		name, issuer.Kind, issuer.Name, name))
}
