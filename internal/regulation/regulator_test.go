package regulation_test

import (
	"database/sql"
	"fmt"
	"net"
	"testing"
	"time"

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
	s.mock.Ctx.Clock = &s.mock.Clock

	s.mock.Ctx.Configuration.Regulation = schema.Regulation{
		Mode:       "both",
		MaxRetries: 3,
		BanTime:    time.Second * 180,
		FindTime:   time.Second * 30,
	}

	s.mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedFor, "127.0.0.1")
}

func (s *RegulatorSuite) Regulator() *regulation.Regulator {
	return regulation.NewRegulator(s.mock.Ctx.Configuration.Regulation, s.mock.StorageMock, &s.mock.Clock)
}

func (s *RegulatorSuite) AssertLogEntryMessageAndError(message, err string) {
	entry := s.mock.Hook.LastEntry()

	s.Require().NotNil(entry)

	s.Equal(message, entry.Message)

	v, ok := entry.Data["error"]

	if err == "" {
		s.False(ok)
		s.Nil(v)
	} else {
		s.True(ok)
		s.Require().NotNil(v)

		theErr, ok := v.(error)
		s.True(ok)
		s.Require().NotNil(theErr)

		s.EqualError(theErr, err)
	}
}

func (s *RegulatorSuite) TearDownTest() {
	s.mock.Ctrl.Finish()
}

func (s *RegulatorSuite) TestShouldHandleBanCheckIPError() {
	regulator := s.Regulator()

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(net.ParseIP("127.0.0.1")))).
			Return(nil, fmt.Errorf("failed to load ip bans")),
	)

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeNone, ban)
	s.Equal("", value)
	s.Equal((*time.Time)(nil), expires)
	s.EqualError(err, "failed to load ip bans")
}

func (s *RegulatorSuite) TestShouldHandleBanCheckIPBanned() {
	regulator := s.Regulator()

	ip := net.ParseIP("127.0.0.1")

	result := []model.BannedIP{
		{
			ID:      1,
			Time:    s.mock.Clock.Now().Add(-5 * time.Minute),
			Expires: sql.NullTime{Valid: true, Time: s.mock.Clock.Now().Add(1 * time.Minute)},
			Expired: sql.NullTime{},
			Revoked: false,
			IP:      model.IP{IP: ip},
			Source:  "regulation",
			Reason:  sql.NullString{Valid: true, String: "Exceeding Maximum Retries"},
		},
	}

	attempt := model.AuthenticationAttempt{
		Time:       s.mock.Clock.Now(),
		Successful: false,
		Banned:     true,
		Username:   "john",
		Type:       regulation.AuthType1FA,
		RemoteIP:   model.NewNullIP(ip),
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip))).
			Return(result, nil),
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
	)

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeIP, ban)
	s.Equal("127.0.0.1", value)
	s.Equal(&result[0].Expires.Time, expires)
	s.EqualError(err, "user is banned")

	b := regulation.NewBan(ban, value, expires)

	s.Equal("expires at 12:01:00AM on February 3 2013 (+00:00)", b.FormatExpires())
	s.Equal(regulation.BanTypeIP, b.Type())
	s.Equal("127.0.0.1", b.Value())

	regulator.HandleAttempt(s.mock.Ctx, false, true, "john", "", "", regulation.AuthType1FA)
}

func (s *RegulatorSuite) TestShouldHandleBanCheckIPBannedPermanent() {
	regulator := s.Regulator()

	ip := net.ParseIP("127.0.0.1")

	result := []model.BannedIP{
		{
			ID:      1,
			Time:    s.mock.Clock.Now().Add(-5 * time.Minute),
			Expires: sql.NullTime{},
			Expired: sql.NullTime{},
			Revoked: false,
			IP:      model.IP{IP: ip},
			Source:  "regulation",
			Reason:  sql.NullString{Valid: true, String: "Exceeding Maximum Retries"},
		},
	}

	attempt := model.AuthenticationAttempt{
		Time:       s.mock.Clock.Now(),
		Successful: false,
		Banned:     true,
		Username:   "john",
		Type:       regulation.AuthType1FA,
		RemoteIP:   model.NewNullIP(ip),
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip))).
			Return(result, nil),
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
	)

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeIP, ban)
	s.Equal("127.0.0.1", value)
	s.Equal((*time.Time)(nil), expires)
	s.EqualError(err, "user is banned")

	b := regulation.NewBan(ban, value, expires)

	s.Equal("never expires", b.FormatExpires())
	s.Equal(regulation.BanTypeIP, b.Type())
	s.Equal("127.0.0.1", b.Value())

	regulator.HandleAttempt(s.mock.Ctx, false, true, "john", "", "", regulation.AuthType1FA)
}

