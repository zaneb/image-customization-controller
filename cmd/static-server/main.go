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
	"flag"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/openshift/image-customization-controller/pkg/ignition"
	"github.com/openshift/image-customization-controller/pkg/imagehandler"
	"github.com/openshift/image-customization-controller/pkg/version"
	// +kubebuilder:scaffold:imports
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

func loadStaticNMState(nmstateDir string, imageServer imagehandler.ImageHandler) error {
	files, err := ioutil.ReadDir(nmstateDir)
	if err != nil {
		return errors.WithMessagef(err, "problem reading %s", nmstateDir)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		b, err := os.ReadFile(path.Join(nmstateDir, f.Name()))
		if err != nil {
			return errors.WithMessagef(err, "problem reading %s", path.Join(nmstateDir, f.Name()))
		}
		igBuilder := ignition.New(b,
			os.Getenv("IRONIC_BASE_URL"),
			os.Getenv("IRONIC_AGENT_IMAGE"),
			os.Getenv("IRONIC_AGENT_PULL_SECRET"),
			os.Getenv("IRONIC_RAMDISK_SSH_KEY"),
		)
		ign, err := igBuilder.Generate()
		if err != nil {
			return errors.WithMessagef(err, "problem generating ignition %s", f.Name())
		}
		imageName := strings.Replace(f.Name(), ".yaml", ".iso", 1) // master-1.yaml -> master-1.iso
		url, err := imageServer.ServeImage(imageName, ign)
		if err != nil {
			return err
		}
		setupLog.Info("serving", "image", imageName, "url", url)
	}
	return nil
}

func main() {
	var devLogging bool
	var imagesBindAddr string
	var imagesPublishAddr string
	var nmstateDir string

	flag.StringVar(&imagesBindAddr, "images-bind-addr", ":8084",
		"The address the images endpoint binds to.")
	flag.StringVar(&imagesPublishAddr, "images-publish-addr", "http://127.0.0.1:8084",
		"The address clients would access the images endpoint from.")
	flag.StringVar(&nmstateDir, "nmstate-dir", "",
		"location of static nmstate files (named with the target image - master-0.yaml).")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(devLogging)))

	version.Print(setupLog)

	for _, env := range []string{"IRONIC_BASE_URL", "DEPLOY_ISO"} {
		val := os.Getenv(env)
		if val == "" {
			setupLog.Info("Missing environment", "variable", env)
			os.Exit(1)
		}
	}

	_, err := url.Parse(imagesPublishAddr)
	if err != nil {
		setupLog.Error(err, "imagesPublishAddr is not parsable")
		os.Exit(1)
	}

	imageServer := imagehandler.NewImageHandler(ctrl.Log.WithName("ImageHandler"), os.Getenv("DEPLOY_ISO"), imagesPublishAddr)
	http.Handle("/", http.FileServer(imageServer.FileSystem()))

	if nmstateDir != "" {
		if err := loadStaticNMState(nmstateDir, imageServer); err != nil {
			setupLog.Error(err, "problem loading static ignitions")
			os.Exit(1)
		}
	}
	if err := http.ListenAndServe(imagesBindAddr, nil); err != nil {
		setupLog.Error(err, "problem serving images")
		os.Exit(1)
	}
}
