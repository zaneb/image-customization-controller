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
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestImageHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/host-xyz-45.iso", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	imageServer := &imageFileSystem{
		log:     zap.New(zap.UseDevMode(true)),
		isoFile: &baseIso{baseFileData{filename: "dummyfile.iso", size: 12345}},
		baseURL: "http://localhost:8080",
		images: []*imageFile{
			{
				name:            "host-xyz-45.iso",
				size:            12345,
				ignitionContent: []byte("asietonarst"),
				imageReader:     strings.NewReader("aiosetnarsetin"),
			},
		},
		mu: &sync.Mutex{},
	}

	handler := http.FileServer(imageServer.FileSystem())
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `aiosetnarsetin`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
