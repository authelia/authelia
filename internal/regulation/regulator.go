package regulation

import (
	"database/sql"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewRegulator create a regulator instance.
func NewRegulator(config schema.Regulation, store storage.RegulatorProvider, clock clock.Provider) *Regulator {
	return &Regulator{
		users:  config.MaxRetries > 0 && utils.IsStringInSlice(typeUser, config.Modes),
		ips:    config.MaxRetries > 0 && utils.IsStringInSlice(typeIP, config.Modes),
		store:  store,
		clock:  clock,
		config: config,
	}
}

func (r *Regulator) HandleAttempt(ctx Context, successful, banned bool, username, requestURI, requestMethod, authType string) {
	ctx.RecordAuthn(successful, banned, strings.ToLower(authType))

	attempt := model.AuthenticationAttempt{
		Time:          r.clock.Now(),
		Successful:    successful,
		Banned:        banned,
		Username:      username,
		Type:          authType,
		RemoteIP:      model.NewNullIP(ctx.RemoteIP()),
		RequestURI:    requestURI,
		RequestMethod: requestMethod,
	}

	var err error
	if err = r.store.AppendAuthenticationLog(ctx, attempt); err != nil {
		ctx.GetLogger().WithFields(map[string]any{fieldUsername: username, "successful": successful}).WithError(err).Errorf("Failed to record %s authentication attempt", authType)
	}

	// We only need to perform the ban checks when; the attempt is unsuccessful, there is not an effective ban in place,
	// regulation is enabled, and the authentication type is 1FA. Thus if this is not the case we can return here.
	if successful || banned || (!r.ips && !r.users) || authType != AuthType1FA {
		return
	}

	since := r.clock.Now().Add(-r.config.FindTime)

	r.handleAttemptPossibleBannedIP(ctx, since)
	r.handleAttemptPossibleBannedUser(ctx, since, username)
}

func (r *Regulator) handleAttemptPossibleBannedIP(ctx Context, since time.Time) {
	if !r.ips {
		return
	}

	var (
		records []model.RegulationRecord
		err     error
	)

	ip := model.NewIP(ctx.RemoteIP())

	log := ctx.GetLogger()

	if records, err = r.store.LoadRegulationRecordsByIP(ctx, ip, since, r.config.MaxRetries); err != nil {
		log.WithFields(map[string]any{fieldRecordType: typeIP}).WithError(err).Error("Failed to load regulation records")

		return
	}

	banexp := r.expires(since, records)

	if banexp == nil {
		return
	}

	sqlban := &model.BannedIP{
		Expires: sql.NullTime{Valid: true, Time: *banexp},
		IP:      ip,
		Source:  "regulation",
		Reason:  sql.NullString{Valid: true, String: "Exceeding Maximum Retries"},
	}

	if err = r.store.SaveBannedIP(ctx, sqlban); err != nil {
		log.WithFields(map[string]any{fieldBanType: typeIP}).WithError(err).Error("Failed to save ban")

		return
	}
}

func (r *Regulator) handleAttemptPossibleBannedUser(ctx Context, since time.Time, username string) {
	if !r.users || username == "" {
		return
	}

	var (
		records []model.RegulationRecord
		err     error
	)

	log := ctx.GetLogger()

	if records, err = r.store.LoadRegulationRecordsByUser(ctx, username, since, r.config.MaxRetries); err != nil {
		log.WithFields(map[string]any{fieldRecordType: typeUser, fieldUsername: username}).WithError(err).Error("Failed to load regulation records")

		return
	}

	banexp := r.expires(since, records)

	if banexp == nil {
		return
	}

	sqlban := &model.BannedUser{
		Expires:  sql.NullTime{Valid: true, Time: *banexp},
		Username: username,
		Source:   "regulation",
		Reason:   sql.NullString{Valid: true, String: "Exceeding Maximum Retries"},
	}

	if err = r.store.SaveBannedUser(ctx, sqlban); err != nil {
		log.WithFields(map[string]any{fieldBanType: typeUser, fieldUsername: username}).WithError(err).Error("Failed to save ban")

		return
	}
}

func (r *Regulator) BanCheck(ctx Context, username string) (ban BanType, value string, expires *time.Time, err error) {
	ip := model.NewIP(ctx.RemoteIP())

	var bansIP []model.BannedIP

	if bansIP, err = r.store.LoadBannedIP(ctx, ip); err != nil {
		return BanTypeNone, "", nil, err
	}

	if len(bansIP) != 0 {
		b := bansIP[0]

		return returnBanResult(BanTypeIP, ip.String(), b.Expires)
	}

	var bansUser []model.BannedUser

	if bansUser, err = r.store.LoadBannedUser(ctx, username); err != nil {
		return BanTypeNone, "", nil, err
	}

	if len(bansUser) != 0 {
		b := bansUser[0]

		return returnBanResult(BanTypeUser, username, b.Expires)
	}

	return BanTypeNone, "", nil, nil
}

func (r *Regulator) expires(since time.Time, records []model.RegulationRecord) *time.Time {
	failures := make([]model.RegulationRecord, 0, len(records))

loop:
	for _, record := range records {
		switch {
		case record.Successful:
			break loop
		case len(failures) >= r.config.MaxRetries:
			continue
		case record.Time.Before(since):
			continue
		default:
			// We stop appending failed attempts once we find the first successful attempts or we reach
			// the configured number of retries, meaning the user is already banned.
			failures = append(failures, record)
		}
	}

	// If the number of failed attempts within the ban time is less than the max number of retries
	// then the user is not banned.
	if len(failures) < r.config.MaxRetries {
		return nil
	}

	expires := failures[0].Time.Add(r.config.BanTime)

	return &expires
}
