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
	//PipelineFinalizerName is used to mark which objects have been processed for deletion.
	PipelineFinalizerName = "pipeline.finalizers.armory.io"
)

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	SpinnakerClient spinnaker.Client
	Recorder        record.EventRecorder
	EventClient     events.EventClient
}

// +kubebuilder:rbac:groups=pacrd.armory.spinnaker.io,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pacrd.armory.spinnaker.io,resources=pipelines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=patch

// Reconcile the status of a pipeline between K8s and Spinnaker.
func (r *PipelineReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("pipeline", req.NamespacedName)

	// Get a handle on the current pipeline object
	var pipeline pacrdv1alpha1.Pipeline
	if err := r.Get(ctx, req.NamespacedName, &pipeline); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling pipeline")

	// If this is the first time we're seeing this pipeline, set it's status.
	if pipeline.Status.Phase == "" {
		_ = r.setPhase(pipeline, pacrdv1alpha1.PipelineCreating)
	}

	if pipeline.ShouldDelete() {
		logger.Info("Deleting pipeline")
		_ = r.setPhase(pipeline, pacrdv1alpha1.PipelineDeleting)

		if err := r.deletePipeline(pipeline); err != nil {
			r.EventClient.SendError(string(pacrdv1alpha1.PipelineDeletionFailed), err)
			return r.complete(pipeline, pacrdv1alpha1.PipelineDeletionFailed, err)
		}

		logger.Info("Successfully deleted pipeline")
		return ctrl.Result{}, nil

	} else if err := r.setFinalizer(pipeline); err != nil {
		return ctrl.Result{}, err
	}

	// At this point we can start validating pipelines. Let's start by making
	// sure for any artifact reference, we can assert that there is an equivalent
	// artifact somewhere else in the pipeline.

	for _, s := range pipeline.Spec.Stages {
		if s.Type == "bakeManifest" {
			for _, ia := range s.GetStage().(*pacrdv1alpha1.BakeManifest).InputArtifacts {
				id, err := pacrdv1alpha1.GetArtifactID(*pipeline.Spec.ExpectedArtifacts, *ia)
				if err != nil {

					// Let the user know via describe calls that something went wrong.
					r.Recorder.Eventf(
						&pipeline,
						"Warning",
						string(pacrdv1alpha1.PipelineValidationFailed),
						err.Error(),
					)

					r.EventClient.SendError(string(pacrdv1alpha1.PipelineValidationFailed), err)
					return r.complete(pipeline, pacrdv1alpha1.PipelineValidationFailed, err)
				}

				ia.ID = id
			}
		}
	}

	// TODO differentiate between app not found and pipeline not found
	// FIXME get pipelines does not seem to honor the application name
	pipeRes, err := r.SpinnakerClient.GetPipelines(pipeline.Spec.Application)

	if err != nil {
		return r.complete(pipeline, pacrdv1alpha1.PipelineOrApplicationNotFound, err)
	}

	upstreamPipeline, err := findPipeline(pipeRes, pipeline.Name, pipeline.Status.ID)

	if err != nil {
		logger.Info("Did not find pipeline in Spinnaker, creating...")

		if err := r.createPipeline(pipeline); err != nil {
			r.EventClient.SendError(string(pacrdv1alpha1.PipelineCreationFailed), err)
			return r.complete(pipeline, pacrdv1alpha1.PipelineCreationFailed, err)
		}

		logger.Info("Created spinnaker pipeline")

		// Do we need this information?
		r.EventClient.SendEvent("Pipeline" + string(pacrdv1alpha1.PipelineCreated), pipelineToMap(pipeline))

		return r.complete(pipeline, pacrdv1alpha1.PipelineCreated, nil)
	}

	logger.Info("Updating pipeline")
	if err := r.updatePipeline(ctx, pipeline, upstreamPipeline); err != nil {
		logger.Info("Failed to update pipeline")
		r.EventClient.SendError(string(pacrdv1alpha1.PipelineUpdateFailed), err)
		return r.complete(pipeline, pacrdv1alpha1.PipelineUpdateFailed, err)
	}
	logger.Info("Done updating pipeline")

	// Do we need this information?
	r.EventClient.SendEvent("Pipeline" + string(pacrdv1alpha1.PipelineUpdated), pipelineToMap(pipeline))

	return r.complete(pipeline, pacrdv1alpha1.PipelineUpdated, nil)

	// TODO watches on Application types
}

