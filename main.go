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
	"github.com/armory-io/pacrd/events"
	"os"

	pacrdv1alpha1 "github.com/armory-io/pacrd/api/v1alpha1"
	"github.com/armory-io/pacrd/controllers"
	"github.com/armory/plank"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = pacrdv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	setupLog.Info("Initializing PaCRD configuration")
	pacrdConfig, err := InitConfig()
	if err != nil {
		setupLog.Error(err, "Unable to parse configuration")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Create a new relic app
	eventClient, errevent  := events.NewNewRelicEventClient(events.EventClientSettings{
		AppName: "pacrd",
		ApiKey:  pacrdConfig.NewRelicLicense,
	})
	if errevent != nil {
		fmt.Println("unable to create New Relic Application", err)
		eventClient = new(events.DefaultClient)
	}


	spinnakerClient := plank.New()
	spinnakerClient.URLs["orca"] = pacrdConfig.SpinnakerServices.Orca
	spinnakerClient.URLs["front50"] = pacrdConfig.SpinnakerServices.Front50
	// Optionally set the fiat user if it is provided in the config.
	if pacrdConfig.FiatServiceAccount != "" {
		spinnakerClient.FiatUser = pacrdConfig.FiatServiceAccount
	}

	if err = (&controllers.ApplicationReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("Application"),
		Scheme:          mgr.GetScheme(),
		SpinnakerClient: spinnakerClient,
		Recorder:        mgr.GetEventRecorderFor("applications"),
		EventClient:	 eventClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Application")
		os.Exit(1)
	}
	if err = (&controllers.PipelineReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("Pipeline"),
		Scheme:          mgr.GetScheme(),
		SpinnakerClient: spinnakerClient,
		Recorder:        mgr.GetEventRecorderFor("pipelines"),
		EventClient:	 eventClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Pipeline")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
