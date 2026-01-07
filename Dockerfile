# Copyright (c) 2024 VEXXHOST, Inc.
# SPDX-License-Identifier: Apache-2.0

ARG RELEASE=bookworm

FROM golang:1.25.5 AS builder
WORKDIR /go/src/app
RUN \
  apt-get update && \
  apt-get install -qq -y --no-install-recommends \
    libvirt-dev && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/*
COPY pkg /go/src/app/pkg
COPY go.mod go.sum main.go /go/src/app/
RUN go build -o /libvirt-tls-sidecar main.go

FROM debian:bookworm
RUN \
  apt-get update -qq && \
  apt-get install -qq -y --no-install-recommends \
    libvirt0 libvirt-clients && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/*
COPY --from=builder /libvirt-tls-sidecar /usr/bin/libvirt-tls-sidecar
ENTRYPOINT ["/usr/bin/libvirt-tls-sidecar"]
