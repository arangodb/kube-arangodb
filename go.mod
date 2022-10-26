module github.com/arangodb/kube-arangodb

go 1.17

replace (
	github.com/arangodb/go-driver => github.com/arangodb/go-driver v1.2.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring => github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.46.0
	github.com/prometheus-operator/prometheus-operator/pkg/client => github.com/prometheus-operator/prometheus-operator/pkg/client v0.46.0
	github.com/stretchr/testify => github.com/stretchr/testify v1.8.1
	github.com/ugorji/go => github.com/ugorji/go v0.0.0-20181209151446-772ced7fd4c2

	k8s.io/api => k8s.io/api v0.21.14
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.14
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.14
	k8s.io/apiserver => k8s.io/apiserver v0.21.14
	k8s.io/client-go => k8s.io/client-go v0.21.14
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.14
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.14
	k8s.io/code-generator => ./deps/k8s.io/code-generator
	k8s.io/component-base => k8s.io/component-base v0.21.14
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20211110012726-3cc51fd1e909
	k8s.io/kubernetes => k8s.io/kubernetes v0.21.14
	k8s.io/metrics => k8s.io/metrics v0.21.14
)

require (
	github.com/arangodb-helper/go-certificates v0.0.0-20180821055445-9fca24fc2680
	github.com/arangodb/arangosync-client v0.8.0
	github.com/arangodb/go-driver v1.4.0
	github.com/arangodb/go-driver/v2 v2.0.0-20221018153419-f67492fedbba
	github.com/arangodb/go-upgrade-rules v0.0.0-20200605091205-439fb1ee86e7
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/dchest/uniuri v1.2.0
	github.com/gin-gonic/gin v1.8.1
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/jessevdk/go-assets v0.0.0-20160921144138-4f4301a06e15
	github.com/josephburnett/jd v1.6.1
	github.com/julienschmidt/httprouter v1.3.0
	github.com/magiconair/properties v1.8.5
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.60.1
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.60.1
	github.com/prometheus/client_golang v1.13.0
	github.com/prometheus/client_model v0.3.0
	github.com/robfig/cron v1.2.0
	github.com/rs/zerolog v1.28.0
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.1
	golang.org/x/sync v0.1.0
	golang.org/x/sys v0.1.0
	golang.org/x/time v0.1.0
	google.golang.org/grpc v1.50.1
	google.golang.org/protobuf v1.28.1
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.25.3
	k8s.io/apiextensions-apiserver v0.25.3
	k8s.io/apimachinery v0.25.3
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/arangodb/go-velocypack v0.0.0-20200318135517-5af53c29c67e // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.11.1 // indirect
	github.com/goccy/go-json v0.9.11 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pavel-v-chernykh/keystore-go v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/net v0.1.0 // indirect
	golang.org/x/oauth2 v0.1.0 // indirect
	golang.org/x/term v0.1.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20221025140454-527a21cfbd71 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/kube-openapi v0.0.0-20221012153701-172d655c2280 // indirect
	k8s.io/utils v0.0.0-20221012122500-cfd413dd9e85 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)
