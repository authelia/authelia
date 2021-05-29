package clientset

import (
	"github.com/authelia/authelia/internal/kubernetes/v1/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// RootInterface is the root of the client set.
type RootInterface interface {
	AccessControlRules() AccessControlRuleInterface
}

// RootClient is the base client for the client set.
type RootClient struct {
	client rest.Interface
}

// NewClient creates a new client based on the given configuration.
func NewClient(restConfig *rest.Config) (*RootClient, error) {
	config := *restConfig
	config.ContentConfig.GroupVersion = &types.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &RootClient{client: client}, nil
}
