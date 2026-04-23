// Copyright (c) 2026 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"os/exec"

	log "github.com/sirupsen/logrus"
	"github.com/vexxhost/pod-tls-sidecar/pkg/tls"
	"libvirt.org/go/libvirt"

	"github.com/vexxhost/libvirt-tls-sidecar/pkg/template"
)

// CreateCertificateSpec creates and watches a certificate specification
func CreateCertificateSpec(ctx context.Context, name string, issuer *template.IssuerInfo, opts ...tls.Option) {
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

// APIUpdateCallback returns the callback for API certificate updates
func APIUpdateCallback() func() {
	return func() {
		cmd := exec.Command("virt-admin", "server-update-tls", "libvirtd")
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.WithError(err).WithField("output", string(out)).Error("failed to reload tls configuration for api")
			return
		}

		log.WithField("output", string(out)).Info("reloaded tls configuration for api")
	}
}

// VNCUpdateCallback returns the callback for VNC certificate updates
func VNCUpdateCallback() func() {
	return func() {
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
	}
}

// APIPaths returns the certificate paths for API
func APIPaths() *tls.WritePathConfig {
	return &tls.WritePathConfig{
		CertificateAuthorityPaths: []string{"/etc/pki/CA/cacert.pem", "/etc/pki/qemu/ca-cert.pem"},
		CertificatePaths:          []string{"/etc/pki/libvirt/servercert.pem", "/etc/pki/libvirt/clientcert.pem", "/etc/pki/qemu/server-cert.pem", "/etc/pki/qemu/client-cert.pem"},
		CertificateKeyPaths:       []string{"/etc/pki/libvirt/private/serverkey.pem", "/etc/pki/libvirt/private/clientkey.pem", "/etc/pki/qemu/server-key.pem", "/etc/pki/qemu/client-key.pem"},
	}
}

// VNCPaths returns the certificate paths for VNC
func VNCPaths() *tls.WritePathConfig {
	return &tls.WritePathConfig{
		CertificateAuthorityPaths: []string{"/etc/pki/libvirt-vnc/ca-cert.pem"},
		CertificatePaths:          []string{"/etc/pki/libvirt-vnc/server-cert.pem"},
		CertificateKeyPaths:       []string{"/etc/pki/libvirt-vnc/server-key.pem"},
	}
}