func (s *RegulatorSuite) TestShouldHandleBanCheckIPBannedFailToAppend() {
	regulator := s.Regulator()

	ip := net.ParseIP("127.0.0.1")

	result := []model.BannedIP{
		{
			ID:      1,
			Time:    s.mock.Clock.Now().Add(-5 * time.Minute),
			Expires: sql.NullTime{Valid: true, Time: s.mock.Clock.Now().Add(1 * time.Minute)},
			Expired: sql.NullTime{},
			Revoked: false,
			IP:      model.IP{IP: ip},
			Source:  "regulation",
			Reason:  sql.NullString{Valid: true, String: "Exceeding Maximum Retries"},
		},
	}

	attempt := model.AuthenticationAttempt{
		Time:       s.mock.Clock.Now(),
		Successful: false,
		Banned:     true,
		Username:   "john",
		Type:       regulation.AuthType1FA,
		RemoteIP:   model.NewNullIP(ip),
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip))).
			Return(result, nil),
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(fmt.Errorf("failed to log")),
	)

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeIP, ban)
	s.Equal("127.0.0.1", value)
	s.Equal(&result[0].Expires.Time, expires)
	s.EqualError(err, "user is banned")

	regulator.HandleAttempt(s.mock.Ctx, false, true, "john", "", "", regulation.AuthType1FA)

	s.AssertLogEntryMessageAndError("Failed to record 1FA authentication attempt", "failed to log")
}

func (s *RegulatorSuite) TestShouldHandleBanCheckIPNotBanned() {
	regulator := s.Regulator()

	ip := net.ParseIP("127.0.0.1")

	attempt := model.AuthenticationAttempt{
		Time:       s.mock.Clock.Now(),
		Successful: false,
		Banned:     false,
		Username:   "john",
		Type:       regulation.AuthType1FA,
		RemoteIP:   model.NewNullIP(ip),
	}

	since := s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)

	records := []model.RegulationRecord{
		{
			Time:       s.mock.Clock.Now().Add(-10 * time.Second),
			Successful: false,
		},
		{
			Time:       s.mock.Clock.Now().Add(-12 * time.Second),
			Successful: false,
		},
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip))).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadBannedUser(s.mock.Ctx, gomock.Eq("john")).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByIP(s.mock.Ctx, model.NewIP(ip), since, s.mock.Ctx.Configuration.Regulation.MaxRetries).
			Return(records, nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByUser(s.mock.Ctx, "john", since, s.mock.Ctx.Configuration.Regulation.MaxRetries).
			Return(nil, nil),
	)

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeNone, ban)
	s.Equal("", value)
	s.Equal((*time.Time)(nil), expires)
	s.NoError(err)

	regulator.HandleAttempt(s.mock.Ctx, false, false, "john", "", "", regulation.AuthType1FA)
}

func (s *RegulatorSuite) TestShouldHandleBanCheckIPNotBannedButFailedAttempt() {
	regulator := s.Regulator()

	ip := net.ParseIP("127.0.0.1")

	attempt := model.AuthenticationAttempt{
		Time:       s.mock.Clock.Now(),
		Successful: false,
		Banned:     false,
		Username:   "john",
		Type:       regulation.AuthType1FA,
		RemoteIP:   model.NewNullIP(ip),
	}

	since := s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)

	records := []model.RegulationRecord{
		{
			Time:       s.mock.Clock.Now().Add(-10 * time.Second),
			Successful: false,
		},
		{
			Time:       s.mock.Clock.Now().Add(-12 * time.Second),
			Successful: false,
		},
		{
			Time:       s.mock.Clock.Now().Add(-12 * time.Second),
			Successful: false,
		},
	}

	sqlban := &model.BannedIP{
		Expires: sql.NullTime{Valid: true, Time: records[0].Time.Add(s.mock.Ctx.Configuration.Regulation.BanTime)},
		IP:      model.NewIP(ip),
		Source:  "regulation",
		Reason:  sql.NullString{Valid: true, String: "Exceeding Maximum Retries"},
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip))).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadBannedUser(s.mock.Ctx, gomock.Eq("john")).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByIP(s.mock.Ctx, model.NewIP(ip), since, s.mock.Ctx.Configuration.Regulation.MaxRetries).
			Return(records, nil),
		s.mock.StorageMock.EXPECT().
			SaveBannedIP(s.mock.Ctx, gomock.Eq(sqlban)).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByUser(s.mock.Ctx, "john", since, s.mock.Ctx.Configuration.Regulation.MaxRetries).
			Return(nil, nil),
	)

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeNone, ban)
	s.Equal("", value)
	s.Equal((*time.Time)(nil), expires)
	s.NoError(err)

	regulator.HandleAttempt(s.mock.Ctx, false, false, "john", "", "", regulation.AuthType1FA)
}

