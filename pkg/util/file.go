//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package util

import (
	"io"
	"os"

	"github.com/pkg/errors"
)

func OpenWithRead(path string, size int) ([]byte, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	buff := make([]byte, size+1)
	if s, err := Read(f, buff); err != nil {
		if nerr := f.Close(); nerr != nil {
			return nil, nerr
		}

		return nil, err
	} else if s == 0 {
		return nil, io.ErrUnexpectedEOF
	} else {
		return buff[:s], nil
	}
}

func Read(in io.Reader, buff []byte) (int, error) {
	readed := 0
	for {
		s, err := in.Read(buff)
		readed += s
		if err != nil {
			if errors.Is(err, io.EOF) {
				return readed, nil
			}

			return readed, err
		}
	}
}
