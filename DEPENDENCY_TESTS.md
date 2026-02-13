# Dependency Test Coverage Matrix

## Overview
This document maps each key dependency to the tests that validate it, ensuring that dependency updates (via Renovate) can be safely validated.

## Key Dependencies and Test Coverage

### 1. cert-manager (v1.19.2)
**Tests validating this dependency:**
- `TestDependencyCertManagerImport` - Validates cert-manager can be imported
- `TestDependencyCertManagerUsages` - Tests KeyUsage types
- `TestDependencyCertManagerIssuerRef` - Validates IssuerReference structure
- `TestDependencyCertManagerCertificateConditions` - Tests certificate conditions
- `TestDependencyCertManagerCertificateSpec` - Validates full CertificateSpec
- `TestDependencyIntegrationFullCertificate` - Tests complete certificate creation
- `TestIntegrationCertManagerTypes` - Validates cert-manager types compatibility
- `TestIntegrationCertManagerAPIVersion` - Ensures correct API version usage
- `TestIntegrationCertificateUsages` - Validates certificate usage types
- `TestIntegrationIssuerRef` - Tests IssuerRef for both Issuer and ClusterIssuer
- All template tests validate Certificate generation

**What's tested:**
- ✓ Certificate type and structure
- ✓ IssuerReference handling (cmmeta.IssuerReference)
- ✓ KeyUsage types (ClientAuth, ServerAuth, DigitalSignature, KeyEncipherment)
- ✓ Certificate conditions and status (ConditionStatus, CertificateCondition)
- ✓ API version compatibility (cert-manager.io/v1)
- ✓ Full certificate spec creation with all fields

### 2. pod-tls-sidecar (v1.0.0)
**Tests validating this dependency:**
- `TestDependencyPodTLSSidecarTemplate` - Validates template package
- `TestDependencyPodTLSSidecarTemplateExecution` - Tests template execution
- `TestDependencyPodTLSSidecarPodInfo` - Validates podinfo package
- `TestDependencyTemplateWithMultipleValues` - Tests various input scenarios
- `TestIntegrationPodTLSSidecarTemplate` - Tests template package integration
- All template tests use pod-tls-sidecar template package

**What's tested:**
- ✓ Template package integration
- ✓ Template creation with `template.New()`
- ✓ Template execution with `template.Execute()`
- ✓ PodInfo structure (Name, Namespace, IP)
- ✓ Template.Values with various inputs
- ✓ Complex namespace and pod names
- ✓ IPv6 address handling

### 3. envconfig (v1.4.0)
**Tests validating this dependency:**
- `TestDependencyEnvconfig` - Validates envconfig package functionality
- `TestDependencyEnvconfigRequired` - Tests required field handling
- `TestIssuerInfoEnvconfig` - Tests API and VNC issuer configuration
- `TestIssuerInfoStructure` - Validates IssuerInfo struct
- `TestIssuerInfoWithDifferentTypes` - Tests various issuer configurations
- `TestIssuerInfoEnvconfigTags` - Verifies envconfig tag processing
- `TestIssuerInfoEmptyValues` - Tests validation of empty values
- `TestIssuerInfoCaseSensitivity` - Tests case sensitivity
- `BenchmarkEnvconfigProcess` - Performance benchmark

**What's tested:**
- ✓ Environment variable parsing with `envconfig.Process()`
- ✓ Required field validation
- ✓ Struct tag processing (`envconfig:"NAME" required:"true"`)
- ✓ Error handling for missing values
- ✓ Case sensitivity handling
- ✓ Default values
- ✓ Performance benchmarking

### 4. k8s.io/client-go (v0.35.0)
**Tests validating this dependency:**
- `TestDependencyK8sClientGoRest` - Validates k8s.io/client-go/rest package
- All integration tests use REST config structures

**What's tested:**
- ✓ REST config structures (`rest.Config`)
- ✓ Kubernetes API client compatibility
- ✓ Config host and authentication fields

### 5. k8s.io/apimachinery (v0.35.0)
**Tests validating this dependency:**
- `TestDependencyK8sApimachineryMetav1` - Tests ObjectMeta
- `TestDependencyK8sApimachineryTypeMeta` - Tests TypeMeta structure
- `TestIntegrationMetadataFields` - Validates k8s metadata integration
- All template tests use ObjectMeta and TypeMeta

**What's tested:**
- ✓ ObjectMeta structure and fields (Name, Namespace, Labels)
- ✓ TypeMeta structure (APIVersion, Kind)
- ✓ Metadata labels and annotations
- ✓ API version and kind handling

### 6. libvirt.org/go/libvirt (v1.11010.0)
**Build compatibility:**
- Main package imports libvirt for runtime functionality
- Tests use `//go:build cgo` tags for CGO-dependent tests
- Tests can run without libvirt installed
- CI workflow installs libvirt-dev for full compatibility

**What's tested:**
- ✓ Build compatibility with CGO
- ✓ Tests work without runtime libvirt dependency
- ✓ Conditional compilation with build tags

## Test Statistics
- **Total Test Functions**: 47
- **Total Test Cases**: 82 (including subtests)
- **Template Package Coverage**: 100% of statements
- **Benchmark Tests**: 4
- **Dependencies Validated**: All 6 key dependencies

## Running Tests by Dependency

### Test cert-manager integration
```bash
go test ./pkg/template/... ./test/... -v -run TestDependencyCertManager
go test ./pkg/template/... -v -run TestIntegrationCertManager
```

### Test pod-tls-sidecar integration
```bash
go test ./pkg/template/... ./test/... -v -run TestDependencyPodTLSSidecar
go test ./pkg/template/... -v -run TestIntegrationPodTLSSidecar
```

### Test envconfig integration
```bash
go test ./test/... -v -run TestDependencyEnvconfig
go test . -v -run TestIssuerInfo -tags cgo
```

### Test k8s.io integration
```bash
go test ./test/... -v -run TestDependencyK8s
go test ./pkg/template/... -v -run TestIntegrationMetadata
```

### Test all integration tests
```bash
go test ./pkg/template/... -v -run TestIntegration
```

### Run all tests with coverage
```bash
go test ./pkg/template/... ./test/... -cover
```

### Run benchmarks
```bash
go test ./pkg/template/... ./test/... -bench=. -benchmem
```

## Benefits for Dependency Updates

These tests ensure that when Renovate (or manual updates) updates dependencies:

1. **Breaking Changes are Detected**: If a dependency introduces breaking changes in their API, tests will fail
2. **Type Compatibility**: Tests validate that types from different packages work together correctly
3. **API Changes**: Tests catch when method signatures or return types change
4. **Performance Regressions**: Benchmark tests help identify performance degradation
5. **Edge Cases**: Tests cover IPv6, special characters, various configurations
6. **CI Integration**: Tests run automatically in GitHub Actions

## CI Workflow Integration

The existing `.github/workflows/ci.yaml` workflow:
1. Installs `libvirt-dev` for full test compatibility
2. Runs all tests including those requiring CGO
3. Uses `robherley/go-test-action` for test reporting
4. Runs on every pull request, including Renovate PRs

This ensures that dependency updates are automatically validated before merging.
