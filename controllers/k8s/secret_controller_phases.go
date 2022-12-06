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
	"regexp"
	"strings"

	clusterV1alpha1 "github.com/tmax-cloud/hypercloud-multi-operator/apis/cluster/v1alpha1"
	"github.com/tmax-cloud/hypercloud-multi-operator/controllers/util"

	coreV1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"

	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *SecretReconciler) UpdateClusterManagerControlPlaneEndpoint(ctx context.Context, secret *coreV1.Secret) (ctrl.Result, error) {
	key := types.NamespacedName{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	}
	log := r.Log.WithValues("secret", key)
	log.Info("Start to reconcile phase for UpdateClusterManagerControlPlaneEndpoint... ")

	kubeConfig, err := clientcmd.Load(secret.Data["value"])
	if err != nil {
		log.Error(err, "Failed to get kubeconfig data from secret")
		return ctrl.Result{}, err
	}

	key = types.NamespacedName{
		Name:      strings.Split(secret.Name, util.KubeconfigSuffix)[0],
		Namespace: secret.Namespace,
	}
	clm := &clusterV1alpha1.ClusterManager{}
	err = r.Get(context.TODO(), key, clm)
	if errors.IsNotFound(err) {
		log.Info("Cannot find clusterManager")
		// return ctrl.Result{RequeueAfter: requeueAfter5Sec}, nil
	} else if err != nil {
		log.Error(err, "Failed to get clusterManager + ["+clm.Name+"]")
		return ctrl.Result{}, err
	} else {
		server := kubeConfig.Clusters[kubeConfig.Contexts[kubeConfig.CurrentContext].Cluster].Server
		if !strings.EqualFold(clm.Status.ControlPlaneEndpoint, server) {
			log.Info("Update clustermanager status. add ControlPlane endpoint")
			helper, _ := patch.NewHelper(clm, r.Client)
			defer func() {
				if err := helper.Patch(context.TODO(), clm); err != nil {
					log.Error(err, "ClusterManager patch error")
				}
			}()
			clm.Status.ControlPlaneEndpoint = server
		}
	}

	return ctrl.Result{}, nil
}

