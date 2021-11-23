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

package main

import (
	"io/fs"
	"net/http"
	"reflect"
	"testing"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/openshift/image-customization-controller/pkg/env"
	"github.com/openshift/image-customization-controller/pkg/imagehandler"
)

type fakeImageFileSystem struct {
	imagesServed []string
}

var _ imagehandler.ImageHandler = &fakeImageFileSystem{}
var _ http.FileSystem = &fakeImageFileSystem{}

func (f *fakeImageFileSystem) Stat() (fs.FileInfo, error)                   { return nil, nil }
func (f *fakeImageFileSystem) Read(p []byte) (n int, err error)             { return 0, nil }
func (f *fakeImageFileSystem) Write(p []byte) (n int, err error)            { return 0, nil }
func (f *fakeImageFileSystem) Close() error                                 { return nil }
func (f *fakeImageFileSystem) Seek(offset int64, whence int) (int64, error) { return 0, nil }
func (f *fakeImageFileSystem) Readdir(n int) ([]fs.FileInfo, error)         { return nil, nil }
func (f *fakeImageFileSystem) Open(name string) (http.File, error)          { return nil, nil }
func (f *fakeImageFileSystem) FileSystem() http.FileSystem                  { return f }
func (f *fakeImageFileSystem) ServeImage(name string, ignitionContent []byte, initrd, static bool) (string, error) {
	f.imagesServed = append(f.imagesServed, name)
	return "", nil
}
func (f *fakeImageFileSystem) RemoveImage(name string) {}

func TestLoadStaticNMState(t *testing.T) {
	fifs := &fakeImageFileSystem{imagesServed: []string{}}
	env := &env.EnvInputs{
		DeployISO:     "foo.iso",
		IronicBaseURL: "http://example.com",
	}
	if err := loadStaticNMState(env, "../../test/data", fifs); err != nil {
		t.Errorf("loadStaticNMState() error = %v", err)
	}
	if !reflect.DeepEqual(fifs.imagesServed, []string{"master-0.iso", "master-1.iso", "master-2.iso"}) {
		t.Errorf("loadStaticNMState() images = %v", fifs.imagesServed)
	}
}
