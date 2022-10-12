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
	"fmt"
	"os"
	"regexp"
	"strings"

	argocdV1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	certmanagerV1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmanagerMetaV1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	clusterV1alpha1 "github.com/tmax-cloud/hypercloud-multi-operator/apis/cluster/v1alpha1"
	hyperauthCaller "github.com/tmax-cloud/hypercloud-multi-operator/controllers/hyperAuth"
	util "github.com/tmax-cloud/hypercloud-multi-operator/controllers/util"
	dynamicv2 "github.com/traefik/traefik/v2/pkg/config/dynamic"
	traefikV1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"

	coreV1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func SetApplicationLink(c *clusterV1alpha1.ClusterManager, subdomain string) {
	c.Status.ApplicationLink = strings.Join(
		[]string{
			"https://",
			subdomain,
			".",
			os.Getenv(util.HC_DOMAIN),
			"/applications/",
			c.GetNamespacedPrefix(),
			"-applications?node=argoproj.io/Application/argocd/",
			c.GetNamespacedPrefix(),
			"-applications/0&resource=",
			"",
		},
		"",
	)
}

func (r *ClusterManagerReconciler) GetKubeconfigSecret(clusterManager *clusterV1alpha1.ClusterManager) (*coreV1.Secret, error) {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.Name + util.KubeconfigSuffix,
		Namespace: clusterManager.Namespace,
	}
	kubeconfigSecret := &coreV1.Secret{}
	if err := r.Get(context.TODO(), key, kubeconfigSecret); errors.IsNotFound(err) {
		log.Info("kubeconfig secret is not found")
		return nil, err
	} else if err != nil {
		log.Error(err, "Failed to get kubeconfig secret")
		return nil, err

	}
	return kubeconfigSecret, nil
}

func (r *ClusterManagerReconciler) CreateCertificate(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-certificate",
		Namespace: clusterManager.Namespace,
	}
	err := r.Get(context.TODO(), key, &certmanagerV1.Certificate{})
	if errors.IsNotFound(err) {
		certificate := &certmanagerV1.Certificate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
				Annotations: map[string]string{
					util.AnnotationKeyOwner:   clusterManager.Annotations[util.AnnotationKeyCreator],
					util.AnnotationKeyCreator: clusterManager.Annotations[util.AnnotationKeyCreator],
				},
				Labels: map[string]string{
					clusterV1alpha1.LabelKeyClmName: clusterManager.Name,
				},
			},
			Spec: certmanagerV1.CertificateSpec{
				SecretName: clusterManager.Name + "-service-cert",
				IsCA:       false,
				Usages: []certmanagerV1.KeyUsage{
					certmanagerV1.UsageDigitalSignature,
					certmanagerV1.UsageKeyEncipherment,
					certmanagerV1.UsageServerAuth,
					certmanagerV1.UsageClientAuth,
				},
				DNSNames: []string{
					"multicluster." + clusterManager.Annotations[clusterV1alpha1.AnnotationKeyClmDomain],
				},
				IssuerRef: certmanagerMetaV1.ObjectReference{
					Name:  "tmaxcloud-issuer",
					Kind:  certmanagerV1.ClusterIssuerKind,
					Group: certmanagerV1.SchemeGroupVersion.Group,
				},
			},
		}
		if err := r.Create(context.TODO(), certificate); err != nil {
			log.Error(err, "Failed to Create Certificate")
			return err
		}

		log.Info("Create Certificate successfully")
		ctrl.SetControllerReference(clusterManager, certificate, r.Scheme)
		return nil
	}

	return err
}

