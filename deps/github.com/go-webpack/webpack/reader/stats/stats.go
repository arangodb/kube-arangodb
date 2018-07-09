package stats

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-webpack/webpack/util"
	"github.com/pkg/errors"
)

// Read stats plugin manifest from HTTP for development or from file for production
func Read(isDev bool, host, fsPath, webPath string) (map[string][]string, error) {
	//log.Println("stats reads", isDev)
	var data []byte
	var err error

	if isDev {
		data, err = devManifest(host, webPath)
	} else {
		data, err = prodManifest(fsPath)
	}

	if err != nil {
		return map[string][]string{}, errors.Wrap(err, "go-webpack: Error reading manifest")
	}

	return parseManifest(data)
}

// ParseManifest Get webpack manifest according to current environment
func parseManifest(data []byte) (map[string][]string, error) {
	var err error

	resp := statsResponse{}
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return map[string][]string{}, errors.Wrap(err, "go-webpack: Error parsing manifest - json decode")
	}
	webpackBase := resp.PublicPath
	//log.Println("webpackBase", webpackBase)

	assets := make(map[string][]string, len(resp.AssetsByChunkName))

	for akey, aval := range resp.AssetsByChunkName {
		var d []string
		err = json.Unmarshal(*aval, &d)
		if err != nil {
			return assets, errors.Wrap(err, fmt.Sprintf("go-webpack: Error when parsing manifest for %s: %s %s", akey, err, string(*aval)))
		}
		for i, v := range d {
			d[i] = webpackBase + v
		}

		assets[akey+".js"] = util.Filter(d, func(v string) bool {
			return strings.HasSuffix(v, ".js")
		})

		assets[akey+".css"] = util.Filter(d, func(v string) bool {
			return strings.HasSuffix(v, ".css")
		})
	}
	return assets, nil
}
