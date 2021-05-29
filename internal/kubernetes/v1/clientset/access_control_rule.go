package clientset

import (
	"context"
	"fmt"

	"github.com/authelia/authelia/internal/kubernetes/v1/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type AccessControlRuleInterface interface {
	List(options metav1.ListOptions) (*types.AccessControlRuleList, error)
	Get(name string, options metav1.GetOptions) (*types.AccessControlRule, error)
	Namespace(name string) AccessControlRuleInterface
	Watch(options metav1.ListOptions) (watch.Interface, error)
	CreateInformer() *AccessControlRuleInformer
}

type accessControlRuleClient struct {
	client    rest.Interface
	namespace string
}

type Informer interface {
	Start()
	WaitForSync() error
	Stop()
}

type AccessControlRuleInformer struct {
	stopSignal chan struct{}
	Store      cache.Store
	Controller cache.Controller
	AddFunc    func(rule *types.AccessControlRule)
	UpdateFunc func(oldRule *types.AccessControlRule, newRule *types.AccessControlRule)
	DeleteFunc func(rule *types.AccessControlRule)
}

// AccessControlRules scopes th request to AccessControlRules.
func (client *RootClient) AccessControlRules() AccessControlRuleInterface {
	return &accessControlRuleClient{
		client:    client.client,
		namespace: metav1.NamespaceAll,
	}
}

// Namespace scopes all further requests made to the specified namespace.
func (client *accessControlRuleClient) Namespace(namespace string) AccessControlRuleInterface {
	return &accessControlRuleClient{
		client:    client.client,
		namespace: namespace,
	}
}

// List lists all of the AccessControlRules.
func (client *accessControlRuleClient) List(options metav1.ListOptions) (*types.AccessControlRuleList, error) {
	result := types.AccessControlRuleList{}
	err := client.client.Get().Namespace(client.namespace).Resource("accesscontrolrules").VersionedParams(&options, scheme.ParameterCodec).Do(context.TODO()).Into(&result)

	return &result, err
}

// Get retrieves an AccessControlRule by name.
func (client *accessControlRuleClient) Get(name string, options metav1.GetOptions) (*types.AccessControlRule, error) {
	result := types.AccessControlRule{}
	err := client.client.Get().Namespace(client.namespace).Resource("accesscontrolrule").Name(name).VersionedParams(&options, scheme.ParameterCodec).Do(context.TODO()).Into(&result)

	return &result, err
}

// Watch watches AccessControlRules for updates.
func (client *accessControlRuleClient) Watch(options metav1.ListOptions) (watch.Interface, error) {
	options.Watch = true
	return client.client.Get().Namespace(client.namespace).Resource("accesscontrolrules").VersionedParams(&options, scheme.ParameterCodec).Watch(context.TODO())
}

// CreateInformer creates an informer for the resource. The run time is controlled by the input stop channel.
// The controller is started immediately after the call.
func (client *accessControlRuleClient) CreateInformer() *AccessControlRuleInformer {
	informer := &AccessControlRuleInformer{}
	// Create a proxy to pass all events to the informer in order to not expose
	// non-type safe objects
	proxy := &cache.ResourceEventHandlerFuncs{
		AddFunc: func(object interface{}) {
			if informer.AddFunc != nil {
				informer.AddFunc(object.(*types.AccessControlRule))
			}
		},
		UpdateFunc: func(oldObject interface{}, newObject interface{}) {
			if informer.UpdateFunc != nil {
				informer.UpdateFunc(oldObject.(*types.AccessControlRule), newObject.(*types.AccessControlRule))
			}
		},
		DeleteFunc: func(object interface{}) {
			if informer.DeleteFunc != nil {
				informer.DeleteFunc(object.(*types.AccessControlRule))
			}
		},
	}

	store, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (result runtime.Object, err error) {
				return client.List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.Watch = true
				return client.client.Get().Namespace(client.namespace).Resource("accesscontrolrules").VersionedParams(&options, scheme.ParameterCodec).Watch(context.TODO())
			},
		},
		&types.AccessControlRule{},
		0,
		proxy,
	)

	informer.stopSignal = make(chan struct{})
	informer.Store = store
	informer.Controller = controller

	return informer
}

func (informer *AccessControlRuleInformer) Start() {
	go informer.Controller.Run(informer.stopSignal)
}

func (informer *AccessControlRuleInformer) WaitForSync() error {
	if !cache.WaitForCacheSync(informer.stopSignal, informer.Controller.HasSynced) {
		return fmt.Errorf("Failed to sync")
	}
	return nil
}

func (informer *AccessControlRuleInformer) Stop() {
	close(informer.stopSignal)
}
