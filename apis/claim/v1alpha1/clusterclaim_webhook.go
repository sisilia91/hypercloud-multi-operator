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
	"reflect"
	"regexp"
	"strconv"
	"strings"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var ClusterClaimWebhookLogger = logf.Log.WithName("clusterclaim-resource")

func (r *ClusterClaim) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-claim-tmax-io-v1alpha1-clusterclaim,mutating=true,failurePolicy=fail,groups=claim.tmax.io,resources=clusterclaims,verbs=update,versions=v1alpha1,name=mutation.webhook.clusterclaim,admissionReviewVersions=v1beta1;v1,sideEffects=NoneOnDryRun

var _ webhook.Defaulter = &ClusterClaim{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ClusterClaim) Default() {
	ClusterClaimWebhookLogger.Info("default", "name", r.Name)
	// if len(r.Name) > maxGeneratedNameLength {
	// r.Name = r.Name[:maxGeneratedNameLength]
	// }
	// return fmt.Sprintf("%s%s", base, utilrand.String(randomLength))

	// r.Name = r.Name + "-" + utilrand.String(randomLength)
	// r.GenerateName = r.GenerateName + "-"
	// utilrand.String(randomLength)
	// r.Name = r.Name + r.Annotations["creator"]
	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-claim-tmax-io-v1alpha1-clusterclaim,mutating=false,failurePolicy=fail,groups=claim.tmax.io,resources=clusterclaims;clusterclaims/status,versions=v1alpha1,name=validation.webhook.clusterclaim,admissionReviewVersions=v1beta1;v1,sideEffects=NoneOnDryRun

var _ webhook.Validator = &ClusterClaim{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterClaim) ValidateCreate() error {
	ClusterClaimWebhookLogger.Info("validate create", "name", r.Name)

	// k8s 리소스들의 이름은 기본적으로 DNS-1123의 룰을 따라야 함
	// 자세한 내용은 https://kubernetes.io/ko/docs/concepts/overview/working-with-objects/names/ 참조
	// cluster manager 리소스는 cluster claim의 spec.clusterName을 metadata.name으로 가지게 되므로
	// spec.clusterName도 DNS-1123 룰을 따르게 해야할 필요가 있으므로 웹훅을 통해 validation
	// 22.07.28 업데이트 사항
	// 몇몇 리소스들은 DNS-1035룰을 따르게 되어 있는데, service가 이에 해당함
	// cluster 이름을 이용하여 service를 생성하는 로직이 있기 때문에, cluster 이름도 DNS-1123이 아닌 DNS-1035를 따라야 함
	// service 이름은 (( cluster name ))-gateway-service로 정해지는데, 63자 이하여야 하므로 cluster name은 45자 이하여야 함
	reg, _ := regexp.Compile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)
	// reg, _ := regexp.Compile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	if !reg.MatchString(r.Spec.ClusterName) {
		errList := []*field.Error{
			{
				Type:     field.ErrorTypeInvalid,
				Field:    "spec.clusterName",
				BadValue: r.Spec.ClusterName,
				Detail: strings.Join(
					[]string{
						"a DNS-1035 label must consist of lower case alphanumeric characters or '-',",
						"start with an alphabetic character, and end with an alphanumeric character",
						"(e.g. 'my-name',  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?')",
						// "a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.',",
						// "and must start and end with an alphanumeric character (e.g. 'example.com',",
						// "regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')",
					},
					" ",
				),
			},
		}
		return k8sErrors.NewInvalid(r.GroupVersionKind().GroupKind(), "InvalidSpecClusterName", errList)
	}

	maxLength := 63 - len("-gateway-service")
	if len(r.Spec.ClusterName) > maxLength {
		errList := []*field.Error{
			{
				Type:     field.ErrorTypeInvalid,
				Field:    "spec.clusterName",
				BadValue: r.Spec.ClusterName,
				Detail:   "must be no more than " + strconv.Itoa(maxLength) + " characters",
			},
		}
		return k8sErrors.NewInvalid(r.GroupVersionKind().GroupKind(), "InvalidSpecClusterName", errList)
	}

	if r.Spec.MasterNum%2 == 0 {
		return errors.New("Cannot be an even number when using managed etcd")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterClaim) ValidateUpdate(old runtime.Object) error {
	ClusterClaimWebhookLogger.Info("validate update", "name", r.Name)
	oldClusterClaim := old.(*ClusterClaim).DeepCopy()

	if !r.ObjectMeta.DeletionTimestamp.IsZero() {
		return nil
	}

	if oldClusterClaim.Status.Phase == ClusterClaimPhaseApproved || oldClusterClaim.Status.Phase == ClusterClaimPhaseClusterDeleted {
		if !reflect.DeepEqual(oldClusterClaim.Spec, r.Spec) {
			return errors.New("cannot modify clusterClaim after approval")
		}
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterClaim) ValidateDelete() error {
	ClusterClaimWebhookLogger.Info("validate delete", "name", r.Name)

	// cluster가 남아있으면 cluster claim을 삭제하지 못하도록 처리
	if r.Status.Phase == ClusterClaimPhaseApproved {
		return k8sErrors.NewBadRequest("Deleting cluster must precedes deleting cluster claim.")
	}

	return nil
}