func (r *ClusterManagerReconciler) CreateIngress(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-ingress",
		Namespace: clusterManager.Namespace,
	}
	err := r.Get(context.TODO(), key, &networkingv1.Ingress{})
	if errors.IsNotFound(err) {
		provider := "tmax-cloud"
		pathType := networkingv1.PathTypePrefix
		prefixMiddleware := clusterManager.GetNamespacedPrefix() + "-prefix@kubernetescrd"
		multiclusterDNS := "multicluster." + clusterManager.Annotations[clusterV1alpha1.AnnotationKeyClmDomain]
		urlPath := "/api/" + clusterManager.Namespace + "/" + clusterManager.Name
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
				Annotations: map[string]string{
					util.AnnotationKeyTraefikEntrypoints: "websecure",
					util.AnnotationKeyTraefikMiddlewares: "api-gateway-system-jwt-decode-auth@kubernetescrd," + prefixMiddleware,
					util.AnnotationKeyOwner:              clusterManager.Annotations[util.AnnotationKeyCreator],
					util.AnnotationKeyCreator:            clusterManager.Annotations[util.AnnotationKeyCreator],
				},
				Labels: map[string]string{
					util.LabelKeyHypercloudIngress:  "multicluster",
					clusterV1alpha1.LabelKeyClmName: clusterManager.Name,
				},
			},
			Spec: networkingv1.IngressSpec{
				IngressClassName: &provider,
				Rules: []networkingv1.IngressRule{
					{
						Host: multiclusterDNS,
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										Path:     urlPath + "/api/kubernetes",
										PathType: &pathType,
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: clusterManager.Name + "-gateway-service",
												Port: networkingv1.ServiceBackendPort{
													Number: 443,
												},
											},
										},
									},
									{
										Path:     urlPath + "/api/prometheus",
										PathType: &pathType,
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: clusterManager.Name + "-gateway-service",
												Port: networkingv1.ServiceBackendPort{
													Number: 443,
												},
											},
										},
									},
								},
							},
						},
					},
				},
				TLS: []networkingv1.IngressTLS{
					{
						Hosts: []string{
							multiclusterDNS,
						},
					},
				},
			},
		}
		if err := r.Create(context.TODO(), ingress); err != nil {
			log.Error(err, "Failed to Create Ingress")
			return err
		}

		log.Info("Create Ingress successfully")
		ctrl.SetControllerReference(clusterManager, ingress, r.Scheme)
		return nil
	}

	return err
}

func (r *ClusterManagerReconciler) CreateGatewayService(clusterManager *clusterV1alpha1.ClusterManager, annotationKey string) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-gateway-service",
		Namespace: clusterManager.Namespace,
	}
	err := r.Get(context.TODO(), key, &coreV1.Service{})
	if errors.IsNotFound(err) {
		service := &coreV1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
				Annotations: map[string]string{
					util.AnnotationKeyOwner:                  clusterManager.Annotations[util.AnnotationKeyCreator],
					util.AnnotationKeyCreator:                clusterManager.Annotations[util.AnnotationKeyCreator],
					util.AnnotationKeyTraefikServerScheme:    "https",
					util.AnnotationKeyTraefikServerTransport: "insecure@file",
				},
				Labels: map[string]string{
					clusterV1alpha1.LabelKeyClmName: clusterManager.Name,
				},
			},
			Spec: coreV1.ServiceSpec{
				ExternalName: clusterManager.Annotations[annotationKey],
				Ports: []coreV1.ServicePort{
					{
						Port:       443,
						Protocol:   coreV1.ProtocolTCP,
						TargetPort: intstr.FromInt(443),
					},
				},
				Type: coreV1.ServiceTypeExternalName,
			},
		}
		if err := r.Create(context.TODO(), service); err != nil {
			log.Error(err, "Failed to Create Service for gateway")
			return err
		}

		log.Info("Create Service for gateway successfully")
		ctrl.SetControllerReference(clusterManager, service, r.Scheme)
		return nil
	}

	return err
}

