package storage_test

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestSQLProviderConn(t *testing.T) {
	testCases := []struct {
		name       string
		setupCtx   func(p *storage.SQLProvider, conn *mocks.MockSQLXConnection) context.Context
		expectConn func(p *storage.SQLProvider, conn *mocks.MockSQLXConnection) storage.SQLXConnection
	}{
		{
			name: "ShouldReturnDBWhenNoConnectionInContext",
			setupCtx: func(p *storage.SQLProvider, conn *mocks.MockSQLXConnection) context.Context {
				return context.Background()
			},
			expectConn: func(p *storage.SQLProvider, conn *mocks.MockSQLXConnection) storage.SQLXConnection {
				return nil
			},
		},
		{
			name: "ShouldReturnConnectionFromContext",
			setupCtx: func(p *storage.SQLProvider, conn *mocks.MockSQLXConnection) context.Context {
				return context.WithValue(context.Background(), storage.CtxKeyConnection, storage.SQLXConnection(conn))
			},
			expectConn: func(p *storage.SQLProvider, conn *mocks.MockSQLXConnection) storage.SQLXConnection {
				return conn
			},
		},
		{
			name: "ShouldReturnDBWhenContextValueIsNilConnection",
			setupCtx: func(p *storage.SQLProvider, conn *mocks.MockSQLXConnection) context.Context {
				return context.WithValue(context.Background(), storage.CtxKeyConnection, storage.SQLXConnection(nil))
			},
			expectConn: func(p *storage.SQLProvider, conn *mocks.MockSQLXConnection) storage.SQLXConnection {
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)
			conn := mocks.NewMockSQLXConnection(ctrl)
			p := storage.NewSQLProviderForTesting(db)

			ctx := tc.setupCtx(p, conn)
			actual := p.Conn(ctx)

			expected := tc.expectConn(p, conn)
			if expected == nil {
				assert.Equal(t, storage.SQLXConnection(db), actual)
			} else {
				assert.Equal(t, expected, actual)
			}
		})
	}
}

func TestSQLProviderTransactionHelpers(t *testing.T) {
	testCases := []struct {
		name   string
		runner func(t *testing.T, db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx, p *storage.SQLProvider)
	}{
		{
			name: "ShouldErrBeginTXWhenBeginTxxFails",
			runner: func(t *testing.T, db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx, p *storage.SQLProvider) {
				db.EXPECT().BeginTxx(gomock.Any(), gomock.Nil()).Return(nil, errors.New("begin failed"))

				ctx, err := p.BeginTX(context.Background())

				assert.Nil(t, ctx)
				assert.EqualError(t, err, "begin failed")
			},
		},
		{
			name: "ShouldBeginTXAndStoreInContext",
			runner: func(t *testing.T, db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx, p *storage.SQLProvider) {
				db.EXPECT().BeginTxx(gomock.Any(), gomock.Nil()).Return(tx, nil)

				ctx, err := p.BeginTX(context.Background())

				require.NoError(t, err)
				require.NotNil(t, ctx)

				stored, ok := ctx.Value(storage.CtxKeyTransaction).(storage.SQLXTx)
				require.True(t, ok)
				assert.Equal(t, storage.SQLXTx(tx), stored)
			},
		},
		{
			name: "ShouldNotCommitWhenNoTransactionInContext",
			runner: func(t *testing.T, db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx, p *storage.SQLProvider) {
				assert.EqualError(t, p.Commit(context.Background()), "could not retrieve tx")
			},
		},
		{
			name: "ShouldCommitWhenTransactionInContext",
			runner: func(t *testing.T, db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx, p *storage.SQLProvider) {
				tx.EXPECT().Commit().Return(nil)

				ctx := context.WithValue(context.Background(), storage.CtxKeyTransaction, storage.SQLXTx(tx))

				assert.NoError(t, p.Commit(ctx))
			},
		},
		{
			name: "ShouldPropagateCommitError",
			runner: func(t *testing.T, db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx, p *storage.SQLProvider) {
				tx.EXPECT().Commit().Return(errors.New("commit failed"))

				ctx := context.WithValue(context.Background(), storage.CtxKeyTransaction, storage.SQLXTx(tx))

				assert.EqualError(t, p.Commit(ctx), "commit failed")
			},
		},
		{
			name: "ShouldNotRollbackWhenNoTransactionInContext",
			runner: func(t *testing.T, db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx, p *storage.SQLProvider) {
				assert.EqualError(t, p.Rollback(context.Background()), "could not retrieve tx")
			},
		},
		{
			name: "ShouldRollbackWhenTransactionInContext",
			runner: func(t *testing.T, db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx, p *storage.SQLProvider) {
				tx.EXPECT().Rollback().Return(nil)

				ctx := context.WithValue(context.Background(), storage.CtxKeyTransaction, storage.SQLXTx(tx))

				assert.NoError(t, p.Rollback(ctx))
			},
		},
		{
			name: "ShouldPropagateRollbackError",
			runner: func(t *testing.T, db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx, p *storage.SQLProvider) {
				tx.EXPECT().Rollback().Return(errors.New("rollback failed"))

				ctx := context.WithValue(context.Background(), storage.CtxKeyTransaction, storage.SQLXTx(tx))

				assert.EqualError(t, p.Rollback(ctx), "rollback failed")
			},
		},
		{
			name: "ShouldCloseDelegatingToDB",
			runner: func(t *testing.T, db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx, p *storage.SQLProvider) {
				db.EXPECT().Close().Return(errors.New("close failed"))

				assert.EqualError(t, p.Close(), "close failed")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)
			tx := mocks.NewMockSQLXTx(ctrl)
			p := storage.NewSQLProviderForTesting(db)

			tc.runner(t, db, tx, p)
		})
	}
}

