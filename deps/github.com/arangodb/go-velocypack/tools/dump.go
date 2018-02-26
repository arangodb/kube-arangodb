package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"strings"

	velocypack "github.com/arangodb/go-velocypack"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		log.Fatalln("Usage: dump <hex encoded slice>")
	}
	slice, err := hex.DecodeString(strings.TrimSpace(args[0]))
	if err != nil {
		log.Fatalf("Failed to decode hex slice: %#v\n", err)
	}
	json, err := velocypack.Slice(slice).JSONString()
	if err != nil {
		log.Fatalf("Failed to convert slice: %#v\n", err)
	}
	fmt.Println(json)
}
