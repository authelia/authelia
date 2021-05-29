package types

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// AccessControlRuleList is a list of AccessControlRule custom resources. Implements runtime.Object.
type AccessControlRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessControlRule `json:"items"`
}

// AccessControlRuleLis is a AccessControlRule custom resource. Implements runtime.Object.
type AccessControlRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              struct {
		Domains   []string   `json:"domains"`
		Policy    string     `json:"policy"`
		Subjects  [][]string `json:"subjects"`
		Networks  []string   `json:"networks"`
		Resources []string   `json:"resources"`
		Methods   []string   `json:"methods"`
	} `json:"spec"`
}

func (in *AccessControlRule) DeepCopyInto(out *AccessControlRule) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec.Policy = in.Spec.Policy

	if in.Spec.Domains != nil {
		out.Spec.Domains = make([]string, len(in.Spec.Domains))
		for i := range in.Spec.Domains {
			out.Spec.Domains[i] = in.Spec.Domains[i]
		}
	}

	if in.Spec.Subjects != nil {
		out.Spec.Subjects = make([][]string, len(in.Spec.Subjects))
		for i := range in.Spec.Subjects {
			out.Spec.Subjects[i] = make([]string, len(in.Spec.Subjects[i]))
			for j := range in.Spec.Subjects[i] {
				out.Spec.Subjects[i][j] = in.Spec.Subjects[i][j]
			}
		}
	}

	if in.Spec.Networks != nil {
		out.Spec.Networks = make([]string, len(in.Spec.Networks))
		for i := range in.Spec.Networks {
			out.Spec.Networks[i] = in.Spec.Networks[i]
		}
	}

	if in.Spec.Resources != nil {
		out.Spec.Resources = make([]string, len(in.Spec.Resources))
		for i := range in.Spec.Resources {
			out.Spec.Resources[i] = in.Spec.Resources[i]
		}
	}

	if in.Spec.Methods != nil {
		out.Spec.Methods = make([]string, len(in.Spec.Methods))
		for i := range in.Spec.Methods {
			out.Spec.Methods[i] = in.Spec.Methods[i]
		}
	}
}

func (in *AccessControlRule) DeepCopyObject() runtime.Object {
	out := AccessControlRule{}
	in.DeepCopyInto(&out)

	return &out
}

func (in *AccessControlRuleList) DeepCopyObject() runtime.Object {
	out := AccessControlRuleList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]AccessControlRule, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}
