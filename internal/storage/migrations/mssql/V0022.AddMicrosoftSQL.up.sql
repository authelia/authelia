IF OBJECT_ID(N'dbo.authentication_logs', N'U') IS NULL
CREATE TABLE [dbo].[authentication_logs] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [time] DATETIME2(7) NOT NULL DEFAULT GETDATE(),
    [successful] BIT NOT NULL,
    [banned] BIT NOT NULL DEFAULT FALSE,
    [username] VARCHAR(100) NOT NULL,
    [auth_type] VARCHAR(8) NOT NULL DEFAULT '1FA',
    [remote_ip] VARCHAR(39) NULL DEFAULT NULL,
    [request_uri] VARCHAR(MAX),
    [request_method] VARCHAR(8) NOT NULL DEFAULT ''
);

CREATE INDEX [authentication_logs_username_idx] ON [dbo].[authentication_logs] ([time], [username], [auth_type]);
CREATE INDEX [authentication_logs_remote_ip_idx] ON [dbo].[authentication_logs] ([time], [remote_ip], [auth_type]);

IF OBJECT_ID(N'dbo.totp_configurations', N'U') IS NULL
CREATE TABLE [dbo].[totp_configurations] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [username] VARCHAR(100) NOT NULL,
    [issuer] VARCHAR(100),
    [algorithm] VARCHAR(6) NOT NULL DEFAULT 'SHA1',
    [digits] INT NOT NULL DEFAULT 6,
    [period] INT NOT NULL DEFAULT 30,
    [secret] VARBINARY(MAX) NOT NULL
);

CREATE UNIQUE INDEX [totp_configurations_username_key] ON [dbo].[totp_configurations] ([username]);

IF OBJECT_ID(N'dbo.totp_history', N'U') IS NULL
CREATE TABLE [dbo].[totp_history] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
	created_at DATETIME2(7) NOT NULL DEFAULT GETDATE(),
    [username] VARCHAR(100) NOT NULL,
	[step] CHAR(64) NOT NULL
);

CREATE UNIQUE INDEX [totp_history_lookup_key] ON [dbo].[totp_history] ([username], [step]);

IF OBJECT_ID(N'dbo.duo_devices', N'U') IS NULL
CREATE TABLE [dbo].[duo_devices] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [username] VARCHAR(100) NOT NULL,
    [device] VARCHAR(32) NOT NULL,
    [method] VARCHAR(16) NOT NULL
);

CREATE UNIQUE INDEX [duo_devices_username_key] ON [dbo].[duo_devices] ([username]);

IF OBJECT_ID(N'dbo.user_preferences', N'U') IS NULL
CREATE TABLE [dbo].[user_preferences] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [username] VARCHAR(100) NOT NULL,
    [second_factor_method] VARCHAR(11) NOT NULL
);

CREATE UNIQUE INDEX [user_preferences_username_key] ON [dbo].[user_preferences] ([username]);

IF OBJECT_ID(N'dbo.migrations', N'U') IS NULL
CREATE TABLE [dbo].[migrations] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [applied] DATETIME2(7) NOT NULL DEFAULT GETDATE(),
    [version_before] INT NULL DEFAULT NULL,
    [version_after] INT NOT NULL,
    [application_version] VARCHAR(128) NOT NULL
);

IF OBJECT_ID(N'dbo.encryption', N'U') IS NULL
CREATE TABLE [dbo].[encryption] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [name] VARCHAR(100),
    [value] VARBINARY(MAX) NOT NULL
);

CREATE UNIQUE INDEX [encryption_name_key] ON [dbo].[encryption] ([name]);

