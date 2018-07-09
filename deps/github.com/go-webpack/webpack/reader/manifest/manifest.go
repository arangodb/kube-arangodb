package manifest

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

// Read webpack-manifest-plugin format manifest
func Read(path string) (map[string][]string, error) {
	assets := make(map[string][]string, 0)
	data, err := ioutil.ReadFile("./public/webpack/manifest.json")
	if err != nil {
		return assets, errors.Wrap(err, "go-webpack: Error when loading manifest from file")
	}

	response := make(map[string]string, 0)
	json.Unmarshal(data, &response)
	for key, value := range response {
		//log.Println("found asset", key, value)
		if !strings.Contains(value, ".map") {
			assets[key] = []string{value}
		}
	}
	return assets, nil
}
