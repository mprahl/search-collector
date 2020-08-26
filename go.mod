// Copyright (c) 2020 Red Hat, Inc.

module github.com/open-cluster-management/search-collector

go 1.13

require (
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e
	github.com/golang/protobuf v1.4.2
	github.com/kennygrant/sanitize v1.2.4
	github.com/open-cluster-management/governance-policy-propagator v0.0.0-20200602150427-d0f4af8aba9d
	github.com/open-cluster-management/multicloud-operators-channel v1.0.0
	github.com/open-cluster-management/multicloud-operators-deployable v1.0.0
	github.com/open-cluster-management/multicloud-operators-foundation v1.0.0
	github.com/open-cluster-management/multicloud-operators-placementrule v1.0.0
	github.com/open-cluster-management/multicloud-operators-subscription v1.0.0
	github.com/open-cluster-management/multicloud-operators-subscription-release v1.0.1-0.20200603160156-4d66bd136ba3 //Use 2.0 when available
	github.com/tkanos/gonfig v0.0.0-20181112185242-896f3d81fadf
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v13.0.0+incompatible
	k8s.io/helm v2.16.7+incompatible
	sigs.k8s.io/application v0.8.3
)

replace (
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible
	github.com/docker/docker => github.com/docker/docker v1.13.1
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.7.4
	github.com/influxdata/influxdb => github.com/influxdata/influxdb v1.8.2
	github.com/mholt/caddy => github.com/mholt/caddy v0.11.5
	github.com/openshift/origin => github.com/openshift/origin v1.2.0
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v2.7.1+incompatible
	k8s.io/api => k8s.io/api v0.17.11
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.11
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.11
	k8s.io/apiserver => k8s.io/apiserver v0.17.11
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.11
	k8s.io/client-go => k8s.io/client-go v0.17.11
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.11
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.11
	k8s.io/code-generator => k8s.io/code-generator v0.17.11
	k8s.io/component-base => k8s.io/component-base v0.17.11
	k8s.io/cri-api => k8s.io/cri-api v0.17.11
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.11
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.11
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.11
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.11
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.11
	k8s.io/kubectl => k8s.io/kubectl v0.17.11
	k8s.io/kubelet => k8s.io/kubelet v0.17.11
	k8s.io/kubernetes => k8s.io/kubernetes v1.17.11
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.11
	k8s.io/metrics => k8s.io/metrics v0.17.11
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.11
)