func TestSQLProviderSimpleExecErrors(t *testing.T) {
	testCases := []struct {
		name      string
		setup     func(db *mocks.MockSQLXDB)
		invoke    func(p *storage.SQLProvider) error
		expectErr string
	}{
		{
			name: "ShouldReturnErrSavePreferred2FAMethod",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "john", "totp").Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SavePreferred2FAMethod(context.Background(), "john", "totp")
			},
			expectErr: "error upserting preferred two factor method for user 'john': boom",
		},
		{
			name: "ShouldReturnErrUpdateTOTPConfigurationSignIn",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), 1).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.UpdateTOTPConfigurationSignIn(context.Background(), 1, sql.NullTime{Valid: true, Time: time.Now()})
			},
			expectErr: "error updating TOTP configuration id 1: boom",
		},
		{
			name: "ShouldReturnErrDeleteTOTPConfiguration",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "john").Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.DeleteTOTPConfiguration(context.Background(), "john")
			},
			expectErr: "error deleting TOTP configuration for user 'john': boom",
		},
		{
			name: "ShouldReturnErrSaveTOTPHistory",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "alice", gomock.Any()).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveTOTPHistory(context.Background(), "alice", 1)
			},
			expectErr: "error inserting TOTP history for user 'alice': boom",
		},
		{
			name: "ShouldReturnErrSaveWebAuthnUser",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "example.com", "john", "uid").Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveWebAuthnUser(context.Background(), model.WebAuthnUser{RPID: "example.com", Username: "john", UserID: "uid"})
			},
			expectErr: "error inserting WebAuthn user 'john' with relying party id 'example.com': boom",
		},
		{
			name: "ShouldReturnErrUpdateWebAuthnCredentialDescription",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "new-desc", "john", 42).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.UpdateWebAuthnCredentialDescription(context.Background(), "john", 42, "new-desc")
			},
			expectErr: "error updating WebAuthn credential description to 'new-desc' for credential id '42': boom",
		},
		{
			name: "ShouldReturnErrDeleteWebAuthnCredential",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "mykid").Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.DeleteWebAuthnCredential(context.Background(), "mykid")
			},
			expectErr: "error deleting WebAuthn credential with kid 'mykid': boom",
		},
		{
			name:  "ShouldReturnErrDeleteWebAuthnCredentialByUsernameNoUsername",
			setup: nil,
			invoke: func(p *storage.SQLProvider) error {
				return p.DeleteWebAuthnCredentialByUsername(context.Background(), "", "")
			},
			expectErr: "error deleting WebAuthn credential with username '' and displayname '': username must not be empty",
		},
		{
			name: "ShouldReturnErrDeleteWebAuthnCredentialByUsernameOnly",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "john").Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.DeleteWebAuthnCredentialByUsername(context.Background(), "john", "")
			},
			expectErr: "error deleting WebAuthn credential for username 'john': boom",
		},
		{
			name: "ShouldReturnErrDeleteWebAuthnCredentialByUsernameAndDisplayName",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "john", "mykey").Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.DeleteWebAuthnCredentialByUsername(context.Background(), "john", "mykey")
			},
			expectErr: "error deleting WebAuthn credential with username 'john' and displayname 'mykey': boom",
		},
		{
			name: "ShouldReturnErrSavePreferredDuoDevice",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "john", "DXYZ", "push").Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SavePreferredDuoDevice(context.Background(), model.DuoDevice{Username: "john", Device: "DXYZ", Method: "push"})
			},
			expectErr: "error upserting preferred duo device for user 'john': boom",
		},
		{
			name: "ShouldReturnErrDeletePreferredDuoDevice",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "john").Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.DeletePreferredDuoDevice(context.Background(), "john")
			},
			expectErr: "error deleting preferred duo device for user 'john': boom",
		},
		{
			name: "ShouldReturnErrConsumeIdentityVerification",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "jti").Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.ConsumeIdentityVerification(context.Background(), "jti", model.NullIP{})
			},
			expectErr: "error updating identity verification: boom",
		},
		{
			name: "ShouldReturnErrRevokeIdentityVerification",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "jti").Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.RevokeIdentityVerification(context.Background(), "jti", model.NullIP{})
			},
			expectErr: "error updating identity verification: boom",
		},
		{
			name: "ShouldReturnErrSaveOAuth2ConsentSessionGranted",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), 7).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveOAuth2ConsentSessionGranted(context.Background(), 7)
			},
			expectErr: "error updating oauth2 consent session (granted) with id '7': boom",
		},
		{
			name: "ShouldReturnErrRevokeOAuth2PARContext",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "sig").Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.RevokeOAuth2PARContext(context.Background(), "sig")
			},
			expectErr: "error revoking oauth2 pushed authorization request context with signature 'sig': boom",
		},
		{
			name: "ShouldReturnErrSaveOAuth2BlacklistedJTI",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "sig", gomock.Any()).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveOAuth2BlacklistedJTI(context.Background(), model.OAuth2BlacklistedJTI{Signature: "sig", ExpiresAt: time.Now()})
			},
			expectErr: "error inserting oauth2 blacklisted JTI with signature 'sig': boom",
		},
		{
			name: "ShouldReturnErrAppendAuthenticationLog",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "john", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.AppendAuthenticationLog(context.Background(), model.AuthenticationAttempt{Username: "john"})
			},
			expectErr: "error inserting authentication attempt for user 'john': boom",
		},
		{
			name: "ShouldReturnErrSaveBannedUser",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), "baduser", "cli", gomock.Any()).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveBannedUser(context.Background(), &model.BannedUser{Username: "baduser", Source: "cli"})
			},
			expectErr: "error inserting banned user with username 'baduser' and source 'cli' and reason '': boom",
		},
		{
			name: "ShouldReturnErrRevokeBannedUser",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), 3).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.RevokeBannedUser(context.Background(), 3, time.Now())
			},
			expectErr: "error revoking banned user with id '3': boom",
		},
		{
			name: "ShouldReturnErrSaveBannedIP",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "cli", gomock.Any()).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveBannedIP(context.Background(), &model.BannedIP{IP: model.NewIP(net.ParseIP("10.0.0.1")), Source: "cli"})
			},
			expectErr: "error inserting banned ip with ip '10.0.0.1' and source 'cli' and reason '': boom",
		},
		{
			name: "ShouldReturnErrRevokeBannedIP",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), 5).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.RevokeBannedIP(context.Background(), 5, time.Now())
			},
			expectErr: "error revoking banned ip with id '5': boom",
		},
		{
			name: "ShouldReturnErrSaveUserOpaqueIdentifier",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "openid", "example.com", "john", uuid.Nil).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveUserOpaqueIdentifier(context.Background(), model.UserOpaqueIdentifier{
					Service: "openid", SectorID: "example.com", Username: "john", Identifier: uuid.Nil,
				})
			},
			expectErr: "error inserting user opaque id for user 'john' with opaque id '00000000-0000-0000-0000-000000000000': boom",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)

			if tc.setup != nil {
				tc.setup(db)
			}

			p := storage.NewSQLProviderForTesting(db)

			assert.EqualError(t, tc.invoke(p), tc.expectErr)
		})
	}
}

