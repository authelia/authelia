package suites

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// ExternalSuite the definition of an external suite.
type ExternalSuite struct {
	Description string
	TestTimeout time.Duration
}

// ExternalRegistry represents a registry of external suites by name.
type ExternalRegistry struct {
	registry map[string]ExternalSuite
}

// ExternalGlobalRegistry a global registry of external suites. It is disjoint from GlobalRegistry
// so external suites do not appear in authelia-scripts suites list.
var ExternalGlobalRegistry *ExternalRegistry

func init() {
	ExternalGlobalRegistry = NewExternalSuitesRegistry()
}

// NewExternalSuitesRegistry create an external suites registry.
func NewExternalSuitesRegistry() *ExternalRegistry {
	return &ExternalRegistry{make(map[string]ExternalSuite)}
}

// Register register an external suite by name.
func (sr *ExternalRegistry) Register(name string, suite ExternalSuite) {
	if _, found := sr.registry[name]; found {
		log.Fatal(fmt.Sprintf("Trying to register the external suite %s multiple times", name))
	}

	sr.registry[name] = suite
}

// Get return an external suite by name.
func (sr *ExternalRegistry) Get(name string) ExternalSuite {
	s, found := sr.registry[name]
	if !found {
		log.Fatal(fmt.Sprintf("The external suite %s does not exist", name))
	}

	return s
}

// Suites list available external suites.
func (sr *ExternalRegistry) Suites() []string {
	suites := make([]string, 0, len(sr.registry))
	for k := range sr.registry {
		suites = append(suites, k)
	}

	return suites
}
