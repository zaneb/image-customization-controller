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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	metal3iov1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/metal3-io/baremetal-operator/pkg/secretutils"
	"github.com/metal3-io/baremetal-operator/pkg/version"
	metal3iocontroller "github.com/openshift/image-customization-controller/controllers/metal3.io"
	"github.com/openshift/image-customization-controller/pkg/ignition"
	"github.com/openshift/image-customization-controller/pkg/imagehandler"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = k8sruntime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = metal3iov1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func printVersion() {
	setupLog.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	setupLog.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	setupLog.Info(fmt.Sprintf("Git commit: %s", version.Commit))
	setupLog.Info(fmt.Sprintf("Build time: %s", version.BuildTime))
	setupLog.Info(fmt.Sprintf("Component: %s", version.String))
}

func setupChecks(mgr ctrl.Manager) error {
	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to create ready check")
		return err
	}

	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to create health check")
		return err
	}
	return nil
}

func runController(watchNamespace string, imageServer imagehandler.ImageHandler) error {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                scheme,
		Port:                  0, // Add flag with default of 9443 when adding webhooks
		Namespace:             watchNamespace,
		ClientDisableCacheFor: []client.Object{&corev1.Secret{}},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return err
	}

	imgReconciler := metal3iocontroller.PreprovisioningImageReconciler{
		Client:       mgr.GetClient(),
		Log:          ctrl.Log.WithName("controllers").WithName("PreprovisioningImage"),
		APIReader:    mgr.GetAPIReader(),
		Scheme:       mgr.GetScheme(),
		ImageHandler: imageServer,
	}
	if err = (&imgReconciler).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PreprovisioningImage")
		return err
	}

	// +kubebuilder:scaffold:builder

	if err := setupChecks(mgr); err != nil {
		return err
	}

	setupLog.Info("starting manager")
	return mgr.Start(ctrl.SetupSignalHandler())
}

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
	var watchNamespace string
	var devLogging bool
	var imagesBindAddr string
	var imagesPublishAddr string
	var startController bool
	var nmstateDir string

	// From CAPI point of view, BMO should be able to watch all namespaces
	// in case of a deployment that is not multi-tenant. If the deployment
	// is for multi-tenancy, then the BMO should watch only the provided
	// namespace.
	flag.StringVar(&watchNamespace, "namespace", os.Getenv("WATCH_NAMESPACE"),
		"Namespace that the controller watches to reconcile preprovisioningimage resources.")
	flag.StringVar(&imagesBindAddr, "images-bind-addr", ":8084",
		"The address the images endpoint binds to.")
	flag.StringVar(&imagesPublishAddr, "images-publish-addr", "http://127.0.0.1:8084",
		"The address clients would access the images endpoint from.")
	flag.BoolVar(&startController, "start-controller", true,
		"Start the controller to reconcile preprovisioningimage resources.")
	flag.StringVar(&nmstateDir, "nmstate-dir", "",
		"location of static nmstate files (named with the target image - master-0.yaml).")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(devLogging)))

	printVersion()

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

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:    scheme,
		Port:      0, // Add flag with default of 9443 when adding webhooks
		Namespace: watchNamespace,
		NewCache: cache.BuilderWithOptions(cache.Options{
			SelectorsByObject: secretutils.AddSecretSelector(nil),
		}),
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	imgReconciler := metal3iocontroller.PreprovisioningImageReconciler{
		Client:       mgr.GetClient(),
		Log:          ctrl.Log.WithName("controllers").WithName("PreprovisioningImage"),
		APIReader:    mgr.GetAPIReader(),
		Scheme:       mgr.GetScheme(),
		ImageHandler: imageServer,
	}
	if err = (&imgReconciler).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PreprovisioningImage")
		os.Exit(1)
	}
	if err := http.ListenAndServe(imagesBindAddr, nil); err != nil {
		setupLog.Error(err, "problem serving images")
		os.Exit(1)
	}
}
