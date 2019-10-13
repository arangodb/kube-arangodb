module github.com/arangodb/kube-arangodb

go 1.12

replace (
	github.com/ugorji/go => github.com/ugorji/go v0.0.0-20181209151446-772ced7fd4c2

	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Azure/go-autorest/autorest v0.1.0 // indirect
	github.com/aktau/github-release v0.7.2 // indirect
	github.com/arangodb-helper/go-certificates v0.0.0-20180821055445-9fca24fc2680
	github.com/arangodb/arangosync-client v0.6.3
	github.com/arangodb/go-driver v0.0.0-20191002124627-11b6bfc64f67
	github.com/arangodb/go-upgrade-rules v0.0.0-20180809110947-031b4774ff21
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/bugagazavr/go-gitlab-client v0.0.0-20150830002541-e5999f934dc4 // indirect
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/coreos/bbolt v1.3.2 // indirect
	github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/coreos/go-semver v0.3.0
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/coreos/prometheus-operator v0.31.1
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815 // indirect
	github.com/dustin/go-broadcast v0.0.0-20171205050544-f664265f5a66 // indirect
	github.com/evanphx/json-patch v4.2.0+incompatible
	github.com/ewoutp/go-gitlab-client v0.0.0-20150214183219-6e4464cd3221 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gin-contrib/sse v0.0.0-20190301062529-5545eab6dad3 // indirect
	github.com/gin-gonic/autotls v0.0.0-20190406003154-fb31fc47f521 // indirect
	github.com/gin-gonic/gin v1.3.0
	github.com/go-openapi/analysis v0.19.0 // indirect
	github.com/go-openapi/errors v0.19.0 // indirect
	github.com/go-openapi/loads v0.19.0 // indirect
	github.com/go-openapi/runtime v0.19.0 // indirect
	github.com/go-openapi/strfmt v0.19.0 // indirect
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gophercloud/gophercloud v0.0.0-20190504011306-6f9faf57fddc // indirect
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75 // indirect
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/helm/helm v2.14.3+incompatible // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jessevdk/go-assets v0.0.0-20160921144138-4f4301a06e15
	github.com/jessevdk/go-assets-builder v0.0.0-20130903091706-b8483521738f // indirect
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/juju/errgo v0.0.0-20140925100237-08cceb5d0b53 // indirect
	github.com/julienschmidt/httprouter v1.2.0
	github.com/magiconair/properties v1.8.0
	github.com/manucorporat/stats v0.0.0-20180402194714-3ba42d56d227 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/mattn/go-isatty v0.0.7 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.0.0
	github.com/pulcy/pulsar v0.0.0-20180915062927-71ea24b0ec2f // indirect
	github.com/robfig/cron v1.2.0
	github.com/rs/zerolog v1.14.3
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sourcegraph/go-vcsurl v0.0.0-20161114165620-2305ecca26ab // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/thinkerou/favicon v0.1.0 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80 // indirect
	github.com/ugorji/go v1.1.4 // indirect
	github.com/voxelbrain/goptions v0.0.0-20180630082107-58cddc247ea2 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/sys v0.0.0-20190506115046-ca7f33d4116e
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	google.golang.org/api v0.4.0 // indirect
	google.golang.org/genproto v0.0.0-20190502173448-54afdca5d873 // indirect
	google.golang.org/grpc v1.20.1 // indirect
	gopkg.in/sourcegraph/go-vcsurl.v1 v1.0.0-20131114132947-6b12603ea6fd // indirect
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v0.3.0
	k8s.io/kube-openapi v0.0.0-20190502190224-411b2483e503 // indirect
	k8s.io/utils v0.0.0-20190506122338-8fab8cb257d5 // indirect
)
