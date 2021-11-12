/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package imagehandler

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/go-logr/logr"
)

// imageFileSystem is an http.FileSystem that creates a virtual filesystem of
// host images.
type imageFileSystem struct {
	isoFile *baseIso
	baseURL string
	images  []*imageFile
	mu      *sync.Mutex
	log     logr.Logger
}

var _ ImageHandler = &imageFileSystem{}
var _ http.FileSystem = &imageFileSystem{}

type ImageHandler interface {
	FileSystem() http.FileSystem
	ServeImage(name string, ignitionContent []byte) (string, error)
}

func NewImageHandler(logger logr.Logger, isoFile, baseURL string) ImageHandler {
	return &imageFileSystem{
		log:     logger,
		isoFile: newBaseIso(isoFile),
		baseURL: baseURL,
		images:  []*imageFile{},
		mu:      &sync.Mutex{},
	}
}

func (f *imageFileSystem) FileSystem() http.FileSystem {
	return f
}

func (f *imageFileSystem) ServeImage(name string, ignitionContent []byte) (string, error) {
	size, err := f.isoFile.Size()
	if err != nil {
		return "", err
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.images = append(f.images, &imageFile{
		name:            name,
		size:            size,
		ignitionContent: ignitionContent,
	})
	u, err := url.Parse(f.baseURL)
	if err != nil {
		return "", err
	}
	u.Path = name
	return u.String(), nil
}

func (f *imageFileSystem) imageFileByName(name string) *imageFile {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, im := range f.images {
		if im.name == name {
			return im
		}
	}
	return nil
}