func (r *ClusterManagerReconciler) CreateGatewayEndpoint(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-gateway-service",
		Namespace: clusterManager.Namespace,
	}
	err := r.Get(context.TODO(), key, &coreV1.Endpoints{})
	if errors.IsNotFound(err) {
		endpoint := &coreV1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
				Annotations: map[string]string{
					util.AnnotationKeyOwner:   clusterManager.Annotations[util.AnnotationKeyCreator],
					util.AnnotationKeyCreator: clusterManager.Annotations[util.AnnotationKeyCreator],
				},
				Labels: map[string]string{
					clusterV1alpha1.LabelKeyClmName: clusterManager.Name,
				},
			},
			Subsets: []coreV1.EndpointSubset{
				{
					Addresses: []coreV1.EndpointAddress{
						{
							IP: clusterManager.Annotations[clusterV1alpha1.AnnotationKeyClmGateway],
						},
					},
					Ports: []coreV1.EndpointPort{
						{
							Port:     443,
							Protocol: coreV1.ProtocolTCP,
						},
					},
				},
			},
		}
		if err := r.Create(context.TODO(), endpoint); err != nil {
			log.Error(err, "Failed to Create Endpoint for gateway")
			return err
		}

		log.Info("Create Endpoint for gateway successfully")
		ctrl.SetControllerReference(clusterManager, endpoint, r.Scheme)
		return nil
	}

	return err
}

func (r *ClusterManagerReconciler) CreateMiddleware(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-prefix",
		Namespace: clusterManager.Namespace,
	}
	err := r.Get(context.TODO(), key, &traefikV1alpha1.Middleware{})
	if errors.IsNotFound(err) {
		middleware := &traefikV1alpha1.Middleware{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
				Annotations: map[string]string{
					util.AnnotationKeyOwner:   clusterManager.Annotations[util.AnnotationKeyCreator],
					util.AnnotationKeyCreator: clusterManager.Annotations[util.AnnotationKeyCreator],
				},
				Labels: map[string]string{
					clusterV1alpha1.LabelKeyClmName: clusterManager.Name,
				},
			},
			Spec: traefikV1alpha1.MiddlewareSpec{
				StripPrefix: &dynamicv2.StripPrefix{
					Prefixes: []string{
						"/api/" + clusterManager.Namespace + "/" + clusterManager.Name,
					},
				},
			},
		}
		if err := r.Create(context.TODO(), middleware); err != nil {
			log.Error(err, "Failed to Create Middleware")
			return err
		}

		log.Info("Create Middleware successfully")
		ctrl.SetControllerReference(clusterManager, middleware, r.Scheme)
		return nil
	}

	return err
}

func (r *ClusterManagerReconciler) CreateServiceAccountSecret(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	re, _ := regexp.Compile("[" + regexp.QuoteMeta(`!#$%&'"*+-/=?^_{|}~().,:;<>[]\`) + "`\\s" + "]")
	email := clusterManager.Annotations[util.AnnotationKeyOwner]
	adminServiceAccountName := re.ReplaceAllString(strings.Replace(email, "@", "-at-", -1), "-")
	kubeconfigSecret, err := r.GetKubeconfigSecret(clusterManager)
	if err != nil {
		log.Error(err, "Failed to get kubeconfig secret")
		return err
	}

	remoteClientset, err := util.GetRemoteK8sClient(kubeconfigSecret)
	if err != nil {
		log.Error(err, "Failed to get remoteK8sClient")
		return err
	}

	tokenSecret, err := remoteClientset.
		CoreV1().
		Secrets(util.KubeNamespace).
		Get(context.TODO(), adminServiceAccountName+"-token", metav1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Info("Waiting for create service account token secret [" + adminServiceAccountName + "]")
		return err
	} else if err != nil {
		log.Error(err, "Failed to get service account token secret ["+adminServiceAccountName+"-token]")
		return err
	}

	if string(tokenSecret.Data["token"]) == "" {
		log.Info("Waiting for create service account token secret [" + adminServiceAccountName + "]")
		return fmt.Errorf("service account token secret is not found")
	}

	jwtDecodeSecretName := adminServiceAccountName + "-" + clusterManager.Name + "-token"
	key := types.NamespacedName{
		Name:      jwtDecodeSecretName,
		Namespace: clusterManager.Namespace,
	}
	jwtDecodeSecret := &coreV1.Secret{}
	err = r.Get(context.TODO(), key, jwtDecodeSecret)
	if errors.IsNotFound(err) {
		secret := &coreV1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
				Labels: map[string]string{
					util.LabelKeyClmSecretType:           util.ClmSecretTypeSAToken,
					clusterV1alpha1.LabelKeyClmName:      clusterManager.Name,
					clusterV1alpha1.LabelKeyClmNamespace: clusterManager.Namespace,
				},
				Annotations: map[string]string{
					util.AnnotationKeyOwner: clusterManager.Annotations[util.AnnotationKeyOwner],
				},
				Finalizers: []string{
					clusterV1alpha1.ClusterManagerFinalizer,
				},
			},
			Data: map[string][]byte{
				"token": tokenSecret.Data["token"],
			},
		}
		if err := r.Create(context.TODO(), secret); err != nil {
			log.Error(err, "Failed to Create Secret for ServiceAccount token")
			return err
		}

		log.Info("Create Secret for ServiceAccount token successfully")
		ctrl.SetControllerReference(clusterManager, secret, r.Scheme)
		return nil
	}

	if !jwtDecodeSecret.DeletionTimestamp.IsZero() {
		err = fmt.Errorf("secret for service account token is not refreshed yet")
	}

	return err
}

