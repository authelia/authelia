package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type NetworkACLSuite struct {
	suite.Suite

	clients []*WebDriverSession
}

func NewNetworkACLSuite() *NetworkACLSuite {
	return &NetworkACLSuite{clients: make([]*WebDriverSession, 3)}
}

func (s *NetworkACLSuite) createClient(idx int) {
	wds, err := StartWebDriverWithProxy(fmt.Sprintf("http://proxy-client%d.example.com:3128", idx), 4444+idx)

	if err != nil {
		log.Fatal(err)
	}

	s.clients[idx] = wds
}

func (s *NetworkACLSuite) teardownClient(idx int) {
	if err := s.clients[idx].Stop(); err != nil {
		log.Fatal(err)
	}
}

func (s *NetworkACLSuite) SetupSuite() {
	wds, err := StartWebDriver()
	if err != nil {
		log.Fatal(err)
	}
	s.clients[0] = wds

	for i := 1; i <= 2; i++ {
		s.createClient(i)
	}
}

func (s *NetworkACLSuite) TearDownSuite() {
	if err := s.clients[0].Stop(); err != nil {
		log.Fatal(err)
	}
	for i := 1; i <= 2; i++ {
		s.teardownClient(i)
	}
}

func (s *NetworkACLSuite) TestShouldAccessSecretUpon2FA() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html", SecureBaseURL)
	secret := s.clients[0].doRegisterThenLogout(ctx, s.T(), "john", "password")

	s.clients[0].doVisit(s.T(), targetURL)
	s.clients[0].verifyIsFirstFactorPage(ctx, s.T())

	s.clients[0].doLoginOneFactor(ctx, s.T(), "john", "password", false, targetURL)
	s.clients[0].verifyIsSecondFactorPage(ctx, s.T())
	s.clients[0].doValidateTOTP(ctx, s.T(), secret)

	s.clients[0].verifySecretAuthorized(ctx, s.T())
}

// from network 192.168.240.201/32
func (s *NetworkACLSuite) TestShouldAccessSecretUpon1FA() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html", SecureBaseURL)
	s.clients[1].doVisit(s.T(), targetURL)
	s.clients[1].verifyIsFirstFactorPage(ctx, s.T())

	s.clients[1].doLoginOneFactor(ctx, s.T(), "john", "password",
		false, fmt.Sprintf("%s/secret.html", SecureBaseURL))
	s.clients[1].verifySecretAuthorized(ctx, s.T())
}

// from network 192.168.240.202/32
func (s *NetworkACLSuite) TestShouldAccessSecretUpon0FA() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.clients[2].doVisit(s.T(), fmt.Sprintf("%s/secret.html", SecureBaseURL))
	s.clients[2].verifySecretAuthorized(ctx, s.T())
}

func TestNetworkACLSuite(t *testing.T) {
	suite.Run(t, NewNetworkACLSuite())
}
