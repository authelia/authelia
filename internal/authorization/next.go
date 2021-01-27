package authorization

import (
	"net"
	"regexp"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// AccessControlRule controls and represents an ACL internally.
type AccessControlRule struct {
	Domains   []AccessControlDomain
	Resources []*regexp.Regexp
	Methods   []string
	Networks  []*net.IPNet
	Subjects  []AccessControlSubjects
}

// NewAccessControlRule parses a schema ACL and generates an internal ACL.
func NewAccessControlRule(rule schema.ACLRule, networks map[string]*net.IPNet) *AccessControlRule {
	acr := &AccessControlRule{}

	for _, domainRule := range rule.Domains {
		domain := AccessControlDomain{}

		if strings.HasPrefix(domainRule, "*.") {
			domain.Wildcard = true
			domain.Name = domainRule[1:]
		} else {
			domain.Name = domainRule
		}

		acr.Domains = append(acr.Domains, domain)
	}

	for _, resource := range rule.Resources {
		acr.Resources = append(acr.Resources, regexp.MustCompile(resource))
	}

	for _, method := range rule.Methods {
		acr.Methods = append(acr.Methods, strings.ToUpper(method))
	}

	for _, network := range rule.Networks {
		if _, ok := networks[network]; !ok {
			var cidr *net.IPNet

			if !strings.Contains(network, "/") {
				ip := net.ParseIP(network)
				if ip.To4() != nil {
					_, cidr, _ = net.ParseCIDR(network + "/32")
				} else {
					_, cidr, _ = net.ParseCIDR(network + "/128")
				}
			} else {
				_, cidr, _ = net.ParseCIDR(network)
			}

			networks[network] = cidr
		}

		acr.Networks = append(acr.Networks, networks[network])
	}

	return acr
}

// AccessControlDomain represents an ACL domain.
type AccessControlDomain struct {
	Name     string
	Wildcard bool
}

// AccessControlSubjects represents an ACL subject.
type AccessControlSubjects struct {
	Subjects []AccessControlSubject
}

// IsMatch returns true if the subject entry completely matches.
func (acs AccessControlSubjects) IsMatch(subject Subject) (match bool) {
	for _, rule := range acs.Subjects {
		if !rule.IsMatch(subject) {
			return false
		}
	}

	return true
}

// AccessControlUser represents an ACL subject of type `user:`.
type AccessControlUser struct {
	Name string
}

// IsMatch returns true if the user matches the subject username.
func (acu AccessControlUser) IsMatch(subject Subject) (match bool) {
	return subject.Username == acu.Name
}

// AccessControlGroup represents an ACL subject of type `group:`.
type AccessControlGroup struct {
	Name string
}

// IsMatch returns true if the group is inside the subject groups.
func (acg AccessControlGroup) IsMatch(subject Subject) (match bool) {
	return utils.IsStringInSlice(acg.Name, subject.Groups)
}

// AccessControlSubject abstracts an ACL subject of type `group:` or `user:`.
type AccessControlSubject interface {
	IsMatch(subject Subject) (match bool)
}
