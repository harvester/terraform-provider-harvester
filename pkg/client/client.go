package client

import (
	harvnetworkclient "github.com/harvester/harvester-network-controller/pkg/generated/clientset/versioned"
	harvclient "github.com/harvester/harvester/pkg/generated/clientset/versioned"
	"github.com/harvester/harvester/pkg/generated/clientset/versioned/scheme"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	kubeschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	RestConfig                *rest.Config
	KubeVirtSubresourceClient *rest.RESTClient
	KubeClient                *kubernetes.Clientset
	HarvesterClient           *harvclient.Clientset
	HarvesterNetworkClient    *harvnetworkclient.Clientset
}

func NewClient(kubeConfig, kubeContext string) (*Client, error) {
	clientConfig := kubeconfig.GetNonInteractiveClientConfigWithContext(kubeConfig, kubeContext)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	copyConfig := rest.CopyConfig(restConfig)
	copyConfig.GroupVersion = &kubeschema.GroupVersion{Group: "subresources.kubevirt.io", Version: "v1"}
	copyConfig.APIPath = "/apis"
	copyConfig.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	restClient, err := rest.RESTClientFor(copyConfig)
	if err != nil {
		return nil, err
	}
	harvClient, err := harvclient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	harvNetworkClient, err := harvnetworkclient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return &Client{
		RestConfig:                restConfig,
		KubeVirtSubresourceClient: restClient,
		KubeClient:                kubeClient,
		HarvesterClient:           harvClient,
		HarvesterNetworkClient:    harvNetworkClient,
	}, nil
}
