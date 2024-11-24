load('ext://cert_manager', 'deploy_cert_manager')
deploy_cert_manager()

docker_build("libvirt-tls-sidecar", ".")
k8s_yaml(['testdata/clusterissuer.yaml', 'testdata/certificates.yaml', 'testdata/issuers.yaml', 'testdata/daemonset.yaml'])
