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
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"
	clusterV1alpha1 "github.com/tmax-cloud/hypercloud-multi-operator/apis/cluster/v1alpha1"
	"github.com/tmax-cloud/hypercloud-multi-operator/controllers/util"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

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

// ClusterReconciler reconciles a Memcached object
type SecretReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Cluster member information
type ClusterMemberInfo struct {
	Id          int64     `json:"Id"`
	Namespace   string    `json:"Namespace"`
	Cluster     string    `json:"Cluster"`
	MemberId    string    `json:"MemberId"`
	Groups      []string  `json:"Groups"`
	MemberName  string    `json:"MemberName"`
	Attribute   string    `json:"Attribute"`
	Role        string    `json:"Role"`
	Status      string    `json:"Status"`
	CreatedTime time.Time `json:"CreatedTime"`
	UpdatedTime time.Time `json:"UpdatedTime"`
}

// +kubebuilder:rbac:groups="",resources=secrets;namespaces;serviceaccounts,verbs=create;delete;get;list;patch;post;update;watch;

func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	_ = context.Background()
	secretName := req.Name
	isArgoSecret := strings.Contains(secretName, "cluster-")
	isSATokenSecret := strings.Contains(secretName, "-token")
	// cluster manager가 들어온 경우 처리
	if !isArgoSecret && !isSATokenSecret && !strings.Contains(secretName, util.KubeconfigSuffix) {
		secretName += util.KubeconfigSuffix
	}
	key := types.NamespacedName{
		Name:      secretName,
		Namespace: req.Namespace,
	}
	log := r.Log.WithValues("Secret", key)
	log.Info("Start to reconcile secret")

	//get secret
	secret := &coreV1.Secret{}
	if err := r.Get(context.TODO(), key, secret); errors.IsNotFound(err) {
		log.Info("Secret resource not found. Ignoring since object must be deleted")
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get secret")
		return ctrl.Result{}, err
	}

	//set patch helper
	patchHelper, err := patch.NewHelper(secret, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		if err := patchHelper.Patch(context.TODO(), secret); err != nil {
			reterr = err
		}
	}()

	// capi에 의해 생성된 kubeconfig secret에는 secret type에 대한 Label이 달려있지 않으므로, 필요한 Label을 달아준다.
	_, isCapiKubeconfig := secret.Labels[util.LabelKeyCapiClusterName]
	if _, ok := secret.Labels[util.LabelKeyClmSecretType]; !ok && isCapiKubeconfig {
		secret.Labels[util.LabelKeyClmSecretType] = util.ClmSecretTypeKubeconfig
		secret.Labels[clusterV1alpha1.LabelKeyClmName] = secret.Labels[util.LabelKeyCapiClusterName]
		secret.Labels[clusterV1alpha1.LabelKeyClmNamespace] = secret.Namespace
	}

	// sjoh
	// _, isCapiKubeconfig := secret.Labels[util.LabelKeyCapiClusterName]
	// Add finalizer first if not exist to avoid the race condition between init and delete
	// capi에 의해 생성된 kubeconfig secret은 capi controller가 처리할 수 있도록 finalizer를 달지 않는다.
	// if !controllerutil.ContainsFinalizer(secret, clusterV1alpha1.ClusterManagerFinalizer) && !isCapiKubeconfig {
	if !controllerutil.ContainsFinalizer(secret, clusterV1alpha1.ClusterManagerFinalizer) {
		controllerutil.AddFinalizer(secret, clusterV1alpha1.ClusterManagerFinalizer)
		return ctrl.Result{}, nil
	}

	// Handle deletion reconciliation loop.
	if !secret.GetDeletionTimestamp().IsZero() {
		return r.reconcileDelete(context.TODO(), secret)
	}

	return r.reconcile(context.TODO(), secret)
}