IF OBJECT_ID(N'dbo.identity_verification', N'U') IS NULL
CREATE TABLE [dbo].[identity_verification] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [jti] CHAR(36),
    [iat] DATETIME2(7) NOT NULL DEFAULT GETDATE(),
    [issued_ip] VARCHAR(39) NOT NULL,
    [exp] DATETIME2(7) NOT NULL,
    [username] VARCHAR(100) NOT NULL,
    [action] VARCHAR(50) NOT NULL,
    [consumed] DATETIME2(7) NULL DEFAULT NULL,
    [consumed_ip] VARCHAR(39) NULL DEFAULT NULL,
	[revoked] DATETIME2(7) NULL DEFAULT NULL,
	[revoked_ip] VARCHAR(39) NULL DEFAULT NULL
);

CREATE UNIQUE INDEX [identity_verification_jti_key] ON [dbo].[identity_verification] ([jti]);

IF OBJECT_ID(N'dbo.webauthn_users', N'U') IS NULL
CREATE TABLE [dbo].[webauthn_users] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [rpid] VARCHAR(512) NOT NULL,
    [username] VARCHAR(100) NOT NULL,
    [userid] CHAR(64) NOT NULL
);

CREATE UNIQUE INDEX [webauthn_users_lookup_key] ON [dbo].[webauthn_users] ([rpid], [username]);

IF OBJECT_ID(N'dbo.webauthn_credentials', N'U') IS NULL
CREATE TABLE [dbo].[webauthn_credentials] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [created_at] DATETIME2(7) NOT NULL DEFAULT GETDATE(),
    [last_used_at] DATETIME2(7) NULL DEFAULT NULL,
    [rpid] VARCHAR(512) NOT NULL,
    [username] VARCHAR(100) NOT NULL,
    [description] VARCHAR(30) NOT NULL,
    [kid] VARCHAR(512) NOT NULL,
    [aaguid] CHAR(36) NULL,
    [attestation_type] VARCHAR(32),
    [attachment] VARCHAR(64) NOT NULL,
    [transport] VARCHAR(64) DEFAULT '',
    [sign_count] INT DEFAULT 0,
    [clone_warning] BIT NOT NULL DEFAULT FALSE,
    [legacy] BIT NOT NULL DEFAULT FALSE,
    [discoverable] BIT NOT NULL,
    [present] BIT NOT NULL DEFAULT FALSE,
    [verified] BIT NOT NULL DEFAULT FALSE,
    [backup_eligible] BIT NOT NULL DEFAULT FALSE,
    [backup_state] BIT NOT NULL DEFAULT FALSE,
    [public_key] VARBINARY(MAX) NOT NULL
);

CREATE UNIQUE INDEX [webauthn_credentials_kid_key] ON [dbo].[webauthn_credentials] ([kid]);
CREATE UNIQUE INDEX [webauthn_credentials_lookup_key] ON [dbo].[webauthn_credentials] ([rpid], [username], [description]);

IF OBJECT_ID(N'dbo.one_time_code', N'U') IS NULL
CREATE TABLE [dbo].[one_time_code] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
	[public_id] CHAR(36) NOT NULL,
    [signature] VARCHAR(128) NOT NULL,
	[issued] DATETIME2(7) NOT NULL DEFAULT GETDATE(),
    [issued_ip] VARCHAR(39) NOT NULL,
	[expires] DATETIME2(7) NOT NULL,
    [username] VARCHAR(100) NOT NULL,
    [intent] VARCHAR(100) NOT NULL,
    [consumed] DATETIME2(7) NULL DEFAULT NULL,
    [consumed_ip] VARCHAR(39) NULL DEFAULT NULL,
	[revoked] DATETIME2(7) NULL DEFAULT NULL,
	[revoked_ip] VARCHAR(39) NULL DEFAULT NULL,
    [code] VARBINARY(MAX) NOT NULL
);

CREATE UNIQUE INDEX [one_time_code_signature] ON [dbo].[one_time_code] ([signature], [username]);

IF OBJECT_ID(N'dbo.user_opaque_identifier', N'U') IS NULL
CREATE TABLE [dbo].[user_opaque_identifier] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [service] VARCHAR(20) NOT NULL,
    [sector_id] VARCHAR(255) NOT NULL,
    [username] VARCHAR(100) NOT NULL,
    [identifier] CHAR(36) NOT NULL
);

