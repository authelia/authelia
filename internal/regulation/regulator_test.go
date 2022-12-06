package regulation_test

import (
	"context"
	"testing"
	"time"

	"github.com/authelia/authelia/v4/internal/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
)

type RegulatorSuite struct {
	suite.Suite

	ctx         context.Context
	ctrl        *gomock.Controller
	storageMock *mocks.MockStorage
	config      schema.RegulationConfiguration
	clock       utils.TestingClock
}

func (s *RegulatorSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.storageMock = mocks.NewMockStorage(s.ctrl)
	s.ctx = context.Background()

	s.config = schema.RegulationConfiguration{
		MaxRetries: 3,
		BanTime:    time.Second * 180,
		FindTime:   time.Second * 30,
	}
	s.clock.Set(time.Now())
}

func (s *RegulatorSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *RegulatorSuite) TestShouldNotThrowWhenUserIsLegitimate() {
	attemptsInDB := []model.AuthenticationAttempt{
		{
			Username:   "john",
			Successful: true,
			Time:       s.clock.Now().Add(-4 * time.Minute),
		},
	}

	s.storageMock.EXPECT().
		LoadAuthenticationLogs(s.ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(s.config, s.storageMock, &s.clock)

	_, err := regulator.Regulate(s.ctx, "john")
	assert.NoError(s.T(), err)
}

// This test checks the case in which a user failed to authenticate many times but always
// with a certain amount of time larger than FindTime. Meaning the user should not be banned.
func (s *RegulatorSuite) TestShouldNotThrowWhenFailedAuthenticationNotInFindTime() {
	attemptsInDB := []model.AuthenticationAttempt{
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-1 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-90 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-180 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadAuthenticationLogs(s.ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(s.config, s.storageMock, &s.clock)

	_, err := regulator.Regulate(s.ctx, "john")
	assert.NoError(s.T(), err)
}

// This test checks the case in which a user failed to authenticate many times only a few
// seconds ago (meaning we are checking from now back to now-FindTime).
func (s *RegulatorSuite) TestShouldBanUserIfLatestAttemptsAreWithinFinTime() {
	attemptsInDB := []model.AuthenticationAttempt{
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-1 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-4 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-6 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-180 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadAuthenticationLogs(s.ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(s.config, s.storageMock, &s.clock)

	_, err := regulator.Regulate(s.ctx, "john")
	assert.Equal(s.T(), regulation.ErrUserIsBanned, err)
}

// This test checks the case in which a user failed to authenticate many times only a few
// seconds ago (meaning we are checking from now-FindTime+X back to now-2FindTime+X knowing that
// we are within now and now-BanTime). It means the user has been banned some time ago and is still
// banned right now.
func (s *RegulatorSuite) TestShouldCheckUserIsStillBanned() {
	attemptsInDB := []model.AuthenticationAttempt{
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-31 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-34 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-36 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadAuthenticationLogs(s.ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(s.config, s.storageMock, &s.clock)

	_, err := regulator.Regulate(s.ctx, "john")
	assert.Equal(s.T(), regulation.ErrUserIsBanned, err)
}

func (s *RegulatorSuite) TestShouldCheckUserIsNotYetBanned() {
	attemptsInDB := []model.AuthenticationAttempt{
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-34 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-36 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadAuthenticationLogs(s.ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(s.config, s.storageMock, &s.clock)

	_, err := regulator.Regulate(s.ctx, "john")
	assert.NoError(s.T(), err)
}

func (s *RegulatorSuite) TestShouldCheckUserWasAboutToBeBanned() {
	attemptsInDB := []model.AuthenticationAttempt{
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-14 * time.Second),
		},
		// more than 30 seconds elapsed between this auth and the preceding one.
		// In that case we don't need to regulate the user even though the number
		// of retrieved attempts is 3.
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-94 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-96 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadAuthenticationLogs(s.ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(s.config, s.storageMock, &s.clock)

	_, err := regulator.Regulate(s.ctx, "john")
	assert.NoError(s.T(), err)
}

func (s *RegulatorSuite) TestShouldCheckRegulationHasBeenResetOnSuccessfulAttempt() {
	attemptsInDB := []model.AuthenticationAttempt{
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-90 * time.Second),
		},
		{
			Username:   "john",
			Successful: true,
			Time:       s.clock.Now().Add(-93 * time.Second),
		},
		// The user was almost banned but he did a successful attempt. Therefore, even if the next
		// failure happens within FindTime, he should not be banned.
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-94 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-96 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadAuthenticationLogs(s.ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(s.config, s.storageMock, &s.clock)

	_, err := regulator.Regulate(s.ctx, "john")
	assert.NoError(s.T(), err)
}

func TestRunRegulatorSuite(t *testing.T) {
	s := new(RegulatorSuite)
	suite.Run(t, s)
}

// This test checks that the regulator is disabled when configuration is set to 0.
func (s *RegulatorSuite) TestShouldHaveRegulatorDisabled() {
	attemptsInDB := []model.AuthenticationAttempt{
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-31 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-34 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.clock.Now().Add(-36 * time.Second),
		},
	}

	s.storageMock.EXPECT().
		LoadAuthenticationLogs(s.ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	// Check Disabled Functionality.
	config := schema.RegulationConfiguration{
		MaxRetries: 0,
		FindTime:   time.Second * 180,
		BanTime:    time.Second * 180,
	}

	regulator := regulation.NewRegulator(config, s.storageMock, &s.clock)
	_, err := regulator.Regulate(s.ctx, "john")
	assert.NoError(s.T(), err)

	// Check Enabled Functionality.
	config = schema.RegulationConfiguration{
		MaxRetries: 1,
		FindTime:   time.Second * 180,
		BanTime:    time.Second * 180,
	}

	regulator = regulation.NewRegulator(config, s.storageMock, &s.clock)
	_, err = regulator.Regulate(s.ctx, "john")
	assert.Equal(s.T(), regulation.ErrUserIsBanned, err)
}