func TestSQLProviderLoadErrors(t *testing.T) {
	testCases := []struct {
		name      string
		setup     func(db *mocks.MockSQLXDB)
		invoke    func(p *storage.SQLProvider) error
		expectErr string
	}{
		{
			name: "ShouldReturnNoRowsForLoadPreferred2FAMethod",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "john").Return(sql.ErrNoRows)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadPreferred2FAMethod(context.Background(), "john")
				return err
			},
			expectErr: sql.ErrNoRows.Error(),
		},
		{
			name: "ShouldReturnErrLoadPreferred2FAMethod",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "john").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadPreferred2FAMethod(context.Background(), "john")
				return err
			},
			expectErr: "error selecting preferred two factor method for user 'john': boom",
		},
		{
			name: "ShouldReturnErrLoadUserInfo",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "john", "john", "john", "john").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadUserInfo(context.Background(), "john")
				return err
			},
			expectErr: "error selecting user info for user 'john': boom",
		},
		{
			name: "ShouldReturnErrLoadUserOpaqueIdentifier",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), uuid.Nil).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadUserOpaqueIdentifier(context.Background(), uuid.Nil)
				return err
			},
			expectErr: "error selecting user opaque id with value '00000000-0000-0000-0000-000000000000': boom",
		},
		{
			name: "ShouldReturnErrLoadUserOpaqueIdentifiersQuery",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().QueryxContext(gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadUserOpaqueIdentifiers(context.Background())
				return err
			},
			expectErr: "error selecting user opaque identifiers: boom",
		},
		{
			name: "ShouldReturnErrLoadUserOpaqueIdentifierBySignature",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "openid", "example.com", "john").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadUserOpaqueIdentifierBySignature(context.Background(), "openid", "example.com", "john")
				return err
			},
			expectErr: "error selecting user opaque with service 'openid' and sector 'example.com' for username 'john': boom",
		},
		{
			name: "ShouldReturnErrLoadTOTPConfiguration",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "john").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadTOTPConfiguration(context.Background(), "john")
				return err
			},
			expectErr: "error selecting TOTP configuration for user 'john': boom",
		},
		{
			name: "ShouldReturnErrNoTOTPConfiguration",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "john").Return(sql.ErrNoRows)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadTOTPConfiguration(context.Background(), "john")
				return err
			},
			expectErr: storage.ErrNoTOTPConfiguration.Error(),
		},
		{
			name: "ShouldReturnErrExistsTOTPHistory",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "alice", gomock.Any()).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.ExistsTOTPHistory(context.Background(), "alice", 1)
				return err
			},
			expectErr: "error checking if TOTP history exists: boom",
		},
		{
			name: "ShouldReturnErrLoadTOTPConfigurations",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), 10, 0).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadTOTPConfigurations(context.Background(), 10, 0)
				return err
			},
			expectErr: "error selecting TOTP configurations: boom",
		},
		{
			name: "ShouldReturnErrLoadWebAuthnUser",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "example.com", "john").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnUser(context.Background(), "example.com", "john")
				return err
			},
			expectErr: "error selecting WebAuthn user 'john' with relying party id 'example.com': boom",
		},
		{
			name: "ShouldReturnErrLoadWebAuthnUserByUserID",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "example.com", "uid").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnUserByUserID(context.Background(), "example.com", "uid")
				return err
			},
			expectErr: "error selecting WebAuthn user with user id 'uid' and relying party id 'example.com': boom",
		},
		{
			name: "ShouldReturnErrLoadWebAuthnCredentialByID",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), 42).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnCredentialByID(context.Background(), 42)
				return err
			},
			expectErr: "error selecting WebAuthn credential with id '42': boom",
		},
		{
			name: "ShouldReturnNoRowsLoadWebAuthnCredentialByID",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), 42).Return(sql.ErrNoRows)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnCredentialByID(context.Background(), 42)
				return err
			},
			expectErr: sql.ErrNoRows.Error(),
		},
		{
			name: "ShouldReturnErrLoadWebAuthnCredentials",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), 10, 0).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnCredentials(context.Background(), 10, 0)
				return err
			},
			expectErr: "error selecting WebAuthn credentials: boom",
		},
		{
			name: "ShouldReturnErrLoadWebAuthnCredentialsByUsernameWithRPID",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), "example.com", "john", false).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnCredentialsByUsername(context.Background(), "example.com", "john")
				return err
			},
			expectErr: "error selecting WebAuthn credentials for user 'john': boom",
		},
		{
			name: "ShouldReturnNoCredentialLoadWebAuthnCredentialsByUsername",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), "example.com", "john", false).Return(sql.ErrNoRows)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnCredentialsByUsername(context.Background(), "example.com", "john")
				return err
			},
			expectErr: storage.ErrNoWebAuthnCredential.Error(),
		},
		{
			name: "ShouldReturnErrLoadWebAuthnCredentialsByUsernameNoRPID",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), "john", false).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnCredentialsByUsername(context.Background(), "", "john")
				return err
			},
			expectErr: "error selecting WebAuthn credentials for user 'john': boom",
		},
		{
			name: "ShouldReturnErrLoadWebAuthnPasskeyCredentialsByUsername",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), "example.com", "john", true).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnPasskeyCredentialsByUsername(context.Background(), "example.com", "john")
				return err
			},
			expectErr: "error selecting passkey WebAuthn credentials for user 'john': boom",
		},
		{
			name: "ShouldReturnErrLoadWebAuthnPasskeyCredentialsByUsernameNoRPID",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), "john", true).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnPasskeyCredentialsByUsername(context.Background(), "", "john")
				return err
			},
			expectErr: "error selecting passkey WebAuthn credentials for user 'john': boom",
		},
		{
			name: "ShouldReturnNoCredentialLoadWebAuthnPasskeyCredentialsByUsername",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), "example.com", "john", true).Return(sql.ErrNoRows)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnPasskeyCredentialsByUsername(context.Background(), "example.com", "john")
				return err
			},
			expectErr: storage.ErrNoWebAuthnCredential.Error(),
		},
		{
			name: "ShouldReturnErrFindIdentityVerification",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "jti").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.FindIdentityVerification(context.Background(), "jti")
				return err
			},
			expectErr: "error selecting identity verification exists: boom",
		},
		{
			name: "ShouldReturnErrLoadIdentityVerification",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "jti").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadIdentityVerification(context.Background(), "jti")
				return err
			},
			expectErr: "error selecting identity verification: boom",
		},
		{
			name: "ShouldReturnErrLoadOneTimeCode",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "john").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadOneTimeCode(context.Background(), "john", model.NewIP(net.ParseIP("127.0.0.1")), "reset", "raw")
				return err
			},
			expectErr: "error selecting one-time code: boom",
		},
		{
			name: "ShouldReturnErrLoadOneTimeCodeBySignature",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "sig").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadOneTimeCodeBySignature(context.Background(), "sig")
				return err
			},
			expectErr: "error selecting one-time code: boom",
		},
		{
			name: "ShouldReturnErrLoadOneTimeCodeByID",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), 5).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadOneTimeCodeByID(context.Background(), 5)
				return err
			},
			expectErr: "error selecting one-time code: boom",
		},
		{
			name: "ShouldReturnErrLoadOneTimeCodeByPublicID",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				id, _ := uuid.NewRandom()
				_, err := p.LoadOneTimeCodeByPublicID(context.Background(), id)

				return err
			},
			expectErr: "error selecting one-time code: boom",
		},
		{
			name: "ShouldReturnErrLoadOAuth2ConsentSessionByChallengeID",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), uuid.Nil).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadOAuth2ConsentSessionByChallengeID(context.Background(), uuid.Nil)
				return err
			},
			expectErr: "error selecting oauth2 consent session with challenge id '00000000-0000-0000-0000-000000000000': boom",
		},
		{
			name: "ShouldReturnErrLoadOAuth2ConsentPreConfigurations",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().QueryxContext(gomock.Any(), gomock.Any(), "client", uuid.Nil, gomock.Any()).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadOAuth2ConsentPreConfigurations(context.Background(), "client", uuid.Nil, time.Now())
				return err
			},
			expectErr: "error selecting oauth2 consent pre-configurations by signature with client id 'client' and subject '00000000-0000-0000-0000-000000000000': boom",
		},
		{
			name: "ShouldReturnErrLoadOAuth2PARContext",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "sig").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadOAuth2PARContext(context.Background(), "sig")
				return err
			},
			expectErr: "error selecting oauth2 pushed authorization request context with signature 'sig': boom",
		},
		{
			name: "ShouldReturnErrLoadOAuth2BlacklistedJTI",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "sig").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadOAuth2BlacklistedJTI(context.Background(), "sig")
				return err
			},
			expectErr: "error selecting oauth2 blacklisted JTI with signature 'sig': boom",
		},
		{
			name: "ShouldReturnErrLoadBannedUser",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), "baduser", gomock.Any()).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadBannedUser(context.Background(), "baduser")
				return err
			},
			expectErr: "error selecting banned user records for username 'baduser': boom",
		},
		{
			name: "ShouldReturnErrLoadBannedUserByID",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), 1).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadBannedUserByID(context.Background(), 1)
				return err
			},
			expectErr: "error selecting banned user with id '1': boom",
		},
		{
			name: "ShouldReturnErrLoadBannedUsers",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), false, gomock.Any(), 10, 0).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadBannedUsers(context.Background(), 10, 0)
				return err
			},
			expectErr: "error selecting banned user records: boom",
		},
		{
			name: "ShouldReturnErrLoadBannedIP",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), false, gomock.Any()).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadBannedIP(context.Background(), model.NewIP(net.ParseIP("1.2.3.4")))
				return err
			},
			expectErr: "error selecting banned ip records for ip '1.2.3.4': boom",
		},
		{
			name: "ShouldReturnErrLoadBannedIPByID",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), 1).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadBannedIPByID(context.Background(), 1)
				return err
			},
			expectErr: "error selecting banned ip with id '1': boom",
		},
		{
			name: "ShouldReturnErrLoadBannedIPs",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), 10, 0).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadBannedIPs(context.Background(), 10, 0)
				return err
			},
			expectErr: "error selecting banned ip records: boom",
		},
		{
			name: "ShouldReturnErrLoadCachedData",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "name").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadCachedData(context.Background(), "name")
				return err
			},
			expectErr: "error selecting cached data with name 'name': boom",
		},
		{
			name: "ShouldReturnErrDeleteCachedData",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "name").Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.DeleteCachedData(context.Background(), "name")
			},
			expectErr: "error deleting cached data with name 'name': boom",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)

			if tc.setup != nil {
				tc.setup(db)
			}

			p := storage.NewSQLProviderForTesting(db)

			assert.EqualError(t, tc.invoke(p), tc.expectErr)
		})
	}
}