CREATE UNIQUE INDEX [user_opaque_identifier_service_sector_id_username_key] ON [dbo].[user_opaque_identifier] (service, sector_id, username);
CREATE UNIQUE INDEX [user_opaque_identifier_identifier_key] ON [dbo].[user_opaque_identifier] (identifier);

IF OBJECT_ID(N'dbo.oauth2_consent_preconfiguration', N'U') IS NULL
CREATE TABLE [dbo].[oauth2_consent_preconfiguration] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [client_id] VARCHAR(255) NOT NULL,
    [subject] CHAR(36) NOT NULL,
    [created_at] DATETIME2(7) NOT NULL DEFAULT GETDATE(),
    [expires_at] DATETIME2(7) NULL DEFAULT NULL,
    [revoked] BIT NOT NULL DEFAULT FALSE,
    [scopes] VARCHAR(MAX) NOT NULL,
    [audience] VARCHAR(MAX) NULL
);

ALTER TABLE [dbo].[oauth2_consent_preconfiguration]
    ADD CONSTRAINT [oauth2_consent_preconfiguration_subject_fkey]
        FOREIGN KEY ([subject])
            REFERENCES [dbo].[user_opaque_identifier] ([identifier]) ON UPDATE CASCADE;

IF OBJECT_ID(N'dbo.oauth2_consent_session', N'U') IS NULL
CREATE TABLE [dbo].[oauth2_consent_session] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [challenge_id] CHAR(36) NOT NULL,
    [client_id] VARCHAR(255) NOT NULL,
    [subject] CHAR(36) NULL DEFAULT NULL,
    [authorized] BIT NOT NULL DEFAULT FALSE,
    [granted] BIT NOT NULL DEFAULT FALSE,
    [requested_at] DATETIME2(7) NOT NULL DEFAULT GETDATE(),
    [responded_at] DATETIME2(7) NULL DEFAULT NULL,
    [form_data] VARCHAR(MAX) NOT NULL,
    [requested_scopes] VARCHAR(MAX) NOT NULL,
    [granted_scopes] VARCHAR(MAX) NOT NULL,
    [requested_audience] VARCHAR(MAX) NULL,
    [granted_audience] VARCHAR(MAX) NULL,
    [preconfiguration] INT NULL DEFAULT NULL
);

CREATE UNIQUE INDEX [oauth2_consent_session_challenge_id_key] ON [dbo].[oauth2_consent_session] ([challenge_id]);

ALTER TABLE [dbo].[oauth2_consent_session]
    ADD CONSTRAINT [oauth2_consent_session_subject_fkey]
        FOREIGN KEY ([subject])
            REFERENCES [dbo].[user_opaque_identifier] ([identifier]) ON UPDATE CASCADE,
    ADD CONSTRAINT [oauth2_consent_session_preconfiguration_fkey]
        FOREIGN KEY ([preconfiguration])
            REFERENCES [dbo].[oauth2_consent_preconfiguration] ([id]) ON DELETE CASCADE ON UPDATE CASCADE;

IF OBJECT_ID(N'dbo.oauth2_access_token_session', N'U') IS NULL
CREATE TABLE [dbo].[oauth2_access_token_session] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [challenge_id] CHAR(36) NULL DEFAULT NULL,
    [request_id] VARCHAR(40) NOT NULL,
    [client_id] VARCHAR(255) NOT NULL,
    [signature] VARCHAR(768) NOT NULL,
    [subject] CHAR(36) NULL DEFAULT NULL,
    [requested_at] DATETIME2(7) NOT NULL DEFAULT GETDATE(),
    [requested_scopes] VARCHAR(MAX) NOT NULL,
    [granted_scopes] VARCHAR(MAX) NOT NULL,
    [requested_audience] VARCHAR(MAX) NULL,
    [granted_audience] VARCHAR(MAX) NULL,
    [active] BIT NOT NULL DEFAULT FALSE,
    [revoked] BIT NOT NULL DEFAULT FALSE,
    [form_data] VARCHAR(MAX) NOT NULL,
    [session_data] VARBINARY(MAX) NOT NULL
);

