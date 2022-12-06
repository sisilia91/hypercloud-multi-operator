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

package util

const (
	KubeNamespace          = "kube-system"
	ApiGatewayNamespace    = "api-gateway-system"
	IngressNginxNamespace  = "ingress-nginx"
	ArgoNamespace          = "argocd"
	HyperregistryNamespace = "hyperregistry"
	OpenSearchNamespace    = "kube-logging"
)

const (
	// defunct
	// MultiApiServerServiceSelectorKey   = "hypercloud4"
	// MultiApiServerServiceSelectorValue = "multi-api-server"
	IngressNginxName = "ingress-nginx-controller"
)

const (
	KubeconfigSuffix = "-kubeconfig"
	// HypercloudIngressClass          = "tmax-cloud"
	// HypercloudMultiIngressClass     = "multicluster"
	// HypercloudMultiIngressSubdomain = "multicluster"
)

const (
	ProviderAws     = "AWS"
	ProviderAwsLogo = "AWS"

	ProviderVsphere     = "VSPHERE"
	ProviderVsphereLogo = "vSphere"

	ProviderUnknown = "Unknown"
)

const (
	ClmSecretTypeKubeconfig = "kubeconfig"
	ClmSecretTypeArgo       = "argocd"
	ClmSecretTypeSAToken    = "token"
)

const (
	ArgoApiGroup                  = "argocd.argoproj.io"
	ArgoServiceAccount            = "argocd-manager"
	ArgoServiceAccountTokenSecret = "argocd-manager-token"
	ArgoClusterRole               = "argocd-manager-role"
	ArgoClusterRoleBinding        = "argocd-manager-role-binding"
	ArgoSecretTypeCluster         = "cluster"
	ArgoSecretTypeGit             = "repo"
	ArgoAppTypeAppOfApp           = "app-of-apps"
	ArgoIngressName               = "argocd-server-ingress"
)

const (
	AnnotationKeyOwner   = "owner"
	AnnotationKeyCreator = "creator"

	AnnotationKeyArgoClusterSecret = "argocd.argoproj.io/cluster.secret"
	AnnotationKeyArgoManagedBy     = "managed-by"

	AnnotationKeyTraefikServerTransport = "traefik.ingress.kubernetes.io/service.serverstransport"
	AnnotationKeyTraefikEntrypoints     = "traefik.ingress.kubernetes.io/router.entrypoints"
	AnnotationKeyTraefikMiddlewares     = "traefik.ingress.kubernetes.io/router.middlewares"
	AnnotationKeyTraefikServerScheme    = "traefik.ingress.kubernetes.io/service.serversscheme"
)

const (
	LabelKeyHypercloudIngress = "ingress.tmaxcloud.org/name"

	LabelKeyClmSecretType = "cluster.tmax.io/clm-secret-type"

	LabelKeyArgoSecretType  = "argocd.argoproj.io/secret-type"
	LabelKeyCapiClusterName = "cluster.x-k8s.io/cluster-name"

	// LabelKeyArgoTargetCluster = "cluster.tmax.io/cluster"
	LabelKeyArgoTargetCluster = "cluster"
	LabelKeyArgoAppType       = "appType"
)

const (
	HARBOR_SERVICE_SET_OIDC_CONFIG = "/api/v2.0/configurations"
)

const (
	ArgoDescriptionGlobalDomain                 = "global domain으로 변경 ex) tmaxcloud.org"
	ArgoDescriptionPrivateRegistry              = "target registry 주소로 변경"
	ArgoDescriptionConsoleSubdomain             = "Console의 Subdomain으로 변경 ex) console"
	ArgoDescriptionHyperAuthSubdomain           = "HyperAuth의 Full domain으로 변경 ex) hyperauth.tmaxcloud.org"
	ArgoDescriptionKibanaSubdomain              = "Kibana의 Subdomain으로 변경 ex) kibana"
	ArgoDescriptionGrafanaOperatorSubdomain     = "Grafana의 Subdomain으로 변경 ex) grafana"
	ArgoDescriptionJaegerSubdomain              = "Jaeger의 Subdomain으로 변경 ex) jaeger"
	ArgoDescriptionKialiSubdomain               = "Kiali의 Subdomain으로 변경 ex) kiali"
	ArgoDescriptionCicdSubdomain                = "Cicd webhook의 Subdomain으로 변경 ex) cicd"
	ArgoDescriptionOpensearchSubdomain          = "Opensearch dashboard의 Subdomain으로 변경 ex) opensearch"
	ArgoDescriptionHyperregistrySubdomain       = "Hyperregistry-core의 Subdomain으로 변경 ex) hyperregistry"
	ArgoDescriptionHyperregistryNotarySubdomain = "Hyperregistry-notary의 Subdomain으로 변경 ex) notary"
	ArgoDescriptionHelmApiServerSubdomain       = "Helm api server의 Subdomain으로 변경 ex) helm"
	ArgoDescriptionHyperregistryStorageClass    = "Hyperregistry가 사용할 StorageClass로 변경(aws의 경우 efs-sc-0, 그외에는 nfs)"
	ArgoDescriptionHyperregistryDBStorageClass  = "Hyperregistry의 DB가 사용할 StorageClass로 변경(aws의 경우 efs-sc-999, 그외에는 nfs)"
	ArgoDescriptionGitRepo                      = "Git repo 주소를 입력(gitlab의 경우 주소맨뒤에 .git을 입력)"
	ArgoDescriptionGitRevision                  = "Git target revision(branch, tag)를 입력 ex) main"
)

// multi-operator bootstrap을 위해 필요한 초기 환경변수
// 변수 추가시 GetRequiredEnvPreset에 추가해야 함
const (
	HC_DOMAIN          = "HC_DOMAIN"
	AUTH_CLIENT_SECRET = "AUTH_CLIENT_SECRET"
	AUTH_SUBDOMAIN     = "AUTH_SUBDOMAIN"
	// AUDIT_WEBHOOK_SERVER_PATH = "AUDIT_WEBHOOK_SERVER_PATH"
)

func GetRequiredEnvPreset() []string {
	return []string{
		HC_DOMAIN,
		AUTH_CLIENT_SECRET,
		AUTH_SUBDOMAIN,
		// AUDIT_WEBHOOK_SERVER_PATH,
	}
}