func TestSQLProviderLoadNilOnNoRows(t *testing.T) {
	testCases := []struct {
		name   string
		setup  func(db *mocks.MockSQLXDB)
		invoke func(t *testing.T, p *storage.SQLProvider)
	}{
		{
			name: "ShouldReturnNilLoadUserOpaqueIdentifierForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				id, _ := uuid.NewRandom()
				subj, err := p.LoadUserOpaqueIdentifier(context.Background(), id)
				assert.NoError(t, err)
				assert.Nil(t, subj)
			},
		},
		{
			name: "ShouldReturnNilLoadUserOpaqueIdentifierBySignatureNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "openid", "example.com", "nobody").Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				subj, err := p.LoadUserOpaqueIdentifierBySignature(context.Background(), "openid", "example.com", "nobody")
				assert.NoError(t, err)
				assert.Nil(t, subj)
			},
		},
		{
			name: "ShouldReturnNilLoadTOTPConfigurationsForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), 10, 0).Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				configs, err := p.LoadTOTPConfigurations(context.Background(), 10, 0)
				assert.NoError(t, err)
				assert.Nil(t, configs)
			},
		},
		{
			name: "ShouldReturnNilLoadWebAuthnUserForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "example.com", "john").Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				user, err := p.LoadWebAuthnUser(context.Background(), "example.com", "john")
				assert.NoError(t, err)
				assert.Nil(t, user)
			},
		},
		{
			name: "ShouldReturnNilLoadWebAuthnUserByUserIDForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "example.com", "uid").Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				user, err := p.LoadWebAuthnUserByUserID(context.Background(), "example.com", "uid")
				assert.NoError(t, err)
				assert.Nil(t, user)
			},
		},
		{
			name: "ShouldReturnNilLoadWebAuthnCredentialsForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), 10, 0).Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				creds, err := p.LoadWebAuthnCredentials(context.Background(), 10, 0)
				assert.NoError(t, err)
				assert.Nil(t, creds)
			},
		},
		{
			name: "ShouldReturnFalseFindIdentityVerificationNotFound",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "jti").Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				found, err := p.FindIdentityVerification(context.Background(), "jti")
				assert.NoError(t, err)
				assert.False(t, found)
			},
		},
		{
			name: "ShouldReturnNilLoadOneTimeCodeForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "john").Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				code, err := p.LoadOneTimeCode(context.Background(), "john", model.NewIP(net.ParseIP("127.0.0.1")), "reset", "raw")
				assert.NoError(t, err)
				assert.Nil(t, code)
			},
		},
		{
			name: "ShouldReturnNilLoadOneTimeCodeBySignatureForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "sig").Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				code, err := p.LoadOneTimeCodeBySignature(context.Background(), "sig")
				assert.NoError(t, err)
				assert.Nil(t, code)
			},
		},
		{
			name: "ShouldReturnNilLoadOneTimeCodeByIDForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), 5).Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				code, err := p.LoadOneTimeCodeByID(context.Background(), 5)
				assert.NoError(t, err)
				assert.Nil(t, code)
			},
		},
		{
			name: "ShouldReturnNilLoadOneTimeCodeByPublicIDForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				id, _ := uuid.NewRandom()
				code, err := p.LoadOneTimeCodeByPublicID(context.Background(), id)
				assert.NoError(t, err)
				assert.Nil(t, code)
			},
		},
		{
			name: "ShouldReturnEmptyLoadOAuth2ConsentPreConfigurationsForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().QueryxContext(gomock.Any(), gomock.Any(), "client", gomock.Any(), gomock.Any()).Return(nil, sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				id, _ := uuid.NewRandom()
				rows, err := p.LoadOAuth2ConsentPreConfigurations(context.Background(), "client", id, time.Now())
				assert.NoError(t, err)
				assert.NotNil(t, rows)
			},
		},
		{
			name: "ShouldReturnNilLoadBannedUserForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), "baduser", gomock.Any()).Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				bans, err := p.LoadBannedUser(context.Background(), "baduser")
				assert.NoError(t, err)
				assert.Nil(t, bans)
			},
		},
		{
			name: "ShouldReturnNilLoadBannedUsersForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), false, gomock.Any(), 10, 0).Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				bans, err := p.LoadBannedUsers(context.Background(), 10, 0)
				assert.NoError(t, err)
				assert.Nil(t, bans)
			},
		},
		{
			name: "ShouldReturnNilLoadBannedIPForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), false, gomock.Any()).Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				bans, err := p.LoadBannedIP(context.Background(), model.NewIP(net.ParseIP("1.2.3.4")))
				assert.NoError(t, err)
				assert.Nil(t, bans)
			},
		},
		{
			name: "ShouldReturnNilLoadBannedIPsForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), 10, 0).Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				bans, err := p.LoadBannedIPs(context.Background(), 10, 0)
				assert.NoError(t, err)
				assert.Nil(t, bans)
			},
		},
		{
			name: "ShouldReturnNilLoadCachedDataForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "name").Return(sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				data, err := p.LoadCachedData(context.Background(), "name")
				assert.NoError(t, err)
				assert.Nil(t, data)
			},
		},
		{
			name: "ShouldReturnNilDeleteCachedDataForNoRows",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "name").Return(nil, sql.ErrNoRows)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				assert.NoError(t, p.DeleteCachedData(context.Background(), "name"))
			},
		},
		{
			name: "ShouldReturnNilLoadWebAuthnPasskeyCredentialsByUsernameSuccess",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), "example.com", "john", true).Return(nil)
			},
			invoke: func(t *testing.T, p *storage.SQLProvider) {
				creds, err := p.LoadWebAuthnPasskeyCredentialsByUsername(context.Background(), "example.com", "john")
				assert.NoError(t, err)
				assert.Nil(t, creds)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)

			if tc.setup != nil {
				tc.setup(db)
			}

			p := storage.NewSQLProviderForTesting(db)

			tc.invoke(t, p)
		})
	}
}

func TestSQLProviderOAuth2SessionUnknownType(t *testing.T) {
	testCases := []struct {
		name      string
		invoke    func(p *storage.SQLProvider) error
		expectErr string
	}{
		{
			name: "ShouldErrSaveOAuth2SessionUnknownType",
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveOAuth2Session(context.Background(), storage.OAuth2SessionType(-1), model.OAuth2Session{RequestID: "req"})
			},
			expectErr: "error inserting oauth2 session for subject '' and request id 'req': unknown oauth2 session type 'invalid'",
		},
		{
			name: "ShouldErrRevokeOAuth2SessionUnknownType",
			invoke: func(p *storage.SQLProvider) error {
				return p.RevokeOAuth2Session(context.Background(), storage.OAuth2SessionType(-1), "sig")
			},
			expectErr: "error revoking oauth2 session with signature 'sig': unknown oauth2 session type 'invalid'",
		},
		{
			name: "ShouldErrRevokeOAuth2SessionByRequestIDUnknownType",
			invoke: func(p *storage.SQLProvider) error {
				return p.RevokeOAuth2SessionByRequestID(context.Background(), storage.OAuth2SessionType(-1), "req")
			},
			expectErr: "error revoking oauth2 session with request id 'req': unknown oauth2 session type 'invalid'",
		},
		{
			name: "ShouldErrDeactivateOAuth2SessionUnknownType",
			invoke: func(p *storage.SQLProvider) error {
				return p.DeactivateOAuth2Session(context.Background(), storage.OAuth2SessionType(-1), "sig")
			},
			expectErr: "error deactivating oauth2 session with signature 'sig': unknown oauth2 session type 'invalid'",
		},
		{
			name: "ShouldErrDeactivateOAuth2SessionByRequestIDUnknownType",
			invoke: func(p *storage.SQLProvider) error {
				return p.DeactivateOAuth2SessionByRequestID(context.Background(), storage.OAuth2SessionType(-1), "req")
			},
			expectErr: "error deactivating oauth2 session with request id 'req': unknown oauth2 session type 'invalid'",
		},
		{
			name: "ShouldErrLoadOAuth2SessionUnknownType",
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadOAuth2Session(context.Background(), storage.OAuth2SessionType(-1), "sig")
				return err
			},
			expectErr: "error selecting oauth2 session: unknown oauth2 session type 'invalid'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)
			p := storage.NewSQLProviderForTesting(db)

			assert.EqualError(t, tc.invoke(p), tc.expectErr)
		})
	}
}

func TestSQLProviderUpdateOAuth2PARContextZeroID(t *testing.T) {
	t.Run("ShouldErrUpdateOAuth2PARContextWithZeroID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db := mocks.NewMockSQLXDB(ctrl)
		p := storage.NewSQLProviderForTesting(db)

		err := p.UpdateOAuth2PARContext(context.Background(), model.OAuth2PARContext{Signature: "sig", RequestID: "req"})

		assert.EqualError(t, err, "error updating oauth2 pushed authorization request context data with signature 'sig' and request id 'req': the id was a zero value")
	})
}

