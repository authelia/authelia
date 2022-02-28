package storage

import (
	"context"
	"time"

	"github.com/authelia/authelia/v4/internal/models"
)

// Provider is an interface providing storage capabilities for persisting any kind of data related to Authelia.
type Provider interface {
	models.StartupCheck

	RegulatorProvider

	SavePreferred2FAMethod(ctx context.Context, username string, method string) (err error)
	LoadPreferred2FAMethod(ctx context.Context, username string) (method string, err error)
	LoadUserInfo(ctx context.Context, username string) (info models.UserInfo, err error)

	SaveIdentityVerification(ctx context.Context, verification models.IdentityVerification) (err error)
	ConsumeIdentityVerification(ctx context.Context, jti string, ip models.NullIP) (err error)
	FindIdentityVerification(ctx context.Context, jti string) (found bool, err error)

	SaveTOTPConfiguration(ctx context.Context, config models.TOTPConfiguration) (err error)
	UpdateTOTPConfigurationSignIn(ctx context.Context, id int, lastUsedAt *time.Time) (err error)
	DeleteTOTPConfiguration(ctx context.Context, username string) (err error)
	LoadTOTPConfiguration(ctx context.Context, username string) (config *models.TOTPConfiguration, err error)
	LoadTOTPConfigurations(ctx context.Context, limit, page int) (configs []models.TOTPConfiguration, err error)

	SaveWebauthnDevice(ctx context.Context, device models.WebauthnDevice) (err error)
	UpdateWebauthnDeviceSignIn(ctx context.Context, id int, rpid string, lastUsedAt *time.Time, signCount uint32, cloneWarning bool) (err error)
	LoadWebauthnDevices(ctx context.Context, limit, page int) (devices []models.WebauthnDevice, err error)
	LoadWebauthnDevicesByUsername(ctx context.Context, username string) (devices []models.WebauthnDevice, err error)

	SavePreferredDuoDevice(ctx context.Context, device models.DuoDevice) (err error)
	DeletePreferredDuoDevice(ctx context.Context, username string) (err error)
	LoadPreferredDuoDevice(ctx context.Context, username string) (device *models.DuoDevice, err error)

	SchemaTables(ctx context.Context) (tables []string, err error)
	SchemaVersion(ctx context.Context) (version int, err error)
	SchemaLatestVersion() (version int, err error)

	SchemaMigrate(ctx context.Context, up bool, version int) (err error)
	SchemaMigrationHistory(ctx context.Context) (migrations []models.Migration, err error)
	SchemaMigrationsUp(ctx context.Context, version int) (migrations []models.SchemaMigration, err error)
	SchemaMigrationsDown(ctx context.Context, version int) (migrations []models.SchemaMigration, err error)

	SchemaEncryptionChangeKey(ctx context.Context, encryptionKey string) (err error)
	SchemaEncryptionCheckKey(ctx context.Context, verbose bool) (err error)

	Close() (err error)
}

// RegulatorProvider is an interface providing storage capabilities for persisting any kind of data related to the regulator.
type RegulatorProvider interface {
	AppendAuthenticationLog(ctx context.Context, attempt models.AuthenticationAttempt) (err error)
	LoadAuthenticationLogs(ctx context.Context, username string, fromDate time.Time, limit, page int) (attempts []models.AuthenticationAttempt, err error)
}
