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
	isoFile       *baseIso
	initramfsFile *baseInitramfs
	baseURL       string
	images        map[string]*imageFile
	mu            *sync.Mutex
	log           logr.Logger
}

var _ ImageHandler = &imageFileSystem{}
var _ http.FileSystem = &imageFileSystem{}

type ImageHandler interface {
	FileSystem() http.FileSystem
	ServeImage(name string, ignitionContent []byte, initramfs bool) (string, error)
	RemoveImage(name string)
}

func NewImageHandler(logger logr.Logger, isoFile, initramfsFile, baseURL string) ImageHandler {
	return &imageFileSystem{
		log:           logger,
		isoFile:       newBaseIso(isoFile),
		initramfsFile: newBaseInitramfs(initramfsFile),
		baseURL:       baseURL,
		images:        map[string]*imageFile{},
		mu:            &sync.Mutex{},
	}
}

func (f *imageFileSystem) FileSystem() http.FileSystem {
	return f
}

func (f *imageFileSystem) getBaseImage(initramfs bool) baseFile {
	if initramfs {
		return f.initramfsFile
	} else {
		return f.isoFile
	}
}

func (f *imageFileSystem) ServeImage(name string, ignitionContent []byte, initramfs bool) (string, error) {
	size, err := f.getBaseImage(initramfs).Size()
	if err != nil {
		return "", err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.images[name]; !exists {
		f.images[name] = &imageFile{
			name:            name,
			size:            size,
			ignitionContent: ignitionContent,
			initramfs:       initramfs,
		}
	}
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
	return f.images[name]
}

func (f *imageFileSystem) RemoveImage(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.images, name)
}