func TestSQLProviderConsumeRevokeOneTimeCodeRowsAffected(t *testing.T) {
	testCases := []struct {
		name      string
		setup     func(db *mocks.MockSQLXDB, result *mocks.MockSQLResult)
		invoke    func(p *storage.SQLProvider) error
		expectErr string
	}{
		{
			name: "ShouldErrConsumeWhenExecFails",
			setup: func(db *mocks.MockSQLXDB, result *mocks.MockSQLResult) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "sig").Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.ConsumeOneTimeCode(context.Background(), &model.OneTimeCode{Signature: "sig"})
			},
			expectErr: "error updating one-time code (consume): boom",
		},
		{
			name: "ShouldErrConsumeWhenRowsAffectedErrors",
			setup: func(db *mocks.MockSQLXDB, result *mocks.MockSQLResult) {
				result.EXPECT().RowsAffected().Return(int64(0), errors.New("ra-err"))
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "sig").Return(result, nil)
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.ConsumeOneTimeCode(context.Background(), &model.OneTimeCode{Signature: "sig"})
			},
			expectErr: "error updating one-time code (consume): ra-err",
		},
		{
			name: "ShouldErrConsumeWhenNoRowsAffected",
			setup: func(db *mocks.MockSQLXDB, result *mocks.MockSQLResult) {
				result.EXPECT().RowsAffected().Return(int64(0), nil)
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "sig").Return(result, nil)
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.ConsumeOneTimeCode(context.Background(), &model.OneTimeCode{Signature: "sig"})
			},
			expectErr: "error updating one-time code (consume): no rows affected",
		},
		{
			name: "ShouldErrConsumeWhenMultipleRowsAffected",
			setup: func(db *mocks.MockSQLXDB, result *mocks.MockSQLResult) {
				result.EXPECT().RowsAffected().Return(int64(2), nil)
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "sig").Return(result, nil)
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.ConsumeOneTimeCode(context.Background(), &model.OneTimeCode{Signature: "sig"})
			},
			expectErr: "error updating one-time code (consume): multiple rows affected",
		},
		{
			name: "ShouldSucceedConsumeOneTimeCode",
			setup: func(db *mocks.MockSQLXDB, result *mocks.MockSQLResult) {
				result.EXPECT().RowsAffected().Return(int64(1), nil)
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "sig").Return(result, nil)
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.ConsumeOneTimeCode(context.Background(), &model.OneTimeCode{Signature: "sig"})
			},
		},
		{
			name: "ShouldErrRevokeWhenExecFails",
			setup: func(db *mocks.MockSQLXDB, result *mocks.MockSQLResult) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.RevokeOneTimeCode(context.Background(), uuid.Nil, model.NewIP(net.ParseIP("1.2.3.4")))
			},
			expectErr: "error updating one-time code (revoke): boom",
		},
		{
			name: "ShouldErrRevokeWhenNoRowsAffected",
			setup: func(db *mocks.MockSQLXDB, result *mocks.MockSQLResult) {
				result.EXPECT().RowsAffected().Return(int64(0), nil)
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(result, nil)
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.RevokeOneTimeCode(context.Background(), uuid.Nil, model.NewIP(net.ParseIP("1.2.3.4")))
			},
			expectErr: "error updating one-time code (consume): no rows affected",
		},
		{
			name: "ShouldErrRevokeWhenMultipleRowsAffected",
			setup: func(db *mocks.MockSQLXDB, result *mocks.MockSQLResult) {
				result.EXPECT().RowsAffected().Return(int64(2), nil)
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(result, nil)
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.RevokeOneTimeCode(context.Background(), uuid.Nil, model.NewIP(net.ParseIP("1.2.3.4")))
			},
			expectErr: "error updating one-time code (consume): multiple rows affected",
		},
		{
			name: "ShouldSucceedRevokeOneTimeCode",
			setup: func(db *mocks.MockSQLXDB, result *mocks.MockSQLResult) {
				result.EXPECT().RowsAffected().Return(int64(1), nil)
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(result, nil)
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.RevokeOneTimeCode(context.Background(), uuid.Nil, model.NewIP(net.ParseIP("1.2.3.4")))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)
			result := mocks.NewMockSQLResult(ctrl)
			p := storage.NewSQLProviderForTesting(db)

			if tc.setup != nil {
				tc.setup(db, result)
			}

			err := tc.invoke(p)

			if tc.expectErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectErr)
			}
		})
	}
}

func TestSQLProviderUpdateWebAuthnCredentialSignInUsesConn(t *testing.T) {
	t.Run("ShouldUseConnectionFromContext", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db := mocks.NewMockSQLXDB(ctrl)
		conn := mocks.NewMockSQLXConnection(ctrl)

		conn.EXPECT().ExecContext(gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))

		p := storage.NewSQLProviderForTesting(db)

		ctx := context.WithValue(context.Background(), storage.CtxKeyConnection, storage.SQLXConnection(conn))

		err := p.UpdateWebAuthnCredentialSignIn(ctx, model.WebAuthnCredential{ID: 1})

		assert.EqualError(t, err, "error updating WebAuthn credentials authentication metadata for id '1': boom")
	})
}

func TestSQLProviderOAuth2SessionByType(t *testing.T) {
	sessionTypes := []struct {
		name        string
		sessionType storage.OAuth2SessionType
	}{
		{"AccessToken", storage.OAuth2SessionTypeAccessToken},
		{"AuthorizeCode", storage.OAuth2SessionTypeAuthorizeCode},
		{"OpenIDConnect", storage.OAuth2SessionTypeOpenIDConnect},
		{"PKCEChallenge", storage.OAuth2SessionTypePKCEChallenge},
		{"RefreshToken", storage.OAuth2SessionTypeRefreshToken},
	}

	for _, st := range sessionTypes {
		st := st

		t.Run("ShouldErrRevokeOAuth2Session"+st.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)
			db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "sig").Return(nil, errors.New("boom"))

			p := storage.NewSQLProviderForTesting(db)

			err := p.RevokeOAuth2Session(context.Background(), st.sessionType, "sig")

			assert.EqualError(t, err, "error revoking oauth2 "+st.sessionType.String()+" session with signature 'sig': boom")
		})

		t.Run("ShouldErrRevokeOAuth2SessionByRequestID"+st.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)
			db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "req").Return(nil, errors.New("boom"))

			p := storage.NewSQLProviderForTesting(db)

			err := p.RevokeOAuth2SessionByRequestID(context.Background(), st.sessionType, "req")

			assert.EqualError(t, err, "error revoking oauth2 "+st.sessionType.String()+" session with request id 'req': boom")
		})

		t.Run("ShouldErrDeactivateOAuth2Session"+st.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)
			db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "sig").Return(nil, errors.New("boom"))

			p := storage.NewSQLProviderForTesting(db)

			err := p.DeactivateOAuth2Session(context.Background(), st.sessionType, "sig")

			assert.EqualError(t, err, "error deactivating oauth2 "+st.sessionType.String()+" session with signature 'sig': boom")
		})

		t.Run("ShouldErrDeactivateOAuth2SessionByRequestID"+st.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)
			db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "req").Return(nil, errors.New("boom"))

			p := storage.NewSQLProviderForTesting(db)

			err := p.DeactivateOAuth2SessionByRequestID(context.Background(), st.sessionType, "req")

			assert.EqualError(t, err, "error deactivating oauth2 "+st.sessionType.String()+" session with request id 'req': boom")
		})
	}
}