func (r *SecretReconciler) DeployRBACResources(ctx context.Context, secret *coreV1.Secret) (ctrl.Result, error) {
	log := r.Log.WithValues(
		"secret",
		types.NamespacedName{
			Name:      secret.Name,
			Namespace: secret.Namespace,
		},
	)
	log.Info("Start to reconcile phase for Deploy rolebinding to remote")

	clm := &clusterV1alpha1.ClusterManager{}
	key := types.NamespacedName{
		Name:      strings.Split(secret.Name, util.KubeconfigSuffix)[0],
		Namespace: secret.Namespace,
	}
	if err := r.Get(context.TODO(), key, clm); err != nil {
		log.Error(err, "Failed to get ClusterManager")
		return ctrl.Result{}, err
	}

	remoteClientset, err := util.GetRemoteK8sClient(secret)
	if err != nil {
		log.Error(err, "Failed to get remoteK8sClient")
		return ctrl.Result{}, err
	}

	clusterAdminCRB := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-owner-crb-" + clm.Annotations[util.AnnotationKeyOwner],
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: rbacv1.GroupName,
				Kind:     rbacv1.UserKind,
				Name:     clm.Annotations[util.AnnotationKeyOwner],
			},
		},
	}

	_, err = remoteClientset.
		RbacV1().
		ClusterRoleBindings().
		Get(context.TODO(), clusterAdminCRB.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := remoteClientset.
			RbacV1().
			ClusterRoleBindings().
			Create(context.TODO(), clusterAdminCRB, metav1.CreateOptions{})
		if err != nil {
			log.Error(err, "Cannot create ClusterRoleBinding for cluster-admin")
			return ctrl.Result{}, err
		}
		log.Info("Create ClusterRoleBinding for cluster-admin to remote cluster successfully")
	} else if err != nil {
		log.Error(err, "Failed to get ClusterRoleBinding for cluster-admin from remote cluster")
		return ctrl.Result{}, err
	}

	targetGroup := []string{
		"",
		"apps",
		"autoscaling",
		"batch",
		"extensions",
		"policy",
		"networking.k8s.io",
		"snapshot.storage.k8s.io",
		"storage.k8s.io",
		"apiextensions.k8s.io",
		"metrics.k8s.io",
	}
	crList := []*rbacv1.ClusterRole{
		CreateClusterRole("developer", targetGroup, []string{rbacv1.VerbAll}),
		CreateClusterRole("guest", targetGroup, []string{"get", "list", "watch"}),
	}
	for _, targetCr := range crList {
		_, err := remoteClientset.
			RbacV1().
			ClusterRoles().
			Get(context.TODO(), targetCr.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			_, err := remoteClientset.
				RbacV1().
				ClusterRoles().
				Create(context.TODO(), targetCr, metav1.CreateOptions{})
			if err != nil {
				log.Error(err, "Cannot create ClusterRole ["+targetCr.Name+"] to remote cluster")
				return ctrl.Result{}, err
			}
			log.Info("Create ClusterRole [" + targetCr.Name + "] to remote cluster successfully")
		} else if err != nil {
			log.Error(err, "Failed to get ClusterRole ["+targetCr.Name+"] from remote cluster")
			return ctrl.Result{}, err
		}
	}

	re, _ := regexp.Compile("[" + regexp.QuoteMeta(`!#$%&'"*+-/=?^_{|}~().,:;<>[]\`) + "`\\s" + "]")
	email := clm.Annotations[util.AnnotationKeyOwner]
	adminServiceAccount := &coreV1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      re.ReplaceAllString(strings.Replace(email, "@", "-at-", -1), "-"),
			Namespace: util.KubeNamespace,
		},
	}
	_, err = remoteClientset.
		CoreV1().
		ServiceAccounts(util.KubeNamespace).
		Get(context.TODO(), adminServiceAccount.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := remoteClientset.
			CoreV1().
			ServiceAccounts(util.KubeNamespace).
			Create(context.TODO(), adminServiceAccount, metav1.CreateOptions{})
		if err != nil {
			log.Error(err, "Cannot create ServiceAccount ["+adminServiceAccount.Name+"] to remote cluster")
			return ctrl.Result{}, err
		}
		log.Info("Create ServiceAccount [" + adminServiceAccount.Name + "] to remote cluster successfully")
	} else if err != nil {
		log.Error(err, "Failed to get ServiceAccount ["+adminServiceAccount.Name+"] from remote cluster")
		return ctrl.Result{}, err
	}

	adminServiceAccountTokenSecret := &coreV1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				coreV1.ServiceAccountNameKey: adminServiceAccount.Name,
			},
			Name: adminServiceAccount.Name + "-token",
		},
		Type: coreV1.SecretTypeServiceAccountToken,
	}
	_, err = remoteClientset.
		CoreV1().
		Secrets(util.KubeNamespace).
		Get(context.TODO(), adminServiceAccountTokenSecret.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := remoteClientset.
			CoreV1().
			Secrets(util.KubeNamespace).
			Create(context.TODO(), adminServiceAccountTokenSecret, metav1.CreateOptions{})
		if err != nil {
			log.Error(err, "Cannot create ServiceAccount token secret ["+adminServiceAccount.Name+"] to remote cluster")
			return ctrl.Result{}, err
		}
		log.Info("Create ServiceAccount token secret [" + adminServiceAccount.Name + "] to remote cluster successfully")
	} else if err != nil {
		log.Error(err, "Failed to get ServiceAccount token secret ["+adminServiceAccount.Name+"] from remote cluster")
		return ctrl.Result{}, err
	}

	adminServiceAccountCRB := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-owner-sa-crb-" + clm.Annotations[util.AnnotationKeyOwner],
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      adminServiceAccount.Name,
				Namespace: util.KubeNamespace,
			},
		},
	}
	_, err = remoteClientset.
		RbacV1().
		ClusterRoleBindings().
		Get(context.TODO(), adminServiceAccountCRB.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := remoteClientset.
			RbacV1().
			ClusterRoleBindings().
			Create(context.TODO(), adminServiceAccountCRB, metav1.CreateOptions{})
		if err != nil {
			log.Error(err, "Cannot create ClusterRoleBinding for admin service account")
			return ctrl.Result{}, err
		}
		log.Info("Create ClusterRoleBinding for admin service account to remote cluster successfully")
	} else if err != nil {
		log.Error(err, "Failed to get ClusterRoleBinding for admin service account from remote cluster")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SecretReconciler) DeployArgocdResources(ctx context.Context, secret *coreV1.Secret) (ctrl.Result, error) {
	log := r.Log.WithValues("secret", types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace})
	log.Info("Start to reconcile phase for Deploy argocd resources to remote")

	kubeConfig, err := clientcmd.Load(secret.Data["value"])
	if err != nil {
		log.Error(err, "Failed to get secret")
		return ctrl.Result{}, err
	}

	serverURI := kubeConfig.Clusters[kubeConfig.Contexts[kubeConfig.CurrentContext].Cluster].Server
	argoSecretName, err := util.URIToSecretName("cluster", serverURI)
	if err != nil {
		log.Error(err, "Failed to parse server uri")
		return ctrl.Result{}, err
	}

	if _, ok := secret.Annotations[util.AnnotationKeyArgoClusterSecret]; !ok {
		secret.Annotations[util.AnnotationKeyArgoClusterSecret] = argoSecretName
	}

	remoteClientset, err := util.GetRemoteK8sClient(secret)
	if err != nil {
		log.Error(err, "Failed to get remoteK8sClient")
		return ctrl.Result{}, err
	}

	argocdManagerSA := &coreV1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: util.ArgoServiceAccount,
		},
	}
	_, err = remoteClientset.
		CoreV1().
		ServiceAccounts(util.KubeNamespace).
		Get(context.TODO(), argocdManagerSA.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := remoteClientset.
			CoreV1().
			ServiceAccounts(util.KubeNamespace).
			Create(context.TODO(), argocdManagerSA, metav1.CreateOptions{})
		if err != nil {
			log.Error(err, "Cannot create ServiceAccount for argocd ["+argocdManagerSA.Name+"] to remote cluster")
			return ctrl.Result{}, err
		}
		log.Info("Create ServiceAccount for argocd [" + argocdManagerSA.Name + "] to remote cluster successfully")
	} else if err != nil {
		log.Error(err, "Failed to get ServiceAccount for argocd ["+argocdManagerSA.Name+"] from remote cluster")
		return ctrl.Result{}, err
	}

	// service account 생성시 token secret이 자동으로 생성되지만
	// random suffix가 붙기때문에 조회하는 process가 번잡하므로 시크릿을 수동으로 생성
	argocdManagerTokenSecret := &coreV1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				coreV1.ServiceAccountNameKey: util.ArgoServiceAccount,
			},
			Name: util.ArgoServiceAccountTokenSecret,
		},
		Type: coreV1.SecretTypeServiceAccountToken,
	}
	_, err = remoteClientset.
		CoreV1().
		Secrets(util.KubeNamespace).
		Get(context.TODO(), argocdManagerTokenSecret.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := remoteClientset.
			CoreV1().
			Secrets(util.KubeNamespace).
			Create(context.TODO(), argocdManagerTokenSecret, metav1.CreateOptions{})
		if err != nil {
			log.Error(err, "Cannot create ServiceAccount token secret for argocd ["+argocdManagerTokenSecret.Name+"] to remote cluster")
			return ctrl.Result{}, err
		}
		log.Info("Create ServiceAccount token secret for argocd [" + argocdManagerTokenSecret.Name + "] to remote cluster successfully")
	} else if err != nil {
		log.Error(err, "Failed to get ServiceAccount token secret for argocd ["+argocdManagerTokenSecret.Name+"] from remote cluster")
		return ctrl.Result{}, err
	}

	argocdManagerRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: util.ArgoClusterRole,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{rbacv1.APIGroupAll},
				Resources: []string{rbacv1.ResourceAll},
				Verbs:     []string{rbacv1.VerbAll},
			},
			{
				NonResourceURLs: []string{rbacv1.NonResourceAll},
				Verbs:           []string{rbacv1.VerbAll},
			},
		},
	}
	_, err = remoteClientset.
		RbacV1().
		ClusterRoles().
		Get(context.TODO(), argocdManagerRole.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := remoteClientset.
			RbacV1().
			ClusterRoles().
			Create(context.TODO(), argocdManagerRole, metav1.CreateOptions{})
		if err != nil {
			log.Error(err, "Cannot create ClusterRole for argocd ["+argocdManagerRole.Name+"] to remote cluster")
			return ctrl.Result{}, err
		}
		log.Info("Create ClusterRole for argocd [" + argocdManagerRole.Name + "] to remote cluster successfully")
	} else if err != nil {
		log.Error(err, "Failed to get ClusterRole for argocd ["+argocdManagerRole.Name+"] from remote cluster")
		return ctrl.Result{}, err
	}

	argocdManagerRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: util.ArgoClusterRoleBinding,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     util.ArgoClusterRole,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      util.ArgoServiceAccount,
				Namespace: util.KubeNamespace,
			},
		},
	}
	_, err = remoteClientset.
		RbacV1().
		ClusterRoleBindings().
		Get(context.TODO(), argocdManagerRoleBinding.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := remoteClientset.
			RbacV1().
			ClusterRoleBindings().
			Create(context.TODO(), argocdManagerRoleBinding, metav1.CreateOptions{})
		if err != nil {
			log.Error(err, "Cannot create ClusterRoleBinding for argocd ["+argocdManagerRoleBinding.Name+"] to remote cluster")
			return ctrl.Result{}, err
		}
		log.Info("Create ClusterRoleBinding for argocd [" + argocdManagerRoleBinding.Name + "] to remote cluster successfully")
	} else if err != nil {
		log.Error(err, "Failed to get ClusterRoleBinding for argocd ["+argocdManagerRoleBinding.Name+"] from remote cluster")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