func (r *ClusterManagerReconciler) CreateApplication(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.GetNamespacedPrefix() + "-applications",
		Namespace: util.ArgoNamespace,
	}
	err := r.Get(context.TODO(), key, &argocdV1alpha1.Application{})
	if errors.IsNotFound(err) {
		application := &argocdV1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
				Labels: map[string]string{
					util.LabelKeyArgoTargetCluster: clusterManager.GetNamespacedPrefix(),
					util.LabelKeyArgoAppType:       util.ArgoAppTypeAppOfApp,
				},
			},
			Spec: argocdV1alpha1.ApplicationSpec{
				Destination: argocdV1alpha1.ApplicationDestination{
					Namespace: util.ArgoNamespace,
					Server:    argocdV1alpha1.KubernetesInternalAPIServerAddr,
				},
				Project: argocdV1alpha1.DefaultAppProjectName,
				Source: argocdV1alpha1.ApplicationSource{
					Helm: &argocdV1alpha1.ApplicationSourceHelm{
						ValueFiles: []string{
							"shared-values.yaml",
							"single-values.yaml",
						},
						Parameters: []argocdV1alpha1.HelmParameter{
							{
								Name:  "global.clusterName",
								Value: clusterManager.Name,
							},
							{
								Name:  "global.clusterNamespace",
								Value: clusterManager.Namespace,
							},
							{
								Name:  "global.privateRegistry",
								Value: util.ArgoDescriptionPrivateRegistry,
							},
							{
								Name:  "global.adminUser",
								Value: clusterManager.Annotations[util.AnnotationKeyOwner],
							},
							{
								Name:  "modules.gatewayBootstrap.console.subdomain",
								Value: util.ArgoDescriptionConsoleSubdomain,
							},
							{
								Name:  "modules.hyperAuth.subdomain",
								Value: util.ArgoDescriptionHyperAuthSubdomain,
							},
							{
								Name:  "modules.efk.kibana.subdomain",
								Value: util.ArgoDescriptionKibanaSubdomain,
							},
							{
								Name:  "modules.grafana.subdomain",
								Value: util.ArgoDescriptionGrafanaSubdomain,
							},
							{
								Name:  "modules.serviceMesh.jaeger.subdomain",
								Value: util.ArgoDescriptionJaegerSubdomain,
							},
							{
								Name:  "modules.serviceMesh.kiali.subdomain",
								Value: util.ArgoDescriptionKialiSubdomain,
							},
							{
								Name:  "modules.cicd.subdomain",
								Value: util.ArgoDescriptionCicdSubdomain,
							},
							{
								Name:  "modules.opensearch.dashboard.subdomain",
								Value: util.ArgoDescriptionOpensearchSubdomain,
							},
							{
								Name:  "modules.hyperregistry.core.subdomain",
								Value: util.ArgoDescriptionHyperregistrySubdomain,
							},
							{
								Name:  "modules.hyperregistry.notary.subdomain",
								Value: util.ArgoDescriptionHyperregistryNotarySubdomain,
							},
							{
								Name:  "modules.hyperregistry.storageClass",
								Value: util.ArgoDescriptionHyperregistryStorageClass,
							},
							{
								Name:  "modules.hyperregistry.storageClassDatabase",
								Value: util.ArgoDescriptionHyperregistryDBStorageClass,
							},
						},
					},
					Path:           "application/helm",
					RepoURL:        util.ArgoDescriptionGitRepo,
					TargetRevision: util.ArgoDescriptionGitRevision,
				},
			},
		}
		if err := r.Create(context.TODO(), application); err != nil {
			log.Error(err, "Failed to Create ArgoCD Application")
			return err
		}

		log.Info("Create ArgoCD Application successfully")
		return nil
	}

	return err
}

