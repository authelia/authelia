package regulation_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/clems4ever/authelia/models"

	"github.com/stretchr/testify/assert"

	"github.com/clems4ever/authelia/configuration/schema"
	"github.com/clems4ever/authelia/mocks"
	"github.com/clems4ever/authelia/regulation"
	"github.com/golang/mock/gomock"
)

type RegulatorSuite struct {
	suite.Suite

	ctrl          *gomock.Controller
	storageMock   *mocks.MockStorageProvider
	configuration schema.RegulationConfiguration
	now           time.Time
}

func (s *RegulatorSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.storageMock = mocks.NewMockStorageProvider(s.ctrl)

	s.configuration = schema.RegulationConfiguration{
		MaxRetries: 3,
		BanTime:    180,
		FindTime:   30,
	}
	s.now = time.Now()
}

func (s *RegulatorSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *RegulatorSuite) TestShouldNotThrowWhenUserIsLegitimate() {
	attempts := []models.AuthenticationAttempt{
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Time:       s.now.Add(-4 * time.Minute),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attempts, nil)

	regulator := regulation.NewRegulator(&s.configuration, s.storageMock)

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
			Time:       s.now.Add(-1 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-90 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-180 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(&s.configuration, s.storageMock)

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
			Time:       s.now.Add(-1 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-4 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-6 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-180 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(&s.configuration, s.storageMock)

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
			Time:       s.now.Add(-31 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-34 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-36 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(&s.configuration, s.storageMock)

	_, err := regulator.Regulate("john")
	assert.Equal(s.T(), regulation.ErrUserIsBanned, err)
}

func (s *RegulatorSuite) TestShouldCheckUserIsNotYetBanned() {
	attemptsInDB := []models.AuthenticationAttempt{
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-34 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-36 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(&s.configuration, s.storageMock)

	_, err := regulator.Regulate("john")
	assert.NoError(s.T(), err)
}

func (s *RegulatorSuite) TestShouldCheckUserWasAboutToBeBanned() {
	attemptsInDB := []models.AuthenticationAttempt{
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-14 * time.Second),
		},
		// more than 30 seconds elapsed between this auth and the preceding one.
		// In that case we don't need to regulate the user even though the number
		// of retrieved attempts is 3.
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-94 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-96 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(&s.configuration, s.storageMock)

	_, err := regulator.Regulate("john")
	assert.NoError(s.T(), err)
}

func (s *RegulatorSuite) TestShouldCheckRegulationHasBeenResetOnSuccessfulAttempt() {
	attemptsInDB := []models.AuthenticationAttempt{
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-90 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Time:       s.now.Add(-93 * time.Second),
		},
		// The user was almost banned but he did a successful attempt. Therefore, even if the next
		// failure happens withing FindTime, he should not be banned.
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-94 * time.Second),
		},
		models.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Time:       s.now.Add(-96 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadLatestAuthenticationLogs(gomock.Eq("john"), gomock.Any()).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(&s.configuration, s.storageMock)

	_, err := regulator.Regulate("john")
	assert.NoError(s.T(), err)
}

func TestRunRegulatorSuite(t *testing.T) {
	s := new(RegulatorSuite)
	suite.Run(t, s)
}
