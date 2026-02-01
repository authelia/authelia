package suites

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// Suite the definition of a suite.
type Suite struct {
	SetUp        func(tmpPath string) error
	SetUpTimeout time.Duration

	// Callback called when an error occur during setup phase.
	OnSetupTimeout func() error

	// Callback called when at least one test fail.
	OnError func() error

	TestTimeout time.Duration

	TearDown        func(tmpPath string) error
	TearDownTimeout time.Duration

	// A textual description of the suite purpose.
	Description string
}

// Registry represent a registry of suite by name.
type Registry struct {
	registry map[string]Suite
}

// GlobalRegistry a global registry used by Authelia tooling.
var GlobalRegistry *Registry

func init() {
	GlobalRegistry = NewSuitesRegistry()
}

// NewSuitesRegistry create a suites registry.
func NewSuitesRegistry() *Registry {
	return &Registry{make(map[string]Suite)}
}

// Register register a suite by name.
func (sr *Registry) Register(name string, suite Suite) {
	if _, found := sr.registry[name]; found {
		log.Fatal(fmt.Sprintf("Trying to register the suite %s multiple times", name))
	}

	sr.registry[name] = suite
}

// Get return a suite by name.
func (sr *Registry) Get(name string) Suite {
	s, found := sr.registry[name]
	if !found {
		log.Fatal(fmt.Sprintf("The suite %s does not exist", name))
	}

	return s
}

// Suites list available suites.
func (sr *Registry) Suites() []string {
	suites := make([]string, 0, len(sr.registry))
	for k := range sr.registry {
		suites = append(suites, k)
	}

	return suites
}
