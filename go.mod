module github.com/arangodb/kube-arangodb

go 1.16

replace (
	github.com/arangodb/go-driver => github.com/arangodb/go-driver v1.2.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring => github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.46.0
	github.com/prometheus-operator/prometheus-operator/pkg/client => github.com/prometheus-operator/prometheus-operator/pkg/client v0.46.0
	github.com/stretchr/testify => github.com/stretchr/testify v1.5.1
	github.com/ugorji/go => github.com/ugorji/go v0.0.0-20181209151446-772ced7fd4c2

	k8s.io/api => k8s.io/api v0.19.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.8
	k8s.io/apiserver => k8s.io/apiserver v0.19.8
	k8s.io/client-go => k8s.io/client-go v0.19.8
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.19.8
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.19.8
	k8s.io/code-generator => ./deps/k8s.io/code-generator
	k8s.io/component-base => k8s.io/component-base v0.19.8
	k8s.io/kubernetes => k8s.io/kubernetes v0.19.8
	k8s.io/metrics => k8s.io/metrics v0.19.8
	k8s.io/kubectl => k8s.io/kubectl v0.19.8
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.8
)

require (
	github.com/arangodb-helper/go-certificates v0.0.0-20180821055445-9fca24fc2680
	github.com/arangodb/arangosync-client v0.7.0
	github.com/arangodb/go-driver v1.2.1
	github.com/arangodb/go-driver/v2 v2.0.0-20211021031401-d92dcd5a4c83
	github.com/arangodb/go-upgrade-rules v0.0.0-20180809110947-031b4774ff21
	//github.com/arangodb/rebalancer v0.1.1
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/gin-gonic/gin v1.7.2
	github.com/go-playground/validator/v10 v10.8.0 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/jessevdk/go-assets v0.0.0-20160921144138-4f4301a06e15
	github.com/julienschmidt/httprouter v1.3.0
	github.com/magiconair/properties v1.8.5
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/onsi/gomega v1.10.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.44.1
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model v0.2.0
	github.com/robfig/cron v1.2.0
	github.com/rs/zerolog v1.19.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/ugorji/go/codec v1.2.6 // indirect
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	google.golang.org/protobuf v1.27.1 // indirect
	k8s.io/api v0.22.4
	k8s.io/apiextensions-apiserver v0.18.3
	k8s.io/apimachinery v0.22.4
	k8s.io/cli-runtime v0.22.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.22.4
)
