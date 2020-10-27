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

package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/armory-io/pacrd/events"
	"github.com/mitchellh/mapstructure"
	"time"

	"github.com/armory/plank/v3"
	"github.com/go-logr/logr"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pacrdv1alpha1 "github.com/armory-io/pacrd/api/v1alpha1"
	"github.com/armory-io/pacrd/spinnaker"
)

const (
	// ControllerName represents this controller's name.
	ControllerName = "spinnaker-application-controller"
	// AppFinalizerName is used to mark which objects have been processed for deletion.
	AppFinalizerName = "app.finalizers.armory.io"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	SpinnakerClient spinnaker.Client
	Recorder        record.EventRecorder
	EventClient     events.EventClient
}

// +kubebuilder:rbac:groups=pacrd.armory.spinnaker.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pacrd.armory.spinnaker.io,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=patch

// Reconcile Application state within the cluster.
func (r *ApplicationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("application", req.NamespacedName)
	logger.Info("reconciling application")
	// TODO note that if the app name changes in the object I don't think we'll
	// be able to cleanly remove it from Spinnaker...

	// Get a handle on the CRD application under scrutiny
	var app pacrdv1alpha1.Application
	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// If this is the first time we're seeing this pipeline, set it's status.
	if app.Status.Phase == "" {
		_ = r.setPhase(app, pacrdv1alpha1.ApplicationCreating)
	}

	if app.ShouldDelete() {
		logger.Info("deleting application")
		_ = r.setPhase(app, pacrdv1alpha1.ApplicationDeleting)

		if err := r.deleteApplication(app); err != nil {
			r.Recorder.Eventf(
				&app,
				"Warning",
				string(pacrdv1alpha1.ApplicationDeletionFailed),
				err.Error(),
			)
			r.EventClient.SendError(string(pacrdv1alpha1.ApplicationDeletionFailed), err)
			return r.complete(app, pacrdv1alpha1.ApplicationDeletionFailed, err)
		}

		// Do we need this information?
		r.EventClient.SendEvent("ApplicationDeleted", applicationToMap(app))

		logger.Info("successfully deleted application from spinnaker")
		return ctrl.Result{}, nil
	} else if err := r.setFinalizer(app); err != nil {
		return ctrl.Result{}, err
	}

	// First, attempt to find the Spinnaker app.
	// TODO move to GetOrCreateApplicationFromSpinnaker fn
	_, spinErr := r.SpinnakerClient.GetApplication(req.Name)
	var fr *plank.FailedResponse
	if spinErr != nil {
		// If we didn't find the Spinnaker app, create it.
		if errors.As(spinErr, &fr) && fr.StatusCode == 404 {
			logger.Info("did not find application in Spinnaker, creating...")
			spinApp := app.ToSpinApplication()
			spinErr = r.SpinnakerClient.CreateApplication(&spinApp)
			// If we got an error from Spinnaker we failed to create the app,
			// so requeue the event.
			if spinErr != nil {
				// TODO it's possible this will never succeed, consider backoff
				//      or general failure
				r.EventClient.SendError(string(pacrdv1alpha1.ApplicationCreationFailed), spinErr)
				return r.complete(app, pacrdv1alpha1.ApplicationCreationFailed, spinErr)
			}

			r.Recorder.Eventf(
				&app,
				"Normal",
				string(pacrdv1alpha1.ApplicationCreated),
				"Application created in Spinnaker",
			)

			// Do we need this information?
			r.EventClient.SendEvent("Application" + string(pacrdv1alpha1.ApplicationCreated), applicationToMap(app))

			logger.Info("created spinnaker app")
			return r.complete(app, pacrdv1alpha1.ApplicationCreated, nil)
		}

		// Otherwise something else broke, so requeue the event.
		// TODO understand this failure case

		r.Recorder.Eventf(
			&app,
			"Warning",
			string(pacrdv1alpha1.ApplicationCreationFailed),
			spinErr.Error(),
		)
		r.EventClient.SendError(string(pacrdv1alpha1.ApplicationCreationFailed), spinErr)
		return r.complete(app, pacrdv1alpha1.ApplicationCreationFailed, spinErr)
	}

	logger.Info("updating application in Spinnaker")
	// Apply whatever is in the CRD to Spinnaker
	spinApp := app.ToSpinApplication()
	if err := r.SpinnakerClient.UpdateApplication(spinApp); err != nil {
		r.Recorder.Eventf(
			&app,
			"Warning",
			string(pacrdv1alpha1.ApplicationUpdateFailed),
			err.Error(),
		)
		r.EventClient.SendError(string(pacrdv1alpha1.ApplicationUpdateFailed), err)
		return r.complete(app, pacrdv1alpha1.ApplicationUpdateFailed, err)
	}

	app.Status.LastConfigured = v1.Time{Time: time.Now()}
	// TODO generalize this
	// TODO set phase appropriately
	app.Status.URL = fmt.Sprintf("http://localhost:9000/#/applications/%s/clusters", app.Name)
	if err := r.Status().Update(context.Background(), &app); err != nil {
		return ctrl.Result{}, err
	}

	r.Recorder.Eventf(
		&app,
		"Normal",
		string(pacrdv1alpha1.ApplicationUpdated),
		"Application updated",
	)

	// Do we need this information?
	r.EventClient.SendEvent("Application" + string(pacrdv1alpha1.ApplicationUpdated), applicationToMap(app))

	logger.Info("done reconciling application")
	return r.complete(app, pacrdv1alpha1.ApplicationUpdated, nil)
}