func (s *RegulatorSuite) TestShouldHandleBanCheckUserError() {
	regulator := s.Regulator()

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(net.ParseIP("127.0.0.1")))).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadBannedUser(s.mock.Ctx, gomock.Eq("john")).
			Return(nil, fmt.Errorf("failed to load user bans")),
	)

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeNone, ban)
	s.Equal("", value)
	s.Equal((*time.Time)(nil), expires)
	s.EqualError(err, "failed to load user bans")
}

func (s *RegulatorSuite) TestShouldHandleBanCheckUserBanned() {
	regulator := s.Regulator()

	ip := net.ParseIP("127.0.0.1")

	result := []model.BannedUser{
		{
			ID:       1,
			Time:     s.mock.Clock.Now().Add(-5 * time.Minute),
			Expires:  sql.NullTime{Valid: true, Time: s.mock.Clock.Now().Add(1 * time.Minute)},
			Expired:  sql.NullTime{},
			Revoked:  false,
			Username: "john",
			Source:   "regulation",
			Reason:   sql.NullString{Valid: true, String: "Exceeding Maximum Retries"},
		},
	}

	attempt := model.AuthenticationAttempt{
		Time:       s.mock.Clock.Now(),
		Successful: false,
		Banned:     true,
		Username:   "john",
		Type:       regulation.AuthType1FA,
		RemoteIP:   model.NewNullIP(ip),
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip))).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadBannedUser(s.mock.Ctx, gomock.Eq("john")).
			Return(result, nil),
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
	)

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeUser, ban)
	s.Equal("john", value)
	s.Equal(&result[0].Expires.Time, expires)
	s.EqualError(err, "user is banned")

	b := regulation.NewBan(ban, value, expires)

	s.Equal("expires at 12:01:00AM on February 3 2013 (+00:00)", b.FormatExpires())
	s.Equal(regulation.BanTypeUser, b.Type())
	s.Equal("john", b.Value())

	regulator.HandleAttempt(s.mock.Ctx, false, true, "john", "", "", regulation.AuthType1FA)
}

func (s *RegulatorSuite) TestShouldHandleBanCheckUserBannedPermanent() {
	regulator := s.Regulator()

	ip := net.ParseIP("127.0.0.1")

	result := []model.BannedUser{
		{
			ID:       1,
			Time:     s.mock.Clock.Now().Add(-5 * time.Minute),
			Expires:  sql.NullTime{},
			Expired:  sql.NullTime{},
			Revoked:  false,
			Username: "john",
			Source:   "regulation",
			Reason:   sql.NullString{Valid: true, String: "Exceeding Maximum Retries"},
		},
	}

	attempt := model.AuthenticationAttempt{
		Time:       s.mock.Clock.Now(),
		Successful: false,
		Banned:     true,
		Username:   "john",
		Type:       regulation.AuthType1FA,
		RemoteIP:   model.NewNullIP(ip),
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip))).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadBannedUser(s.mock.Ctx, gomock.Eq("john")).
			Return(result, nil),
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
	)

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeUser, ban)
	s.Equal("john", value)
	s.Equal((*time.Time)(nil), expires)
	s.EqualError(err, "user is banned")

	b := regulation.NewBan(ban, value, expires)

	s.Equal("never expires", b.FormatExpires())
	s.Equal(regulation.BanTypeUser, b.Type())
	s.Equal("john", b.Value())

	regulator.HandleAttempt(s.mock.Ctx, false, true, "john", "", "", regulation.AuthType1FA)
}

