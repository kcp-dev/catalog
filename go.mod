module github.com/kcp-dev/catalog

go 1.19

require (
	github.com/kcp-dev/kcp v0.9.1
	github.com/kcp-dev/kcp/pkg/apis v0.9.1
	github.com/kcp-dev/logicalcluster/v2 v2.0.0-alpha.3
	github.com/onsi/ginkgo/v2 v2.1.4
	github.com/onsi/gomega v1.19.0
	k8s.io/api v0.25.0
	k8s.io/apimachinery v0.25.0
	k8s.io/client-go v0.25.0
	k8s.io/klog/v2 v2.70.1
	sigs.k8s.io/controller-runtime v0.13.0
)

require (
	cloud.google.com/go v0.97.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.27 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.20 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.8.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kcp-dev/apimachinery v0.0.0-20220912132244-efe716c18e43 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.12.2 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/spf13/pflag v1.0.6-0.20210604193023-d5e0c0615ace // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/crypto v0.0.0-20220315160706-3147a52a75dd // indirect
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220609170525-579cf78fd858 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.25.0 // indirect
	k8s.io/apiserver v0.24.3 // indirect
	k8s.io/component-base v0.25.0 // indirect
	k8s.io/kube-openapi v0.0.0-20220803162953-67bda5d908f1 // indirect
	k8s.io/utils v0.0.0-20220728103510-ee6ede2d64ed // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	k8s.io/api => github.com/kcp-dev/kubernetes/staging/src/k8s.io/api v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/apiextensions-apiserver => github.com/kcp-dev/kubernetes/staging/src/k8s.io/apiextensions-apiserver v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/apimachinery => github.com/kcp-dev/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/apiserver => github.com/kcp-dev/kubernetes/staging/src/k8s.io/apiserver v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/cli-runtime => github.com/kcp-dev/kubernetes/staging/src/k8s.io/cli-runtime v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/client-go => github.com/kcp-dev/kubernetes/staging/src/k8s.io/client-go v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/cloud-provider => github.com/kcp-dev/kubernetes/staging/src/k8s.io/cloud-provider v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/cluster-bootstrap => github.com/kcp-dev/kubernetes/staging/src/k8s.io/cluster-bootstrap v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/code-generator => github.com/kcp-dev/kubernetes/staging/src/k8s.io/code-generator v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/component-base => github.com/kcp-dev/kubernetes/staging/src/k8s.io/component-base v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/component-helpers => github.com/kcp-dev/kubernetes/staging/src/k8s.io/component-helpers v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/controller-manager => github.com/kcp-dev/kubernetes/staging/src/k8s.io/controller-manager v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/cri-api => github.com/kcp-dev/kubernetes/staging/src/k8s.io/cri-api v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/csi-translation-lib => github.com/kcp-dev/kubernetes/staging/src/k8s.io/csi-translation-lib v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/kube-aggregator => github.com/kcp-dev/kubernetes/staging/src/k8s.io/kube-aggregator v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/kube-controller-manager => github.com/kcp-dev/kubernetes/staging/src/k8s.io/kube-controller-manager v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/kube-proxy => github.com/kcp-dev/kubernetes/staging/src/k8s.io/kube-proxy v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/kube-scheduler => github.com/kcp-dev/kubernetes/staging/src/k8s.io/kube-scheduler v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/kubectl => github.com/kcp-dev/kubernetes/staging/src/k8s.io/kubectl v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/kubelet => github.com/kcp-dev/kubernetes/staging/src/k8s.io/kubelet v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/kubernetes => github.com/kcp-dev/kubernetes v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/legacy-cloud-providers => github.com/kcp-dev/kubernetes/staging/src/k8s.io/legacy-cloud-providers v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/metrics => github.com/kcp-dev/kubernetes/staging/src/k8s.io/metrics v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/mount-utils => github.com/kcp-dev/kubernetes/staging/src/k8s.io/mount-utils v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/pod-security-admission => github.com/kcp-dev/kubernetes/staging/src/k8s.io/pod-security-admission v0.0.0-20220915135949-eeba459ad2a1
	k8s.io/sample-apiserver => github.com/kcp-dev/kubernetes/staging/src/k8s.io/sample-apiserver v0.0.0-20220915135949-eeba459ad2a1
	sigs.k8s.io/controller-runtime => github.com/kcp-dev/controller-runtime v0.12.2-0.20220808200255-4b60fd66e5de
)
