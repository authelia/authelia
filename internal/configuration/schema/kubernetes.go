package schema

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubernetesConfiguration represents the configuration for the Kubernetes features.
type KubernetesConfiguration struct {
	TrustAccessControlRules bool   `mapstructure:"trust_access_control_rules"`
	MasterURL               string `mapstructure:"master_url"`
	ConfigFilePath          string `mapstructure:"config_file_path"`
	Namespace               string `mapstructure:"namespace"`
}

// DefaultKubernetesConfiguration represents the default values of the KubernetesConfiguration.
var DefaultKubernetesConfiguration = KubernetesConfiguration{
	TrustAccessControlRules: false,
	MasterURL:               "",
	ConfigFilePath:          "",
	Namespace:               metav1.NamespaceAll,
}

// IsEnabled describes whether or not any custom resource is enabled.
func (config KubernetesConfiguration) IsEnabled() bool {
	return config.TrustAccessControlRules
}

// UseFlags describes whether or not configuration flags should be used.
func (config KubernetesConfiguration) UseFlags() bool {
	return config.MasterURL != "" || config.ConfigFilePath != ""
}
