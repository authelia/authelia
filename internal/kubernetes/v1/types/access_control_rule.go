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
		Domain  string `json:"domain"`
		Policy  string `json:"policy"`
		Subject string `json:"subject"`
	} `json:"spec"`
}

func (in *AccessControlRule) DeepCopyInto(out *AccessControlRule) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec.Domain = in.Spec.Domain
	out.Spec.Policy = in.Spec.Policy
	out.Spec.Subject = in.Spec.Subject
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