func (s *RegulatorSuite) TestShouldHandleBanCheckUserBannedFailToAppend() {
	regulator := s.Regulator()

	ip := net.ParseIP("127.0.0.1")

	result := []model.BannedUser{
		{
			ID:       1,
			Time:     s.mock.Clock.Now().Add(-5 * time.Minute),
			Expires:  sql.NullTime{Valid: true, Time: s.mock.Clock.Now().Add(1 * time.Minute)},
			Expired:  sql.NullTime{},
			Revoked:  false,
			Username: "john",
			Source:   "regulation",
			Reason:   sql.NullString{Valid: true, String: "Exceeding Maximum Retries"},
		},
	}

	attempt := model.AuthenticationAttempt{
		Time:       s.mock.Clock.Now(),
		Successful: false,
		Banned:     true,
		Username:   "john",
		Type:       regulation.AuthType1FA,
		RemoteIP:   model.NewNullIP(ip),
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip))).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadBannedUser(s.mock.Ctx, gomock.Eq("john")).
			Return(result, nil),
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(fmt.Errorf("failed to log")),
	)

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeUser, ban)
	s.Equal("john", value)
	s.Equal(&result[0].Expires.Time, expires)
	s.EqualError(err, "user is banned")

	regulator.HandleAttempt(s.mock.Ctx, false, true, "john", "", "", regulation.AuthType1FA)

	s.AssertLogEntryMessageAndError("Failed to record 1FA authentication attempt", "failed to log")
}

func (s *RegulatorSuite) TestShouldHandleBanCheckUserNotBanned() {
	regulator := s.Regulator()

	ip := net.ParseIP("127.0.0.1")

	attempt := model.AuthenticationAttempt{
		Time:       s.mock.Clock.Now(),
		Successful: false,
		Banned:     false,
		Username:   "john",
		Type:       regulation.AuthType1FA,
		RemoteIP:   model.NewNullIP(ip),
	}

	since := s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)

	records := []model.RegulationRecord{
		{
			Time:       s.mock.Clock.Now().Add(-10 * time.Second),
			Successful: false,
		},
		{
			Time:       s.mock.Clock.Now().Add(-12 * time.Second),
			Successful: false,
		},
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip))).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadBannedUser(s.mock.Ctx, gomock.Eq("john")).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByIP(s.mock.Ctx, model.NewIP(ip), since, s.mock.Ctx.Configuration.Regulation.MaxRetries).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByUser(s.mock.Ctx, "john", since, s.mock.Ctx.Configuration.Regulation.MaxRetries).
			Return(records, nil),
	)

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeNone, ban)
	s.Equal("", value)
	s.Equal((*time.Time)(nil), expires)
	s.NoError(err)

	regulator.HandleAttempt(s.mock.Ctx, false, false, "john", "", "", regulation.AuthType1FA)
}

func (s *RegulatorSuite) TestShouldHandleBanCheckUserNotBannedButFailedAttempt() {
	regulator := s.Regulator()

	ip := net.ParseIP("127.0.0.1")

	attempt := model.AuthenticationAttempt{
		Time:       s.mock.Clock.Now(),
		Successful: false,
		Banned:     false,
		Username:   "john",
		Type:       regulation.AuthType1FA,
		RemoteIP:   model.NewNullIP(ip),
	}

	since := s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)

	records := []model.RegulationRecord{
		{
			Time:       s.mock.Clock.Now().Add(-10 * time.Second),
			Successful: false,
		},
		{
			Time:       s.mock.Clock.Now().Add(-12 * time.Second),
			Successful: false,
		},
		{
			Time:       s.mock.Clock.Now().Add(-12 * time.Second),
			Successful: false,
		},
	}

	sqlban := &model.BannedUser{
		Expires:  sql.NullTime{Valid: true, Time: records[0].Time.Add(s.mock.Ctx.Configuration.Regulation.BanTime)},
		Username: "john",
		Source:   "regulation",
		Reason:   sql.NullString{Valid: true, String: "Exceeding Maximum Retries"},
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip))).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadBannedUser(s.mock.Ctx, gomock.Eq("john")).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByIP(s.mock.Ctx, model.NewIP(ip), since, s.mock.Ctx.Configuration.Regulation.MaxRetries).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByUser(s.mock.Ctx, "john", since, s.mock.Ctx.Configuration.Regulation.MaxRetries).
			Return(records, nil),
		s.mock.StorageMock.EXPECT().
			SaveBannedUser(s.mock.Ctx, gomock.Eq(sqlban)).
			Return(nil),
	)

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeNone, ban)
	s.Equal("", value)
	s.Equal((*time.Time)(nil), expires)
	s.NoError(err)

	regulator.HandleAttempt(s.mock.Ctx, false, false, "john", "", "", regulation.AuthType1FA)
}

