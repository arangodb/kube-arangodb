module github.com/arangodb/kube-arangodb

go 1.12

replace (
	github.com/arangodb/arangosync => ./deps/github.com/arangodb/arangosync
	github.com/ugorji/go => github.com/ugorji/go v0.0.0-20181209151446-772ced7fd4c2

	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/code-generator => ./deps/k8s.io/code-generator

)

require (
	cloud.google.com/go v0.34.0
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/Azure/go-autorest/autorest v0.1.0
	github.com/Azure/go-autorest/autorest/adal v0.1.0
	github.com/Azure/go-autorest/autorest/date v0.1.0
	github.com/PuerkitoBio/purell v1.1.1
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578
	github.com/aktau/github-release v0.7.2
	github.com/arangodb-helper/go-certificates v0.0.0-20180821055445-9fca24fc2680
	github.com/arangodb/arangosync v0.0.0-00010101000000-000000000000
	github.com/arangodb/go-driver v0.0.0-20190430103524-b14f41496c3d
	github.com/arangodb/go-upgrade-rules v0.0.0-20180809110947-031b4774ff21
	github.com/arangodb/go-velocypack v0.0.0-20190129082528-7896a965b4ad
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973
	github.com/bugagazavr/go-gitlab-client v0.0.0-20150830002541-e5999f934dc4
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/cockroachdb/cmux v0.0.0-20170110192607-30d10be49292
	github.com/coreos/bbolt v1.3.2
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/coreos/go-semver v0.3.0
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f
	github.com/cpuguy83/go-md2man v1.0.10
	github.com/davecgh/go-spew v1.1.1
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/dustin/go-broadcast v0.0.0-20171205050544-f664265f5a66
	github.com/dustin/go-humanize v1.0.0
	github.com/emicklei/go-restful v0.0.0-20170410110728-ff4f55a20633
	github.com/evanphx/json-patch v4.2.0+incompatible // indirect
	github.com/ewoutp/go-gitlab-client v0.0.0-20150214183219-6e4464cd3221
	github.com/ghodss/yaml v1.0.0
	github.com/gin-contrib/sse v0.0.0-20190301062529-5545eab6dad3
	github.com/gin-gonic/autotls v0.0.0-20190406003154-fb31fc47f521
	github.com/gin-gonic/gin v1.3.0
	github.com/go-kit/kit v0.8.0
	github.com/go-openapi/analysis v0.19.0
	github.com/go-openapi/errors v0.19.0
	github.com/go-openapi/jsonpointer v0.18.0
	github.com/go-openapi/jsonreference v0.18.0
	github.com/go-openapi/loads v0.19.0
	github.com/go-openapi/runtime v0.19.0
	github.com/go-openapi/spec v0.18.0
	github.com/go-openapi/strfmt v0.19.0
	github.com/go-openapi/swag v0.18.0
	github.com/gogo/protobuf v1.2.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef
	github.com/golang/protobuf v1.3.1
	github.com/google/btree v1.0.0
	github.com/google/gofuzz v1.0.0
	github.com/googleapis/gax-go v2.0.2+incompatible // indirect
	github.com/googleapis/gnostic v0.2.0
	github.com/gophercloud/gophercloud v0.0.0-20190504011306-6f9faf57fddc
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.8.5
	github.com/hashicorp/golang-lru v0.5.1
	github.com/jessevdk/go-assets v0.0.0-20160921144138-4f4301a06e15
	github.com/jessevdk/go-assets-builder v0.0.0-20130903091706-b8483521738f // indirect
	github.com/jessevdk/go-flags v1.4.0
	github.com/jonboulle/clockwork v0.1.0
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/juju/errgo v0.0.0-20140925100237-08cceb5d0b53
	github.com/julienschmidt/httprouter v1.2.0
	github.com/kr/pretty v0.1.0
	github.com/mailru/easyjson v0.0.0-20190312143242-1de009706dbe
	github.com/manucorporat/stats v0.0.0-20180402194714-3ba42d56d227
	github.com/mattn/go-colorable v0.1.1
	github.com/mattn/go-isatty v0.0.7
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/mitchellh/go-homedir v1.1.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v1.0.1
	github.com/mwitkow/go-conntrack v0.0.0-20161129095857-cc309e4a2223
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pavel-v-chernykh/keystore-go v2.1.0+incompatible
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/pkg/errors v0.8.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/prometheus/client_model v0.0.0-20190115171406-56726106282f
	github.com/prometheus/common v0.2.0
	github.com/prometheus/procfs v0.0.0-20190117184657-bf6a532e95b1
	github.com/pulcy/pulsar v0.0.0-20180915062927-71ea24b0ec2f
	github.com/rs/zerolog v1.14.3
	github.com/russross/blackfriday v2.0.0+incompatible
	github.com/shurcooL/sanitized_anchor_name v1.0.0
	github.com/sirupsen/logrus v1.4.1
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/sourcegraph/go-vcsurl v0.0.0-20161114165620-2305ecca26ab
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2
	github.com/stretchr/objx v0.1.1
	github.com/stretchr/testify v1.3.0
	github.com/thinkerou/favicon v0.1.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80
	github.com/ugorji/go v1.1.4 // indirect
	github.com/ugorji/go/codec v0.0.0-20181209151446-772ced7fd4c2
	github.com/voxelbrain/goptions v0.0.0-20180630082107-58cddc247ea2
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/crypto v0.0.0-20190426145343-a29dc8fdc734
	golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20190506115046-ca7f33d4116e
	golang.org/x/text v0.3.1-0.20181227161524-e6919f6577db
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	golang.org/x/tools v0.0.0-20190425163242-31fd60d6bfdc
	google.golang.org/api v0.4.0 // indirect
	google.golang.org/genproto v0.0.0-20190502173448-54afdca5d873
	google.golang.org/grpc v1.20.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
	gopkg.in/go-playground/assert.v1 v1.2.1
	gopkg.in/go-playground/validator.v8 v8.18.2
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/sourcegraph/go-vcsurl.v1 v1.0.0-20131114132947-6b12603ea6fd
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/gengo v0.0.0-20190128074634-0689ccc1d7d6
	k8s.io/klog v0.3.0
	k8s.io/kube-openapi v0.0.0-20190502190224-411b2483e503 // indirect
	k8s.io/utils v0.0.0-20190506122338-8fab8cb257d5 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)
