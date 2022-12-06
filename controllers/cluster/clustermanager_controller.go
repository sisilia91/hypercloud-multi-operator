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
	certmanagerV1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	servicecatalogv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	clusterV1alpha1 "github.com/tmax-cloud/hypercloud-multi-operator/apis/cluster/v1alpha1"
	util "github.com/tmax-cloud/hypercloud-multi-operator/controllers/util"
	traefikV1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"

	coreV1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	capiV1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ClusterManagerReconciler reconciles a ClusterManager object
type ClusterManagerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cluster.tmax.io,resources=clustermanagers,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=cluster.tmax.io,resources=clustermanagers/status,verbs=get;list;patch;update;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinedeployments,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinedeployments/status,verbs=get;list;patch;update;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines,verbs=get;list;patch;update;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines/status,verbs=get;list;patch;update;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings;roles;rolebindings,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=kubeadmcontrolplanes,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=kubeadmcontrolplanes/status,verbs=get;list;patch;update;watch
// +kubebuilder:rbac:groups=servicecatalog.k8s.io,resources=serviceinstances,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=servicecatalog.k8s.io,resources=serviceinstances/status,verbs=get;list;patch;update;watch
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups="",resources=services;endpoints,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=traefik.containo.us,resources=middlewares,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=create;delete;get;list;patch;update;watch

func (r *ClusterManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	_ = context.Background()
	log := r.Log.WithValues("clustermanager", req.NamespacedName)

	//get ClusterManager
	clusterManager := &clusterV1alpha1.ClusterManager{}
	if err := r.Get(context.TODO(), req.NamespacedName, clusterManager); errors.IsNotFound(err) {
		log.Info("ClusterManager resource not found. Ignoring since object must be deleted")
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get ClusterManager")
		return ctrl.Result{}, err
	}

	//set patch helper
	patchHelper, err := patch.NewHelper(clusterManager, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always reconcile the Status.Phase field.
		r.reconcilePhase(context.TODO(), clusterManager)

		if err := patchHelper.Patch(context.TODO(), clusterManager); err != nil {
			// if err := patchClusterManager(context.TODO(), patchHelper, clusterManager, patchOpts...); err != nil {
			// reterr = kerrors.NewAggregate([]error{reterr, err})
			reterr = err
		}
	}()

	// Add finalizer first if not exist to avoid the race condition between init and delete
	if !controllerutil.ContainsFinalizer(clusterManager, clusterV1alpha1.ClusterManagerFinalizer) {
		controllerutil.AddFinalizer(clusterManager, clusterV1alpha1.ClusterManagerFinalizer)
		return ctrl.Result{}, nil
	}

	// Handle deletion reconciliation loop.
	if !clusterManager.ObjectMeta.DeletionTimestamp.IsZero() {
		clusterManager.Status.Ready = false
		return r.reconcileDelete(context.TODO(), clusterManager)
	}

	// Handle normal reconciliation loop.
	return r.reconcile(context.TODO(), clusterManager)
}

// func (r *ClusterManagerReconciler) reconcileDeleteForRegisteredClusterManager(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (reconcile.Result, error) {
// 	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
// 	log.Info("Start to reconcile delete for registered ClusterManager")

// 	// cluster member crb 는 db 에 저장되어있는 member 의 정보로 삭제 되어지기 때문에, crb 를 지운후 db 에서 삭제 해야한다.
// 	// if err := util.Delete(clusterManager.Namespace, clusterManager.Name); err != nil {
// 	// 	log.Error(err, "Failed to delete cluster info from cluster_member table")
// 	// 	return ctrl.Result{}, err
// 	// }

// 	key := types.NamespacedName{
// 		Name:      clusterManager.Name + util.KubeconfigSuffix,
// 		Namespace: clusterManager.Namespace,
// 	}
// 	kubeconfigSecret := &coreV1.Secret{}
// 	if err := r.Get(context.TODO(), key, kubeconfigSecret); err != nil && !errors.IsNotFound(err) {
// 		log.Error(err, "Failed to get kubeconfig Secret")
// 		return ctrl.Result{}, err
// 	} else if err == nil {
// 		if err := r.Delete(context.TODO(), kubeconfigSecret); err != nil {
// 			log.Error(err, "Failed to delete kubeconfig Secret")
// 			return ctrl.Result{}, err
// 		}

