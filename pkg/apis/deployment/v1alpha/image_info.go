//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Ewout Prangsma
//

package v1alpha

import driver "github.com/arangodb/go-driver"

// ImageInfo contains an ID of an image and the ArangoDB version inside the image.
type ImageInfo struct {
	Image           string         `json:"image"`                      // Human provided name of the image
	ImageID         string         `json:"image-id,omitempty"`         // Unique ID (with SHA256) of the image
	ArangoDBVersion driver.Version `json:"arangodb-version,omitempty"` // ArangoDB version within the image
	Enterprise      bool           `json:"enterprise,omitempty"`       // If set, this is an enterprise image
}

// ImageInfoList is a list of image infos
type ImageInfoList []ImageInfo

// GetByImage returns the info in the given list for the image with given name.
// If not found, false is returned.
func (l ImageInfoList) GetByImage(image string) (ImageInfo, bool) {
	for _, x := range l {
		if x.Image == image {
			return x, true
		}
	}
	return ImageInfo{}, false
}

// GetByImageID returns the info in the given list for the image with given id.
// If not found, false is returned.
func (l ImageInfoList) GetByImageID(imageID string) (ImageInfo, bool) {
	for _, x := range l {
		if x.ImageID == imageID {
			return x, true
		}
	}
	return ImageInfo{}, false
}

// AddOrUpdate adds the given info to the given list, if its image does not exist
// in the list. If the image does exist in the list, its entry is replaced by the given info.
// If not found, false is returned.
func (l *ImageInfoList) AddOrUpdate(info ImageInfo) {
	// Look for existing entry
	for i, x := range *l {
		if x.Image == info.Image {
			(*l)[i] = info
			return
		}
	}
	// No existing entry found, add it
	*l = append(*l, info)
}
