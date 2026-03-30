// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"github.com/vexxhost/pod-tls-sidecar/pkg/tls"
	"k8s.io/client-go/rest"

	"github.com/vexxhost/libvirt-tls-sidecar/internal/app"
	"github.com/vexxhost/libvirt-tls-sidecar/pkg/template"
)

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

	go app.CreateCertificateSpec(
		ctx,
		"api",
		&apiIssuer,
		tls.WithRestConfig(config),
		tls.WithPaths(app.APIPaths()),
		tls.WithOnUpdate(app.APIUpdateCallback()),
	)

	go app.CreateCertificateSpec(
		ctx,
		"vnc",
		&vncIssuer,
		tls.WithRestConfig(config),
		tls.WithPaths(app.VNCPaths()),
		tls.WithOnUpdate(app.VNCUpdateCallback()),
	)

	<-ctx.Done()
}
