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
	"io"
	"io/fs"
	"time"

	"github.com/openshift/assisted-image-service/pkg/isoeditor"
)

// imageFile is the http.File use in imageFileSystem.
type imageFile struct {
	io.ReadSeekCloser
	name              string
	size              int64
	ignitionContent   []byte
	rhcosStreamReader io.ReadSeeker
}

// file interface implementation

var _ fs.File = &imageFile{}

func (f *imageFile) Init(isoFile string) error {
	if f.rhcosStreamReader == nil {
		var err error
		f.rhcosStreamReader, err = isoeditor.NewRHCOSStreamReader(
			isoFile,
			&isoeditor.IgnitionContent{Config: f.ignitionContent},
			nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *imageFile) Write(p []byte) (n int, err error)        { return 0, notImplementedFn("Write") }
func (f *imageFile) Stat() (fs.FileInfo, error)               { return fs.FileInfo(f), nil }
func (f *imageFile) Close() error                             { return nil }
func (f *imageFile) Readdir(count int) ([]fs.FileInfo, error) { return []fs.FileInfo{}, nil }
func (f *imageFile) Read(p []byte) (n int, err error)         { return f.rhcosStreamReader.Read(p) }
func (f *imageFile) Seek(offset int64, whence int) (int64, error) {
	return f.rhcosStreamReader.Seek(offset, whence)
}

// fileInfo interface implementation

var _ fs.FileInfo = &imageFile{}

func (i *imageFile) Name() string       { return i.name }
func (i *imageFile) Size() int64        { return i.size }
func (i *imageFile) Mode() fs.FileMode  { return 0444 }
func (i *imageFile) ModTime() time.Time { return time.Now() }
func (i *imageFile) IsDir() bool        { return false }
func (i *imageFile) Sys() interface{}   { return nil }