// 		log.Info("Delete kubeconfig Secret successfully")
// 	}

// 	// ArgoCD application이 모두 삭제되었는지 테스트
// 	if err := r.CheckApplicationRemains(clusterManager); err != nil {
// 		return ctrl.Result{}, err
// 	}

// 	if err := r.DeleteTraefikResources(clusterManager); err != nil {
// 		return ctrl.Result{}, err
// 	}

// 	if err := r.DeleteHyperAuthResourcesForSingleCluster(clusterManager); err != nil {
// 		return ctrl.Result{}, err
// 	}

// 	controllerutil.RemoveFinalizer(clusterManager, clusterV1alpha1.ClusterManagerFinalizer)
// 	return ctrl.Result{}, nil
// }

// reconcile handles cluster reconciliation.
func (r *ClusterManagerReconciler) reconcile(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (ctrl.Result, error) {
	phases := []func(context.Context, *clusterV1alpha1.ClusterManager) (ctrl.Result, error){}
	phases = append(phases, r.ReadyReconcilePhase)

	if clusterManager.Labels[clusterV1alpha1.LabelKeyClmClusterType] == clusterV1alpha1.ClusterTypeCreated {
		// cluster claim 으로 cluster 를 생성한 경우에만 수행
		phases = append(
			phases,
			// cluster manager 의  metadata 와 provider 정보를 service instance 의 parameter 값에 넣어 service instance 를 생성한다.
			r.CreateServiceInstance,
			// cluster manager 가 바라봐야 할 cluster 의 endpoint 를 annotation 으로 달아준다.
			r.SetEndpoint,
			// cluster claim 을 통해, cluster 의 spec 을 변경한 경우, 그에 맞게 master 노드의 spec 을 업데이트 해준다.
			r.kubeadmControlPlaneUpdate,
			// cluster claim 을 통해, cluster 의 spec 을 변경한 경우, 그에 맞게 worker 노드의 spec 을 업데이트 해준다.
			r.machineDeploymentUpdate,
		)
	} else {
		// cluster 를 등록한 경우에만 수행
		// cluster registration 으로 cluster 를 등록한 경우에는, kubeadm-config 를 가져와 그에 맞게 cluster manager 의 spec 을 업데이트 해야한다.
		// kubeadm-config 에 따라,
		// cluster manager 에 k8s version을 업데이트 해주고,
		// single cluster 의 nodes 를 가져와 ready 상태의 worker node 와 master node의 개수를 업데이트해준다.
		// 또한, 해당 cluster 의 provider 이름 (Aws/Vsphere) 을 업데이트 해주는 과정을 진행한다.
		phases = append(phases, r.UpdateClusterManagerStatus)
	}
	// 공통적으로 수행
	phases = append(
		phases,
		// Argocd 연동을 위해 필요한 정보를 kube-config 로 부터 가져와 secret을 생성한다.
		r.CreateArgocdResources,
		// single cluster 의 api gateway service 의 주소로 gateway service 생성
		r.CreateGatewayResources,
		// Kibana, Grafana, Kiali 등 모듈과 HyperAuth oidc 연동을 위한 resource 생성 작업 (HyperAuth 계정정보로 여러 모듈에 로그인 가능)
		// HyperAuth caller 를 통해 admin token 을 가져와 각 모듈 마다 HyperAuth client 를 생성후, 모듈에 따른 resource들을 추가한다.
		// HyperRegistry를 위한 admin group 또한 생성해준다.
		r.CreateHyperAuthResources,
		// // hyperregistry domain 을 single cluster 의 ingress 로 부터 가져와 oidc 연동설정
		// r.SetHyperregistryOidcConfig,
		// Traefik 을 통하기 위한 리소스인 certificate, ingress, middleware를 생성한다.
		// 콘솔에서 ingress를 조회하여 LNB에 cluster를 listing 해주므로 cluster가 완전히 join되고 나서
		// LNB에 리스팅 될 수 있게 해당 프로세스를 가장 마지막에 수행한다.
		r.CreateTraefikResources,
	)

	// special case- capi upgrade/master scaling/worker scaling
	if clusterManager.Labels[clusterV1alpha1.LabelKeyClmClusterType] == clusterV1alpha1.ClusterTypeCreated {
		if clusterManager.Status.Version != "" && clusterManager.Spec.Version != clusterManager.Status.Version {
			phases = []func(context.Context, *clusterV1alpha1.ClusterManager) (ctrl.Result, error){}
			if clusterManager.Spec.Provider == clusterV1alpha1.ProviderVSphere {
				phases = append(phases, r.CreateUpgradeServiceInstance)
			}
			phases = append(phases, r.ClusterUpgrade)
		} else if clusterManager.Status.MasterNum != 0 && clusterManager.Spec.MasterNum != clusterManager.Status.MasterNum {
			phases = []func(context.Context, *clusterV1alpha1.ClusterManager) (ctrl.Result, error){r.ControlplaneScaling}
		} else if clusterManager.Status.WorkerNum != 0 && clusterManager.Spec.WorkerNum != clusterManager.Status.WorkerNum {
			phases = []func(context.Context, *clusterV1alpha1.ClusterManager) (ctrl.Result, error){r.WorkerScaling}
		}
	}

	res := ctrl.Result{}
	errs := []error{}
	// phases 를 돌면서, append 한 함수들을 순차적으로 수행하고,
	// error가 있는지 체크하여 error가 있으면 무조건 requeue
	// 이때는 가장 최초로 error가 발생한 phase의 requeue after time을 따라감
	// 모든 error를 최종적으로 aggregate하여 반환할 수 있도록 리스트로 반환
	// error는 없지만 다시 requeue 가 되어야 하는 phase들이 존재하는 경우
	// LowestNonZeroResult 함수를 통해 requeueAfter time 이 가장 짧은 함수를 찾는다.
	for _, phase := range phases {
		// Call the inner reconciliation methods.
		phaseResult, err := phase(ctx, clusterManager)
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

func (r *ClusterManagerReconciler) reconcileDelete(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (reconcile.Result, error) {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	log.Info("Start reconcile phase for delete")

	// sjoh - kubeconfig가 없으면 cluster가 삭제된 것을 가정, 없으면 skip한다.
	// key := types.NamespacedName{
	// 	Name:      clusterManager.Name + util.KubeconfigSuffix,
	// 	Namespace: clusterManager.Namespace,
	// }
	// if err := r.Get(context.TODO(), key, kubeconfigSecret); err != nil && !errors.IsNotFound(err) {
	// 	log.Error(err, "Failed to get kubeconfig secret")
	// 	return ctrl.Result{}, err
	// }

	// ArgoCD application이 모두 삭제되었는지 테스트
	if err := r.CheckApplicationRemains(clusterManager); err != nil {
		return ctrl.Result{RequeueAfter: requeueAfter10Second}, err
	}

	// ClusterAPI-provider-aws의 경우, lb type의 svc가 남아있으면 infra nlb deletion이 stuck걸리면서 클러스터가 지워지지 않는 버그가 있음
	// 이를 해결하기 위해 클러스터를 삭제하기 전에 lb type의 svc를 전체 삭제한 후 클러스터를 삭제
	if err := r.DeleteLoadBalancerServices(clusterManager); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.DeleteTraefikResources(clusterManager); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.DeleteHyperAuthResourcesForSingleCluster(clusterManager); err != nil {
		return ctrl.Result{}, err
	}

	// delete serviceinstance
	key := types.NamespacedName{
		Name:      clusterManager.Name + "-" + clusterManager.Annotations[clusterV1alpha1.AnnotationKeyClmSuffix],
		Namespace: clusterManager.Namespace,
	}
	serviceInstance := &servicecatalogv1beta1.ServiceInstance{}
	if err := r.Get(context.TODO(), key, serviceInstance); errors.IsNotFound(err) {
		log.Info("ServiceInstance is already deleted. Waiting cluster to be deleted")
	} else if err != nil {
		log.Error(err, "Failed to get serviceInstance")
		return ctrl.Result{}, err
	} else {
		if err := r.Delete(context.TODO(), serviceInstance); err != nil {
			log.Error(err, "Failed to delete serviceInstance")
			return ctrl.Result{}, err
		}
	}

	// capi가 생성한 kubeconfig는 service instance를 지우면서 삭제되었으므로, registration으로 생성한 경우 또한 kubeconfig를 이 시점에서 삭제한다.
	// kubeconfig가 없으면 skip 한다.
	if clusterManager.Labels[clusterV1alpha1.LabelKeyClmClusterType] == clusterV1alpha1.ClusterTypeRegistered {
		key := types.NamespacedName{
			Name:      clusterManager.Name + util.KubeconfigSuffix,
			Namespace: clusterManager.Namespace,
		}
		regKubeconfigSecret := &coreV1.Secret{}
		if err := r.Get(context.TODO(), key, regKubeconfigSecret); errors.IsNotFound(err) {
			log.Info("Kubeconfig secret for cluster registration was deleted successfully")
		} else if err != nil {
			log.Error(err, "Failed to get kubeconfig secret for cluster registration")
			return ctrl.Result{}, err
		} else {
			if err := r.Delete(context.TODO(), regKubeconfigSecret); err != nil {
				log.Error(err, "Failed to delete kubeconfig secret for cluster registration")
				return ctrl.Result{}, err
			}
			log.Info("Kubeconfig secret for cluster registration is deleting")
		}
	}

	//delete handling
	key = clusterManager.GetNamespacedName()
	if err := r.Get(context.TODO(), key, &capiV1alpha3.Cluster{}); errors.IsNotFound(err) {
		if err := util.Delete(clusterManager.Namespace, clusterManager.Name); err != nil {
			log.Error(err, "Failed to delete cluster info from cluster_member table")
			return ctrl.Result{}, err
		}
		// kubeconfig secret이 없다면(모든 시크릿이 삭제되었다면) clm을 삭제한다.
		key = types.NamespacedName{
			Name:      clusterManager.Name + util.KubeconfigSuffix,
			Namespace: clusterManager.Namespace,
		}
		if err := r.Get(context.TODO(), key, &coreV1.Secret{}); errors.IsNotFound(err) {
			controllerutil.RemoveFinalizer(clusterManager, clusterV1alpha1.ClusterManagerFinalizer)
			log.Info("Cluster manager was deleted successfully")
			return ctrl.Result{}, nil
		} else if err != nil {
			log.Error(err, "Failed to get kubeconfig secret")
			return ctrl.Result{}, err
		}

	} else if err != nil {
		log.Error(err, "Failed to get cluster")
		return ctrl.Result{}, err
	}

	log.Info("Cluster is deleting. Requeue after 1min")
	return ctrl.Result{RequeueAfter: requeueAfter1Minute}, nil
}

func (r *ClusterManagerReconciler) reconcilePhase(_ context.Context, clusterManager *clusterV1alpha1.ClusterManager) {

	if !clusterManager.DeletionTimestamp.IsZero() {
		clusterManager.Status.SetTypedPhase(clusterV1alpha1.ClusterManagerPhaseDeleting)
		return
	}

	if clusterManager.Status.Phase == "" {
		clusterManager.Status.SetTypedPhase(clusterV1alpha1.ClusterManagerPhaseProcessing)
	}

	if clusterManager.Status.ArgoReady {
		clusterManager.Status.SetTypedPhase(clusterV1alpha1.ClusterManagerPhaseSyncNeeded)
	}

	if clusterManager.Status.GatewayReady {
		clusterManager.Status.SetTypedPhase(clusterV1alpha1.ClusterManagerPhaseProcessing)
	}

	if clusterManager.Status.TraefikReady {
		clusterManager.Status.SetTypedPhase(clusterV1alpha1.ClusterManagerPhaseReady)
	}

	// cluster scaling
	if (clusterManager.Status.MasterNum != 0 && clusterManager.Spec.MasterNum != clusterManager.Status.MasterNum) ||
		(clusterManager.Status.WorkerNum != 0 && clusterManager.Spec.WorkerNum != clusterManager.Status.WorkerNum) {
		clusterManager.Status.SetTypedPhase(clusterV1alpha1.ClusterManagerPhaseScaling)
		return
	}

	// cluster upgrading
	if clusterManager.Status.Version != "" && clusterManager.Status.Version != clusterManager.Spec.Version {
		clusterManager.Status.SetTypedPhase(clusterV1alpha1.ClusterManagerPhaseUpgrading)
		return
	}
}

func (r *ClusterManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	controller, err := ctrl.NewControllerManagedBy(mgr).
		For(&clusterV1alpha1.ClusterManager{}).
		WithEventFilter(
			predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					return true
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					// created clm은 update 필요가 있지만 registered는 clm update가 필요 없다
					// 다만 registered인 경우 deletionTimestamp가 있는경우 delete 수행을 위해 reconcile을 수행하긴 해야한다.
					oldclm := e.ObjectOld.(*clusterV1alpha1.ClusterManager)
					newclm := e.ObjectNew.(*clusterV1alpha1.ClusterManager)

					isFinalized := !controllerutil.ContainsFinalizer(oldclm, clusterV1alpha1.ClusterManagerFinalizer) &&
						controllerutil.ContainsFinalizer(newclm, clusterV1alpha1.ClusterManagerFinalizer)
					isDelete := oldclm.DeletionTimestamp.IsZero() &&
						!newclm.DeletionTimestamp.IsZero()
					isControlPlaneEndpointUpdate := oldclm.Status.ControlPlaneEndpoint == "" &&
						newclm.Status.ControlPlaneEndpoint != ""
					isSubResourceNotReady := !newclm.Status.ArgoReady || !newclm.Status.TraefikReady || !newclm.Status.GatewayReady

					isUpgrade := oldclm.Spec.Version != "" && oldclm.Spec.Version != newclm.Spec.Version
					isScaling := oldclm.Spec.MasterNum != newclm.Spec.MasterNum ||
						oldclm.Spec.WorkerNum != newclm.Spec.WorkerNum
					if isDelete || isControlPlaneEndpointUpdate || isFinalized || isUpgrade || isScaling {
						return true
					} else {
						if newclm.Labels[clusterV1alpha1.LabelKeyClmClusterType] == clusterV1alpha1.ClusterTypeCreated {
							return true
						}
						if newclm.Labels[clusterV1alpha1.LabelKeyClmClusterType] == clusterV1alpha1.ClusterTypeRegistered {
							return isSubResourceNotReady
						}
					}
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

	controller.Watch(
		&source.Kind{Type: &capiV1alpha3.Cluster{}},
		handler.EnqueueRequestsFromMapFunc(r.requeueClusterManagersForCluster),
		predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldc := e.ObjectOld.(*capiV1alpha3.Cluster)
				newc := e.ObjectNew.(*capiV1alpha3.Cluster)

				return !oldc.Status.ControlPlaneInitialized && newc.Status.ControlPlaneInitialized
			},
			CreateFunc: func(e event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return true
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return false
			},
		},
	)

	controller.Watch(
		&source.Kind{Type: &controlplanev1.KubeadmControlPlane{}},
		handler.EnqueueRequestsFromMapFunc(r.requeueClusterManagersForKubeadmControlPlane),
		predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldKcp := e.ObjectOld.(*controlplanev1.KubeadmControlPlane)
				newKcp := e.ObjectNew.(*controlplanev1.KubeadmControlPlane)

				return oldKcp.Status.Replicas != newKcp.Status.Replicas
			},
			CreateFunc: func(e event.CreateEvent) bool {
				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return false
			},
		},
	)

	controller.Watch(
		&source.Kind{Type: &capiV1alpha3.MachineDeployment{}},
		handler.EnqueueRequestsFromMapFunc(r.requeueClusterManagersForMachineDeployment),
		predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldMd := e.ObjectOld.(*capiV1alpha3.MachineDeployment)
				newMd := e.ObjectNew.(*capiV1alpha3.MachineDeployment)

				return oldMd.Status.Replicas != newMd.Status.Replicas
			},
			CreateFunc: func(e event.CreateEvent) bool {
				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return false
			},
		},
	)

	subResources := []client.Object{
		&certmanagerV1.Certificate{},
		&networkingv1.Ingress{},
		&coreV1.Service{},
		&traefikV1alpha1.Middleware{},
	}
	for _, resource := range subResources {
		controller.Watch(
			&source.Kind{Type: resource},
			handler.EnqueueRequestsFromMapFunc(r.requeueClusterManagersForSubresources),
			predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					return false
				},
				CreateFunc: func(e event.CreateEvent) bool {
					return false
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					_, ok := e.Object.GetLabels()[clusterV1alpha1.LabelKeyClmName]
					return ok
				},
				GenericFunc: func(e event.GenericEvent) bool {
					return false
				},
			},
		)
	}

	return nil
}
