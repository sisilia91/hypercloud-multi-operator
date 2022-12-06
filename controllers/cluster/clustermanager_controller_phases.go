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
	"fmt"
	"regexp"
	"strings"

	argocdV1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	servicecatalogv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	clusterV1alpha1 "github.com/tmax-cloud/hypercloud-multi-operator/apis/cluster/v1alpha1"
	hyperauthCaller "github.com/tmax-cloud/hypercloud-multi-operator/controllers/hyperAuth"
	util "github.com/tmax-cloud/hypercloud-multi-operator/controllers/util"

	coreV1 "k8s.io/api/core/v1"
	networkingV1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"

	// cpavV1alpha3 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1alpha3"
	capiV1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)

// reconcile시작 전 필요한 동작들을 수행
func (r *ClusterManagerReconciler) ReadyReconcilePhase(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (ctrl.Result, error) {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	log.Info("Start to reconcile phase for ReadyReconcilePhase")

	// Label migration for old version
	if _, ok := clusterManager.GetLabels()[clusterV1alpha1.LabelKeyClmClusterTypeDefunct]; ok {
		clusterManager.Labels[clusterV1alpha1.LabelKeyClmClusterType] =
			clusterManager.Labels[clusterV1alpha1.LabelKeyClmClusterTypeDefunct]
	}

	// Status migration for old version
	if !clusterManager.Status.GatewayReadyMigration {
		clusterManager.Status.GatewayReady = clusterManager.Status.PrometheusReady
		clusterManager.Status.GatewayReadyMigration = true
	}

	// ApplicationLink migration for old version
	if clusterManager.Status.ApplicationLink == "" {
		argoIngress := &networkingV1.Ingress{}
		key := types.NamespacedName{
			Name:      util.ArgoIngressName,
			Namespace: util.ArgoNamespace,
		}
		if err := r.Get(context.TODO(), key, argoIngress); err != nil {
			log.Error(err, "Can not get argocd ingress information.")
		} else {
			subdomain := strings.Split(argoIngress.Spec.Rules[0].Host, ".")[0]
			SetApplicationLink(clusterManager, subdomain)
		}
	}

	if clusterManager.Status.Version == "" {
		clusterManager.Status.Version = clusterManager.Spec.Version
	}
	if clusterManager.Status.MasterNum == 0 {
		clusterManager.Status.MasterNum = clusterManager.Spec.MasterNum
	}
	if clusterManager.Status.WorkerNum == 0 {
		clusterManager.Status.WorkerNum = clusterManager.Spec.WorkerNum
	}

	return ctrl.Result{}, nil
}

func (r *ClusterManagerReconciler) UpdateClusterManagerStatus(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (ctrl.Result, error) {
	if clusterManager.Status.ControlPlaneReady {
		return ctrl.Result{}, nil
	}
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	log.Info("Start to reconcile phase for UpdateClusterManagerStatus")

	kubeconfigSecret, err := r.GetKubeconfigSecret(clusterManager)
	if err != nil {
		log.Error(err, "Failed to get kubeconfig secret")
		return ctrl.Result{RequeueAfter: requeueAfter10Second}, nil
	}

	remoteClientset, err := util.GetRemoteK8sClient(kubeconfigSecret)
	if err != nil {
		log.Error(err, "Failed to get remoteK8sClient")
		return ctrl.Result{}, err
	}

	// cluster registration의 경우에는 k8s version을 parameter로 받지 않기 때문에,
	// k8s version을 single cluster의 kube-system 네임스페이스의 kubeadm-config ConfigMap으로 부터 조회
	kubeadmConfig, err := remoteClientset.
		CoreV1().
		ConfigMaps(util.KubeNamespace).
		Get(context.TODO(), "kubeadm-config", metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Failed to get kubeadm-config ConfigMap from remote cluster")
		return ctrl.Result{}, err
	}

	jsonData, _ := yaml.YAMLToJSON([]byte(kubeadmConfig.Data["ClusterConfiguration"]))
	data := make(map[string]interface{})
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return ctrl.Result{}, err
	}
	clusterManager.Spec.Version = fmt.Sprintf("%v", data["kubernetesVersion"])

	nodeList, err := remoteClientset.
		CoreV1().
		Nodes().
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error(err, "Failed to list remote K8s nodeList")
		return ctrl.Result{}, err
	}

	clusterManager.Spec.MasterNum = 0
	clusterManager.Status.MasterRun = 0
	clusterManager.Spec.WorkerNum = 0
	clusterManager.Status.WorkerRun = 0
	clusterManager.Spec.Provider = util.ProviderUnknown
	clusterManager.Status.Provider = util.ProviderUnknown
	for _, node := range nodeList.Items {
		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			clusterManager.Spec.MasterNum++
			if node.Status.Conditions[len(node.Status.Conditions)-1].Type == "Ready" {
				clusterManager.Status.MasterRun++
			}
		} else {
			clusterManager.Spec.WorkerNum++
			if node.Status.Conditions[len(node.Status.Conditions)-1].Type == "Ready" {
				clusterManager.Status.WorkerRun++
			}
		}

		if clusterManager.Spec.Provider == util.ProviderUnknown && node.Spec.ProviderID != "" {
			providerID, err := util.GetProviderName(
				strings.Split(node.Spec.ProviderID, "://")[0],
			)
			if err != nil {
				log.Error(err, "Cannot find given provider name.")
			}
			clusterManager.Status.Provider = providerID
			clusterManager.Spec.Provider = providerID
		}
	}

	if clusterManager.Spec.Provider == util.ProviderUnknown {
		reg, _ := regexp.Compile(`cloud-provider: [a-zA-Z-_ ]+`)
		matchString := reg.FindString(kubeadmConfig.Data["ClusterConfiguration"])
		if matchString != "" {
			cloudProvider, err := util.GetProviderName(
				matchString[len("cloud-provider: "):],
			)
			if err != nil {
				log.Error(err, "Cannot find given provider name.")
			}
			clusterManager.Status.Provider = cloudProvider
			clusterManager.Spec.Provider = cloudProvider
		}
	}

	// health check
	resp, err := remoteClientset.
		RESTClient().
		Get().
		AbsPath("/readyz").
		DoRaw(context.TODO())
	if err != nil {
		log.Error(err, "Failed to get remote cluster status")
		return ctrl.Result{}, err
	}
	if string(resp) == "ok" {
		clusterManager.Status.ControlPlaneReady = true
		clusterManager.Status.Ready = true
	} else {
		log.Info("Remote cluster is not ready... wait...")
		return ctrl.Result{RequeueAfter: requeueAfter30Second}, nil
	}

	log.Info("Update status of ClusterManager successfully")
	generatedSuffix := util.CreateSuffixString()
	clusterManager.Annotations[clusterV1alpha1.AnnotationKeyClmSuffix] = generatedSuffix
	return ctrl.Result{}, nil
}