func (r *ClusterManagerReconciler) DeleteCertificate(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-certificate",
		Namespace: clusterManager.Namespace,
	}
	certificate := &certmanagerV1.Certificate{}
	err := r.Get(context.TODO(), key, certificate)
	if errors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		log.Error(err, "Failed to get Certificate")
		return err
	}

	if err := r.Delete(context.TODO(), certificate); err != nil {
		log.Error(err, "Failed to delete Certificate")
		return err
	}

	log.Info("Delete Certificate successfully")
	return nil
}

func (r *ClusterManagerReconciler) DeleteCertSecret(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-service-cert",
		Namespace: clusterManager.Namespace,
	}
	secret := &coreV1.Secret{}
	err := r.Get(context.TODO(), key, secret)
	if errors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		log.Error(err, "Failed to get Secret for certificate")
		return err
	}

	if err := r.Delete(context.TODO(), secret); err != nil {
		log.Error(err, "Failed to delete Secret for certificate")
		return err
	}

	log.Info("Delete Secret for certificate successfully")
	return nil
}

func (r *ClusterManagerReconciler) DeleteIngress(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-ingress",
		Namespace: clusterManager.Namespace,
	}
	ingress := &networkingv1.Ingress{}
	err := r.Get(context.TODO(), key, ingress)
	if errors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		log.Error(err, "Failed to get Ingress")
		return err
	}

	if err := r.Delete(context.TODO(), ingress); err != nil {
		log.Error(err, "Failed to delete Ingress")
		return err
	}

	log.Info("Delete Ingress successfully")
	return nil
}

func (r *ClusterManagerReconciler) DeleteService(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-service",
		Namespace: clusterManager.Namespace,
	}
	service := &coreV1.Service{}
	err := r.Get(context.TODO(), key, service)
	if errors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		log.Error(err, "Failed to get Service")
		return err
	}

	if err := r.Delete(context.TODO(), service); err != nil {
		log.Error(err, "Failed to delete Service")
		return err
	}

	log.Info("Delete Service successfully")
	return nil
}

func (r *ClusterManagerReconciler) DeleteEndpoint(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-service",
		Namespace: clusterManager.Namespace,
	}
	endpoint := &coreV1.Endpoints{}
	err := r.Get(context.TODO(), key, endpoint)
	if errors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		log.Error(err, "Failed to get Endpoint")
		return err
	}

	if err := r.Delete(context.TODO(), endpoint); err != nil {
		log.Error(err, "Failed to delete Endpoint")
		return err
	}

	log.Info("Delete Endpoint successfully")
	return nil
}

func (r *ClusterManagerReconciler) DeleteMiddleware(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-prefix",
		Namespace: clusterManager.Namespace,
	}
	middleware := &traefikV1alpha1.Middleware{}
	err := r.Get(context.TODO(), key, middleware)
	if errors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		log.Error(err, "Failed to get Middleware")
		return err
	}

	if err := r.Delete(context.TODO(), middleware); err != nil {
		log.Error(err, "Failed to delete Middleware")
		return err
	}

	log.Info("Delete Middleware successfully")
	return nil
}

func (r *ClusterManagerReconciler) DeleteGatewayService(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-gateway-service",
		Namespace: clusterManager.Namespace,
	}
	service := &coreV1.Service{}
	err := r.Get(context.TODO(), key, service)
	if errors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		log.Error(err, "Failed to get Service")
		return err
	}

	if err := r.Delete(context.TODO(), service); err != nil {
		log.Error(err, "Failed to delete Service")
		return err
	}

	log.Info("Delete Service successfully")
	return nil
}