func TestSQLProviderRemainingExecErrors(t *testing.T) {
	testCases := []struct {
		name      string
		setup     func(db *mocks.MockSQLXDB)
		invoke    func(p *storage.SQLProvider) error
		expectErr string
	}{
		{
			name: "ShouldReturnErrSaveOAuth2ConsentSession",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(
					gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveOAuth2ConsentSession(context.Background(), &model.OAuth2ConsentSession{
					ChallengeID: uuid.Nil,
					ClientID:    "client",
				})
			},
			expectErr: "error inserting oauth2 consent session with challenge id '00000000-0000-0000-0000-000000000000' for subject '00000000-0000-0000-0000-000000000000': boom",
		},
		{
			name: "ShouldReturnErrSaveOAuth2ConsentSessionResponse",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(
					gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), 5,
				).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveOAuth2ConsentSessionResponse(context.Background(), &model.OAuth2ConsentSession{
					ID:          5,
					ChallengeID: uuid.Nil,
				}, true)
			},
			expectErr: "error updating oauth2 consent session (authorized  'true') with id '5' and challenge id '00000000-0000-0000-0000-000000000000' for subject '00000000-0000-0000-0000-000000000000': boom",
		},
		{
			name: "ShouldReturnErrSaveOAuth2PARContext",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(
					gomock.Any(), gomock.Any(),
					"sig", "req", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveOAuth2PARContext(context.Background(), model.OAuth2PARContext{Signature: "sig", RequestID: "req"})
			},
			expectErr: "error inserting oauth2 pushed authorization request context data for with signature 'sig' and request id 'req': boom",
		},
		{
			name: "ShouldReturnErrUpdateOAuth2PARContext",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(
					gomock.Any(), gomock.Any(),
					"sig", "req", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), 9,
				).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.UpdateOAuth2PARContext(context.Background(), model.OAuth2PARContext{ID: 9, Signature: "sig", RequestID: "req"})
			},
			expectErr: "error updating oauth2 pushed authorization request context data with id '9' and signature 'sig' and request id 'req': boom",
		},
		{
			name: "ShouldReturnErrSaveOAuth2DeviceCodeSession",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(
					gomock.Any(), gomock.Any(),
					gomock.Any(), "req", gomock.Any(), "sig", gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveOAuth2DeviceCodeSession(context.Background(), &model.OAuth2DeviceCodeSession{Signature: "sig", RequestID: "req"})
			},
			expectErr: "error inserting oauth2 device code session with device code signature 'sig' and user code signature '' for subject '' and request id 'req': boom",
		},
		{
			name: "ShouldReturnErrUpdateOAuth2DeviceCodeSession",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(
					gomock.Any(), gomock.Any(),
					gomock.Any(), "req", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "sig",
				).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.UpdateOAuth2DeviceCodeSession(context.Background(), &model.OAuth2DeviceCodeSession{Signature: "sig", RequestID: "req"})
			},
			expectErr: "error updating oauth2 device code session with device code signature 'sig': boom",
		},
		{
			name: "ShouldReturnErrUpdateOAuth2DeviceCodeSessionData",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(
					gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), "sig",
				).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.UpdateOAuth2DeviceCodeSessionData(context.Background(), &model.OAuth2DeviceCodeSession{Signature: "sig"})
			},
			expectErr: "error updating oauth2 device code session data with device code signature 'sig': boom",
		},
		{
			name: "ShouldReturnErrDeactivateOAuth2DeviceCodeSession",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "sig").Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.DeactivateOAuth2DeviceCodeSession(context.Background(), "sig")
			},
			expectErr: "error deactivating oauth2 device code session with device code signature 'sig': boom",
		},
		{
			name: "ShouldReturnErrLoadOAuth2DeviceCodeSession",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "sig").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadOAuth2DeviceCodeSession(context.Background(), "sig")
				return err
			},
			expectErr: "error selecting oauth2 device code session with device code signature 'sig': boom",
		},
		{
			name: "ShouldReturnErrLoadOAuth2DeviceCodeSessionByUserCode",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "sig").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadOAuth2DeviceCodeSessionByUserCode(context.Background(), "sig")
				return err
			},
			expectErr: "error selecting oauth2 device code session with user code signature 'sig': boom",
		},
		{
			name: "ShouldReturnErrSaveTOTPConfiguration",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(
					gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(),
					"john", "Authelia",
					"SHA1", gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveTOTPConfiguration(context.Background(), model.TOTPConfiguration{
					Username:  "john",
					Issuer:    "Authelia",
					Algorithm: "SHA1",
					Digits:    6,
					Period:    30,
				})
			},
			expectErr: "error upserting TOTP configuration for user 'john': boom",
		},
		{
			name: "ShouldReturnErrSaveWebAuthnCredential",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(
					gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveWebAuthnCredential(context.Background(), model.WebAuthnCredential{
					Username:  "john",
					KID:       model.NewBase64([]byte("kid")),
					PublicKey: []byte("pk"),
				})
			},
			expectErr: "error inserting WebAuthn credential for user 'john' kid '61326c6b': boom",
		},
		{
			name: "ShouldReturnErrSaveIdentityVerification",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(
					gomock.Any(), gomock.Any(),
					uuid.Nil, gomock.Any(), gomock.Any(), gomock.Any(), "john", "reset",
				).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveIdentityVerification(context.Background(), model.IdentityVerification{
					JTI:      uuid.Nil,
					Username: "john",
					Action:   "reset",
				})
			},
			expectErr: "error inserting identity verification for user 'john' with uuid '00000000-0000-0000-0000-000000000000': boom",
		},
		{
			name: "ShouldReturnErrSaveOneTimeCode",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(
					gomock.Any(), gomock.Any(),
					uuid.Nil, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "john", "reset", gomock.Any(),
				).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.SaveOneTimeCode(context.Background(), model.OneTimeCode{
					PublicID: uuid.Nil,
					Username: "john",
					Intent:   "reset",
				})

				return err
			},
			expectErr: "error inserting one-time code for user 'john' with signature '33700c24ce79aab7d5233eae0535db72e0fabbd48161dc5bdb852aa059aacdb83703d06b26c2378ef66d4ddcbb4186703a38d46c8dbe75c5ee0d94eb43d3b012': boom",
		},
		{
			name: "ShouldReturnErrSaveCachedDataNoEncryption",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "key", gomock.Any(), false, gomock.Any()).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveCachedData(context.Background(), model.CachedData{Name: "key", Value: []byte("val")})
			},
			expectErr: "error inserting cached data with name 'key': boom",
		},
		{
			name: "ShouldReturnErrSaveCachedDataEncrypted",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "key", gomock.Any(), true, gomock.Any()).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				return p.SaveCachedData(context.Background(), model.CachedData{Name: "key", Value: []byte("val"), Encrypted: true})
			},
			expectErr: "error inserting cached data with name 'key': boom",
		},
		{
			name: "ShouldReturnErrLoadRegulationRecordsByUserBan",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "john").Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadRegulationRecordsByUser(context.Background(), "john", time.Now(), 10)
				return err
			},
			expectErr: "error selecting last banned user time for username 'john': boom",
		},
		{
			name: "ShouldReturnErrLoadRegulationRecordsByUserSelect",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "john").Return(sql.ErrNoRows)
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "john", false, 10).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadRegulationRecordsByUser(context.Background(), "john", time.Now(), 10)
				return err
			},
			expectErr: "error selecting regulation records for username 'john': boom",
		},
		{
			name: "ShouldReturnErrLoadRegulationRecordsByIPBan",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadRegulationRecordsByIP(context.Background(), model.NewIP(net.ParseIP("1.2.3.4")), time.Now(), 10)
				return err
			},
			expectErr: "error selecting last banned time for ip '1.2.3.4': boom",
		},
		{
			name: "ShouldReturnErrLoadRegulationRecordsByIPSelect",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sql.ErrNoRows)
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), false, 10).Return(errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadRegulationRecordsByIP(context.Background(), model.NewIP(net.ParseIP("1.2.3.4")), time.Now(), 10)
				return err
			},
			expectErr: "error selecting regulation records for ip '1.2.3.4': boom",
		},
		{
			name: "ShouldReturnErrSaveOAuth2ConsentPreConfiguration",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().ExecContext(
					gomock.Any(), gomock.Any(),
					"client", uuid.Nil, gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(nil, errors.New("boom"))
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.SaveOAuth2ConsentPreConfiguration(context.Background(), model.OAuth2ConsentPreConfig{
					ClientID: "client",
					Subject:  uuid.Nil,
				})

				return err
			},
			expectErr: "error inserting oauth2 consent pre-configuration for subject '00000000-0000-0000-0000-000000000000' with client id 'client' and scopes '': boom",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)

			if tc.setup != nil {
				tc.setup(db)
			}

			p := storage.NewSQLProviderForTesting(db)

			assert.EqualError(t, tc.invoke(p), tc.expectErr)
		})
	}
}

