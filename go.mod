module github.com/arangodb/kube-arangodb

go 1.13

replace (
	github.com/coreos/prometheus-operator => github.com/coreos/prometheus-operator v0.37.0
	github.com/stretchr/testify => github.com/stretchr/testify v1.5.1
	github.com/ugorji/go => github.com/ugorji/go v0.0.0-20181209151446-772ced7fd4c2

	k8s.io/api => k8s.io/api v0.15.9
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.15.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.15.9
	k8s.io/client-go => k8s.io/client-go v0.15.9
	k8s.io/code-generator => ./deps/k8s.io/code-generator
)

require (
	github.com/aktau/github-release v0.7.2 // indirect
	github.com/arangodb-helper/go-certificates v0.0.0-20180821055445-9fca24fc2680
	github.com/arangodb/arangosync-client v0.6.3
	github.com/arangodb/go-driver v0.0.0-20191002124627-11b6bfc64f67
	github.com/arangodb/go-upgrade-rules v0.0.0-20180809110947-031b4774ff21
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/coreos/go-semver v0.3.0
	github.com/coreos/prometheus-operator v0.37.0
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/gin-contrib/sse v0.0.0-20190301062529-5545eab6dad3 // indirect
	github.com/gin-gonic/gin v1.3.0
	github.com/jessevdk/go-assets v0.0.0-20160921144138-4f4301a06e15
	github.com/jessevdk/go-assets-builder v0.0.0-20130903091706-b8483521738f
	github.com/julienschmidt/httprouter v1.3.0
	github.com/magiconair/properties v1.8.0
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.2.1
	github.com/robfig/cron v1.2.0
	github.com/rs/zerolog v1.14.3
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80 // indirect
	github.com/ugorji/go/codec v0.0.0-20181209151446-772ced7fd4c2 // indirect
	github.com/voxelbrain/goptions v0.0.0-20180630082107-58cddc247ea2 // indirect
	golang.org/x/sys v0.0.0-20200116001909-b77594299b42
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
)