func (r *ClusterManagerReconciler) DeleteGatewayEndpoint(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	key := types.NamespacedName{
		Name:      clusterManager.Name + "-gateway-service",
		Namespace: clusterManager.Namespace,
	}
	endpoint := &coreV1.Endpoints{}
	err := r.Get(context.TODO(), key, endpoint)
	if errors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		log.Error(err, "Failed to get Endpoint")
		return err
	}

	if err := r.Delete(context.TODO(), endpoint); err != nil {
		log.Error(err, "Failed to delete Endpoint")
		return err
	}

	log.Info("Delete Endpoint successfully")
	return nil
}

func (r *ClusterManagerReconciler) DeleteDeprecatedTraefikResources(clusterManager *clusterV1alpha1.ClusterManager) (bool, error) {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	ready := true
	key := types.NamespacedName{
		Name:      clusterManager.Name + "-ingress",
		Namespace: clusterManager.Namespace,
	}
	ingress := &networkingv1.Ingress{}
	if err := r.Get(context.TODO(), key, ingress); errors.IsNotFound(err) {
		log.Info("Not found: " + key.Name)
	} else if err != nil {
		log.Error(err, "Failed to get: "+key.Name)
		return ready, err
	} else {
		if err := r.Delete(context.TODO(), ingress); err != nil {
			log.Error(err, "Failed to delete: "+key.Name)
			return ready, err
		}
		ready = false
	}

	key = types.NamespacedName{
		Name:      clusterManager.Name + "-service",
		Namespace: clusterManager.Namespace,
	}
	service := &coreV1.Service{}
	if err := r.Get(context.TODO(), key, service); errors.IsNotFound(err) {
		log.Info("Not found: " + key.Name)
	} else if err != nil {
		log.Error(err, "Failed to get: "+key.Name)
		return ready, err
	} else {
		if err := r.Delete(context.TODO(), service); err != nil {
			log.Error(err, "Failed to delete: "+key.Name)
			return ready, err
		}
		ready = false
	}

	endpoint := &coreV1.Endpoints{}
	if err := r.Get(context.TODO(), key, endpoint); errors.IsNotFound(err) {
		log.Info("Not found: " + key.Name)
	} else if err != nil {
		log.Error(err, "Failed to get: "+key.Name)
		return ready, err
	} else {
		if err := r.Delete(context.TODO(), endpoint); err != nil {
			log.Error(err, "Failed to delete: "+key.Name)
			return ready, err
		}
		ready = false
	}

	return ready, nil
}

func (r *ClusterManagerReconciler) DeleteDeprecatedPrometheusResources(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	key := types.NamespacedName{
		Name:      clusterManager.Name + "-prometheus-service",
		Namespace: clusterManager.Namespace,
	}
	service := &coreV1.Service{}
	if err := r.Get(context.TODO(), key, service); errors.IsNotFound(err) {
		log.Info("Not found: " + key.Name)
	} else if err != nil {
		log.Error(err, "Failed to get: "+key.Name)
		return err
	} else {
		if err := r.Delete(context.TODO(), service); err != nil {
			log.Error(err, "Failed to delete: "+key.Name)
			return err
		}
	}

	endpoint := &coreV1.Endpoints{}
	if err := r.Get(context.TODO(), key, endpoint); errors.IsNotFound(err) {
		log.Info("Not found: " + key.Name)
	} else if err != nil {
		log.Error(err, "Failed to get: "+key.Name)
		return err
	} else {
		if err := r.Delete(context.TODO(), endpoint); err != nil {
			log.Error(err, "Failed to delete: "+key.Name)
			return err
		}
	}

	return nil
}

