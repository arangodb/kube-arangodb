package main

import (
	"github.com/arangodb/kube-arangodb/pkg/backup/handlers/arango/backup"
	"github.com/arangodb/kube-arangodb/pkg/backup/handlers/arango/policy"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	arangoInformer "github.com/arangodb/kube-arangodb/pkg/generated/informers/externalversions"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"k8s.io/client-go/tools/clientcmd"
	"math/rand"
	"net/http"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	operatorName = "operator"
)

func main() {
	operator := operator.NewOperator(operatorName)

	rand.Seed(time.Now().Unix())

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	restClient, err := clientcmd.BuildConfigFromFlags("", "/home/ubuntu/.kube/config")
	if err != nil {
		panic(err)
	}

	arangoClientSet, err := arangoClientSet.NewForConfig(restClient)
	if err != nil {
		panic(err)
	}

	arangoInformer := arangoInformer.NewSharedInformerFactoryWithOptions(arangoClientSet, 30*time.Second, arangoInformer.WithNamespace("test"))

	if err =backup.RegisterInformer(operator, arangoClientSet, arangoInformer); err != nil {
		panic(err)
	}

	if err = policy.RegisterInformer(operator, arangoClientSet, arangoInformer); err != nil {
		panic(err)
	}

	if err = operator.RegisterStarter(arangoInformer); err != nil {
		panic(err)
	}

	stopCh := make(chan struct{})

	operator.Start(2, stopCh)

	prometheus.MustRegister(operator)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe("127.0.0.1:5000", nil)
	}()

	<-stopCh
}