// This test checks the case in which a user failed to authenticate many times but always
// with a certain amount of time larger than FindTime. Meaning the user should not be banned.
func (s *RegulatorSuite) TestShouldNotBanWhenFailedAuthenticationNotInFindTime() {
	attemptsInDB := []model.RegulationRecord{
		{
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-1 * time.Second),
		},
		{
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-90 * time.Second),
		},
		{
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-180 * time.Second),
		},
	}

	ip := net.ParseIP("127.0.0.1")

	attempt := model.AuthenticationAttempt{
		Time:       s.mock.Clock.Now(),
		Successful: false,
		Banned:     false,
		Username:   "john",
		Type:       regulation.AuthType1FA,
		RemoteIP:   model.NewNullIP(ip),
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip))).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadBannedUser(s.mock.Ctx, gomock.Eq("john")).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByIP(s.mock.Ctx, gomock.Eq(model.NewIP(net.ParseIP("127.0.0.1"))), gomock.Eq(s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)), gomock.Eq(3)).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByUser(s.mock.Ctx, gomock.Eq("john"), gomock.Eq(s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)), gomock.Eq(3)).
			Return(attemptsInDB, nil),
	)

	regulator := s.Regulator()

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeNone, ban)
	s.Equal("", value)
	s.Equal((*time.Time)(nil), expires)
	s.NoError(err)

	regulator.HandleAttempt(s.mock.Ctx, false, false, "john", "", "", regulation.AuthType1FA)
}

// This test checks the case in which a user failed to authenticate many times only a few
// seconds ago (meaning we are checking from now back to now-FindTime).
func (s *RegulatorSuite) TestShouldBanUserIfLatestAttemptsAreWithinFindTime() {
	attemptsInDB := []model.RegulationRecord{
		{
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-1 * time.Second),
		},
		{
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-4 * time.Second),
		},
		{
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-6 * time.Second),
		},
		{
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-180 * time.Second),
		},
	}

	banexp := s.mock.Clock.Now().Add(-1 * time.Second).Add(s.mock.Ctx.Configuration.Regulation.BanTime)

	banned := &model.BannedUser{
		Expires:  sql.NullTime{Time: banexp, Valid: true},
		Username: "john",
		Source:   "regulation",
		Reason:   sql.NullString{Valid: true, String: "Exceeding Maximum Retries"},
	}

	ip := net.ParseIP("127.0.0.1")

	attempt := model.AuthenticationAttempt{
		Time:       s.mock.Clock.Now(),
		Successful: false,
		Banned:     false,
		Username:   "john",
		Type:       regulation.AuthType1FA,
		RemoteIP:   model.NewNullIP(ip),
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(net.ParseIP("127.0.0.1")))).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadBannedUser(s.mock.Ctx, gomock.Eq("john")).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByIP(s.mock.Ctx, gomock.Eq(model.NewIP(net.ParseIP("127.0.0.1"))), gomock.Eq(s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)), gomock.Eq(3)).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByUser(s.mock.Ctx, gomock.Eq("john"), gomock.Eq(s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)), gomock.Eq(3)).
			Return(attemptsInDB, nil),
		s.mock.StorageMock.EXPECT().
			SaveBannedUser(s.mock.Ctx, gomock.Eq(banned)).
			Return(nil),
	)

	regulator := s.Regulator()

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeNone, ban)
	s.Equal("", value)
	s.Equal((*time.Time)(nil), expires)
	s.NoError(err)

	regulator.HandleAttempt(s.mock.Ctx, false, false, "john", "", "", regulation.AuthType1FA)
}