func (r *ApplicationReconciler) deleteApplication(app pacrdv1alpha1.Application) error {
	// Determine if we're actively deleting the CRD and make sure we clean up
	// any Spinnaker resources that are related. Finalizers are initialized
	// further down the method.
	finalizers := app.GetObjectMeta().GetFinalizers()

	if containsString(finalizers, AppFinalizerName) {
		if err := r.SpinnakerClient.DeleteApplication(app.Name); err != nil {
			var fr *plank.ErrUnsupportedStatusCode
			if fr != nil && fr.Code == 404 {
				// This situation implies the application was already deleted, either
				// by something else or a previously failed delete reconcile. Either
				// way we only care that it's gone, so do nothing here.
			} else {
				return err
			}
		}

		// Make sure we clean up the finalizer so k8s knows we're done with
		// this resource.
		app.SetFinalizers(removeString(finalizers, AppFinalizerName))
		if err := r.Update(context.Background(), &app); err != nil {
			return err
		}
	}

	return nil
}

// Register finalizers so that deletes propagate properly.
func (r *ApplicationReconciler) setFinalizer(app pacrdv1alpha1.Application) error {
	finalizers := app.GetObjectMeta().GetFinalizers()
	if !containsString(finalizers, AppFinalizerName) {
		app.SetFinalizers(append(finalizers, AppFinalizerName))
		if err := r.Update(context.Background(), &app); err != nil {
			return err
		}
	}

	return nil
}

func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pacrdv1alpha1.Application{}).
		Complete(r)
}

func (r *ApplicationReconciler) setPhase(app pacrdv1alpha1.Application, p pacrdv1alpha1.ApplicationPhase) error {
	if app.Status.Phase == p {
		return nil // Nothing to update
	}
	app.Status.Phase = p
	app.Status.LastConfigured = v1.Time{Time: time.Now()}
	return r.Status().Update(context.Background(), &app)
}

// Complete a reconciliation loop for the current pipeline and update phase.
// TODO this is a good candidate for some interfaces so it can be shared w/ app controller
func (r *ApplicationReconciler) complete(app pacrdv1alpha1.Application, phase pacrdv1alpha1.ApplicationPhase, e error) (ctrl.Result, error) {
	_ = r.setPhase(app, phase)
	return ctrl.Result{}, e
}

func applicationToMap( app pacrdv1alpha1.Application) map[string]interface{} {
	applicationMap := make(map[string]interface{})
	err := mapstructure.Decode(app, &applicationMap)
	if err != nil {
		return nil
	}
	return applicationMap
}