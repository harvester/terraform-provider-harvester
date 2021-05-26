package client

import (
	harvnetworkclient "github.com/harvester/harvester-network-controller/pkg/generated/clientset/versioned"
	harvclient "github.com/harvester/harvester/pkg/generated/clientset/versioned"
	"github.com/harvester/harvester/pkg/generated/clientset/versioned/scheme"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	kubeschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type Client struct {
	RestConfig                *rest.Config
	KubeVirtSubresourceClient *rest.RESTClient
	HarvesterClient           *harvclient.Clientset
	HarvesterNetworkClient    *harvnetworkclient.Clientset
}

func NewClient(kubeConfig string) (*Client, error) {
	clientConfig := kubeconfig.GetNonInteractiveClientConfig(kubeConfig)
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
	harvNetworkClient, err := harvnetworkclient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return &Client{
		RestConfig:                restConfig,
		KubeVirtSubresourceClient: restClient,
		HarvesterClient:           harvClient,
		HarvesterNetworkClient:    harvNetworkClient,
	}, nil
}
