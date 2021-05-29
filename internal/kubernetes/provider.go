package kubernetes

import (
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/kubernetes/v1/clientset"
	"github.com/authelia/authelia/internal/kubernetes/v1/types"
	"github.com/authelia/authelia/internal/logging"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Provider struct {
	informers             []clientset.Informer
	client                *clientset.RootClient
	config                *schema.Configuration
	AccessControlRuleFunc func(config *schema.AccessControlConfiguration)
}

// CreateProvider creates an Kubernetes provider based on custom resources.
func CreateProvider(config *schema.Configuration) (*Provider, error) {
	logger := logging.Logger()

	var kubernetesConfig *rest.Config
	var err error
	if config.Kubernetes.UseFlags() {
		kubernetesConfig, err = clientcmd.BuildConfigFromFlags(config.Kubernetes.MasterURL, config.Kubernetes.ConfigFilePath)
	} else {
		kubernetesConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		logger.Errorf("Unable to configure Kubernetes client: %+v", err)
		return nil, err
	}

	types.AddToScheme(scheme.Scheme)

	client, err := clientset.NewClient(kubernetesConfig)
	if err != nil {
		logger.Errorf("Unable to create Kubernetes client: %+v", err)
		return nil, err
	}

	provider := &Provider{
		config: config,
		client: client,
	}

	provider.configureInformers()

	return provider, nil
}

func (provider *Provider) configureInformers() {
	logger := logging.Logger()

	if provider.config.Kubernetes.TrustAccessControlRules {
		logger.Debug("Enabling Kubernetes AccessControlRule watcher")
		informer := provider.client.AccessControlRules().Namespace(provider.config.Kubernetes.Namespace).CreateInformer()
		informer.AddFunc = func(rule *types.AccessControlRule) {
			provider.emitAccessControlConfiguration(informer)
		}
		informer.UpdateFunc = func(oldRule *types.AccessControlRule, newRule *types.AccessControlRule) {
			provider.emitAccessControlConfiguration(informer)
		}
		informer.DeleteFunc = func(rule *types.AccessControlRule) {
			provider.emitAccessControlConfiguration(informer)
		}

		provider.informers = append(provider.informers, informer)
	}
}

// Start starts all informers and waits for the initial synchronization.
func (provider *Provider) Start() error {
	logger := logging.Logger()

	// Run all synchronizations in parallel
	logger.Debug("Waiting for initial Kubernetes sync")
	var group errgroup.Group
	for i, informer := range provider.informers {
		// Give up the looping variable
		index := i
		informer.Start()
		group.Go(func() error {
			return provider.informers[index].WaitForSync()
		})
	}

	return group.Wait()
}

func (provider *Provider) emitAccessControlConfiguration(informer *clientset.AccessControlRuleInformer) {
	logger := logging.Logger()

	logger.Debug("Changes made to deployed Access Control Rules, compiling new configuration")

	// TODO: Validate rules, needless to allocate memory items which don't pass initial validation
	items := informer.Store.List()
	rules := make([]schema.ACLRule, len(items))
	for i, item := range items {
		rule := item.(*types.AccessControlRule)
		rules[i].Domains = rule.Spec.Domains
		rules[i].Methods = rule.Spec.Methods
		rules[i].Networks = rule.Spec.Networks
		rules[i].Policy = rule.Spec.Policy
		rules[i].Resources = rule.Spec.Resources
		rules[i].Subjects = rule.Spec.Subjects
	}

	config := &schema.AccessControlConfiguration{
		DefaultPolicy: provider.config.AccessControl.DefaultPolicy,
		Networks:      provider.config.AccessControl.Networks,
		Rules:         rules,
	}

	logger.Debug("Compiled new configuration, emitting")
	if provider.AccessControlRuleFunc != nil {
		provider.AccessControlRuleFunc(config)
	}
}
