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

	"github.com/go-logr/logr"
	clusterV1alpha1 "github.com/tmax-cloud/hypercloud-multi-operator/apis/cluster/v1alpha1"
	util "github.com/tmax-cloud/hypercloud-multi-operator/controllers/util"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ClusterRegistrationReconciler reconciles a ClusterRegistration object
type ClusterRegistrationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cluster.tmax.io,resources=clusterregistrations,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=cluster.tmax.io,resources=clusterregistrations/status,verbs=get;patch;update

func (r *ClusterRegistrationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	_ = context.Background()
	log := r.Log.WithValues("clusterregistration", req.NamespacedName)

	// get ClusterRegistration
	clusterRegistration := &clusterV1alpha1.ClusterRegistration{}
	if err := r.Get(context.TODO(), req.NamespacedName, clusterRegistration); errors.IsNotFound(err) {
		log.Info("ClusterRegistration not found. Ignoring since object must be deleted")
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get ClusterRegistration")
		return ctrl.Result{}, err
	}

	patchHelper, err := patch.NewHelper(clusterRegistration, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	// status.phase migration for old version
	if clusterRegistration.Status.Phase == clusterV1alpha1.ClusterRegistrationDeprecatedPhaseSuccess {
		clusterRegistration.Status.Phase = clusterV1alpha1.ClusterRegistrationPhaseRegistered
	} else if clusterRegistration.Status.Phase == clusterV1alpha1.ClusterRegistrationDeprecatedPhaseDeleted {
		clusterRegistration.Status.Phase = clusterV1alpha1.ClusterRegistrationPhaseClusterDeleted
	}

	defer func() {
		// Always reconcile the Status.Phase field.
		r.reconcilePhase(context.TODO(), clusterRegistration)

		if err := patchHelper.Patch(context.TODO(), clusterRegistration); err != nil {
			// if err := patchClusterRegistration(context.TODO(), patchHelper, ClusterRegistration, patchOpts...); err != nil {
			// reterr = kerrors.NewAggregate([]error{reterr, err})
			reterr = err
		}
	}()

	// Handle normal reconciliation loop.
	return r.reconcile(context.TODO(), clusterRegistration)
}

// reconcile handles cluster reconciliation.
func (r *ClusterRegistrationReconciler) reconcile(ctx context.Context, ClusterRegistration *clusterV1alpha1.ClusterRegistration) (ctrl.Result, error) {
	phases := []func(context.Context, *clusterV1alpha1.ClusterRegistration) (ctrl.Result, error){
		// cluster 등록전, validation 을 체크하는 과정으로
		// single cluster 의 kube-config 가 올바른지 체크하기 위해, kube-config 를 사용해 node 들을 가져올수있는지 확인한다.
		// 또한, 중복성 체크를 위해 해당 name 과 namespace 를 가지는 cluster manager 가 이미 있는지 확인한다.
		r.CheckValidation,
		// 해당 cluster 에 대한 cluster manager 를 생성한다.
		r.CreateClusterManager,
		// kube-config 를 secret 으로 생성한다.
		r.CreateKubeconfigSecret,
	}

	res := ctrl.Result{}
	errs := []error{}
	for _, phase := range phases {
		// Call the inner reconciliation methods.
		phaseResult, err := phase(ctx, ClusterRegistration)
		if err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			continue
		}

		// Aggregate phases which requeued without err
		res = util.LowestNonZeroResult(res, phaseResult)
	}

	return res, kerrors.NewAggregate(errs)
}

func (r *ClusterRegistrationReconciler) reconcilePhase(_ context.Context, ClusterRegistration *clusterV1alpha1.ClusterRegistration) {
	if ClusterRegistration.Status.ClusterValidated {
		ClusterRegistration.Status.SetTypedPhase(clusterV1alpha1.ClusterRegistrationPhaseRegistered)
	}
}

func (r *ClusterRegistrationReconciler) requeueClusterRegistrationsForClusterManager(o client.Object) []ctrl.Request {
	clm := o.DeepCopyObject().(*clusterV1alpha1.ClusterManager)
	log := r.Log.WithValues("ClusterRegistration-ObjectMapper", "clusterManagerToClusterClusterRegistrations", "ClusterRegistration", clm.Name)

	//get clusterRegistration
	key := types.NamespacedName{
		Name:      clm.Labels[clusterV1alpha1.LabelKeyClrName],
		Namespace: clm.Namespace,
	}
	clr := &clusterV1alpha1.ClusterRegistration{}
	if err := r.Get(context.TODO(), key, clr); errors.IsNotFound(err) {
		log.Info("ClusterRegistration resource not found. Ignoring since object must be deleted")
		return nil
	} else if err != nil {
		log.Error(err, "Failed to get ClusterRegistration")
		return nil
	}

	if clr.Status.Phase != clusterV1alpha1.ClusterRegistrationPhaseRegistered {
		log.Info("ClusterRegistration for ClusterManager [" + clr.Spec.ClusterName + "] is already delete... Do not update cluster registration status to delete ")
		return nil
	}

	clr.Status.Phase = clusterV1alpha1.ClusterRegistrationPhaseClusterDeleted
	clr.Status.Reason = "cluster is deleted"
	if err := r.Status().Update(context.TODO(), clr); err != nil {
		log.Error(err, "Failed to update ClusterRegistration status")
		return nil //??
	}
	return nil
}

func (r *ClusterRegistrationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	controller, err := ctrl.NewControllerManagedBy(mgr).
		For(&clusterV1alpha1.ClusterRegistration{}).
		WithEventFilter(
			predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					// phase success 일 때 한번 들어오는데.. 왜 그러냐... controller 재기동 돼서? by 조상원
					clr := e.Object.(*clusterV1alpha1.ClusterRegistration)
					if clr.Status.Phase == "" {
						return true
					} else {
						return false
					}
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldClr := e.ObjectOld.(*clusterV1alpha1.ClusterRegistration)
					newClr := e.ObjectNew.(*clusterV1alpha1.ClusterRegistration)
					if oldClr.DeletionTimestamp.IsZero() && !newClr.DeletionTimestamp.IsZero() {
						return true
					}
					// if oldClr.Status.Phase == "Success" && newClr.Status.Phase == "Validated" {
					// 	return true
					// }
					return false
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return false
				},
				GenericFunc: func(e event.GenericEvent) bool {
					return false
				},
			},
		).
		Build(r)

	if err != nil {
		return err
	}

	return controller.Watch(
		&source.Kind{Type: &clusterV1alpha1.ClusterManager{}},
		handler.EnqueueRequestsFromMapFunc(r.requeueClusterRegistrationsForClusterManager),
		predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				return false
			},
			CreateFunc: func(e event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				clm := e.Object.(*clusterV1alpha1.ClusterManager)
				val, ok := clm.Labels[clusterV1alpha1.LabelKeyClmClusterType]
				if ok && val == clusterV1alpha1.ClusterTypeRegistered {
					return true
				}
				return false
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return false
			},
		},
	)
}