func (r *ClusterManagerReconciler) CreateServiceInstance(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (ctrl.Result, error) {
	if clusterManager.Annotations[clusterV1alpha1.AnnotationKeyClmSuffix] != "" {
		return ctrl.Result{}, nil
	}
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	log.Info("Start to reconcile phase for CreateServiceInstance")

	// hypercloud5-api-server를 audit webhook server로 사용한다.
	// api-server의 CA certificate를 추출하여 service instance에 넣어준다.
	// key = types.NamespacedName{
	// 	Name:      "hypercloud5-api-server-certs",
	// 	Namespace: "hypercloud5-system",
	// }
	// auditWebhookServerSecret := &coreV1.Secret{}
	// if err := r.Get(context.TODO(), key, auditWebhookServerSecret); errors.IsNotFound(err) {
	// 	log.Error(err, "hypercloud5-api-server-certs secret not created . Waiting for secret to be created")
	// 	return ctrl.Result{RequeueAfter: requeueAfter10Second}, err
	// } else if err != nil {
	// 	log.Error(err, "Failed to get hypercloud5-api-server-certs secret")
	// 	return ctrl.Result{}, err
	// }
	// webhookServerCACert, err := base64.StdEncoding.DecodeString(string(hyperauthHttpsSecret.Data["ca.crt"]))
	// if err != nil {
	// 	log.Error(err, "Failed to decode hypercloud5-api-server-certs")
	// 	return ctrl.Result{}, err
	// }

	// hyperauth certificate를 가져와서 service instance에 넣어주어야 한다.
	// key := types.NamespacedName{
	// 	Name:      hyperAuth.HYPERAUTH_HTTPS_SECRET,
	// 	Namespace: hyperAuth.HYPERAUTH_NAMESPACE,
	// }

	// hyperauthHttpsSecret := &coreV1.Secret{}
	// if err := r.Get(context.TODO(), key, hyperauthHttpsSecret); errors.IsNotFound(err) {
	// 	log.Error(err, "Hyperauth-https-secret not created . Waiting for secret to be created")
	// 	return ctrl.Result{RequeueAfter: requeueAfter10Second}, err
	// } else if err != nil {
	// 	log.Error(err, "Failed to get hyperauth-https-secret")
	// 	return ctrl.Result{}, err
	// }

	// hyperauthTlsCert := hyperAuth.GetHyperAuthTLSCertificate(hyperauthHttpsSecret)

	key := types.NamespacedName{
		Name:      clusterManager.Name + clusterManager.Annotations[clusterV1alpha1.AnnotationKeyClmSuffix],
		Namespace: clusterManager.Namespace,
	}

	if err := r.Get(context.TODO(), key, &servicecatalogv1beta1.ServiceInstance{}); errors.IsNotFound(err) {
		clusterJson, err := Marshaling(&ClusterParameter{}, *clusterManager)
		if err != nil {
			log.Error(err, "Failed to marshal cluster parameters")
		}

		var providerJson []byte
		switch strings.ToUpper(clusterManager.Spec.Provider) {
		case util.ProviderAws:
			providerJson, err = Marshaling(&AwsParameter{}, *clusterManager)
			if err != nil {
				log.Error(err, "Failed to marshal cluster parameters")
				return ctrl.Result{}, err
			}
		case util.ProviderVsphere:
			providerJson, err = Marshaling(&VsphereParameter{}, *clusterManager)
			if err != nil {
				log.Error(err, "Failed to marshal cluster parameters")
				return ctrl.Result{}, err
			}
		}
		clusterJson = util.MergeJson(clusterJson, providerJson)
		generatedSuffix := util.CreateSuffixString()
		serviceInstanceName := clusterManager.Name + "-" + generatedSuffix
		serviceInstance := MakeServiceInstance(clusterManager, serviceInstanceName, clusterJson, false)
		if err = r.Create(context.TODO(), serviceInstance); err != nil {
			log.Error(err, "Failed to create ServiceInstance")
			return ctrl.Result{}, err
		}
		clusterManager.Annotations[clusterV1alpha1.AnnotationKeyClmSuffix] = generatedSuffix
	} else if err != nil {
		log.Error(err, "Failed to get ServiceInstance")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterManagerReconciler) CreateUpgradeServiceInstance(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (ctrl.Result, error) {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	log.Info("Start to reconcile phase for CreateUpgradeServiceInstance")

	clmSuffix, ok := clusterManager.Annotations[clusterV1alpha1.AnnotationKeyClmSuffix]
	if !ok {
		err := fmt.Errorf("Cannot find cluster suffix from cluster manager")
		log.Error(err, "failed to upgrade vsphere cluster")
		return ctrl.Result{}, err
	}
	serviceInstanceName := fmt.Sprintf("%s-%s-%s", clusterManager.Name, clmSuffix, clusterManager.Spec.Version)
	key := types.NamespacedName{
		Name:      serviceInstanceName,
		Namespace: clusterManager.Namespace,
	}

	if err := r.Get(context.TODO(), key, &servicecatalogv1beta1.ServiceInstance{}); errors.IsNotFound(err) {
		upgradeJson, err := Marshaling(&VsphereUpgradeParameter{}, *clusterManager)
		if err != nil {
			log.Error(err, "Failed to marshal upgrade parameters")
		}
		serviceInstance := MakeServiceInstance(clusterManager, serviceInstanceName, upgradeJson, true)
		ctrl.SetControllerReference(clusterManager, serviceInstance, r.Scheme)
		if err = r.Create(context.TODO(), serviceInstance); err != nil {
			log.Error(err, "Failed to create ServiceInstance")
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get ServiceInstance")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterManagerReconciler) SetEndpoint(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (ctrl.Result, error) {
	if clusterManager.Annotations[clusterV1alpha1.AnnotationKeyClmApiserver] != "" {
		return ctrl.Result{}, nil
	}
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	log.Info("Start to reconcile phase for SetEndpoint")

	key := clusterManager.GetNamespacedName()
	cluster := &capiV1alpha3.Cluster{}
	if err := r.Get(context.TODO(), key, cluster); errors.IsNotFound(err) {
		log.Info("Cluster is not found. Requeue after 20sec")
		return ctrl.Result{RequeueAfter: requeueAfter20Second}, err
	} else if err != nil {
		log.Error(err, "Failed to get cluster")
		return ctrl.Result{}, err
	}

	if cluster.Spec.ControlPlaneEndpoint.Host == "" {
		log.Info("ControlPlane endpoint is not ready yet. requeue after 20sec")
		return ctrl.Result{RequeueAfter: requeueAfter20Second}, nil
	}
	clusterManager.Annotations[clusterV1alpha1.AnnotationKeyClmApiserver] = cluster.Spec.ControlPlaneEndpoint.Host

	return ctrl.Result{}, nil
}

// controlplane을 scaling한다.
func (r *ClusterManagerReconciler) ControlplaneScaling(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (ctrl.Result, error) {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	log.Info("Start to reconcile phase for ControlplaneScaling")

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-control-plane",
		Namespace: clusterManager.Namespace,
	}

	kcp := &controlplanev1.KubeadmControlPlane{}
	if err := r.Get(context.TODO(), key, kcp); errors.IsNotFound(err) {
		log.Error(err, "Cannot find kubeadmcontrolplane")
		return ctrl.Result{}, err
	} else if err != nil {
		log.Error(err, "Failed to get kubeadmcontrolplane")
		return ctrl.Result{}, err
	}

	expectedNum := int32(clusterManager.Spec.MasterNum)
	if *kcp.Spec.Replicas != expectedNum {
		*kcp.Spec.Replicas = expectedNum
		if err := r.Update(context.TODO(), kcp); err != nil {
			log.Info("Failed to update kubadmcontrolplane")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: requeueAfter20Second}, nil
	}

	if kcp.Status.ReadyReplicas == *kcp.Spec.Replicas {
		log.Info("Controlplane scaling is completed successfully")
		clusterManager.Status.MasterNum = clusterManager.Spec.MasterNum
		return ctrl.Result{}, nil
	}
	if clusterManager.Spec.MasterNum > clusterManager.Status.MasterNum {
		log.Info("Waiting for Controlplane nodes to be scaled out. Requeue after 1 min.")
	} else {
		log.Info("Waiting for Controlplane nodes to be scaled in. Requeue after 1 min.")
	}
	return ctrl.Result{RequeueAfter: requeueAfter1Minute}, nil
}

// worker를 scaling한다.
func (r *ClusterManagerReconciler) WorkerScaling(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (ctrl.Result, error) {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	log.Info("Start to reconcile phase for WorkerScaling")

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-md-0",
		Namespace: clusterManager.Namespace,
	}

	md := &capiV1alpha3.MachineDeployment{}
	if err := r.Get(context.TODO(), key, md); errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get machineDeployment")
		return ctrl.Result{}, err
	}

	expectedNum := int32(clusterManager.Spec.WorkerNum)
	if *md.Spec.Replicas != expectedNum {
		*md.Spec.Replicas = expectedNum
		if err := r.Update(context.TODO(), md); err != nil {
			log.Info("Failed to update machineDeployment")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: requeueAfter20Second}, nil
	}

	if md.Status.ReadyReplicas == *md.Spec.Replicas {
		log.Info("Worker scaling is completed successfully")
		clusterManager.Status.WorkerNum = clusterManager.Spec.WorkerNum
		return ctrl.Result{}, nil
	}
	if clusterManager.Spec.WorkerNum > clusterManager.Spec.WorkerNum {
		log.Info("Waiting for Worker nodes to be scaled out. Requeue after 20sec.")
	} else {
		log.Info("Waiting for Worker nodes to be scaled in. Requeue after 20sec.")
	}
	return ctrl.Result{RequeueAfter: requeueAfter20Second}, nil
}

// cluster를 upgrade한다. vsphere의 경우, serviceinstance 생성이 필요하다.
func (r *ClusterManagerReconciler) ClusterUpgrade(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (ctrl.Result, error) {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	log.Info("Start to reconcile phase for ClusterUpgrade")

	if clusterManager.Spec.Provider == clusterV1alpha1.ProviderVSphere {
		// service instance 체크
		clmSuffix, ok := clusterManager.Annotations[clusterV1alpha1.AnnotationKeyClmSuffix]
		if !ok {
			err := fmt.Errorf("Cannot find cluster suffix from cluster manager")
			log.Error(err, "failed to upgrade vsphere cluster")
			return ctrl.Result{}, err
		}
		serviceInstanceName := fmt.Sprintf("%s-%s-%s", clusterManager.Name, clmSuffix, clusterManager.Spec.Version)

		key := types.NamespacedName{
			Name:      serviceInstanceName,
			Namespace: clusterManager.Namespace,
		}

		serviceinstance := &servicecatalogv1beta1.ServiceInstance{}
		if err := r.Get(context.TODO(), key, serviceinstance); errors.IsNotFound(err) {
			log.Info("Waiting for vsphere upgrade service instance to be created")
			return ctrl.Result{RequeueAfter: requeueAfter10Second}, nil
		} else if err != nil {
			log.Error(err, "Failed to get service instance")
			return ctrl.Result{}, err
		}

		if serviceinstance.Status.ProvisionStatus != servicecatalogv1beta1.ServiceInstanceProvisionStatusProvisioned {
			log.Info("Waiting for vsphere upgrade service instance to be provisioned")
			return ctrl.Result{RequeueAfter: requeueAfter10Second}, nil
		}
	}

	// kcp 업데이트
	key := types.NamespacedName{
		Name:      clusterManager.Name + "-control-plane",
		Namespace: clusterManager.Namespace,
	}

	kcp := &controlplanev1.KubeadmControlPlane{}
	if err := r.Get(context.TODO(), key, kcp); errors.IsNotFound(err) {
		log.Error(err, "Cannot find kubeadmcontrolplane")
		return ctrl.Result{}, err
	} else if err != nil {
		log.Error(err, "Failed to get kubeadmcontrolplane")
		return ctrl.Result{}, err
	}

	// 단일 트랜잭션으로 업데이트 필요
	if kcp.Spec.Version != clusterManager.Spec.Version {
		kcp.Spec.Version = clusterManager.Spec.Version
		if clusterManager.Spec.Provider == clusterV1alpha1.ProviderVSphere {
			kcp.Spec.InfrastructureTemplate.Name = fmt.Sprintf("%s-%s", clusterManager.Name, clusterManager.Spec.Version)
		}
		if err := r.Update(context.TODO(), kcp); err != nil {
			log.Error(err, "Failed to update kubeadmcontrolplane")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: requeueAfter10Second}, nil
	}

	// upgrade 완료한 machine 찾기
	machines, err := r.GetUpgradeControlplaneMachines(clusterManager)
	if err != nil {
		log.Error(err, "Failed to list machines")
		return ctrl.Result{RequeueAfter: requeueAfter10Second}, nil
	}

	if len(machines.NewMachineRunningList) == clusterManager.Spec.MasterNum {
		log.Info(fmt.Sprintf("Controlplane nodes upgraded successfully (%d/%d)", len(machines.NewMachineRunningList), clusterManager.Spec.MasterNum))
	} else {
		log.Info(fmt.Sprintf("Controlplane nodes are upgrading (%d/%d)", len(machines.NewMachineRunningList), clusterManager.Spec.MasterNum))
		log.Info(fmt.Sprintf("Need to upgrade machine: [%s]. Requeue After 1 min", strings.Join(machines.OldMachineList, ", ")))
		return ctrl.Result{RequeueAfter: requeueAfter1Minute}, nil
	}

	// machineDeployment 업데이트
	key = types.NamespacedName{
		Name:      clusterManager.Name + "-md-0",
		Namespace: clusterManager.Namespace,
	}

	md := &capiV1alpha3.MachineDeployment{}
	if err := r.Get(context.TODO(), key, md); errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get machineDeployment")
		return ctrl.Result{}, err
	}

	if *md.Spec.Template.Spec.Version != clusterManager.Spec.Version {
		*md.Spec.Template.Spec.Version = clusterManager.Spec.Version
		if clusterManager.Spec.Provider == clusterV1alpha1.ProviderVSphere {
			md.Spec.Template.Spec.InfrastructureRef.Name = fmt.Sprintf("%s-%s", clusterManager.Name, clusterManager.Spec.Version)
		}
		if err := r.Update(context.TODO(), md); err != nil {
			log.Error(err, "Failed to update machinedeployment")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: requeueAfter10Second}, nil
	}

	// upgrade 완료한 machine 찾기
	machines, err = r.GetUpgradeWorkerMachines(clusterManager)
	if err != nil {
		log.Error(err, "Failed to list machines")
		return ctrl.Result{RequeueAfter: requeueAfter10Second}, nil
	}

	if len(machines.NewMachineRunningList) == clusterManager.Spec.WorkerNum {
		log.Info(fmt.Sprintf("worker nodes upgraded successfully (%d/%d)", len(machines.NewMachineRunningList), clusterManager.Spec.WorkerNum))
	} else {
		log.Info(fmt.Sprintf("worker nodes are upgrading (%d/%d)", len(machines.NewMachineRunningList), clusterManager.Spec.WorkerNum))
		log.Info(fmt.Sprintf("Need to upgrade machine: [%s]. Requeue After 1 min", strings.Join(machines.OldMachineList, ", ")))
		return ctrl.Result{RequeueAfter: requeueAfter1Minute}, nil
	}

	clusterManager.Status.Version = clusterManager.Spec.Version
	log.Info("Cluster upgradeded successfully")
	return ctrl.Result{}, nil
}

func (r *ClusterManagerReconciler) kubeadmControlPlaneUpdate(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (ctrl.Result, error) {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	log.Info("Start to reconcile phase for kubeadmControlPlaneUpdate")

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-control-plane",
		Namespace: clusterManager.Namespace,
	}
	kcp := &controlplanev1.KubeadmControlPlane{}
	if err := r.Get(context.TODO(), key, kcp); errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get kubeadmControlPlane")
		return ctrl.Result{}, err
	}

	//create helper for patch
	// helper, _ := patch.NewHelper(kcp, r.Client)
	// defer func() {
	// 	if err := helper.Patch(context.TODO(), kcp); err != nil {
	// 		r.Log.Error(err, "KubeadmControlPlane patch error")
	// 	}
	// }()

	// if *kcp.Spec.Replicas != int32(clusterManager.Spec.MasterNum) {
	// 	*kcp.Spec.Replicas = int32(clusterManager.Spec.MasterNum)
	// }

	clusterManager.Status.Ready = true
	return ctrl.Result{}, nil
}

func (r *ClusterManagerReconciler) machineDeploymentUpdate(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (ctrl.Result, error) {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	log.Info("Start to reconcile phase for machineDeploymentUpdate")

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-md-0",
		Namespace: clusterManager.Namespace,
	}
	md := &capiV1alpha3.MachineDeployment{}
	if err := r.Get(context.TODO(), key, md); errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get machineDeployment")
		return ctrl.Result{}, err
	}

	//create helper for patch
	// helper, _ := patch.NewHelper(md, r.Client)
	// defer func() {
	// 	if err := helper.Patch(context.TODO(), md); err != nil {
	// 		r.Log.Error(err, "machineDeployment patch error")
	// 	}
	// }()

	// if *md.Spec.Replicas != int32(clusterManager.Spec.WorkerNum) {
	// 	*md.Spec.Replicas = int32(clusterManager.Spec.WorkerNum)
	// }

	return ctrl.Result{}, nil
}

func (r *ClusterManagerReconciler) CreateArgocdResources(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (ctrl.Result, error) {
	if !clusterManager.Status.ControlPlaneReady || !clusterManager.Status.Ready || clusterManager.Status.ArgoReady {
		return ctrl.Result{}, nil
	}
	log := r.Log.WithValues("ClusterManager", clusterManager.GetNamespacedName())
	log.Info("Start to reconcile phase for CreateArgocdResources")

	kubeconfigSecret, err := r.GetKubeconfigSecret(clusterManager)
	if err != nil {
		log.Error(err, "Failed to get kubeconfig secret")
		return ctrl.Result{RequeueAfter: requeueAfter10Second}, nil
	}

	kubeConfig, err := clientcmd.Load(kubeconfigSecret.Data["value"])
	if err != nil {
		log.Error(err, "Failed to get kubeconfig data from secret")
		return ctrl.Result{}, err
	}

	remoteClientset, err := util.GetRemoteK8sClient(kubeconfigSecret)
	if err != nil {
		log.Error(err, "Failed to get remoteK8sClient")
		return ctrl.Result{}, err
	}

	// single cluster에서 secret을 조회
	// argocd-manager service account의 token을 얻기 위한 secret
	tokenSecret, err := remoteClientset.
		CoreV1().
		Secrets(util.KubeNamespace).
		Get(context.TODO(), util.ArgoServiceAccountTokenSecret, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Info("Service account secret not found. Wait for creating")
		return ctrl.Result{RequeueAfter: requeueAfter10Second}, nil
	} else if err != nil {
		log.Error(err, "Failed to get service account secret")
		return ctrl.Result{}, err
	}

	// token secret이 잘들어가 있는지 check
	if string(tokenSecret.Data["token"]) == "" {
		log.Info("Service account secret token data not found. Wait for creating")
		return ctrl.Result{Requeue: true}, nil
	}

	// ArgoCD single cluster 연동을 위한 secret에 들어가야 할 데이터를 생성
	configJson, err := json.Marshal(
		&argocdV1alpha1.ClusterConfig{
			BearerToken: string(tokenSecret.Data["token"]),
			TLSClientConfig: argocdV1alpha1.TLSClientConfig{
				Insecure: false,
				CAData:   kubeConfig.Clusters[kubeConfig.Contexts[kubeConfig.CurrentContext].Cluster].CertificateAuthorityData,
			},
		},
	)
	if err != nil {
		log.Error(err, "Failed to marshal cluster authorization parameters")
		return ctrl.Result{}, err
	}

	// master cluster에 secret 생성
	// ArgoCD에서 single cluster를 연동하기 위한 secret
	clusterName := strings.Split(kubeconfigSecret.Name, util.KubeconfigSuffix)[0]
	key := types.NamespacedName{
		Name:      kubeconfigSecret.Annotations[util.AnnotationKeyArgoClusterSecret],
		Namespace: util.ArgoNamespace,
	}
	argocdClusterSecret := &coreV1.Secret{}
	if err := r.Get(context.TODO(), key, argocdClusterSecret); errors.IsNotFound(err) {
		argocdClusterSecret = &coreV1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
				Annotations: map[string]string{
					util.AnnotationKeyOwner:         kubeconfigSecret.Annotations[util.AnnotationKeyOwner],
					util.AnnotationKeyCreator:       kubeconfigSecret.Annotations[util.AnnotationKeyCreator],
					util.AnnotationKeyArgoManagedBy: util.ArgoApiGroup,
				},
				Labels: map[string]string{
					util.LabelKeyClmSecretType:           util.ClmSecretTypeArgo,
					util.LabelKeyArgoSecretType:          util.ArgoSecretTypeCluster,
					clusterV1alpha1.LabelKeyClmName:      clusterManager.Name,
					clusterV1alpha1.LabelKeyClmNamespace: clusterManager.Namespace,
				},
				Finalizers: []string{
					clusterV1alpha1.ClusterManagerFinalizer,
				},
			},
			StringData: map[string]string{
				"config": string(configJson),
				"name":   clusterName,
				"server": kubeConfig.Clusters[kubeConfig.Contexts[kubeConfig.CurrentContext].Cluster].Server,
			},
		}
		if err := r.Create(context.TODO(), argocdClusterSecret); err != nil {
			log.Error(err, "Cannot create Argocd Secret for remote cluster")
			return ctrl.Result{}, err
		}
		log.Info("Create Argocd Secret for remote cluster successfully")
	} else if err != nil {
		log.Error(err, "Failed to get Argocd Secret for remote cluster")
		return ctrl.Result{}, err
	} else if !argocdClusterSecret.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{Requeue: true}, nil
	}

	if err := r.CreateApplication(clusterManager); err != nil {
		return ctrl.Result{}, err
	}

	argoIngress := &networkingV1.Ingress{}
	key = types.NamespacedName{
		Name:      util.ArgoIngressName,
		Namespace: util.ArgoNamespace,
	}
	if err := r.Get(context.TODO(), key, argoIngress); err != nil {
		log.Error(err, "Can not get argocd ingress information.")
	} else {
		subdomain := strings.Split(argoIngress.Spec.Rules[0].Host, ".")[0]
		SetApplicationLink(clusterManager, subdomain)
	}

	log.Info("Create argocd cluster secret successfully")
	clusterManager.Status.ArgoReady = true

	return ctrl.Result{}, nil
}

