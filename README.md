`libvirt-tls-sidecar`

This is a tool which is built on top of the `pod-tls-sidecar` framework however
it has extra tooling which will make sure to reload the TLS certificates in the
libvirt daemon when they are updated.

## Development

This project uses Tilt for development. To start the development environment run:

```bash
tilt up
```

In order to trigger a manual renewal, you can try something like this:

```bash
cmctl renew libvirt-api
cmctl renew libvirt-vnc
```

You can then view the logs after:

```bash
kubectl -n default logs ds/libvirt-libvirt-default -c tls-sidecar
```