func (s *RegulatorSuite) TestShouldCheckUserWasAboutToBeBanned() {
	attemptsInDB := []model.RegulationRecord{
		{
			Successful: false,
			Time:       s.mock.Clock.Now(),
		},
		{
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-14 * time.Second),
		},
		// more than 30 seconds elapsed between this auth and the preceding one.
		// In that case we don't need to regulate the user even though the number
		// of retrieved attempts is 3.
		{
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-94 * time.Second),
		},
		{
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-96 * time.Second),
		},
	}

	ip := net.ParseIP("127.0.0.1")

	attempt := model.AuthenticationAttempt{
		Time:       s.mock.Clock.Now(),
		Successful: false,
		Banned:     false,
		Username:   "john",
		Type:       regulation.AuthType1FA,
		RemoteIP:   model.NewNullIP(ip),
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip))).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadBannedUser(s.mock.Ctx, gomock.Eq("john")).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip)), gomock.Eq(s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)), gomock.Eq(3)).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByUser(s.mock.Ctx, gomock.Eq("john"), gomock.Eq(s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)), gomock.Eq(3)).
			Return(attemptsInDB, nil),
	)

	regulator := s.Regulator()

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeNone, ban)
	s.Equal("", value)
	s.Equal((*time.Time)(nil), expires)
	s.NoError(err)

	regulator.HandleAttempt(s.mock.Ctx, false, false, "john", "", "", regulation.AuthType1FA)
}

func (s *RegulatorSuite) TestShouldCheckRegulationHasBeenResetOnSuccessfulAttempt() {
	attemptsInDB := []model.RegulationRecord{
		{
			Successful: false,
			Time:       s.mock.Clock.Now(),
		},
		{
			Successful: true,
			Time:       s.mock.Clock.Now().Add(-5 * time.Second),
		},
		// The user was almost banned but he did a successful attempt. Therefore, even if the next
		// failure happens within FindTime, he should not be banned.
		{
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-10 * time.Second),
		},
		{
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-15 * time.Second),
		},
		{
			Successful: false,
			Time:       s.mock.Clock.Now().Add(-20 * time.Second),
		},
	}

	ip := net.ParseIP("127.0.0.1")

	attempt := model.AuthenticationAttempt{
		Time:       s.mock.Clock.Now(),
		Successful: false,
		Banned:     false,
		Username:   "john",
		Type:       regulation.AuthType1FA,
		RemoteIP:   model.NewNullIP(ip),
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadBannedIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip))).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadBannedUser(s.mock.Ctx, gomock.Eq("john")).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByIP(s.mock.Ctx, gomock.Eq(model.NewIP(ip)), gomock.Eq(s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)), gomock.Eq(3)).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByUser(s.mock.Ctx, gomock.Eq("john"), gomock.Eq(s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)), gomock.Eq(3)).
			Return(attemptsInDB, nil),
	)

	regulator := s.Regulator()

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeNone, ban)
	s.Equal("", value)
	s.Equal((*time.Time)(nil), expires)
	s.NoError(err)

	regulator.HandleAttempt(s.mock.Ctx, false, false, "john", "", "", regulation.AuthType1FA)
}

// This test checks that the regulator is disabled when configuration is set to 0.
func (s *RegulatorSuite) TestShouldHaveRegulatorDisabled() {
	// Check Disabled Functionality.
	config := schema.Regulation{
		MaxRetries: 0,
		FindTime:   time.Second * 180,
		BanTime:    time.Second * 180,
	}

	regulator := regulation.NewRegulator(config, s.mock.StorageMock, &s.mock.Clock)

	s.mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil)
	s.mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq("john")).Return(nil, nil)

	ban, value, expires, err := regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeNone, ban)
	s.Equal("", value)
	s.Equal((*time.Time)(nil), expires)
	s.NoError(err)

	// Check Enabled Functionality.
	config = schema.Regulation{
		MaxRetries: 1,
		FindTime:   time.Second * 180,
		BanTime:    time.Second * 180,
	}

	regulator = regulation.NewRegulator(config, s.mock.StorageMock, &s.mock.Clock)

	s.mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil)
	s.mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq("john")).Return(nil, nil)

	ban, value, expires, err = regulator.BanCheck(s.mock.Ctx, "john")

	s.Equal(regulation.BanTypeNone, ban)
	s.Equal("", value)
	s.Equal((*time.Time)(nil), expires)
	s.NoError(err)
}

func TestRunRegulatorSuite(t *testing.T) {
	s := new(RegulatorSuite)
	suite.Run(t, s)
}
