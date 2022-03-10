//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package v2alpha1

import (
	"fmt"
	"reflect"

	driver "github.com/arangodb/go-driver"
)

// ImageInfo contains an ID of an image and the ArangoDB version inside the image.
type ImageInfo struct {
	Image           string                                      `json:"image"`                      // Human provided name of the image
	ImageID         string                                      `json:"image-id,omitempty"`         // @deprecated -> use ArchImageID | Unique ID (with SHA256) of the image for default Arch type
	ArangoDBVersion driver.Version                              `json:"arangodb-version,omitempty"` // ArangoDB version within the image
	Enterprise      bool                                        `json:"enterprise,omitempty"`       // If set, this is an enterprise image
	ArchImageID     map[ArangoDeploymentArchitectureType]string `json:"archImageID,omitempty"`      // Unique ID (with SHA256) of the image by Arch type
}

func (in *ImageInfo) String() string {
	if in == nil {
		return "undefined"
	}

	e := "Community"

	if in.Enterprise {
		e = "Enterprise"
	}

	return fmt.Sprintf("ArangoDB %s %s (%s)", e, string(in.ArangoDBVersion), in.Image)
}

// ImageInfoList is a list of image infos
type ImageInfoList []ImageInfo

func (in ImageInfoList) Add(i ...ImageInfo) ImageInfoList {
	return append(in, i...)
}

// GetByImage returns the info in the given list for the image with given name.
// If not found, false is returned.
func (in ImageInfoList) GetByImage(image string) (ImageInfo, bool) {
	for _, x := range in {
		if x.Image == image {
			return x, true
		}
	}
	return ImageInfo{}, false
}

// GetByImageAndArch returns the ImageInfo in the given list for the given image and arch
// If not found, false is returned.
func (in ImageInfoList) GetByImageAndArch(image string, arch ArangoDeploymentArchitectureType) (ImageInfo, bool) {
	for _, x := range in {
		if x.Image == image {
			if _, ok := x.ArchImageID[arch]; ok {
				return x, true
			}
			return x, false
		}
	}
	return ImageInfo{}, false
}

// GetByImageID returns the info in the given list for the image with given id.
// If not found, false is returned.
func (in ImageInfoList) GetByImageID(imageID string, arch ArangoDeploymentArchitectureType) (ImageInfo, bool) {
	for _, x := range in {
		if foundImageID, ok := x.ArchImageID[arch]; ok {
			if foundImageID == imageID {
				return x, true
			}
		}
	}
	return ImageInfo{}, false
}

// AddOrUpdate adds the given info to the given list, if its image does not exist
// in the list. If the image does exist in the list, its entry is replaced by the given info.
// If not found, false is returned.
func (in *ImageInfoList) AddOrUpdate(info ImageInfo) {
	// Look for existing entry
	for i, x := range *in {
		if x.Image == info.Image {
			for arch, imgID := range x.ArchImageID {
				info.ArchImageID[arch] = imgID

				// set ImageID for backward compatibility
				if imgID, ok := info.ArchImageID[ArangoDeploymentArchitectureDefault]; ok {
					info.ImageID = imgID
				}
			}
			(*in)[i] = info
			return
		}
	}
	// No existing entry found, add it
	*in = append(*in, info)
}

// Equal compares to ImageInfo
func (in *ImageInfo) Equal(other *ImageInfo) bool {
	if in == nil && other == nil {
		return true
	} else if in == nil || other == nil {
		return false
	} else if in == other {
		return true
	}

	return in.ArangoDBVersion == other.ArangoDBVersion &&
		in.Enterprise == other.Enterprise &&
		in.Image == other.Image &&
		reflect.DeepEqual(in.ArchImageID, other.ArchImageID)
}

// Equal compares to ImageInfoList
func (in ImageInfoList) Equal(other ImageInfoList) bool {
	if len(in) != len(other) {
		return false
	}

	for i := 0; i < len(in); i++ {
		for arch, imgID := range in[i].ArchImageID {
			ii, found := in.GetByImageID(imgID, arch)

			if !found {
				return false
			}

			if !in[i].Equal(&ii) {
				return false
			}
		}
	}

	return true
}