func TestSQLProviderSchemaEncryptionRotateHMACKey(t *testing.T) {
	testCases := []struct {
		name      string
		hmacName  string
		setup     func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx)
		expectErr string
	}{
		{
			name:      "ShouldErrUnknownKeyName",
			hmacName:  "unknown",
			setup:     nil,
			expectErr: "unknown key name 'unknown'",
		},
		{
			name:     "ShouldErrBeginTransactionForOneTimeCode",
			hmacName: "otc",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx) {
				db.EXPECT().Beginx().Return(nil, errors.New("begin failed"))
			},
			expectErr: "error beginning transaction to rotate hmac key: begin failed",
		},
		{
			name:     "ShouldErrSetCryptographyKeyForOneTimeCodeAndRollback",
			hmacName: "otc",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx) {
				db.EXPECT().Beginx().Return(tx, nil)
				tx.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "hmac_key_otc", gomock.Any()).Return(nil, errors.New("upsert failed"))
				tx.EXPECT().Rollback().Return(nil)
			},
			expectErr: "error setting the hmac key: upsert failed",
		},
		{
			name:     "ShouldErrSetCryptographyKeyForOneTimeCodeWithRollbackErr",
			hmacName: "otc",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx) {
				db.EXPECT().Beginx().Return(tx, nil)
				tx.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "hmac_key_otc", gomock.Any()).Return(nil, errors.New("upsert failed"))
				tx.EXPECT().Rollback().Return(errors.New("rollback failed"))
			},
			expectErr: "error rolling back transaction to rotate hmac key: rollback failed",
		},
		{
			name:     "ShouldErrTruncateForOneTimeCodeAndRollback",
			hmacName: "otc",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx) {
				db.EXPECT().Beginx().Return(tx, nil)
				tx.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "hmac_key_otc", gomock.Any()).Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "DELETE FROM one_time_code;").Return(nil, errors.New("delete failed"))
				tx.EXPECT().Rollback().Return(nil)
			},
			expectErr: "error truncating one time-codes: error occurred truncating table 'one_time_code': error occurred performing the delete: delete failed",
		},
		{
			name:     "ShouldErrTruncateForOneTimeCodeWithRollbackErr",
			hmacName: "otc",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx) {
				db.EXPECT().Beginx().Return(tx, nil)
				tx.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "hmac_key_otc", gomock.Any()).Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "DELETE FROM one_time_code;").Return(nil, errors.New("delete failed"))
				tx.EXPECT().Rollback().Return(errors.New("rollback failed"))
			},
			expectErr: "error rolling back transaction to rotate hmac key: rollback failed",
		},
		{
			name:     "ShouldErrCommitForOneTimeCode",
			hmacName: "otc",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx) {
				db.EXPECT().Beginx().Return(tx, nil)
				tx.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "hmac_key_otc", gomock.Any()).Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "DELETE FROM one_time_code;").Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "DELETE FROM sqlite_sequence WHERE name = ?;", "one_time_code").Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "VACUUM;").Return(nil, nil)
				tx.EXPECT().Commit().Return(errors.New("commit failed"))
			},
			expectErr: "error committing transaction to rotate hmac key: commit failed",
		},
		{
			name:     "ShouldSucceedForOneTimeCode",
			hmacName: "otc",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx) {
				db.EXPECT().Beginx().Return(tx, nil)
				tx.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "hmac_key_otc", gomock.Any()).Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "DELETE FROM one_time_code;").Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "DELETE FROM sqlite_sequence WHERE name = ?;", "one_time_code").Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "VACUUM;").Return(nil, nil)
				tx.EXPECT().Commit().Return(nil)
			},
		},
		{
			name:     "ShouldErrBeginTransactionForOneTimePassword",
			hmacName: "otp",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx) {
				db.EXPECT().Beginx().Return(nil, errors.New("begin failed"))
			},
			expectErr: "error beginning transaction to rotate hmac key: begin failed",
		},
		{
			name:     "ShouldErrSetCryptographyKeyForOneTimePasswordAndRollback",
			hmacName: "otp",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx) {
				db.EXPECT().Beginx().Return(tx, nil)
				tx.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "hmac_key_otp", gomock.Any()).Return(nil, errors.New("upsert failed"))
				tx.EXPECT().Rollback().Return(nil)
			},
			expectErr: "error setting the hmac key: upsert failed",
		},
		{
			name:     "ShouldErrSetCryptographyKeyForOneTimePasswordWithRollbackErr",
			hmacName: "otp",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx) {
				db.EXPECT().Beginx().Return(tx, nil)
				tx.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "hmac_key_otp", gomock.Any()).Return(nil, errors.New("upsert failed"))
				tx.EXPECT().Rollback().Return(errors.New("rollback failed"))
			},
			expectErr: "error rolling back transaction to rotate hmac key: rollback failed",
		},
		{
			name:     "ShouldErrTruncateForOneTimePasswordAndRollback",
			hmacName: "otp",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx) {
				db.EXPECT().Beginx().Return(tx, nil)
				tx.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "hmac_key_otp", gomock.Any()).Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "DELETE FROM totp_history;").Return(nil, errors.New("delete failed"))
				tx.EXPECT().Rollback().Return(nil)
			},
			expectErr: "error truncating totp history: error occurred truncating table 'totp_history': error occurred performing the delete: delete failed",
		},
		{
			name:     "ShouldErrTruncateForOneTimePasswordWithRollbackErr",
			hmacName: "otp",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx) {
				db.EXPECT().Beginx().Return(tx, nil)
				tx.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "hmac_key_otp", gomock.Any()).Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "DELETE FROM totp_history;").Return(nil, errors.New("delete failed"))
				tx.EXPECT().Rollback().Return(errors.New("rollback failed"))
			},
			expectErr: "error rolling back transaction to rotate hmac key: rollback failed",
		},
		{
			name:     "ShouldErrCommitForOneTimePassword",
			hmacName: "otp",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx) {
				db.EXPECT().Beginx().Return(tx, nil)
				tx.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "hmac_key_otp", gomock.Any()).Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "DELETE FROM totp_history;").Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "DELETE FROM sqlite_sequence WHERE name = ?;", "totp_history").Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "VACUUM;").Return(nil, nil)
				tx.EXPECT().Commit().Return(errors.New("commit failed"))
			},
			expectErr: "error committing transaction to rotate hmac key: commit failed",
		},
		{
			name:     "ShouldSucceedForOneTimePassword",
			hmacName: "otp",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx) {
				db.EXPECT().Beginx().Return(tx, nil)
				tx.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "hmac_key_otp", gomock.Any()).Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "DELETE FROM totp_history;").Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "DELETE FROM sqlite_sequence WHERE name = ?;", "totp_history").Return(nil, nil)
				tx.EXPECT().ExecContext(gomock.Any(), "VACUUM;").Return(nil, nil)
				tx.EXPECT().Commit().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)
			tx := mocks.NewMockSQLXTx(ctrl)

			if tc.setup != nil {
				tc.setup(db, tx)
			}

			p := storage.NewSQLProviderForTesting(db)

			err := p.SchemaEncryptionRotateHMACKey(context.Background(), tc.hmacName)

			if tc.expectErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectErr)
			}
		})
	}
}

func TestSQLProviderRevokeOneTimeCodeRowsAffectedError(t *testing.T) {
	t.Run("ShouldErrWhenRowsAffectedReturnsError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db := mocks.NewMockSQLXDB(ctrl)
		result := mocks.NewMockSQLResult(ctrl)

		result.EXPECT().RowsAffected().Return(int64(0), errors.New("ra-err"))
		db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(result, nil)

		p := storage.NewSQLProviderForTesting(db)

		err := p.RevokeOneTimeCode(context.Background(), uuid.Nil, model.NewIP(net.ParseIP("1.2.3.4")))

		assert.EqualError(t, err, "error updating one-time code (consume): ra-err")
	})
}

func TestSQLProviderFindIdentityVerificationBranches(t *testing.T) {
	testCases := []struct {
		name      string
		populate  func(v *model.IdentityVerification)
		expectErr string
		expectOk  bool
	}{
		{
			name: "ShouldReturnFalseWhenRevoked",
			populate: func(v *model.IdentityVerification) {
				v.RevokedAt = sql.NullTime{Valid: true, Time: time.Now()}
				v.ExpiresAt = time.Now().Add(time.Hour)
			},
			expectErr: "the token has been revoked",
		},
		{
			name: "ShouldReturnFalseWhenConsumed",
			populate: func(v *model.IdentityVerification) {
				v.ConsumedAt = sql.NullTime{Valid: true, Time: time.Now()}
				v.ExpiresAt = time.Now().Add(time.Hour)
			},
			expectErr: "the token has already been consumed",
		},
		{
			name: "ShouldReturnFalseWhenExpired",
			populate: func(v *model.IdentityVerification) {
				v.ExpiresAt = time.Now().Add(-time.Hour)
			},
			expectErr: "the token expired ",
		},
		{
			name: "ShouldReturnTrueWhenValid",
			populate: func(v *model.IdentityVerification) {
				v.ExpiresAt = time.Now().Add(time.Hour)
			},
			expectOk: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)

			db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "jti").DoAndReturn(
				func(_ context.Context, dest any, _ string, _ ...any) error {
					v := dest.(*model.IdentityVerification)
					tc.populate(v)

					return nil
				},
			)

			p := storage.NewSQLProviderForTesting(db)

			found, err := p.FindIdentityVerification(context.Background(), "jti")

			if tc.expectErr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectOk, found)
			} else {
				assert.False(t, found)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErr)
			}
		})
	}
}

