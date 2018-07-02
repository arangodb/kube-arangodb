package stats

import "encoding/json"

type statsResponse struct {
	Errors            []string                    `json:"errors"`
	Warning           []string                    `json:"warnings"`
	Version           string                      `json:"version"`
	Hash              string                      `json:"hash"`
	PublicPath        string                      `json:"publicPath"`
	AssetsByChunkName map[string]*json.RawMessage `json:"assetsByChunkName"`
	Assets            []*json.RawMessage          `json:"assets"`
}
