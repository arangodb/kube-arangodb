package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os/exec"
)

var (
	// addresses
	localPort  = flag.String("lport", "4444", "proxy local port")
	targetPort = flag.String("rport", "8080", "proxy remote port")

	// tls configuration for proxy as a server (listen)
	localTLS  = flag.Bool("ltls", false, "tls/ssl between client and proxy, you must set 'lcert' and 'lkey'")
	localCert = flag.String("lcert", "", "certificate file for proxy server side")
	localKey  = flag.String("lkey", "", "key x509 file for proxy server side")

	// tls configuration for proxy as a client (connection to target)
	targetTLS  = flag.Bool("rtls", false, "tls/ssl between proxy and target, you must set 'rcert' and 'rkey'")
	targetCert = flag.String("rcert", "", "certificate file for proxy client side")
	targetKey  = flag.String("rkey", "", "key x509 file for proxy client side")
)

func main() {
	flag.Parse()

	p := Server{
		Addr:   fmt.Sprintf(":%s", *localPort),
		Target: fmt.Sprintf(":%s", *targetPort),
	}

	go runPortForward(*localPort, *targetPort)

	if *targetTLS {
		cert, err := tls.LoadX509KeyPair(*targetCert, *targetKey)
		if err != nil {
			log.Fatalf("configuration tls for target connection: %v", err)
		}
		p.TLSConfigTarget = &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	}

	log.Println("Proxying from " + p.Addr + " to " + p.Target)
	if *localTLS {
		p.ListenAndServeTLS(*localCert, *localKey)
	} else {
		p.ListenAndServe()
	}
}

func runPortForward(portLocal, portRemote string) {
	cmd := exec.Command("kubectl",
		"--kubeconfig", "/Users/jwierzbo/.kube/config_arango",
		"port-forward", "-n", "jakubwierzbowski",
		"service/triton-triton-inference-server-metrics", fmt.Sprintf("%s:%s", portLocal, portRemote))

	_, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