func TestSQLProviderDecryptErrorPaths(t *testing.T) {
	invalidCipher := []byte("not-valid-ciphertext")

	testCases := []struct {
		name      string
		setup     func(db *mocks.MockSQLXDB)
		invoke    func(p *storage.SQLProvider) error
		expectErr string
	}{
		{
			name: "ShouldErrLoadTOTPConfigurationDecrypt",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "john").DoAndReturn(
					func(_ context.Context, dest any, _ string, _ ...any) error {
						c := dest.(*model.TOTPConfiguration)
						c.Username = "john"
						c.Secret = invalidCipher

						return nil
					},
				)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadTOTPConfiguration(context.Background(), "john")
				return err
			},
			expectErr: "error decrypting TOTP secret for user 'john': ",
		},
		{
			name: "ShouldErrLoadTOTPConfigurationsDecrypt",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), 10, 0).DoAndReturn(
					func(_ context.Context, dest any, _ string, _ ...any) error {
						configs := dest.(*[]model.TOTPConfiguration)
						*configs = []model.TOTPConfiguration{{Username: "john", Secret: invalidCipher}}

						return nil
					},
				)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadTOTPConfigurations(context.Background(), 10, 0)
				return err
			},
			expectErr: "error decrypting TOTP configuration for user 'john': ",
		},
		{
			name: "ShouldErrLoadWebAuthnCredentialByIDDecryptPublicKey",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), 7).DoAndReturn(
					func(_ context.Context, dest any, _ string, _ ...any) error {
						c := dest.(*model.WebAuthnCredential)
						c.ID = 7
						c.Username = "john"
						c.PublicKey = invalidCipher

						return nil
					},
				)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnCredentialByID(context.Background(), 7)
				return err
			},
			expectErr: "error decrypting WebAuthn credential public key of credential with id '7' for user 'john': ",
		},
		{
			name: "ShouldErrLoadWebAuthnCredentialByIDDecryptAttestation",
			setup: func(db *mocks.MockSQLXDB) {
				validPK, err := encryptForTesting([]byte("public-key"))
				require.NoError(t, err)

				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), 7).DoAndReturn(
					func(_ context.Context, dest any, _ string, _ ...any) error {
						c := dest.(*model.WebAuthnCredential)
						c.ID = 7
						c.Username = "john"
						c.PublicKey = validPK
						c.Attestation = invalidCipher

						return nil
					},
				)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnCredentialByID(context.Background(), 7)
				return err
			},
			expectErr: "error decrypting WebAuthn credential attestation of credential with id '7' for user 'john': ",
		},
		{
			name: "ShouldErrLoadWebAuthnCredentialsDecryptPublicKey",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), 10, 0).DoAndReturn(
					func(_ context.Context, dest any, _ string, _ ...any) error {
						creds := dest.(*[]model.WebAuthnCredential)
						*creds = []model.WebAuthnCredential{{ID: 9, Username: "john", PublicKey: invalidCipher}}

						return nil
					},
				)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnCredentials(context.Background(), 10, 0)
				return err
			},
			expectErr: "error decrypting WebAuthn credential public key of credential with id '9' for user 'john': ",
		},
		{
			name: "ShouldErrLoadWebAuthnCredentialsDecryptAttestation",
			setup: func(db *mocks.MockSQLXDB) {
				validPK, err := encryptForTesting([]byte("public-key"))
				require.NoError(t, err)

				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), 10, 0).DoAndReturn(
					func(_ context.Context, dest any, _ string, _ ...any) error {
						creds := dest.(*[]model.WebAuthnCredential)
						*creds = []model.WebAuthnCredential{{ID: 9, Username: "john", PublicKey: validPK, Attestation: invalidCipher}}

						return nil
					},
				)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnCredentials(context.Background(), 10, 0)
				return err
			},
			expectErr: "error decrypting WebAuthn credential attestation of credential with id '9' for user 'john': ",
		},
		{
			name: "ShouldErrLoadWebAuthnCredentialsByUsernameDecryptPublicKey",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), "example.com", "john", false).DoAndReturn(
					func(_ context.Context, dest any, _ string, _ ...any) error {
						creds := dest.(*[]model.WebAuthnCredential)
						*creds = []model.WebAuthnCredential{{ID: 9, Username: "john", PublicKey: invalidCipher}}

						return nil
					},
				)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnCredentialsByUsername(context.Background(), "example.com", "john")
				return err
			},
			expectErr: "error decrypting WebAuthn credential public key of credential with id '9' for user 'john': ",
		},
		{
			name: "ShouldErrLoadWebAuthnCredentialsByUsernameDecryptAttestation",
			setup: func(db *mocks.MockSQLXDB) {
				validPK, err := encryptForTesting([]byte("public-key"))
				require.NoError(t, err)

				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), "example.com", "john", false).DoAndReturn(
					func(_ context.Context, dest any, _ string, _ ...any) error {
						creds := dest.(*[]model.WebAuthnCredential)
						*creds = []model.WebAuthnCredential{{ID: 9, Username: "john", PublicKey: validPK, Attestation: invalidCipher}}

						return nil
					},
				)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnCredentialsByUsername(context.Background(), "example.com", "john")
				return err
			},
			expectErr: "error decrypting WebAuthn credential attestation of credential with id '9' for user 'john': ",
		},
		{
			name: "ShouldErrLoadWebAuthnPasskeyCredentialsByUsernameDecryptPublicKey",
			setup: func(db *mocks.MockSQLXDB) {
				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), "example.com", "john", true).DoAndReturn(
					func(_ context.Context, dest any, _ string, _ ...any) error {
						creds := dest.(*[]model.WebAuthnCredential)
						*creds = []model.WebAuthnCredential{{ID: 9, Username: "john", PublicKey: invalidCipher}}

						return nil
					},
				)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnPasskeyCredentialsByUsername(context.Background(), "example.com", "john")
				return err
			},
			expectErr: "error decrypting passkey WebAuthn credential public key of credential with id '9' for user 'john': ",
		},
		{
			name: "ShouldErrLoadWebAuthnPasskeyCredentialsByUsernameDecryptAttestation",
			setup: func(db *mocks.MockSQLXDB) {
				validPK, err := encryptForTesting([]byte("public-key"))
				require.NoError(t, err)

				db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), "example.com", "john", true).DoAndReturn(
					func(_ context.Context, dest any, _ string, _ ...any) error {
						creds := dest.(*[]model.WebAuthnCredential)
						*creds = []model.WebAuthnCredential{{ID: 9, Username: "john", PublicKey: validPK, Attestation: invalidCipher}}

						return nil
					},
				)
			},
			invoke: func(p *storage.SQLProvider) error {
				_, err := p.LoadWebAuthnPasskeyCredentialsByUsername(context.Background(), "example.com", "john")
				return err
			},
			expectErr: "error decrypting passkey WebAuthn credential attestation of credential with id '9' for user 'john': ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)

			if tc.setup != nil {
				tc.setup(db)
			}

			p := storage.NewSQLProviderForTesting(db)

			err := tc.invoke(p)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectErr)
		})
	}
}

func TestSQLProviderSaveWebAuthnCredentialAttestation(t *testing.T) {
	t.Run("ShouldExerciseAttestationEncryptionWhenPresent", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db := mocks.NewMockSQLXDB(ctrl)
		db.EXPECT().ExecContext(
			gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		).Return(nil, errors.New("boom"))

		p := storage.NewSQLProviderForTesting(db)

		err := p.SaveWebAuthnCredential(context.Background(), model.WebAuthnCredential{
			Username:    "john",
			KID:         model.NewBase64([]byte("kid")),
			PublicKey:   []byte("pk"),
			Attestation: []byte("attestation-payload"),
		})

		assert.EqualError(t, err, "error inserting WebAuthn credential for user 'john' kid '61326c6b': boom")
	})
}

func TestSQLProviderStartupCheckOpenErr(t *testing.T) {
	t.Run("ShouldReturnErrorWhenDBOpenFailed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db := mocks.NewMockSQLXDB(ctrl)
		p := storage.NewSQLProviderForTesting(db).WithOpenErr(errors.New("dsn invalid"))

		err := p.StartupCheck()

		assert.EqualError(t, err, "error opening database: dsn invalid")
	})
}

func encryptForTesting(clearText []byte) ([]byte, error) {
	return storage.NewSQLProviderForTesting(nil).Encrypt(clearText)
}