CREATE INDEX [oauth2_access_token_session_request_id_idx] ON [dbo].[oauth2_access_token_session] ([request_id]);
CREATE INDEX [oauth2_access_token_session_client_id_idx] ON [dbo].[oauth2_access_token_session] ([client_id]);
CREATE INDEX [oauth2_access_token_session_client_id_subject_idx] ON [dbo].[oauth2_access_token_session] ([client_id], [subject]);

ALTER TABLE [dbo].[oauth2_access_token_session]
    ADD CONSTRAINT [oauth2_access_token_session_challenge_id_fkey]
        FOREIGN KEY ([challenge_id])
            REFERENCES [dbo].[oauth2_consent_session] ([challenge_id]) ON DELETE CASCADE ON UPDATE CASCADE,
    ADD CONSTRAINT [oauth2_access_token_session_subject_fkey]
        FOREIGN KEY ([subject])
            REFERENCES [dbo].[user_opaque_identifier] ([identifier]) ON UPDATE CASCADE;

IF OBJECT_ID(N'dbo.oauth2_authorization_code_session', N'U') IS NULL
CREATE TABLE [dbo].[oauth2_authorization_code_session] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [challenge_id] CHAR(36) NOT NULL,
    [request_id] VARCHAR(40) NOT NULL,
    [client_id] VARCHAR(255) NOT NULL,
    [signature] VARCHAR(255) NOT NULL,
    [subject] CHAR(36) NOT NULL,
    [requested_at] DATETIME2(7) NOT NULL DEFAULT GETDATE(),
    [requested_scopes] VARCHAR(MAX) NOT NULL,
    [granted_scopes] VARCHAR(MAX) NOT NULL,
    [requested_audience] VARCHAR(MAX) NULL,
    [granted_audience] VARCHAR(MAX) NULL,
    [active] BIT NOT NULL DEFAULT FALSE,
    [revoked] BIT NOT NULL DEFAULT FALSE,
    [form_data] VARCHAR(MAX) NOT NULL,
    [session_data] VARBINARY(MAX) NOT NULL
);

CREATE INDEX [oauth2_authorization_code_session_request_id_idx] ON [dbo].[oauth2_authorization_code_session] ([request_id]);
CREATE INDEX [oauth2_authorization_code_session_client_id_idx] ON [dbo].[oauth2_authorization_code_session] ([client_id]);
CREATE INDEX [oauth2_authorization_code_session_client_id_subject_idx] ON [dbo].[oauth2_authorization_code_session] ([client_id], [subject]);

ALTER TABLE [dbo].[oauth2_authorization_code_session]
    ADD CONSTRAINT [oauth2_authorization_code_session_challenge_id_fkey]
        FOREIGN KEY ([challenge_id])
            REFERENCES [dbo].[oauth2_consent_session] ([challenge_id]) ON DELETE CASCADE ON UPDATE CASCADE,
    ADD CONSTRAINT [oauth2_authorization_code_session_subject_fkey]
        FOREIGN KEY ([subject])
            REFERENCES [dbo].[user_opaque_identifier] ([identifier]) ON UPDATE CASCADE;

IF OBJECT_ID(N'dbo.oauth2_openid_connect_session', N'U') IS NULL
CREATE TABLE [dbo].[oauth2_openid_connect_session] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [challenge_id] CHAR(36) NOT NULL,
    [request_id] VARCHAR(40) NOT NULL,
    [client_id] VARCHAR(255) NOT NULL,
    [signature] VARCHAR(255) NOT NULL,
    [subject] CHAR(36) NOT NULL,
    [requested_at] DATETIME2(7) NOT NULL DEFAULT GETDATE(),
    [requested_scopes] VARCHAR(MAX) NOT NULL,
    [granted_scopes] VARCHAR(MAX) NOT NULL,
    [requested_audience] VARCHAR(MAX) NULL,
    [granted_audience] VARCHAR(MAX) NULL,
    [active] BIT NOT NULL DEFAULT FALSE,
    [revoked] BIT NOT NULL DEFAULT FALSE,
    [form_data] VARCHAR(MAX) NOT NULL,
    [session_data] VARBINARY(MAX) NOT NULL
);