// Complete a reconciliation loop for the current pipeline and update phase.
// TODO this is a good candidate for some interfaces so it can be shared w/ app controller
func (r *PipelineReconciler) complete(pipeline pacrdv1alpha1.Pipeline, phase pacrdv1alpha1.PipelinePhase, e error) (ctrl.Result, error) {
	_ = r.setPhase(pipeline, phase)
	return ctrl.Result{}, e
}

// SetPhase sets the phase for the current pipeline.
// NOTE: this triggers another event in the reconciliation queue so use this method conservatively.
// TODO this is a good candidate for some interfaces so it can be shared w/ app controller
func (r *PipelineReconciler) setPhase(pipeline pacrdv1alpha1.Pipeline, p pacrdv1alpha1.PipelinePhase) error {
	if pipeline.Status.Phase == p {
		return nil // Nothing to update
	}
	pipeline.Status.Phase = p
	pipeline.Status.LastConfigured = v1.Time{Time: time.Now()}
	return r.Status().Update(context.Background(), &pipeline)
}

// SetFinalzier registers a finalizer so that deletes propagate properly.
func (r *PipelineReconciler) setFinalizer(pipeline pacrdv1alpha1.Pipeline) error {
	finalizers := pipeline.GetObjectMeta().GetFinalizers()
	if !containsString(finalizers, PipelineFinalizerName) {
		pipeline.SetFinalizers(append(finalizers, PipelineFinalizerName))
		if err := r.Update(context.Background(), &pipeline); err != nil {
			return err
		}
	}

	return nil
}

func (r *PipelineReconciler) getPipeline(pipeline pacrdv1alpha1.Pipeline) (plank.Pipeline, error) {

	// FIXME get pipelines does not seem to honor the application name
	pipeRes, err := r.SpinnakerClient.GetPipelines(pipeline.Spec.Application)

	if err != nil {
		r.Recorder.Eventf(
			&pipeline,
			"Warning",
			string(pacrdv1alpha1.PipelineOrApplicationNotFound),
			err.Error(),
		)
		r.EventClient.SendError(string(pacrdv1alpha1.PipelineOrApplicationNotFound), err)
		return plank.Pipeline{}, err
	}

	return findPipeline(pipeRes, pipeline.Name, pipeline.Status.ID)

}

// CreatePipeline attempts to create a pipeline in the upstream Spinnaker.
// This method differs from updatePipeline in its propagated Events and Phases.
func (r *PipelineReconciler) createPipeline(pipeline pacrdv1alpha1.Pipeline) error {
	// Note that passing the empty string here is denoting that the pipeline
	// doesn't have a valid ID. It will be created according to the app name
	// provided in the spec.
	pipe, err := pipeline.ToSpinnakerPipeline()
	if err != nil {
		// Let the user know via describe calls that something went wrong.
		r.Recorder.Eventf(
			&pipeline,
			"Warning",
			string(pacrdv1alpha1.PipelineCreationFailed),
			err.Error(),
		)
		r.EventClient.SendError(string(pacrdv1alpha1.PipelineCreationFailed), err)
		return err
	}

	// Log pipeline stages
	r.EventClient.SendPipelineStages(pipe)

	if err := r.SpinnakerClient.UpsertPipeline(pipe, ""); err != nil {

		err = spinnaker.UnwrapFrontFiftyBadResponse(err)

		// Let the user know via describe calls that something went wrong.
		r.Recorder.Eventf(
			&pipeline,
			"Warning",
			string(pacrdv1alpha1.PipelineCreationFailed),
			err.Error(),
		)
		r.EventClient.SendError(string(pacrdv1alpha1.PipelineCreationFailed), err)
		return err
	}
	r.Recorder.Eventf(
		&pipeline,
		"Normal",
		string(pacrdv1alpha1.PipelineUpdated),
		"Pipeline successfully created in Spinnaker.",
	)

	// Do we need this information?
	r.EventClient.SendEvent("Pipeline" + string(pacrdv1alpha1.PipelineUpdated), pipelineToMap(pipeline))
	
	return nil
}