// reconcile handles cluster reconciliation.
func (r *SecretReconciler) reconcile(ctx context.Context, secret *coreV1.Secret) (ctrl.Result, error) {
	phases := []func(context.Context, *coreV1.Secret) (ctrl.Result, error){
		// cluster manager 가 바라봐야 할 single cluster 의 api-server 를 설정해주는 작업을 진행한다.
		// 해당 secret 으로 부터 kubeconfig data 를 가져와 kubeconfig 의 server 를 cluster manager 의 control plane endpoint 로 설정해준다.
		r.UpdateClusterManagerControlPlaneEndpoint,
		// single cluster 에 admin/developer/guest 에 따른 cluster role 을 생성하고,
		// cluster owner 에 대해 admin role 을 가지는 cluster rolebinding 을 생성한다.
		r.DeployRBACResources,
		// single cluster 에 Argocd 연동을 위한 리소스 배포작업을 진행한다.
		// Argocd 용 service account 를 생성하고,
		// cluster role 과 cluster rolebinding 을 생성한다.
		r.DeployArgocdResources,
		// r.DeployOpensearchResources,
	}

	res := ctrl.Result{}
	errs := []error{}
	for _, phase := range phases {
		// Call the inner reconciliation methods.
		phaseResult, err := phase(ctx, secret)
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

func (r *SecretReconciler) reconcileDelete(ctx context.Context, secret *coreV1.Secret) (reconcile.Result, error) {
	key := types.NamespacedName{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	}
	log := r.Log.WithValues("secret", key)
	log.Info("Start to reconcile delete")

	key = types.NamespacedName{
		Name:      secret.Labels[clusterV1alpha1.LabelKeyClmName],
		Namespace: secret.Labels[clusterV1alpha1.LabelKeyClmNamespace],
	}
	clm := &clusterV1alpha1.ClusterManager{}
	if err := r.Get(context.TODO(), key, clm); errors.IsNotFound(err) {
		log.Info("Not found cluster manager. Already deleted")
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get ClusterManager")
		return ctrl.Result{}, err
	}

	// cluster registration의 경우, clr status를 update한다.
	if clm.Labels[clusterV1alpha1.ClusterTypeCreated] == clusterV1alpha1.ClusterTypeRegistered {
		key = types.NamespacedName{
			Name:      clm.Labels[clusterV1alpha1.LabelKeyClrName],
			Namespace: secret.Labels[clusterV1alpha1.LabelKeyClmNamespace],
		}
		clr := &clusterV1alpha1.ClusterRegistration{}
		if err := r.Get(context.TODO(), key, clr); err != nil {
			log.Error(err, "Failed to get ClusterRegistration")
			return ctrl.Result{}, err
		}

		helper, _ := patch.NewHelper(clr, r.Client)
		defer func() {
			if err := helper.Patch(context.TODO(), clr); err != nil {
				r.Log.Error(err, "ClusterRegistration patch error")
			}
		}()
		clr.Status.Reason = "kubeconfig secret is deleted"
		clr.Status.ClusterValidated = false
		clr.Status.Ready = false
	}

	// sjoh
	// if !clm.GetDeletionTimestamp().IsZero() {
	// 	if secret.Labels[util.LabelKeyClmSecretType] == util.ClmSecretTypeArgo {
	// 		helper, _ := patch.NewHelper(clm, r.Client)
	// 		defer func() {
	// 			if err := helper.Patch(context.TODO(), clm); err != nil {
	// 				r.Log.Error(err, "ClusterManager patch error")
	// 			}
	// 		}()

	// 		clm.Status.ArgoReady = false
	// 		controllerutil.RemoveFinalizer(secret, clusterV1alpha1.ClusterManagerFinalizer)
	// 	} else if secret.Labels[util.LabelKeyClmSecretType] == util.ClmSecretTypeKubeconfig {
	// 		key = types.NamespacedName{
	// 			Name:      clm.Labels[clusterV1alpha1.LabelKeyClrName],
	// 			Namespace: secret.Labels[clusterV1alpha1.LabelKeyClmNamespace],
	// 		}
	// 		clr := &clusterV1alpha1.ClusterRegistration{}
	// 		if err := r.Get(context.TODO(), key, clr); err != nil {
	// 			log.Error(err, "Failed to get ClusterRegistration")
	// 			return ctrl.Result{}, err
	// 		}

	// 		helper, _ := patch.NewHelper(clr, r.Client)
	// 		defer func() {
	// 			if err := helper.Patch(context.TODO(), clr); err != nil {
	// 				r.Log.Error(err, "ClusterRegistration patch error")
	// 			}
	// 		}()

	// 		// clr.Status.Phase = "Validated"
	// 		// clr.Status.ClusterValidated = true
	// 		clr.Status.Reason = "kubeconfig secret is deleted"
	// 		controllerutil.RemoveFinalizer(secret, clusterV1alpha1.ClusterManagerFinalizer)
	// 	} else if secret.Labels[util.LabelKeyClmSecretType] == util.ClmSecretTypeSAToken {
	// 		helper, _ := patch.NewHelper(clm, r.Client)
	// 		defer func() {
	// 			if err := helper.Patch(context.TODO(), clm); err != nil {
	// 				r.Log.Error(err, "ClusterManager patch error")
	// 			}
	// 		}()

	// 		clm.Status.TraefikReady = false
	// 		controllerutil.RemoveFinalizer(secret, clusterV1alpha1.ClusterManagerFinalizer)
	// 	}

	// 	controllerutil.RemoveFinalizer(secret, clusterV1alpha1.ClusterManagerFinalizer)
	// 	return ctrl.Result{}, nil
	// }

	// kubeconfig secret에 대해서만 처리하도록 한다.
	// capi가 생성한 kubeconfig secret이 들어오는 경우, single cluster에는 접근할 수 없다.
	if secret.Labels[util.LabelKeyClmSecretType] == util.ClmSecretTypeArgo ||
		secret.Labels[util.LabelKeyClmSecretType] == util.ClmSecretTypeSAToken {
		controllerutil.RemoveFinalizer(secret, clusterV1alpha1.ClusterManagerFinalizer)
		return ctrl.Result{}, nil
	}

	re, _ := regexp.Compile("[" + regexp.QuoteMeta(`!#$%&'"*+-/=?^_{|}~().,:;<>[]\`) + "`\\s" + "]")
	email := clm.Annotations[util.AnnotationKeyOwner]
	adminServiceAccountName := re.ReplaceAllString(strings.Replace(email, "@", "-at-", -1), "-")

	// registration인 경우, single cluster 내부에 설치된 리소스를 삭제한다.
	// capi를 통해서 생성한 single cluster의 경우, 내부에 설치된 리소스들은 무시한다.
	if clm.Labels[clusterV1alpha1.LabelKeyClmClusterType] == clusterV1alpha1.ClusterTypeRegistered {
		remoteClientset, err := util.GetRemoteK8sClient(secret)
		if err != nil {
			log.Error(err, "Failed to get remoteK8sClient")
			return ctrl.Result{}, err
		}

		saList := []types.NamespacedName{
			{
				Name:      util.ArgoServiceAccount,
				Namespace: util.KubeNamespace,
			},
			{
				Name:      adminServiceAccountName,
				Namespace: util.KubeNamespace,
			},
		}
		for _, targetSa := range saList {
			_, err := remoteClientset.
				CoreV1().
				ServiceAccounts(targetSa.Namespace).
				Get(context.TODO(), targetSa.Name, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				log.Info("Cannot find ServiceAccount [" + targetSa.String() + "] from remote cluster. Maybe already deleted")
			} else if err != nil {
				log.Error(err, "Failed to get ServiceAccount ["+targetSa.String()+"] from remote cluster")
				return ctrl.Result{}, err
			} else {
				err := remoteClientset.
					CoreV1().
					ServiceAccounts(targetSa.Namespace).
					Delete(context.TODO(), targetSa.Name, metav1.DeleteOptions{})
				if err != nil {
					log.Error(err, "Cannot delete ServiceAccount ["+targetSa.String()+"] from remote cluster")
					return ctrl.Result{}, err
				}
				log.Info("Delete ServiceAccount [" + targetSa.String() + "] from remote cluster successfully")
			}
		}

		secretList := []types.NamespacedName{
			{
				Name:      util.ArgoServiceAccountTokenSecret,
				Namespace: util.KubeNamespace,
			},
			{
				Name:      adminServiceAccountName + "-token",
				Namespace: util.KubeNamespace,
			},
		}
		for _, targetSecret := range secretList {
			_, err := remoteClientset.
				CoreV1().
				Secrets(targetSecret.Namespace).
				Get(context.TODO(), targetSecret.Name, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				log.Info("Cannot find Secret [" + targetSecret.String() + "] from remote cluster. Maybe already deleted")
			} else if err != nil {
				log.Error(err, "Failed to get Secret ["+targetSecret.String()+"] from remote cluster")
				return ctrl.Result{}, err
			} else {
				err := remoteClientset.
					CoreV1().
					Secrets(targetSecret.Namespace).
					Delete(context.TODO(), targetSecret.Name, metav1.DeleteOptions{})
				if err != nil {
					log.Error(err, "Cannot delete Secret ["+targetSecret.String()+"] from remote cluster")
					return ctrl.Result{}, err
				}
				log.Info("Delete Secret [" + targetSecret.String() + "] from remote cluster successfully")
			}
		}

		// db 로 부터 클러스터에 초대 된 member 들의 info 가져오기
		jsonData, _ := util.List(clm.Namespace, clm.Name)
		memberList := []ClusterMemberInfo{}
		if err := json.Unmarshal(jsonData, &memberList); err != nil {
			return ctrl.Result{}, err
		}

		crbList := []string{
			"cluster-owner-crb-" + secret.Annotations[util.AnnotationKeyOwner],
			"cluster-owner-sa-crb-" + secret.Annotations[util.AnnotationKeyOwner],
			util.ArgoClusterRoleBinding,
		}
		for _, member := range memberList {
			if member.Status == "invited" && member.Attribute == "user" {
				// user 로 초대 된 member crb
				crbList = append(crbList, member.MemberId+"-user-rolebinding")
			} else if member.Status == "invited" && member.Attribute == "group" {
				// group 으로 초대 된 member crb
				crbList = append(crbList, member.MemberId+"-group-rolebinding")
			}
		}
		for _, targetCrb := range crbList {
			_, err := remoteClientset.
				RbacV1().
				ClusterRoleBindings().
				Get(context.TODO(), targetCrb, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				log.Info("Cannot find ClusterRoleBinding [" + targetCrb + "] from remote cluster. Maybe already deleted")
			} else if err != nil {
				log.Error(err, "Failed to get ClusterRoleBinding ["+targetCrb+"] from remote cluster")
				return ctrl.Result{}, err
			} else {
				err := remoteClientset.
					RbacV1().
					ClusterRoleBindings().
					Delete(context.TODO(), targetCrb, metav1.DeleteOptions{})
				if err != nil {
					log.Error(err, "Cannot delete ClusterRoleBinding ["+targetCrb+"] from remote cluster")
					return ctrl.Result{}, err
				}
				log.Info("Delete ClusterRoleBinding [" + targetCrb + "] from remote cluster successfully")
			}
		}

		crList := []string{
			"developer",
			"guest",
			util.ArgoClusterRole,
		}
		for _, targetCr := range crList {
			_, err := remoteClientset.
				RbacV1().
				ClusterRoles().
				Get(context.TODO(), targetCr, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				log.Info("Cannot find ClusterRole [" + targetCr + "] from remote cluster. Maybe already deleted")
			} else if err != nil {
				log.Error(err, "Failed to get ClusterRole ["+targetCr+"] from remote cluster")
				return ctrl.Result{}, err
			} else {
				err := remoteClientset.
					RbacV1().
					ClusterRoles().
					Delete(context.TODO(), targetCr, metav1.DeleteOptions{})
				if err != nil {
					log.Error(err, "Cannot delete ClusterRole ["+targetCr+"] from remote cluster")
					return ctrl.Result{}, err
				}
				log.Info("Delete ClusterRole [" + targetCr + "] from remote cluster successfully")
			}
		}

	}

	// db 에서 member 삭제
	if err := util.Delete(clm.Namespace, clm.Name); err != nil {
		log.Error(err, "Failed to delete cluster info from cluster_member table")
		return ctrl.Result{}, err
	}

	// master cluster에 있는 시크릿 리소스 삭제
	// argocd cluster secret
	key = types.NamespacedName{
		Name:      secret.Annotations[util.AnnotationKeyArgoClusterSecret],
		Namespace: util.ArgoNamespace,
	}
	argoClusterSecret := &coreV1.Secret{}
	if err := r.Get(context.TODO(), key, argoClusterSecret); errors.IsNotFound(err) {
		log.Info("Cannot find Secret for argocd external cluster [" + argoClusterSecret.Name + "]. Maybe already deleted")
	} else if err != nil {
		log.Error(err, "Failed to get Secret for argocd external cluster ["+argoClusterSecret.Name+"]")
		return ctrl.Result{}, err
	} else {
		if err := r.Delete(context.TODO(), argoClusterSecret); err != nil {
			log.Error(err, "Cannot delete Secret for argocd external cluster ["+argoClusterSecret.Name+"]")
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(argoClusterSecret, clusterV1alpha1.ClusterManagerFinalizer)
		log.Info("Delete Secret for argocd external cluster [" + argoClusterSecret.Name + "] successfully")

	}

	// SA token secret
	key = types.NamespacedName{
		Name:      adminServiceAccountName + "-" + clm.Name + "-token",
		Namespace: secret.Namespace,
	}
	saTokenSecret := &coreV1.Secret{}
	if err := r.Get(context.TODO(), key, saTokenSecret); errors.IsNotFound(err) {
		log.Info("Cannot find Secret for ServiceAccount [" + saTokenSecret.Name + "]. Maybe already deleted")
	} else if err != nil {
		log.Error(err, "Failed to get Secret for ServiceAccount ["+saTokenSecret.Name+"]")
		return ctrl.Result{}, err
	} else {
		if err := r.Delete(context.TODO(), saTokenSecret); err != nil {
			log.Error(err, "Cannot delete Secret for ServiceAccount ["+saTokenSecret.Name+"]")
			return ctrl.Result{}, err
		}
		log.Info("Delete Secret for ServiceAccount [" + saTokenSecret.Name + "] successfully")
	}

	// kubeconfig finalizer 제거
	key = types.NamespacedName{
		Name:      clm.Name + util.KubeconfigSuffix,
		Namespace: clm.Namespace,
	}
	kubeconfigSecret := &coreV1.Secret{}
	if err := r.Get(context.TODO(), key, kubeconfigSecret); errors.IsNotFound(err) {
		log.Info("Cannot find secret for secret [" + kubeconfigSecret.Name + "]. Maybe already deleted")
	} else if err != nil {
		log.Error(err, "Failed to get Secret for secret ["+kubeconfigSecret.Name+"]")
		return ctrl.Result{}, err
	} else {
		controllerutil.RemoveFinalizer(secret, clusterV1alpha1.ClusterManagerFinalizer)
		log.Info("Delete Secret for secret [" + kubeconfigSecret.Name + "] successfully")
	}

	return ctrl.Result{}, nil
}

func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	controller, err := ctrl.NewControllerManagedBy(mgr).
		For(&coreV1.Secret{}).
		WithEventFilter(
			predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					return false
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldSecret := e.ObjectOld.(*coreV1.Secret)
					newSecret := e.ObjectNew.(*coreV1.Secret)
					_, oldTarget := oldSecret.Labels[util.LabelKeyClmSecretType]
					_, newTarget := newSecret.Labels[util.LabelKeyClmSecretType]
					isTarget := oldTarget || newTarget
					isDelete := oldSecret.GetDeletionTimestamp().IsZero() && !newSecret.GetDeletionTimestamp().IsZero()
					isFinalized := !controllerutil.ContainsFinalizer(oldSecret, clusterV1alpha1.ClusterManagerFinalizer) &&
						controllerutil.ContainsFinalizer(newSecret, clusterV1alpha1.ClusterManagerFinalizer)
					if isTarget && (isDelete || isFinalized) {
						return true
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

	return controller.Watch(
		&source.Kind{Type: &clusterV1alpha1.ClusterManager{}},
		&handler.EnqueueRequestForObject{},
		predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldClm := e.ObjectOld.(*clusterV1alpha1.ClusterManager)
				newClm := e.ObjectNew.(*clusterV1alpha1.ClusterManager)
				if !oldClm.Status.ControlPlaneReady && newClm.Status.ControlPlaneReady {
					return true
				}
				return false
			},
			CreateFunc: func(e event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return false
			},
		},
	)
}