func (r *ClusterManagerReconciler) CheckApplicationRemains(clusterManager *clusterV1alpha1.ClusterManager) error {
	appList := &argocdV1alpha1.ApplicationList{}
	if err := r.List(context.TODO(), appList); err != nil {
		return err
	}
	for _, app := range appList.Items {
		if app.Labels[util.LabelKeyArgoTargetCluster] == clusterManager.GetNamespacedPrefix() {
			return fmt.Errorf("application still remains")
		}
	}

	return nil
}

func (r *ClusterManagerReconciler) DeleteLoadBalancerServices(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	kubeconfigSecret, err := r.GetKubeconfigSecret(clusterManager)
	if errors.IsNotFound(err) {
		log.Info("Cluster is already deleted")
		return nil
	} else if err != nil {
		log.Error(err, "Failed to get kubeconfig secret")
		return err
	}

	remoteClientset, err := util.GetRemoteK8sClient(kubeconfigSecret)
	if err != nil {
		log.Error(err, "Failed to get remoteK8sClient")
		return err
	}

	if _, err := remoteClientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{}); err != nil {
		log.Info("Failed to get node for remote cluster. Skip delete LoadBalancer services process")
		return nil
	}

	nsList, err := remoteClientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error(err, "Failed to list namespaces")
		return err
	}

	for _, ns := range nsList.Items {
		if ns.Name == util.KubeNamespace {
			continue
		}

		svcList, err := remoteClientset.CoreV1().Services(ns.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error(err, "Failed to list services in namespace ["+ns.Name+"]")
			return err
		}

		for _, svc := range svcList.Items {
			if svc.Spec.Type != coreV1.ServiceTypeLoadBalancer {
				continue
			}

			delErr := remoteClientset.CoreV1().Services(ns.Name).Delete(context.TODO(), svc.Name, metav1.DeleteOptions{})
			if delErr != nil {
				log.Error(err, "Failed to delete service ["+svc.Name+"]in namespace ["+ns.Name+"]")
				return err
			}
		}
	}

	log.Info("Delete LoadBalancer services in single cluster successfully")
	return nil
}

func (r *ClusterManagerReconciler) DeleteTraefikResources(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())

	if err := r.DeleteCertificate(clusterManager); err != nil {
		return err
	}

	if err := r.DeleteCertSecret(clusterManager); err != nil {
		return err
	}

	if err := r.DeleteIngress(clusterManager); err != nil {
		return err
	}

	if err := r.DeleteMiddleware(clusterManager); err != nil {
		return err
	}

	if err := r.DeleteGatewayService(clusterManager); err != nil {
		return err
	}

	if err := r.DeleteGatewayEndpoint(clusterManager); err != nil {
		return err
	}

	log.Info("Delete traefik resources successfully")
	return nil
}

func (r *ClusterManagerReconciler) DeleteHyperAuthResourcesForSingleCluster(clusterManager *clusterV1alpha1.ClusterManager) error {
	log := r.Log.WithValues("clustermanager", clusterManager.GetNamespacedName())
	key := types.NamespacedName{
		Name:      "passwords",
		Namespace: "hyperauth",
	}
	secret := &coreV1.Secret{}
	if err := r.Get(context.TODO(), key, secret); errors.IsNotFound(err) {
		log.Info("HyperAuth password secret is not found")
		return err
	} else if err != nil {
		log.Error(err, "Failed to get HyperAuth password secret")
		return err
	}

	clientConfigs := hyperauthCaller.GetClientConfigPreset(clusterManager.GetNamespacedPrefix())
	for _, config := range clientConfigs {
		err := hyperauthCaller.DeleteClient(config, secret)
		if err != nil {
			log.Error(err, "Failed to delete HyperAuth client ["+config.ClientId+"] for single cluster")
			return err
		}
	}

	groupConfigs := hyperauthCaller.GetGroupConfigPreset(clusterManager.GetNamespacedPrefix())
	for _, config := range groupConfigs {
		err := hyperauthCaller.DeleteGroup(config, secret)
		if err != nil {
			log.Error(err, "Failed to delete HyperAuth group ["+config.Name+"] for single cluster")
			return err
		}
	}

	log.Info("Delete HyperAuth resources for single cluster successfully")
	return nil
}
