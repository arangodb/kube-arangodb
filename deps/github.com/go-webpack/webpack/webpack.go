package webpack

import (
	"errors"
	"html/template"
	"log"
	"strings"

	"github.com/go-webpack/webpack/helper"
	"github.com/go-webpack/webpack/reader"
)

// DevHost webpack-dev-server host:port
var DevHost = "localhost:3808"

// FsPath filesystem path to public webpack dir
var FsPath = "public/webpack"

// WebPath http path to public webpack dir
var WebPath = "webpack"

// Plugin webpack plugin to use, can be stats or manifest
var Plugin = "stats"

// IgnoreMissing ignore assets missing on manifest or fail on them
var IgnoreMissing = true

// Verbose error messages to console (even if error is ignored)
var Verbose = true

var isDev = false
var initDone = false
var preloadedAssets map[string][]string

func readManifest() (map[string][]string, error) {
	return reader.Read(Plugin, DevHost, FsPath, WebPath, isDev)
}

// Init Set current environment and preload manifest
func Init(dev bool) {
	var err error
	isDev = dev
	if isDev {
		// Try to preload manifest, so we can show an error if webpack-dev-server is not running
		_, err = readManifest()
	} else {
		preloadedAssets, err = readManifest()
	}
	if err != nil {
		log.Println(err)
	}
	initDone = true
}

// AssetHelper renders asset tag with url from webpack manifest to the page
func AssetHelper(key string) (template.HTML, error) {
	var err error

	if !initDone {
		return "", errors.New("Please call webpack.Init() first (see readme)")
	}

	var assets map[string][]string
	if isDev {
		assets, err = readManifest()
		if err != nil {
			return template.HTML(""), err
		}
	} else {
		assets = preloadedAssets
	}

	parts := strings.Split(key, ".")
	kind := parts[len(parts)-1]
	//log.Println("showing assets:", key, parts, kind)

	v, ok := assets[key]
	if !ok {
		message := "go-webpack: Asset file '" + key + "' not found in manifest"
		if Verbose {
			log.Printf("%s. Manifest contents: %+v", message, assets)
		}
		if IgnoreMissing {
			return template.HTML(""), nil
		}
		return template.HTML(""), errors.New(message)
	}

	buf := []string{}
	for _, s := range v {
		if strings.HasSuffix(s, "."+kind) {
			buf = append(buf, helper.AssetTag(kind, s))
		} else {
			log.Println("skip asset", s, ": bad type")
		}
	}
	return template.HTML(strings.Join(buf, "\n")), nil
}
