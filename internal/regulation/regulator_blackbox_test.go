package regulation_test

import (
	"database/sql"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	s.mock.Ctx.Providers.Clock = &s.mock.Clock

	s.mock.Ctx.Configuration.Regulation = schema.Regulation{
		Modes:      []string{"ip", "user"},
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

	s.Regexp(`\d{2}:\d{2}:\d{2}(AM|PM) on \w+ \d{1,2} \d{4} \(\+\d{2}:\d{2}\)`, b.FormatExpires())
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

	s.Regexp(`\d{2}:\d{2}:\d{2}(AM|PM) on \w+ \d{1,2} \d{4} \(\+\d{2}:\d{2}\)`, b.FormatExpires())
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

func (s *RegulatorSuite) TestShouldHaveRegulatorDisabled() {
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

func (s *RegulatorSuite) TestShouldHandleLoadRegulationRecordsByIPError() {
	regulator := s.Regulator()

	ip := net.ParseIP("127.0.0.1")
	since := s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)

	attempt := model.AuthenticationAttempt{
		Time:          s.mock.Clock.Now(),
		Successful:    false,
		Banned:        false,
		Username:      "john",
		Type:          regulation.AuthType1FA,
		RemoteIP:      model.NewNullIP(ip),
		RequestURI:    "",
		RequestMethod: "",
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByIP(s.mock.Ctx, model.NewIP(ip), since, s.mock.Ctx.Configuration.Regulation.MaxRetries).
			Return(nil, fmt.Errorf("load ip records failed")),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByUser(s.mock.Ctx, "john", since, s.mock.Ctx.Configuration.Regulation.MaxRetries).
			Return(nil, nil),
	)

	regulator.HandleAttempt(s.mock.Ctx, false, false, "john", "", "", regulation.AuthType1FA)

	s.AssertLogEntryMessageAndError("Failed to load regulation records", "load ip records failed")
}

func (s *RegulatorSuite) TestShouldHandleSaveBannedIPError() {
	regulator := s.Regulator()

	ip := net.ParseIP("127.0.0.1")
	since := s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)

	records := []model.RegulationRecord{
		{Time: s.mock.Clock.Now().Add(-10 * time.Second), Successful: false},
		{Time: s.mock.Clock.Now().Add(-12 * time.Second), Successful: false},
		{Time: s.mock.Clock.Now().Add(-14 * time.Second), Successful: false},
	}

	sqlban := &model.BannedIP{
		Expires: sql.NullTime{Valid: true, Time: records[0].Time.Add(s.mock.Ctx.Configuration.Regulation.BanTime)},
		IP:      model.NewIP(ip),
		Source:  "regulation",
		Reason:  sql.NullString{Valid: true, String: "Exceeding Maximum Retries"},
	}

	attempt := model.AuthenticationAttempt{
		Time:          s.mock.Clock.Now(),
		Successful:    false,
		Banned:        false,
		Username:      "john",
		Type:          regulation.AuthType1FA,
		RemoteIP:      model.NewNullIP(ip),
		RequestURI:    "",
		RequestMethod: "",
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByIP(s.mock.Ctx, model.NewIP(ip), since, s.mock.Ctx.Configuration.Regulation.MaxRetries).
			Return(records, nil),
		s.mock.StorageMock.EXPECT().
			SaveBannedIP(s.mock.Ctx, gomock.Eq(sqlban)).
			Return(fmt.Errorf("save ip ban failed")),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByUser(s.mock.Ctx, "john", since, s.mock.Ctx.Configuration.Regulation.MaxRetries).
			Return(nil, nil),
	)

	regulator.HandleAttempt(s.mock.Ctx, false, false, "john", "", "", regulation.AuthType1FA)

	s.AssertLogEntryMessageAndError("Failed to save ban", "save ip ban failed")
}

func (s *RegulatorSuite) TestShouldHandleLoadRegulationRecordsByUserError() {
	regulator := s.Regulator()

	ip := net.ParseIP("127.0.0.1")
	since := s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)

	attempt := model.AuthenticationAttempt{
		Time:          s.mock.Clock.Now(),
		Successful:    false,
		Banned:        false,
		Username:      "john",
		Type:          regulation.AuthType1FA,
		RemoteIP:      model.NewNullIP(ip),
		RequestURI:    "",
		RequestMethod: "",
	}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByIP(s.mock.Ctx, model.NewIP(ip), since, s.mock.Ctx.Configuration.Regulation.MaxRetries).
			Return(nil, nil),
		s.mock.StorageMock.EXPECT().
			LoadRegulationRecordsByUser(s.mock.Ctx, "john", since, s.mock.Ctx.Configuration.Regulation.MaxRetries).
			Return(nil, fmt.Errorf("load user records failed")),
	)

	regulator.HandleAttempt(s.mock.Ctx, false, false, "john", "", "", regulation.AuthType1FA)

	s.AssertLogEntryMessageAndError("Failed to load regulation records", "load user records failed")
}

