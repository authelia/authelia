package regulation_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
)

type RegulatorSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *RegulatorSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	s.mock.Ctx.Configuration.Regulation = schema.Regulation{
		MaxRetries: 3,
		BanTime:    time.Second * 180,
		FindTime:   time.Second * 30,
	}

	s.mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedFor, "127.0.0.1")
}

func (s *RegulatorSuite) TearDownTest() {
	s.mock.Ctrl.Finish()
}

func (s *RegulatorSuite) TestShouldMark() {
	regulator := regulation.NewRegulator(s.mock.Ctx.Configuration.Regulation, s.mock.StorageMock, &s.mock.Clock)

	s.mock.StorageMock.EXPECT().AppendAuthenticationLog(s.mock.Ctx, model.AuthenticationAttempt{
		Time:          s.mock.Clock.Now(),
		Successful:    true,
		Banned:        false,
		Username:      "john",
		Type:          "1fa",
		RemoteIP:      model.NewNullIP(net.ParseIP("127.0.0.1")),
		RequestURI:    "https://google.com",
		RequestMethod: fasthttp.MethodGet,
	})

	s.NoError(regulator.Mark(s.mock.Ctx, true, false, "john", "https://google.com", fasthttp.MethodGet, "1fa"))
}

func (s *RegulatorSuite) TestShouldHandleRegulateError() {
	regulator := regulation.NewRegulator(s.mock.Ctx.Configuration.Regulation, s.mock.StorageMock, &s.mock.Clock)

	s.mock.StorageMock.EXPECT().LoadAuthenticationLogs(s.mock.Ctx, "john", s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.BanTime), 10, 0).Return(nil, fmt.Errorf("failed"))

	until, err := regulator.Regulate(s.mock.Ctx, "john")

	s.NoError(err)
	s.Equal(time.Time{}, until)
}

func (s *RegulatorSuite) TestShouldNotThrowWhenUserIsLegitimate() {
	attemptsInDB := []model.AuthenticationAttempt{
		{
			Username:   "john",
			Successful: true,
			Time:       s.mock.Clock.Now().Add(-4 * time.Minute),
		},
	}

	s.mock.StorageMock.EXPECT().
		LoadAuthenticationLogs(s.mock.Ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(s.mock.Ctx.Configuration.Regulation, s.mock.StorageMock, &s.mock.Clock)

	_, err := regulator.Regulate(s.mock.Ctx, "john")
	assert.NoError(s.T(), err)
}

// This test checks the case in which a user failed to authenticate many times but always
// with a certain amount of time larger than FindTime. Meaning the user should not be banned.
func (s *RegulatorSuite) TestShouldNotThrowWhenFailedAuthenticationNotInFindTime() {
	attemptsInDB := []model.AuthenticationAttempt{
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-1 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-90 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-180 * time.Second),
		},
	}

	s.mock.StorageMock.EXPECT().
		LoadAuthenticationLogs(s.mock.Ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(s.mock.Ctx.Configuration.Regulation, s.mock.StorageMock, &s.mock.Clock)

	_, err := regulator.Regulate(s.mock.Ctx, "john")
	assert.NoError(s.T(), err)
}

// This test checks the case in which a user failed to authenticate many times only a few
// seconds ago (meaning we are checking from now back to now-FindTime).
func (s *RegulatorSuite) TestShouldBanUserIfLatestAttemptsAreWithinFinTime() {
	attemptsInDB := []model.AuthenticationAttempt{
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-1 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-4 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-6 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-180 * time.Second),
		},
	}

	s.mock.StorageMock.EXPECT().
		LoadAuthenticationLogs(s.mock.Ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(s.mock.Ctx.Configuration.Regulation, s.mock.StorageMock, &s.mock.Clock)

	_, err := regulator.Regulate(s.mock.Ctx, "john")
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
			Time:       s.mock.Clock.Now().Add(-31 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-34 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-36 * time.Second),
		},
	}

	s.mock.StorageMock.EXPECT().
		LoadAuthenticationLogs(s.mock.Ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(s.mock.Ctx.Configuration.Regulation, s.mock.StorageMock, &s.mock.Clock)

	_, err := regulator.Regulate(s.mock.Ctx, "john")
	assert.Equal(s.T(), regulation.ErrUserIsBanned, err)
}

func (s *RegulatorSuite) TestShouldCheckUserIsNotYetBanned() {
	attemptsInDB := []model.AuthenticationAttempt{
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-34 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-36 * time.Second),
		},
	}

	s.mock.StorageMock.EXPECT().
		LoadAuthenticationLogs(s.mock.Ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(s.mock.Ctx.Configuration.Regulation, s.mock.StorageMock, &s.mock.Clock)

	_, err := regulator.Regulate(s.mock.Ctx, "john")
	assert.NoError(s.T(), err)
}

func (s *RegulatorSuite) TestShouldCheckUserWasAboutToBeBanned() {
	attemptsInDB := []model.AuthenticationAttempt{
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-14 * time.Second),
		},
		// more than 30 seconds elapsed between this auth and the preceding one.
		// In that case we don't need to regulate the user even though the number
		// of retrieved attempts is 3.
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-94 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-96 * time.Second),
		},
	}

	s.mock.StorageMock.EXPECT().
		LoadAuthenticationLogs(s.mock.Ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(s.mock.Ctx.Configuration.Regulation, s.mock.StorageMock, &s.mock.Clock)

	_, err := regulator.Regulate(s.mock.Ctx, "john")
	assert.NoError(s.T(), err)
}

func (s *RegulatorSuite) TestShouldCheckRegulationHasBeenResetOnSuccessfulAttempt() {
	attemptsInDB := []model.AuthenticationAttempt{
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-90 * time.Second),
		},
		{
			Username:   "john",
			Successful: true,
			Time:       s.mock.Clock.Now().Add(-93 * time.Second),
		},
		// The user was almost banned but he did a successful attempt. Therefore, even if the next
		// failure happens within FindTime, he should not be banned.
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-94 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-96 * time.Second),
		},
	}

	s.mock.StorageMock.EXPECT().
		LoadAuthenticationLogs(s.mock.Ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	regulator := regulation.NewRegulator(s.mock.Ctx.Configuration.Regulation, s.mock.StorageMock, &s.mock.Clock)

	_, err := regulator.Regulate(s.mock.Ctx, "john")
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
			Time:       s.mock.Clock.Now().Add(-31 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-34 * time.Second),
		},
		{
			Username:   "john",
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-36 * time.Second),
		},
	}

	s.mock.StorageMock.EXPECT().
		LoadAuthenticationLogs(s.mock.Ctx, gomock.Eq("john"), gomock.Any(), gomock.Eq(10), gomock.Eq(0)).
		Return(attemptsInDB, nil)

	// Check Disabled Functionality.
	config := schema.Regulation{
		MaxRetries: 0,
		FindTime:   time.Second * 180,
		BanTime:    time.Second * 180,
	}

	regulator := regulation.NewRegulator(config, s.mock.StorageMock, &s.mock.Clock)
	_, err := regulator.Regulate(s.mock.Ctx, "john")
	assert.NoError(s.T(), err)

	// Check Enable Functionality.
	config = schema.Regulation{
		MaxRetries: 1,
		FindTime:   time.Second * 180,
		BanTime:    time.Second * 180,
	}

	regulator = regulation.NewRegulator(config, s.mock.StorageMock, &s.mock.Clock)
	_, err = regulator.Regulate(s.mock.Ctx, "john")
	assert.Equal(s.T(), regulation.ErrUserIsBanned, err)
}
