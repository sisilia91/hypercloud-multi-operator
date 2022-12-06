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

package v1alpha1

import (
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var ClusterManagerWebhookLogger = logf.Log.WithName("clustermanager-resource")

func (r *ClusterManager) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// // +kubebuilder:webhook:path=/mutate-cluster-tmax-io-v1alpha1-clustermanager,mutating=true,failurePolicy=fail,groups=cluster.tmax.io,resources=clustermanagers,verbs=create;update,versions=v1alpha1,name=mutation.webhook.clustermanager,admissionReviewVersions=v1beta1;v1,sideEffects=NoneOnDryRun

// var _ webhook.Defaulter = &ClusterManager{}

// // Default implements webhook.Defaulter so a webhook will be registered for the type
// func (r *ClusterManager) Default() {
// 	ClusterManagerWebhookLogger.Info("default", "name", r.Name)

// 	// TODO(user): fill in your defaulting logic.
// }

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=update,path=/validate-cluster-tmax-io-v1alpha1-clustermanager,mutating=false,failurePolicy=fail,groups=cluster.tmax.io,resources=clustermanagers,versions=v1alpha1,name=validation.webhook.clustermanager,admissionReviewVersions=v1beta1;v1,sideEffects=NoneOnDryRun

var _ webhook.Validator = &ClusterManager{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterManager) ValidateCreate() error {
	ClusterManagerWebhookLogger.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterManager) ValidateUpdate(old runtime.Object) error {

	ClusterManagerWebhookLogger.Info("validate update", "name", r.Name)
	oldClusterManager := old.(*ClusterManager).DeepCopy()

	if r.Annotations["owner"] != oldClusterManager.Annotations["owner"] {
		return errors.New("cannot modify clusterManager.Annotations.owner")
	}

	// if r.Status.Ready == false {
	// 	if !reflect.DeepEqual(r.Status.Members, oldClusterClaim.Status.Members) {
	// 		return errors.New("Cannot modify members when cluster status is not ready")
	// 	}
	// 	if !reflect.DeepEqual(r.Status.Groups, oldClusterClaim.Status.Groups) {
	// 		return errors.New("Cannot modify groups when cluster status is not ready")
	// 	}
	// }

	if oldClusterManager.Labels[LabelKeyClmClusterType] == ClusterTypeCreated {
		// cluster 생성 중에 version upgrade 또는 scaling을 진행할 수 없음
		if oldClusterManager.Status.Phase == ClusterManagerPhaseProcessing {
			if r.Spec.Version != oldClusterManager.Spec.Version ||
				r.Spec.MasterNum != oldClusterManager.Spec.MasterNum ||
				r.Spec.WorkerNum != oldClusterManager.Spec.WorkerNum {
				return errors.New("Cannot upgrade or scaling, when cluster's phase is progressing")
			}

		}

		// version upgrade의 경우
		if r.Spec.Version != oldClusterManager.Spec.Version {
			// vsphere의 경우, version과 template을 함께 업데이트해야 함
			if r.Spec.Provider == ProviderVSphere {
				if r.VsphereSpec.VcenterTemplate == oldClusterManager.VsphereSpec.VcenterTemplate {
					return errors.New("For vsphere provider, must update spec.version and vsphereSpec.vcetnerTemplate")
				}
			}
		}

		// scaling을 못하는 경우
		if (r.Spec.MasterNum != oldClusterManager.Spec.MasterNum ||
			r.Spec.WorkerNum != oldClusterManager.Spec.WorkerNum) &&
			(oldClusterManager.Status.Phase == ClusterManagerPhaseProcessing ||
				oldClusterManager.Status.Phase == ClusterManagerPhaseScaling ||
				oldClusterManager.Status.Phase == ClusterManagerPhaseDeleting ||
				oldClusterManager.Status.Phase == ClusterManagerPhaseUpgrading) {
			return errors.New("Cannot update MasterNum or WorkerNum at Processing, Scaling, Upgrading or Deleting phases")
		}

		// version upgrade를 못하는 경우
		if r.Spec.Version != oldClusterManager.Spec.Version &&
			(oldClusterManager.Status.Phase == ClusterManagerPhaseProcessing ||
				oldClusterManager.Status.Phase == ClusterManagerPhaseScaling ||
				oldClusterManager.Status.Phase == ClusterManagerPhaseDeleting ||
				oldClusterManager.Status.Phase == ClusterManagerPhaseUpgrading) {
			return errors.New("Cannot update version at Progressing, Scaling, Upgrading or Deleting phases")
		}

		if r.Spec.MasterNum%2 == 0 {
			return errors.New("Cannot be an even number when using managed etcd")
		}
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterManager) ValidateDelete() error {

	ClusterManagerWebhookLogger.Info("validate delete", "name", r.Name)

	return nil
}