func (s *RegulatorSuite) TestShouldHandleSaveBannedUserError() {
	regulator := s.Regulator()

	ip := net.ParseIP("127.0.0.1")
	since := s.mock.Clock.Now().Add(-s.mock.Ctx.Configuration.Regulation.FindTime)

	records := []model.RegulationRecord{
		{Time: s.mock.Clock.Now().Add(-5 * time.Second), Successful: false},
		{Time: s.mock.Clock.Now().Add(-10 * time.Second), Successful: false},
		{Time: s.mock.Clock.Now().Add(-12 * time.Second), Successful: false},
	}

	sqlban := &model.BannedUser{
		Expires:  sql.NullTime{Valid: true, Time: records[0].Time.Add(s.mock.Ctx.Configuration.Regulation.BanTime)},
		Username: "john",
		Source:   "regulation",
		Reason:   sql.NullString{Valid: true, String: "Exceeding Maximum Retries"},
	}

	attempt := model.AuthenticationAttempt{
		Time:          s.mock.Clock.Now(),
		Successful:    false,
		Banned:        false,
		Username:      "john",
		Type:          regulation.AuthType1FA,
		RemoteIP:      model.NewNullIP(ip),
		RequestURI:    "",
		RequestMethod: "",
	}

	gomock.InOrder(
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
			Return(fmt.Errorf("save user ban failed")),
	)

	regulator.HandleAttempt(s.mock.Ctx, false, false, "john", "", "", regulation.AuthType1FA)

	s.AssertLogEntryMessageAndError("Failed to save ban", "save user ban failed")
}

func TestRunRegulatorSuite(t *testing.T) {
	s := new(RegulatorSuite)
	suite.Run(t, s)
}

func TestHandleAttemptShortCircuits(t *testing.T) {
	defaultCfg := schema.Regulation{
		Modes:      []string{"ip", "user"},
		MaxRetries: 3,
		BanTime:    3 * time.Minute,
		FindTime:   30 * time.Second,
	}
	ip := net.ParseIP("127.0.0.1")

	testCases := []struct {
		name            string
		config          schema.Regulation
		successful      bool
		banned          bool
		username        string
		authType        string
		expectAppendErr string
	}{
		{
			name:       "ShouldSkipBanChecksOnSuccessfulAttempt",
			config:     defaultCfg,
			successful: true,
			banned:     false,
			username:   "john",
			authType:   regulation.AuthType1FA,
		},
		{
			name:       "ShouldSkipBanChecksWhenAlreadyBanned",
			config:     defaultCfg,
			successful: false,
			banned:     true,
			username:   "john",
			authType:   regulation.AuthType1FA,
		},
		{
			name: "ShouldSkipBanChecksWhenRegulationDisabled",
			config: schema.Regulation{
				Modes:      []string{"ip", "user"},
				MaxRetries: 0,
				BanTime:    3 * time.Minute,
				FindTime:   30 * time.Second,
			},
			successful: false,
			banned:     false,
			username:   "john",
			authType:   regulation.AuthType1FA,
		},
		{
			name:       "ShouldSkipBanChecksForNon1FAType",
			config:     defaultCfg,
			successful: false,
			banned:     false,
			username:   "john",
			authType:   "2FA",
		},
		{
			name: "ShouldSkipUserChecksWhenUsernameEmptyAndIPModeDisabled",
			config: schema.Regulation{
				Modes:      []string{"user"},
				MaxRetries: 3,
				BanTime:    3 * time.Minute,
				FindTime:   30 * time.Second,
			},
			successful: false,
			banned:     false,
			username:   "",
			authType:   regulation.AuthType1FA,
		},
		{
			name: "ShouldLogErrorWhenAppendAuthenticationLogFails",
			config: schema.Regulation{
				Modes:      []string{"ip", "user"},
				MaxRetries: 0,
				BanTime:    3 * time.Minute,
				FindTime:   30 * time.Second,
			},
			successful:      false,
			banned:          false,
			username:        "john",
			authType:        regulation.AuthType1FA,
			expectAppendErr: "failed to append",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewMockAutheliaCtx(t)
			m.Ctx.Providers.Clock = &m.Clock
			m.Ctx.Configuration.Regulation = tc.config
			m.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedFor, "127.0.0.1")

			reg := regulation.NewRegulator(m.Ctx.Configuration.Regulation, m.StorageMock, &m.Clock)

			attempt := model.AuthenticationAttempt{
				Time:          m.Clock.Now(),
				Successful:    tc.successful,
				Banned:        tc.banned,
				Username:      tc.username,
				Type:          tc.authType,
				RemoteIP:      model.NewNullIP(ip),
				RequestURI:    "",
				RequestMethod: "",
			}

			if tc.expectAppendErr != "" {
				m.StorageMock.EXPECT().
					AppendAuthenticationLog(m.Ctx, gomock.Eq(attempt)).
					Return(assert.AnError).DoAndReturn(func(_ any, _ any) error {
					return assert.AnError
				})
			} else {
				m.StorageMock.EXPECT().
					AppendAuthenticationLog(m.Ctx, gomock.Eq(attempt)).
					Return(nil)
			}

			reg.HandleAttempt(m.Ctx, tc.successful, tc.banned, tc.username, "", "", tc.authType)

			if tc.expectAppendErr != "" {
				entry := m.Hook.LastEntry()
				require.NotNil(t, entry)
				assert.Equal(t, "Failed to record "+tc.authType+" authentication attempt", entry.Message)

				errField, ok := entry.Data["error"]
				require.True(t, ok)
				err, ok := errField.(error)
				require.True(t, ok)
				assert.Error(t, err)
			}
		})
	}
}