func (r *PipelineReconciler) updatePipeline(ctx context.Context, pipeline pacrdv1alpha1.Pipeline, upstreamPipeline plank.Pipeline) error {

	r.updatePipelineStatus(pipeline, upstreamPipeline, ctx)

	plankPipe, err := pipeline.ToSpinnakerPipeline()

	if err != nil {
		r.Recorder.Eventf(
			&pipeline,
			"Warning",
			string(pacrdv1alpha1.PipelineUpdateFailed),
			err.Error(),
		)
		r.EventClient.SendError(string(pacrdv1alpha1.PipelineUpdateFailed), err)
		return err

	}

	// Log pipeline stages
	r.EventClient.SendPipelineStages(plankPipe)

	if err := r.SpinnakerClient.UpsertPipeline(plankPipe, upstreamPipeline.ID); err != nil {
		err = spinnaker.UnwrapFrontFiftyBadResponse(err)

		r.Recorder.Eventf(
			&pipeline,
			"Warning",
			string(pacrdv1alpha1.PipelineUpdateFailed),
			err.Error(),
		)
		r.EventClient.SendError(string(pacrdv1alpha1.PipelineUpdateFailed), err)
		return err
	}

	r.Recorder.Eventf(
		&pipeline,
		"Normal",
		string(pacrdv1alpha1.PipelineUpdated),
		"Pipeline successfully updated in Spinnaker.",
	)

	// Do we need this information?
	r.EventClient.SendEvent("Pipeline" + string(pacrdv1alpha1.PipelineUpdated), pipelineToMap(pipeline))
	
	return nil
}

func (r *PipelineReconciler) updatePipelineStatus(pipeline pacrdv1alpha1.Pipeline, upstreamPipeline plank.Pipeline, ctx context.Context) {
	// Persist pipeline id and url if not set already
	if pipeline.Status.ID == "" || pipeline.Status.URL == "" {
		pipeline.Status.ID = upstreamPipeline.ID
		pipeline.Status.URL = fmt.Sprintf("http://localhost:9000/#/applications/%s/executions/configure/%s", pipeline.Spec.Application, pipeline.Status.ID)

		_ = r.Status().Update(ctx, &pipeline)
	}
}

func (r *PipelineReconciler) deletePipeline(pipeline pacrdv1alpha1.Pipeline) error {
	// Determine if we're actively deleting the CRD and make sure we clean up
	// any Spinnaker resources that are related. Finalizers are initialized
	// further down the method.
	finalizers := pipeline.GetObjectMeta().GetFinalizers()

	if containsString(finalizers, PipelineFinalizerName) {
		plankPipe, err := pipeline.ToSpinnakerPipeline()
		if err != nil {
			return err
		}
		if err := r.SpinnakerClient.DeletePipeline(plankPipe); err != nil {
			var ue *plank.ErrUnsupportedStatusCode
			if errors.As(err, &ue) && ue.Code == 404 {
				// This is an acceptable error, do nothing here and continue on to
				// removing finalizers for this crd.
			} else {
				return err
			}
		}

		// Make sure we clean up the finalizer so k8s knows we're done with
		// this resource.
		pipeline.SetFinalizers(removeString(finalizers, PipelineFinalizerName))
		if err := r.Update(context.Background(), &pipeline); err != nil {
			return err
		}
	}

	return nil
}

func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pacrdv1alpha1.Pipeline{}).
		Complete(r)
}

func findPipeline(pipelines []plank.Pipeline, name string, id string) (plank.Pipeline, error) {
	for _, pipeline := range pipelines {
		if pipeline.ID == id || pipeline.Name == name {
			return pipeline, nil
		}
	}
	return plank.Pipeline{}, fmt.Errorf("could not find pipeline in list of returned pipelines")
}

func pipelineToMap( pipeline pacrdv1alpha1.Pipeline) map[string]interface{} {
	pipelineMap := make(map[string]interface{})
	err := mapstructure.Decode(pipeline, &pipelineMap)
	if err != nil {
		return nil
	}
	return pipelineMap
}