CREATE INDEX [oauth2_openid_connect_session_request_id_idx] ON [dbo].[oauth2_openid_connect_session] ([request_id]);
CREATE INDEX [oauth2_openid_connect_session_client_id_idx] ON [dbo].[oauth2_openid_connect_session] ([client_id]);
CREATE INDEX [oauth2_openid_connect_session_client_id_subject_idx] ON [dbo].[oauth2_openid_connect_session] ([client_id], [subject]);

ALTER TABLE [dbo].[oauth2_openid_connect_session]
    ADD CONSTRAINT [oauth2_openid_connect_session_challenge_id_fkey]
        FOREIGN KEY ([challenge_id])
            REFERENCES [dbo].[oauth2_consent_session] ([challenge_id]) ON DELETE CASCADE ON UPDATE CASCADE,
    ADD CONSTRAINT [oauth2_openid_connect_session_subject_fkey]
        FOREIGN KEY ([subject])
            REFERENCES [dbo].[user_opaque_identifier] ([identifier]) ON UPDATE CASCADE;

IF OBJECT_ID(N'dbo.oauth2_pkce_request_session', N'U') IS NULL
CREATE TABLE [dbo].[oauth2_pkce_request_session] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [challenge_id] CHAR(36) NOT NULL,
    [request_id] VARCHAR(40) NOT NULL,
    [client_id] VARCHAR(255) NOT NULL,
    [signature] VARCHAR(255) NOT NULL,
    [subject] CHAR(36) NOT NULL,
    [requested_at] DATETIME2(7) NOT NULL DEFAULT GETDATE(),
    [requested_scopes] VARCHAR(MAX) NOT NULL,
    [granted_scopes] VARCHAR(MAX) NOT NULL,
    [requested_audience] VARCHAR(MAX) NULL,
    [granted_audience] VARCHAR(MAX) NULL,
    [active] BIT NOT NULL DEFAULT FALSE,
    [revoked] BIT NOT NULL DEFAULT FALSE,
    [form_data] VARCHAR(MAX) NOT NULL,
    [session_data] VARBINARY(MAX) NOT NULL
);

CREATE INDEX [oauth2_pkce_request_session_request_id_idx] ON [dbo].[oauth2_pkce_request_session] ([request_id]);
CREATE INDEX [oauth2_pkce_request_session_client_id_idx] ON [dbo].[oauth2_pkce_request_session] ([client_id]);
CREATE INDEX [oauth2_pkce_request_session_client_id_subject_idx] ON [dbo].[oauth2_pkce_request_session] ([client_id], [subject]);

ALTER TABLE [dbo].[oauth2_pkce_request_session]
    ADD CONSTRAINT [oauth2_pkce_request_session_challenge_id_fkey]
        FOREIGN KEY ([challenge_id])
            REFERENCES [dbo].[oauth2_consent_session] ([challenge_id]) ON DELETE CASCADE ON UPDATE CASCADE,
    ADD CONSTRAINT [oauth2_pkce_request_session_subject_fkey]
        FOREIGN KEY ([subject])
            REFERENCES [dbo].[user_opaque_identifier] ([identifier]) ON UPDATE CASCADE;

