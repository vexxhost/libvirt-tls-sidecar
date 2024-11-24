// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os/exec"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"github.com/vexxhost/pod-tls-sidecar/pkg/tls"
	"k8s.io/client-go/rest"
	"libvirt.org/go/libvirt"

	"github.com/vexxhost/libvirt-tls-sidecar/pkg/template"
)

type IssuerInfo struct {
	Kind string `envconfig:"ISSUER_KIND" required:"true"`
	Name string `envconfig:"ISSUER_NAME" required:"true"`
}

func main() {
	var apiIssuer template.IssuerInfo
	if err := envconfig.Process("API", &apiIssuer); err != nil {
		log.Fatal(err)
	}

	var vncIssuer template.IssuerInfo
	if err := envconfig.Process("VNC", &vncIssuer); err != nil {
		log.Fatal(err)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	go createCertificateSpec(ctx, "api", &apiIssuer, tls.WithRestConfig(config), tls.WithPaths(&tls.WritePathConfig{
		CertificateAuthorityPaths: []string{"/etc/pki/CA/cacert.pem", "/etc/pki/qemu/ca-cert.pem"},
		CertificatePaths:          []string{"/etc/pki/libvirt/servercert.pem", "/etc/pki/libvirt/clientcert.pem", "/etc/pki/qemu/server-cert.pem", "/etc/pki/qemu/client-cert.pem"},
		CertificateKeyPaths:       []string{"/etc/pki/libvirt/private/serverkey.pem", "/etc/pki/libvirt/private/clientkey.pem", "/etc/pki/qemu/server-key.pem", "/etc/pki/qemu/client-key.pem"},
	}), tls.WithOnUpdate(func() {
		cmd := exec.Command("virt-admin", "server-update-tls", "libvirtd")
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.WithError(err).WithField("output", string(out)).Error("failed to reload tls configuration for api")
			return
		}

		log.WithField("output", string(out)).Info("reloaded tls configuration for api")
	}))
	go createCertificateSpec(ctx, "vnc", &vncIssuer, tls.WithRestConfig(config), tls.WithPaths(&tls.WritePathConfig{
		CertificateAuthorityPaths: []string{"/etc/pki/libvirt-vnc/ca-cert.pem"},
		CertificatePaths:          []string{"/etc/pki/libvirt-vnc/server-cert.pem"},
		CertificateKeyPaths:       []string{"/etc/pki/libvirt-vnc/server-key.pem"},
	}), tls.WithOnUpdate(func() {
		conn, err := libvirt.NewConnect("qemu:///system")
		if err != nil {
			log.WithError(err).Error("failed to connect to libvirt")
			return
		}
		defer conn.Close()

		// list all domains
		domains, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
		if err != nil {
			log.WithError(err).Error("failed to list domains")
			return
		}

		for _, domain := range domains {
			response, err := domain.QemuMonitorCommand(`{"execute": "display-reload", "arguments":{"type": "vnc", "tls-certs": true}}`, libvirt.DOMAIN_QEMU_MONITOR_COMMAND_DEFAULT)
			if err != nil {
				log.WithError(err).Error("failed to reload tls configuration for vnc")
				continue
			}

			log.WithField("response", response).Info("reloaded tls configuration for vnc")
		}
	}))

	<-ctx.Done()
}

func createCertificateSpec(ctx context.Context, name string, issuer *template.IssuerInfo, opts ...tls.Option) {
	tmpl, err := template.New(name, issuer)
	if err != nil {
		log.Fatal(err)
	}

	args := []tls.Option{
		tls.WithTemplate(tmpl),
	}
	args = append(args, opts...)

	config, err := tls.NewConfig(args...)
	if err != nil {
		log.Fatal(err)
	}

	manager, err := tls.NewManager(config)
	if err != nil {
		log.Fatal(err)
	}

	err = manager.Create(ctx)
	if err != nil {
		log.Fatal(err)
	}

	manager.Watch(ctx)
}