func (r *ClusterManagerReconciler) CreateGatewayResources(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (reconcile.Result, error) {
	if !clusterManager.Status.ArgoReady || clusterManager.Status.GatewayReady {
		return ctrl.Result{}, nil
	}
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	log.Info("Start to reconcile phase for CreateGatewayResources")

	kubeconfigSecret, err := r.GetKubeconfigSecret(clusterManager)
	if err != nil {
		log.Error(err, "Failed to get kubeconfig secret")
		return ctrl.Result{RequeueAfter: requeueAfter10Second}, nil
	}

	remoteClientset, err := util.GetRemoteK8sClient(kubeconfigSecret)
	if err != nil {
		log.Error(err, "Failed to get remoteK8sClient")
		return ctrl.Result{}, err
	}

	// host domain or ip를 얻기 위해 single cluster의
	// api-gateway-system 네임스페이스의 gateway service를 조회
	gatewayService, err := remoteClientset.
		CoreV1().
		Services(util.ApiGatewayNamespace).
		Get(context.TODO(), "gateway", metav1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Info("Cannot find Service for gateway. Wait for installing api-gateway. Requeue after 1 min")
		return ctrl.Result{RequeueAfter: requeueAfter1Minute}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service for gateway")
		return ctrl.Result{}, err
	}

	// single cluster의 gateway service가 LoadBalancer가 아닐 경우에는(시나리오상 NodePort일 경우)
	// k8s api-server의 endpoint도 NodePort로 되어있을 것이므로
	// k8s api-server의 domain host를 gateway service의 endpoint로 사용
	// single cluster의 k8s api-server domain과 gateway service의 domain중
	// 어떤 것을 이용해야 할지 앞의 로직에서 annotation key를 통해 전달
	annotationKey := clusterV1alpha1.AnnotationKeyClmApiserver
	if gatewayService.Spec.Type != coreV1.ServiceTypeNodePort {
		if gatewayService.Status.LoadBalancer.Ingress == nil {
			err := fmt.Errorf("service for gateway's type is not LoadBalancer or not ready")
			log.Error(err, "Service for api-gateway is not Ready. Requeue after 1 min")
			return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter1Minute}, err
		}

		ingress := gatewayService.Status.LoadBalancer.Ingress[0]
		hostnameOrIp := ingress.Hostname + ingress.IP
		if hostnameOrIp == "" {
			err := fmt.Errorf("service for gateway doesn't have both hostname and ip address")
			log.Error(err, "Service for api-gateway is not Ready. Requeue after 1 min")
			return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter1Minute}, err
		}

		clusterManager.Annotations[clusterV1alpha1.AnnotationKeyClmGateway] = hostnameOrIp
		annotationKey = clusterV1alpha1.AnnotationKeyClmGateway
	}

	// master cluster에 service 생성
	// single cluster의 gateway service로 연결시켜줄 external name type의 service
	// 앞에서 받은 annotation key를 이용하여 service의 endpoint가 설정 됨
	// ip address의 경우 k8s 기본 정책상으로는 endpoint resource로 생성하여 연결을 하는게 일반적인데
	// ip address도 external name type service의 external name의 value로 넣을 수 있기 때문에
	// 리소스 관리를 최소화 하기 위해 external name type으로 동일하게 생성
	if err := r.CreateGatewayService(clusterManager, annotationKey); err != nil {
		return ctrl.Result{}, err
	}

	// For migration from b5.0.26.6 > b5.0.26.7
	// 리소스 이름 및 status 이름 변경에 대응하기 위한 migration 코드
	// traefikReady, err := r.DeleteDeprecatedTraefikResources(clusterManager)
	// if err != nil {
	// 	return ctrl.Result{}, err
	// }
	// clusterManager.Status.TraefikReady = traefikReady

	if err := r.DeleteDeprecatedPrometheusResources(clusterManager); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Create gateway resources successfully")
	clusterManager.Status.GatewayReady = true
	return ctrl.Result{}, nil
}

