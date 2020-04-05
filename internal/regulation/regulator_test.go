package regulation_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/mocks"
	"github.com/authelia/authelia/internal/models"
	"github.com/authelia/authelia/internal/regulation"
	"github.com/authelia/authelia/internal/storage"
)

type RegulatorSuite struct {
	suite.Suite

	ctrl          *gomock.Controller
	storageMock   *storage.MockProvider
	configuration schema.RegulationConfiguration
	clock         mocks.TestingClock
}

func (s *RegulatorSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.storageMock = storage.NewMockProvider(s.ctrl)

	s.configuration = schema.RegulationConfiguration{
		MaxRetries: 3,
		BanTime:    "180",
		FindTime:   "30",
	}
	s.clock.Set(time.Now())
}

func (s *RegulatorSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *RegulatorSuite) TestShouldNotThrowWhenUserIsLegitimate() {
	attemptsInDB := []models.AuthenticationAttempt{
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Time:       s.clock.Now().Add(-4 * time.Minute),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(&s.configuration, s.storageMock, &s.clock)

	_, err := regulator.Regulate("john")
	assert.NoError(s.T(), err)
}

// This test checks the case in which a user failed to authenticate many times but always
// with a certain amount of time larger than FindTime. Meaning the user should not be banned.
func (s *RegulatorSuite) TestShouldNotThrowWhenFailedAuthenticationNotInFindTime() {
	attemptsInDB := []models.AuthenticationAttempt{
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-1 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-90 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-180 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(&s.configuration, s.storageMock, &s.clock)

	_, err := regulator.Regulate("john")
	assert.NoError(s.T(), err)
}

// This test checks the case in which a user failed to authenticate many times only a few
// seconds ago (meaning we are checking from now back to now-FindTime).
func (s *RegulatorSuite) TestShouldBanUserIfLatestAttemptsAreWithinFinTime() {
	attemptsInDB := []models.AuthenticationAttempt{
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-1 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-4 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-6 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-180 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(&s.configuration, s.storageMock, &s.clock)

	_, err := regulator.Regulate("john")
	assert.Equal(s.T(), regulation.ErrUserIsBanned, err)
}

// This test checks the case in which a user failed to authenticate many times only a few
// seconds ago (meaning we are checking from now-FindTime+X back to now-2FindTime+X knowing that
// we are within now and now-BanTime). It means the user has been banned some time ago and is still
// banned right now.
func (s *RegulatorSuite) TestShouldCheckUserIsStillBanned() {
	attemptsInDB := []models.AuthenticationAttempt{
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-31 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-34 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-36 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(&s.configuration, s.storageMock, &s.clock)

	_, err := regulator.Regulate("john")
	assert.Equal(s.T(), regulation.ErrUserIsBanned, err)
}

func (s *RegulatorSuite) TestShouldCheckUserIsNotYetBanned() {
	attemptsInDB := []models.AuthenticationAttempt{
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-34 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-36 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(&s.configuration, s.storageMock, &s.clock)

	_, err := regulator.Regulate("john")
	assert.NoError(s.T(), err)
}

func (s *RegulatorSuite) TestShouldCheckUserWasAboutToBeBanned() {
	attemptsInDB := []models.AuthenticationAttempt{
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-14 * time.Second),
		},
		// more than 30 seconds elapsed between this auth and the preceding one.
		// In that case we don't need to regulate the user even though the number
		// of retrieved attempts is 3.
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-94 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-96 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(&s.configuration, s.storageMock, &s.clock)

	_, err := regulator.Regulate("john")
	assert.NoError(s.T(), err)
}

func (s *RegulatorSuite) TestShouldCheckRegulationHasBeenResetOnSuccessfulAttempt() {
	attemptsInDB := []models.AuthenticationAttempt{
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-90 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Time:       s.clock.Now().Add(-93 * time.Second),
		},
		// The user was almost banned but he did a successful attempt. Therefore, even if the next
		// failure happens within FindTime, he should not be banned.
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-94 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-96 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(&s.configuration, s.storageMock, &s.clock)

	_, err := regulator.Regulate("john")
	assert.NoError(s.T(), err)
}

func TestRunRegulatorSuite(t *testing.T) {
	s := new(RegulatorSuite)
	suite.Run(t, s)
}

// This test checks that the regulator is disabled when configuration is set to 0.
func (s *RegulatorSuite) TestShouldHaveRegulatorDisabled() {
	attemptsInDB := []models.AuthenticationAttempt{
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-31 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-34 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-36 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attemptsInDB, nil)

	// Check Disabled Functionality
	configuration := schema.RegulationConfiguration{
		MaxRetries: 0,
		FindTime:   "180",
		BanTime:    "180",
	}

	regulator := regulation.NewRegulator(&configuration, s.storageMock, &s.clock)
	_, err := regulator.Regulate("john")
	assert.NoError(s.T(), err)

	// Check Enabled Functionality
	configuration = schema.RegulationConfiguration{
		MaxRetries: 1,
		FindTime:   "180",
		BanTime:    "180",
	}

	regulator = regulation.NewRegulator(&configuration, s.storageMock, &s.clock)
	_, err = regulator.Regulate("john")
	assert.Equal(s.T(), regulation.ErrUserIsBanned, err)
}
