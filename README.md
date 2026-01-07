# `libvirt-tls-sidecar`

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

## Creating a Release

This project uses [semantic versioning](https://semver.org/) with version tags
prefixed with `v` (e.g., `v1.0.0`, `v1.2.3`).

To create a new release using the [GitHub CLI](https://cli.github.com/):

```bash
gh release create vX.Y.Z --generate-notes
```

This will:

- Create a new tag `vX.Y.Z` on the current branch
- Auto-generate release notes from merged pull requests and commits since the
  last release
- Publish the release to GitHub

### Version Numbering

Follow semantic versioning (`MAJOR.MINOR.PATCH`):

- **MAJOR**: Breaking changes or incompatible API changes
- **MINOR**: New features that are backwards compatible
- **PATCH**: Backwards compatible bug fixes