func (r *ClusterManagerReconciler) CreateHyperAuthResources(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (reconcile.Result, error) {
	if !clusterManager.Status.GatewayReady || clusterManager.Status.AuthClientReady {
		return ctrl.Result{}, nil
	}
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	log.Info("Start to reconcile phase for CreateHyperauthClient")

	// Hyperauth의 password를 가져오기 위한 secret을 조회
	key := types.NamespacedName{
		Name:      "passwords",
		Namespace: "hyperauth",
	}
	secret := &coreV1.Secret{}
	if err := r.Get(context.TODO(), key, secret); errors.IsNotFound(err) {
		log.Info("Hyperauth password secret is not found")
		return ctrl.Result{}, err
	} else if err != nil {
		log.Error(err, "Failed to get hyperauth password secret")
		return ctrl.Result{}, err
	}

	// Hyperauth와 연동해야 하는 module 리스트는 정해져 있으므로, preset.go에서 관리
	// cluster마다 client 이름이 달라야 해서 {namespace}-{cluster name} 를 prefix로
	// 붙여주기로 했기 때문에, preset을 기본토대로 prefix를 추가하여 리턴하도록 구성
	// client 생성 (kibana, grafana, kiali, jaeger, hyperregistry, opensearch)
	clientConfigs := hyperauthCaller.GetClientConfigPreset(clusterManager.GetNamespacedPrefix())
	for _, config := range clientConfigs {
		if err := hyperauthCaller.CreateClient(config, secret); err != nil {
			log.Error(err, "Failed to create hyperauth client ["+config.ClientId+"] for single cluster")
			return ctrl.Result{RequeueAfter: requeueAfter10Second}, err
		}
	}

	// protocol mapper 생성 (kibana, jaeger, hyperregistry, opensearch)
	protocolMapperMappingConfigs := hyperauthCaller.GetMappingProtocolMapperToClientConfigPreset(clusterManager.GetNamespacedPrefix())
	for _, config := range protocolMapperMappingConfigs {
		if err := hyperauthCaller.CreateClientLevelProtocolMapper(config, secret); err != nil {
			log.Error(err, "Failed to create hyperauth protocol mapper ["+config.ClientId+"] for single cluster")
			return ctrl.Result{RequeueAfter: requeueAfter10Second}, err
		}
	}

	// client-level role을 생성하고 role에 cluster admin 계정을 mapping (kibana, jaeger, opensearch)
	clientLevelRoleConfigs := hyperauthCaller.GetClientLevelRoleConfigPreset(clusterManager.GetNamespacedPrefix())
	for _, config := range clientLevelRoleConfigs {
		if err := hyperauthCaller.CreateClientLevelRole(config, secret); err != nil {
			log.Error(err, "Failed to create hyperauth client-level role ["+config.ClientId+"] for single cluster")
			return ctrl.Result{RequeueAfter: requeueAfter10Second}, err
		}

		userEmail := clusterManager.Annotations[util.AnnotationKeyOwner]
		if err := hyperauthCaller.AddClientLevelRolesToUserRoleMapping(config, userEmail, secret); err != nil {
			log.Error(err, "Failed to add client-level role to user role mapping ["+config.ClientId+"] for single cluster")
			return ctrl.Result{RequeueAfter: requeueAfter10Second}, err
		}
	}

	// client와 client scope를 매핑 (kiali)
	clientScopeMappingConfig := hyperauthCaller.GetClientScopeMappingPreset(clusterManager.GetNamespacedPrefix())
	for _, config := range clientScopeMappingConfig {
		err := hyperauthCaller.AddClientScopeToClient(config, secret)
		if err != nil {
			log.Error(err, "Failed to add client scope to client ["+config.ClientId+"] for single cluster")
			return ctrl.Result{RequeueAfter: requeueAfter10Second}, err
		}
	}

	// group을 생성하고 cluster owner에게 group을 mapping
	groupConfig := hyperauthCaller.GetGroupConfigPreset(clusterManager.GetNamespacedPrefix())
	for _, config := range groupConfig {
		err := hyperauthCaller.CreateGroup(config, secret)
		if err != nil {
			log.Error(err, "Failed to create group ["+config.Name+"] for single cluster")
			return ctrl.Result{RequeueAfter: requeueAfter10Second}, err
		}

		err = hyperauthCaller.AddGroupToUser(clusterManager.Annotations[util.AnnotationKeyOwner], config, secret)
		if err != nil {
			log.Error(err, "Failed to add group to user ["+config.Name+"] for single cluster")
			return ctrl.Result{RequeueAfter: requeueAfter10Second}, err
		}
	}

	log.Info("Create clients for single cluster successfully")
	clusterManager.Status.AuthClientReady = true
	return ctrl.Result{}, nil
}

// func (r *ClusterManagerReconciler) SetHyperregistryOidcConfig(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (reconcile.Result, error) {
// 	if !clusterManager.Status.AuthClientReady || clusterManager.Status.HyperregistryOidcReady {
// 		return ctrl.Result{}, nil
// 	}
// 	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

// 	if os.Getenv("HYPERREGISTRY_ENABLED") == "false" {
// 		log.Info("Skip oidc config for hyperregistry")
// 		clusterManager.Status.HyperregistryOidcReady = true
// 		return ctrl.Result{}, nil
// 	}
// 	log.Info("Start to reconcile phase for SetHyperregistryOidcConfig")

// 	kubeconfigSecret, err := r.GetKubeconfigSecret(clusterManager)
// 	if err != nil {
// 		log.Error(err, "Failed to get kubeconfig secret")
// 		return ctrl.Result{RequeueAfter: requeueAfter10Second}, nil
// 	}

// 	remoteClientset, err := util.GetRemoteK8sClient(kubeconfigSecret)
// 	if err != nil {
// 		log.Error(err, "Failed to get remoteK8sClient")
// 		return ctrl.Result{}, err
// 	}

// 	// single cluster의 hyperregistry 계정정보를 조회하기 위해 secret을 조회
// 	secret, err := remoteClientset.
// 		CoreV1().
// 		Secrets(util.HyperregistryNamespace).
// 		Get(context.TODO(), "hyperregistry-harbor-core", metav1.GetOptions{})
// 	if err != nil {
// 		log.Error(err, "Failed to get Secret \"hyperregistry-harbor-core\"")
// 		return ctrl.Result{}, err
// 	}

// 	// single cluster의 hyperregistry 접속 주소를 조회하기 위해 ingress를 조회
// 	ingress, err := remoteClientset.
// 		NetworkingV1().
// 		Ingresses(util.HyperregistryNamespace).
// 		Get(context.TODO(), "hyperregistry-harbor-ingress", metav1.GetOptions{})
// 	if err != nil {
// 		log.Error(err, "Failed to get Ingress \"hyperregistry-harbor-ingress\"")
// 		return ctrl.Result{}, err
// 	}

// 	// hyperregistry의 경우 configmap이나 deploy의 env로 oidc 정보를 줄 수 없게 되어 있어서
// 	// http request를 생성하여 oidc 정보를 put 할 수 있도록 구현
// 	hyperauthDomain := "https://" + os.Getenv("AUTH_SUBDOMAIN") + "." + os.Getenv("HC_DOMAIN") + "/auth/realms/tmax"
// 	config := util.OidcConfig{
// 		AuthMode:         "oidc_auth",
// 		OidcAdminGroup:   "admin",
// 		OidcAutoOnBoard:  true,
// 		OidcClientId:     clusterManager.GetNamespacedPrefix() + "-hyperregistry",
// 		OidcClientSecret: os.Getenv("AUTH_CLIENT_SECRET"),
// 		OidcEndpoint:     hyperauthDomain,
// 		OidcGroupsClaim:  "group",
// 		OidcName:         "hyperregistry",
// 		OidcScope:        "openid",
// 		OidcUserClaim:    "preferred_username",
// 		OidcVerifyCert:   false,
// 	}
// 	hostpath := ingress.Spec.Rules[0].Host
// 	if err := util.SetHyperregistryOIDC(config, secret, hostpath); err != nil {
// 		log.Error(err, "Failed to set oidc configuration for hyperregistry")
// 		return ctrl.Result{}, err
// 	}

// 	log.Info("Set oidc config for hyperregistry successfully")
// 	clusterManager.Status.HyperregistryOidcReady = true
// 	return ctrl.Result{}, nil
// }

func (r *ClusterManagerReconciler) CreateTraefikResources(ctx context.Context, clusterManager *clusterV1alpha1.ClusterManager) (ctrl.Result, error) {
	if !clusterManager.Status.AuthClientReady || clusterManager.Status.TraefikReady {
		return ctrl.Result{}, nil
	}
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	log.Info("Start to reconcile phase for CreateTraefikResources")

	if err := r.CreateCertificate(clusterManager); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.CreateMiddleware(clusterManager); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.CreateServiceAccountSecret(clusterManager); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.CreateIngress(clusterManager); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Create traefik resources successfully")
	clusterManager.Status.TraefikReady = true
	return ctrl.Result{}, nil
}