IF OBJECT_ID(N'dbo.oauth2_refresh_token_session', N'U') IS NULL
CREATE TABLE [dbo].[oauth2_refresh_token_session] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [challenge_id] CHAR(36) NOT NULL,
    [request_id] VARCHAR(40) NOT NULL,
    [client_id] VARCHAR(255) NOT NULL,
    [signature] VARCHAR(255) NOT NULL,
    [subject] CHAR(36) NOT NULL,
    [requested_at] DATETIME2(7) NOT NULL DEFAULT GETDATE(),
    [requested_scopes] VARCHAR(MAX) NOT NULL,
    [granted_scopes] VARCHAR(MAX) NOT NULL,
    [requested_audience] VARCHAR(MAX) NULL,
    [granted_audience] VARCHAR(MAX) NULL,
    [active] BIT NOT NULL DEFAULT FALSE,
    [revoked] BIT NOT NULL DEFAULT FALSE,
    [form_data] VARCHAR(MAX) NOT NULL,
    [session_data] VARBINARY(MAX) NOT NULL
);

CREATE INDEX [oauth2_refresh_token_session_request_id_idx] ON [dbo].[oauth2_refresh_token_session] ([request_id]);
CREATE INDEX [oauth2_refresh_token_session_client_id_idx] ON [dbo].[oauth2_refresh_token_session] ([client_id]);
CREATE INDEX [oauth2_refresh_token_session_client_id_subject_idx] ON [dbo].[oauth2_refresh_token_session] ([client_id], [subject]);

ALTER TABLE [dbo].[oauth2_refresh_token_session]
    ADD CONSTRAINT [oauth2_refresh_token_session_challenge_id_fkey]
        FOREIGN KEY ([challenge_id])
            REFERENCES [dbo].[oauth2_consent_session] ([challenge_id]) ON DELETE CASCADE ON UPDATE CASCADE,
    ADD CONSTRAINT [oauth2_refresh_token_session_subject_fkey]
        FOREIGN KEY ([subject])
            REFERENCES [dbo].[user_opaque_identifier] ([identifier]);

IF OBJECT_ID(N'dbo.oauth2_par_context', N'U') IS NULL
CREATE TABLE [dbo].[oauth2_par_context] (
    [id] INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
    [request_id] VARCHAR(40) NOT NULL,
    [client_id] VARCHAR(255) NOT NULL,
    [signature] VARCHAR(255) NOT NULL,
    [requested_at] DATETIME2(7) NOT NULL DEFAULT GETDATE(),
    [scopes] VARCHAR(MAX) NOT NULL,
    [audience] VARCHAR(MAX) NOT NULL,
    [handled_response_types] VARCHAR(MAX) NOT NULL,
    [response_mode] VARCHAR(MAX) NOT NULL,
    [response_mode_default] VARCHAR(MAX) NOT NULL,
    [revoked] BIT NOT NULL DEFAULT FALSE,
    [form_data] VARCHAR(MAX) NOT NULL,
    [session_data] VARBINARY(MAX) NOT NULL
);

CREATE UNIQUE INDEX [oauth2_par_context_signature_key] ON [dbo].[oauth2_par_context] ([signature]);


CREATE UNIQUE INDEX [duo_devices_username_key] ON [dbo].[duo_devices] ([username]);
CREATE UNIQUE INDEX [encryption_name_key] ON [dbo].[encryption] ([name]);
CREATE UNIQUE INDEX [identity_verification_jti_key] ON [dbo].[identity_verification] ([jti]);
CREATE UNIQUE INDEX [totp_configurations_username_key] ON [dbo].[totp_configurations] ([username]);
CREATE UNIQUE INDEX [user_opaque_identifier_identifier_key] ON [dbo].[user_opaque_identifier] ([identifier]);
CREATE UNIQUE INDEX [user_opaque_identifier_lookup_key] ON [dbo].[user_opaque_identifier] ([service], [sector_id], [username]);
CREATE UNIQUE INDEX [user_preferences_username_key] ON [dbo].[user_preferences] ([username]);
CREATE UNIQUE INDEX [webauthn_devices_kid_key] ON [dbo].[webauthn_devices] ([kid]);
CREATE UNIQUE INDEX [webauthn_devices_lookup_key] ON [dbo].[webauthn_devices] ([username], [description]);
