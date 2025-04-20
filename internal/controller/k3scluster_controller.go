/*
Copyright 2025.

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

package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/toolkit/reconciler"
	step_result "github.com/kloudlite/operator/toolkit/reconciler/step-result"
	v1 "github.com/kloudlite/plugin-k3s-cluster/api/v1"
	"github.com/kloudlite/plugin-k3s-cluster/internal/controller/templates"
	"github.com/kloudlite/plugin-k3s-cluster/internal/env"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// K3sClusterReconciler reconciles a K3sCluster object
type K3sClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Env    *env.Env

	ClusterLifecycleSpecTemplate []byte
}

func (c *K3sClusterReconciler) GetName() string {
	return "k3s-cluster"
}

const (
	createClusterJob = "create-cluster-job"
)

// +kubebuilder:rbac:groups=plugin-k3s-cluster.kloudlite.github.com,resources=k3sclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=plugin-k3s-cluster.kloudlite.github.com,resources=k3sclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=plugin-k3s-cluster.kloudlite.github.com,resources=k3sclusters/finalizers,verbs=update

func (r *K3sClusterReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &v1.K3sCluster{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList([]reconciler.CheckMeta{
		{Name: createClusterJob, Title: "Creates Cluster Installation Job"},
	}); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(reconciler.ForegroundFinalizer, reconciler.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createClusterLifecycleJob(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *K3sClusterReconciler) finalize(req *reconciler.Request[*v1.K3sCluster]) step_result.Result {
	check := reconciler.NewRunningCheck("finalizing", req)

	if step := req.EnsureCheckList([]reconciler.CheckMeta{
		{Name: "finalizing", Title: "Cleanup Owned Resources"},
	}); !step.ShouldProceed() {
		return step
	}

	if result := req.CleanupOwnedResources(check); !result.ShouldProceed() {
		return result
	}

	return req.Finalize()
}

func (r *K3sClusterReconciler) parseSpecToVarFileJson(req *reconciler.Request[*v1.K3sCluster]) ([]byte, error) {
	obj := req.Object
	switch obj.Spec.CloudProvider {
	case v1.CloudProviderAWS:
		{
			if obj.Spec.AWS == nil {
				return nil, fmt.Errorf(".spec.aws must be set when cloud provider is %s", obj.Spec.CloudProvider)
			}

			return json.Marshal(map[string]any{
				"cluster_name":  obj.Name,
				"aws_region":    "",
				"vpc_id":        obj.Spec.AWS.VPC.ID,
				"cluster_state": obj.Spec.ClusterState,
				"master_nodes": func() []map[string]any {
					nodes := make([]map[string]any, 0, len(obj.Spec.AWS.MasterNodes))
					for _, node := range obj.Spec.AWS.MasterNodes {
						nodes = append(nodes, map[string]any{
							"name":              node.Name,
							"ami":               node.AMI,
							"instance_type":     node.InstanceType,
							"availability_zone": node.AvailabilityZone,
							"vpc_subnet_id":     obj.Spec.AWS.VPC.ID,
							"root_volume_size":  node.RootVolumeSize,
							"root_volume_type":  node.RootVolumeType,
						})
					}
					return nodes
				}(),
				"kloudlite_release": "v1.1.6-nightly",
				"base_domain":       "cluster@kloudlite.io",
			})
		}
	default:
		{
			return nil, fmt.Errorf("cloud provider not supported")
		}
	}
}

func (r *K3sClusterReconciler) createClusterLifecycleJob(req *reconciler.Request[*v1.K3sCluster]) step_result.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(createClusterJob, req)

	valuesJSON, err := r.parseSpecToVarFileJson(req)
	if err != nil {
		return check.Failed(err)
	}

	b, err := templates.ParseBytes(r.ClusterLifecycleSpecTemplate, templates.ClusterLifecycleSpecTemplateArgs{
		Tolerations:          []corev1.Toleration{},
		NodeSelector:         map[string]string{},
		JobImage:             r.Env.IACJobImage,
		CloudProvider:        obj.Spec.CloudProvider.String(),
		TFWorkspaceName:      obj.Name,
		TFWorkspaceNamespace: obj.Namespace,
		ValuesJSON:           string(valuesJSON),
	})
	if err != nil {
		return check.Failed(err)
	}

	lf := &crdsv1.Lifecycle{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}
	controllerutil.CreateOrUpdate(ctx, r.Client, lf, func() error {
		return yaml.Unmarshal(b, &lf.Spec)
	})

	if lf.HasCompleted() {
		switch lf.Status.Phase {
		case crdsv1.JobPhaseFailed:
			return check.Failed(fmt.Errorf("job failed"))
		case crdsv1.JobPhaseSucceeded:
			return check.Completed()
		}
	}

	return check.StillRunning(fmt.Errorf("waiting for job to complete"))
}

// SetupWithManager sets up the controller with the Manager.
func (r *K3sClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}

	if r.Scheme == nil {
		r.Scheme = mgr.GetScheme()
	}

	var err error
	r.ClusterLifecycleSpecTemplate, err = templates.Read(templates.ClusterLifeycleSpec)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.K3sCluster{}).Named(r.GetName())
	builder.Owns(&batchv1.Job{})
	builder.Owns(&corev1.Secret{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor(r.GetName())))

	return builder.Complete(r)
}
