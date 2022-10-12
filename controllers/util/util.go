package util

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
)

// LowestNonZeroResult compares two reconciliation results
// and returns the one with lowest requeue time.
func LowestNonZeroResult(i, j ctrl.Result) ctrl.Result {
	switch {
	case i.IsZero():
		return j
	case j.IsZero():
		return i
	case i.Requeue:
		return i
	case j.Requeue:
		return j
	case i.RequeueAfter < j.RequeueAfter:
		return i
	default:
		return j
	}
}

func GetRemoteK8sClient(secret *coreV1.Secret) (*kubernetes.Clientset, error) {
	value, ok := secret.Data["value"]
	if !ok {
		err := errors.NewBadRequest("secret does not have a value")
		return nil, err
	}

	remoteClientConfig, err := clientcmd.NewClientConfigFromBytes(value)
	if err != nil {
		return nil, err
	}

	remoteRestConfig, err := remoteClientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	remoteClientset, err := kubernetes.NewForConfig(remoteRestConfig)
	if err != nil {
		return nil, err
	}

	return remoteClientset, nil
}

func GetRemoteK8sClientByKubeConfig(kubeConfig []byte) (*kubernetes.Clientset, error) {
	remoteClientConfig, err := clientcmd.NewClientConfigFromBytes(kubeConfig)
	if err != nil {
		return nil, err
	}

	remoteRestConfig, err := remoteClientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	remoteClientset, err := kubernetes.NewForConfig(remoteRestConfig)
	if err != nil {
		return nil, err
	}

	return remoteClientset, nil
}

func GetK8sClient() (*kubernetes.Clientset, error) {
	config, err := restclient.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	config.Burst = 100
	config.QPS = 100
	Clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return Clientset, nil
}

func CreateSuffixString() string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyz")

	s := make([]rune, 5)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}

	return string(s)
}

func MergeJson(dest []byte, source []byte) []byte {
	dest = append(dest[0:len(dest)-1], 44)
	dest = append(dest, source[1:]...)

	return dest
}

func URIToSecretName(uriType, uri string) (string, error) {
	parsedURI, err := url.ParseRequestURI(uri)
	if err != nil {
		return "", err
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(uri))
	host := strings.ToLower(strings.Split(parsedURI.Host, ":")[0])

	return fmt.Sprintf("%s-%s-%v", uriType, host, h.Sum32()), nil
}

func GetProviderName(provider string) (string, error) {
	provider = strings.ToUpper(provider)
	providerNameLogo := map[string]string{
		ProviderAws:     ProviderAwsLogo,
		ProviderVsphere: ProviderVsphereLogo,
	}

	if providerNameLogo[provider] == "" {
		return ProviderUnknown, fmt.Errorf("Cannot found provider [" + provider + "]")
	}

	return providerNameLogo[provider], nil
}

func CheckRequiredEnvPreset() error {
	notExistEnvList := []string{}
	for _, env := range GetRequiredEnvPreset() {
		if os.Getenv(env) == "" {
			notExistEnvList = append(notExistEnvList, env)
		}
	}

	if len(notExistEnvList) != 0 {
		return fmt.Errorf("%s env not exist", strings.Join(notExistEnvList, ", "))
	}
	return nil
}
