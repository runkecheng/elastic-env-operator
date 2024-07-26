module github.com/wosai/elastic-env-operator

go 1.21

require (
	github.com/VictoriaMetrics/operator/api v0.0.0-20220220173620-6f1e087b0e5c
	github.com/go-logr/logr v1.2.3
	github.com/go-logr/zapr v1.2.0
	github.com/gogo/protobuf v1.3.2
	github.com/imdario/mergo v0.3.13
	github.com/onsi/ginkgo/v2 v2.1.4
	github.com/onsi/gomega v1.19.0
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.58.0
	github.com/stretchr/testify v1.7.0
	github.com/wosai/elastic-env-operator/api v0.5.5
	go.uber.org/zap v1.19.1
	gotest.tools v2.2.0+incompatible
	istio.io/api v0.0.0-20200812202721-24be265d41c3
	istio.io/client-go v0.0.0-20200814134724-bcbf0ed82b30
	k8s.io/api v0.24.4
	k8s.io/apiextensions-apiserver v0.24.4
	k8s.io/apimachinery v0.24.4
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.12.3
)

require (
	cloud.google.com/go/compute v1.7.0 // indirect
	github.com/VictoriaMetrics/VictoriaMetrics v1.79.1-0.20220722115041-1aa51127716e // indirect
	github.com/VictoriaMetrics/metrics v1.19.3 // indirect
	github.com/VictoriaMetrics/metricsql v0.44.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.8.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.2 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fastrand v1.1.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	github.com/valyala/histogram v1.2.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/net v0.0.0-20220708220712-1185a9018129 // indirect
	golang.org/x/oauth2 v0.0.0-20220718184931-c8730f7fcb92 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	golang.org/x/term v0.0.0-20220526004731-065cf7ba2467 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220609170525-579cf78fd858 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	istio.io/gogo-genproto v0.0.0-20190930162913-45029607206a // indirect
	k8s.io/component-base v0.24.4 // indirect
	k8s.io/klog/v2 v2.70.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220627174259-011e075b9cb8 // indirect
	k8s.io/utils v0.0.0-20220713171938-56c0de1e6f5e // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	github.com/google/gnostic => github.com/googleapis/gnostic v0.6.9
	github.com/googleapis/gnostic => github.com/google/gnostic v0.6.9
	github.com/wosai/elastic-env-operator/api => ./api
	k8s.io/api => k8s.io/api v0.24.4
	k8s.io/client-go => k8s.io/client-go v0.24.4
)
