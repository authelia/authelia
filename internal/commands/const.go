package commands

import (
	"errors"
	"regexp"
)

const (
	cliOutputFmtSuccessfulUserExportFile = "Successfully exported %d %s as %s to the '%s' file\n"
	cliOutputFmtSuccessfulUserImportFile = "Successfully imported %d %s from the %s file '%s' into the database\n"
)

const (
	fmtCmdAutheliaShort = "authelia %s"

	fmtCmdAutheliaLong = `authelia %s

An open-source authentication and authorization server providing
two-factor authentication and single sign-on (SSO) for your
applications via a web portal.

General documentation is available at: https://www.authelia.com/
CLI documentation is available at: https://www.authelia.com/reference/cli/authelia/authelia/`

	cmdAutheliaExample = `authelia --config /etc/authelia/config.yml --config /etc/authelia/access-control.yml
authelia --config /etc/authelia/config.yml,/etc/authelia/access-control.yml
authelia --config /etc/authelia/config/`

	fmtAutheliaBuild = `Last Tag: %s
State: %s
Branch: %s
Commit: %s
Build Number: %s
Build OS: %s
Build Arch: %s
Build Compiler: %s
Build Date: %s
Development: %t
Extra: %s`

	fmtAutheliaBuildGo = `
Go:
    Version: %s
    Module Path: %s
    Executable Path: %s
`

	cmdAutheliaBuildInfoShort = "Show the build information of Authelia"

	cmdAutheliaBuildInfoLong = `Show the build information of Authelia.

This outputs detailed version information about the specific version
of the Authelia binary. This information is embedded into Authelia
by the continuous integration.

This could be vital in debugging if you're not using a particular
tagged build of Authelia. It's suggested to provide it along with
your issue.
`
	cmdAutheliaBuildInfoExample = `authelia build-info`

	cmdAutheliaAccessControlShort = "Helpers for the access control system"

	cmdAutheliaAccessControlLong = `Helpers for the access control system.`

	cmdAutheliaAccessControlExample = `authelia access-control --help`

	cmdAutheliaAccessControlCheckPolicyShort = "Checks a request against the access control rules to determine what policy would be applied"

	cmdAutheliaAccessControlCheckPolicyLong = `
Checks a request against the access control rules to determine what policy would be applied.

Legend:

	#		The rule position in the configuration.
	*		The first fully matched rule.
	~		Potential match i.e. if the user was authenticated they may match this rule.
	hit     The criteria in this column is a match to the request.
	miss    The criteria in this column is not match to the request.
	may     The criteria in this column is potentially a match to the request.

Notes:

	A rule that potentially matches a request will cause a redirection to occur in order to perform one-factor
	authentication. This is so Authelia can adequately determine if the rule actually matches.
`
	cmdAutheliaAccessControlCheckPolicyExample = `authelia access-control check-policy --config config.yml --url https://example.com
authelia access-control check-policy --config config.yml --url https://example.com --username john
authelia access-control check-policy --config config.yml --url https://example.com --groups admin,public
authelia access-control check-policy --config config.yml --url https://example.com --username john --method GET
authelia access-control check-policy --config config.yml --url https://example.com --username john --method GET --verbose`

	cmdAutheliaStorageShort = "Manage the Authelia storage"

	cmdAutheliaStorageLong = `Manage the Authelia storage.

This subcommand has several methods to interact with the Authelia SQL Database. This allows doing several advanced
operations which would be much harder to do manually.
`

	cmdAutheliaStorageExample = `authelia storage --help`

	cmdAutheliaStorageCacheShort = "Manage storage cache"

	cmdAutheliaStorageCacheLong = `Manage storage cache.

This subcommand allows management of the storage cache.`

	cmdAutheliaStorageCacheExample = `authelia storage cache --help`

	cmdAutheliaStorageCacheMDS3Short = "Manage WebAuthn MDS3 cache storage"

	cmdAutheliaStorageCacheMDS3Long = `Manage WebAuthn MDS3 cache storage.

This subcommand allows management of the WebAuthn MDS3 cache storage.`

	cmdAutheliaStorageCacheMDS3Example = `authelia storage cache mds3 --help`

	cmdAutheliaStorageCacheMDS3DeleteShort = "Delete WebAuthn MDS3 cache storage"

	cmdAutheliaStorageCacheMDS3DeleteLong = `Delete WebAuthn MDS3 cache storage.

This subcommand allows deletion of the WebAuthn MDS3 cache storage.`

	cmdAutheliaStorageCacheMDS3DeleteExample = `authelia storage cache mds3 delete`

	cmdAutheliaStorageCacheMDS3UpdateShort = "Update WebAuthn MDS3 cache storage"

	cmdAutheliaStorageCacheMDS3UpdateLong = `Update WebAuthn MDS3 cache storage.

This subcommand allows updating of the WebAuthn MDS3 cache storage.`

	cmdAutheliaStorageCacheMDS3UpdateExample = `authelia storage cache mds3 update`

	cmdAutheliaStorageCacheMDS3DumpShort = "Dump WebAuthn MDS3 cache storage"

	cmdAutheliaStorageCacheMDS3DumpLong = `Dump WebAuthn MDS3 cache storage.

This subcommand allows dumping of the WebAuthn MDS3 cache storage to a file.`

	cmdAutheliaStorageCacheMDS3DumpExample = `authelia storage cache mds3 dump`

	cmdAutheliaStorageCacheMDS3StatusShort = "View WebAuthn MDS3 cache storage status"

	cmdAutheliaStorageCacheMDS3StatusLong = `View WebAuthn MDS3 cache storage status.

This subcommand allows management of the WebAuthn MDS3 cache storage.`

	cmdAutheliaStorageCacheMDS3StatusExample = `authelia storage cache mds3 status`

	cmdAutheliaStorageEncryptionShort = "Manage storage encryption"

	cmdAutheliaStorageEncryptionLong = `Manage storage encryption.

This subcommand allows management of the storage encryption.`

	cmdAutheliaStorageEncryptionExample = `authelia storage encryption --help`

	cmdAutheliaStorageEncryptionCheckShort = "Checks the encryption key against the database data"

	cmdAutheliaStorageEncryptionCheckLong = `Checks the encryption key against the database data.

This is useful for validating all data that can be encrypted is intact.`

	cmdAutheliaStorageEncryptionCheckExample = `authelia storage encryption check
authelia storage encryption check --verbose
authelia storage encryption check --verbose --config config.yml
authelia storage encryption check --verbose --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageEncryptionChangeKeyShort = "Changes the encryption key"

	cmdAutheliaStorageEncryptionChangeKeyLong = `Changes the encryption key.

This subcommand allows you to change the encryption key of an Authelia SQL database.`

	cmdAutheliaStorageEncryptionChangeKeyExample = `authelia storage encryption change-key --config config.yml --new-encryption-key 0e95cb49-5804-4ad9-be82-bb04a9ddecd8
authelia storage encryption change-key --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --new-encryption-key 0e95cb49-5804-4ad9-be82-bb04a9ddecd8 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageBansShort = "Manages user and ip bans"

	cmdAutheliaStorageBansLong = `Manages user and ip bans.

This subcommand allows listing, creating, and revoking user and ip bans from the regulation system.`

	cmdAutheliaStorageBansExample = `authelia storage bans --help`

	cmdAutheliaStorageBansUserShort = "Manages user bans"

	cmdAutheliaStorageBansUserLong = `Manages user bans.

This subcommand allows listing, creating, and revoking user bans from the regulation system.`

	cmdAutheliaStorageBansUserExample = `authelia storage bans user --help`

	cmdAutheliaStorageBansIPShort = "Manages ip bans"

	cmdAutheliaStorageBansIPLong = `Manages ip bans.

This subcommand allows listing, creating, and revoking ip bans from the regulation system.`

	cmdAutheliaStorageBansIPExample = `authelia storage bans ip --help`

	cmdAutheliaStorageBansListShort = "Lists %s bans"

	cmdAutheliaStorageBansListLong = `Lists %s bans.

This subcommand allows listing %s bans from the regulation system.`

	cmdAutheliaStorageBansListExample = `authelia storage bans %s --help`

	cmdAutheliaStorageBansAddShort = "Adds %s bans"

	cmdAutheliaStorageBansAddLong = `Adds %s bans.

This subcommand allows adding %s bans to the regulation system.`

	cmdAutheliaStorageBansAddExample = `authelia storage bans %s add --help`

	cmdAutheliaStorageBansRevokeShort = "Revokes %s bans"

	cmdAutheliaStorageBansRevokeLong = `Revokes %s bans.

This subcommand allows revoking %s bans in the regulation system.`

	cmdAutheliaStorageBansRevokeExample = `authelia storage bans %s revoke --help`

	cmdAutheliaStorageUserShort = "Manages user settings"

	cmdAutheliaStorageUserLong = `Manages user settings.

This subcommand allows modifying and exporting user settings.`

	cmdAutheliaStorageUserExample = `authelia storage user --help`

	cmdAutheliaStorageUserIdentifiersShort = "Manage user opaque identifiers"

	cmdAutheliaStorageUserIdentifiersLong = `Manage user opaque identifiers.

This subcommand allows performing various tasks related to the opaque identifiers for users.`

	cmdAutheliaStorageUserIdentifiersExample = `authelia storage user identifiers --help`

	cmdAutheliaStorageUserIdentifiersExportShort = "Export the identifiers to a YAML file"

	cmdAutheliaStorageUserIdentifiersExportLong = `Export the identifiers to a YAML file.

This subcommand allows exporting the opaque identifiers for users in order to back them up.`

	cmdAutheliaStorageUserIdentifiersExportExample = `authelia storage user identifiers export
authelia storage user identifiers export --file export.yml
authelia storage user identifiers export --file export.yml --config config.yml
authelia storage user identifiers export --file export.yml --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageUserIdentifiersImportShort = "Import the identifiers from a YAML file"

	cmdAutheliaStorageUserIdentifiersImportLong = `Import the identifiers from a YAML file.

This subcommand allows you to import the opaque identifiers for users from a YAML file.

The YAML file can either be automatically generated using the authelia storage user identifiers export command, or
manually provided the file is in the same format.`

	cmdAutheliaStorageUserIdentifiersImportExample = `authelia storage user identifiers import
authelia storage user identifiers import authelia.export.opaque-identifiers.yml
authelia storage user identifiers import --config config.yml export.yml
authelia storage user identifiers import --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw export.yml`

	cmdAutheliaStorageUserIdentifiersGenerateShort = "Generate opaque identifiers in bulk"

	cmdAutheliaStorageUserIdentifiersGenerateLong = `Generate opaque identifiers in bulk.

This subcommand allows various options for generating the opaque identifies for users in bulk.`

	cmdAutheliaStorageUserIdentifiersGenerateExample = `authelia storage user identifiers generate --users john,mary
authelia storage user identifiers generate --users john,mary --services openid
authelia storage user identifiers generate --users john,mary --services openid --sectors=",example.com,test.com"
authelia storage user identifiers generate --users john,mary --services openid --sectors=",example.com,test.com" --config config.yml
authelia storage user identifiers generate --users john,mary --services openid --sectors=",example.com,test.com" --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageUserIdentifiersAddShort = "Add an opaque identifier for a user to the database"

	cmdAutheliaStorageUserIdentifiersAddLong = `Add an opaque identifier for a user to the database.

This subcommand allows manually adding an opaque identifier for a user to the database provided it's in the correct format.`

	cmdAutheliaStorageUserIdentifiersAddExample = `authelia storage user identifiers add john --identifier f0919359-9d15-4e15-bcba-83b41620a073
authelia storage user identifiers add john --identifier f0919359-9d15-4e15-bcba-83b41620a073 --config config.yml
authelia storage user identifiers add john --identifier f0919359-9d15-4e15-bcba-83b41620a073 --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageUserWebAuthnShort = "Manage WebAuthn credentials"

	cmdAutheliaStorageUserWebAuthnLong = `Manage WebAuthn credentials.

This subcommand allows interacting with WebAuthn credentials.`

	cmdAutheliaStorageUserWebAuthnExample = `authelia storage user webauthn --help`

	cmdAutheliaStorageUserWebAuthnImportShort = "Perform imports of the WebAuthn credentials"

	cmdAutheliaStorageUserWebAuthnImportLong = `Perform imports of the WebAuthn credentials.

This subcommand allows importing WebAuthn credentials from the YAML format.`

	cmdAutheliaStorageUserWebAuthnImportExample = `authelia storage user webauthn export
authelia storage user webauthn import --file authelia.export.webauthn.yml
authelia storage user webauthn import --file authelia.export.webauthn.yml --config config.yml
authelia storage user webauthn import --file authelia.export.webauthn.yml --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageUserWebAuthnExportShort = "Perform exports of the WebAuthn credentials"

	cmdAutheliaStorageUserWebAuthnExportLong = `Perform exports of the WebAuthn credentials.

This subcommand allows exporting WebAuthn credentials to various formats.`

	cmdAutheliaStorageUserWebAuthnExportExample = `authelia storage user webauthn export
authelia storage user webauthn export --file authelia.export.webauthn.yml
authelia storage user webauthn export --config config.yml
authelia storage user webauthn export--encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageUserWebAuthnListShort = "List WebAuthn credentials"

	cmdAutheliaStorageUserWebAuthnListLong = `List WebAuthn credentials.

This subcommand allows listing WebAuthn credentials.`

	cmdAutheliaStorageUserWebAuthnListExample = `authelia storage user webauthn list
authelia storage user webauthn list john
authelia storage user webauthn list --config config.yml
authelia storage user webauthn list john --config config.yml
authelia storage user webauthn list --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw
authelia storage user webauthn list john --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageUserWebAuthnVerifyShort = "Verify WebAuthn credentials"

	cmdAutheliaStorageUserWebAuthnVerifyLong = `Verify WebAuthn credentials.

This subcommand allows verifying registered WebAuthn credentials.`

	cmdAutheliaStorageUserWebAuthnVerifyExample = `authelia storage user webauthn verify`

	cmdAutheliaStorageUserWebAuthnDeleteShort = "Delete a WebAuthn credential"

	cmdAutheliaStorageUserWebAuthnDeleteLong = `Delete a WebAuthn credential.

This subcommand allows deleting a WebAuthn credential directly from the database.`

	cmdAutheliaStorageUserWebAuthnDeleteExample = `authelia storage user webauthn delete john --all
authelia storage user webauthn delete john --all --config config.yml
authelia storage user webauthn delete john --all --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw
authelia storage user webauthn delete john --description Primary
authelia storage user webauthn delete john --description Primary --config config.yml
authelia storage user webauthn delete john --description Primary --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw
authelia storage user webauthn delete --kid abc123
authelia storage user webauthn delete --kid abc123 --config config.yml
authelia storage user webauthn delete --kid abc123 --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageUserTOTPShort = "Manage TOTP configurations"

	cmdAutheliaStorageUserTOTPLong = `Manage TOTP configurations.

This subcommand allows deleting, exporting, and creating user TOTP configurations.`

	cmdAutheliaStorageUserTOTPExample = `authelia storage user totp --help`

	cmdAutheliaStorageUserTOTPGenerateShort = "Generate a TOTP configuration for a user"

	cmdAutheliaStorageUserTOTPGenerateLong = `Generate a TOTP configuration for a user.

This subcommand allows generating a new TOTP configuration for a user,
and overwriting the existing configuration if applicable.`

	cmdAutheliaStorageUserTOTPGenerateExample = `authelia storage user totp generate john
authelia storage user totp generate john --period 90
authelia storage user totp generate john --digits 8
authelia storage user totp generate john --algorithm SHA512
authelia storage user totp generate john --algorithm SHA512 --config config.yml
authelia storage user totp generate john --algorithm SHA512 --config config.yml --path john.png`

	cmdAutheliaStorageUserTOTPDeleteShort = "Delete a TOTP configuration for a user"

	cmdAutheliaStorageUserTOTPDeleteLong = `Delete a TOTP configuration for a user.

This subcommand allows deleting a TOTP configuration directly from the database for a given user.`

	cmdAutheliaStorageUserTOTPDeleteExample = `authelia storage user totp delete john
authelia storage user totp delete john --config config.yml
authelia storage user totp delete john --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageUserTOTPImportShort = "Perform imports of the TOTP configurations"

	cmdAutheliaStorageUserTOTPImportLong = `Perform imports of the TOTP configurations.

This subcommand allows importing TOTP configurations from the YAML format.`

	cmdAutheliaStorageUserTOTPImportExample = `authelia storage user totp import authelia.export.totp.yml
authelia storage user totp import --config config.yml authelia.export.totp.yml
authelia storage user totp import --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw authelia.export.totp.yml`

	cmdAutheliaStorageUserTOTPExportShort = "Perform exports of the TOTP configurations"

	cmdAutheliaStorageUserTOTPExportLong = `Perform exports of the TOTP configurations.

This subcommand allows exporting TOTP configurations to importable YAML files, or use the subcommands to export them to other non-importable formats.`

	cmdAutheliaStorageUserTOTPExportExample = `authelia storage user totp export --file example.yml
authelia storage user totp export --config config.yml
authelia storage user totp export --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageUserTOTPExportCSVShort = "Perform exports of the TOTP configurations to a CSV"

	cmdAutheliaStorageUserTOTPExportCSVLong = `Perform exports of the TOTP configurations to a CSV.

This subcommand allows exporting TOTP configurations to a CSV.`

	cmdAutheliaStorageUserTOTPExportCSVExample = `authelia storage user totp export csv --file users.csv
authelia storage user totp export csv --config config.yml
authelia storage user totp export csv --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageUserTOTPExportURIShort = "Perform exports of the TOTP configurations to URIs"

	cmdAutheliaStorageUserTOTPExportURILong = `Perform exports of the TOTP configurations to URIs.

This subcommand allows exporting TOTP configurations to TOTP URIs.`

	cmdAutheliaStorageUserTOTPExportURIExample = `authelia storage user totp export uri
authelia storage user totp export uri --config config.yml
authelia storage user totp export uri --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageUserTOTPExportPNGShort = "Perform exports of the TOTP configurations to QR code PNG images"

	cmdAutheliaStorageUserTOTPExportPNGLong = `Perform exports of the TOTP configurations to QR code PNG images.

This subcommand allows exporting TOTP configurations to PNG images with QR codes which represent the appropriate URI so they can be scanned.`

	cmdAutheliaStorageUserTOTPExportPNGExample = `authelia storage user totp export png
authelia storage user totp export png --directory example/dir
authelia storage user totp export png --config config.yml
authelia storage user totp export png --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageSchemaInfoShort = "Show the storage information"

	cmdAutheliaStorageSchemaInfoLong = `Show the storage information.

This subcommand shows advanced information about the storage schema useful in some diagnostic tasks.`

	cmdAutheliaStorageSchemaInfoExample = `authelia storage schema-info
authelia storage schema-info --config config.yml
authelia storage schema-info --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageMigrateShort = "Perform or list migrations"

	cmdAutheliaStorageMigrateLong = `Perform or list migrations.

This subcommand handles schema migration tasks.`

	cmdAutheliaStorageMigrateExample = `authelia storage migrate --help`

	cmdAutheliaStorageMigrateHistoryShort = "Show migration history"

	cmdAutheliaStorageMigrateHistoryLong = `Show migration history.

This subcommand allows users to list previous migrations.`

	cmdAutheliaStorageMigrateHistoryExample = `authelia storage migrate history
authelia storage migrate history --config config.yml
authelia storage migrate history --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageMigrateListUpShort = "List the up migrations available"

	cmdAutheliaStorageMigrateListUpLong = `List the up migrations available.

This subcommand lists the schema migrations available in this version of Authelia which are greater than the current
schema version of the database.`

	cmdAutheliaStorageMigrateListUpExample = `authelia storage migrate list-up
authelia storage migrate list-up --config config.yml
authelia storage migrate list-up --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageMigrateListDownShort = "List the down migrations available"

	cmdAutheliaStorageMigrateListDownLong = `List the down migrations available.

This subcommand lists the schema migrations available in this version of Authelia which are less than the current
schema version of the database.`

	cmdAutheliaStorageMigrateListDownExample = `authelia storage migrate list-down
authelia storage migrate list-down --config config.yml
authelia storage migrate list-down --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageMigrateUpShort = "Perform a migration up"

	cmdAutheliaStorageMigrateUpLong = `Perform a migration up.

This subcommand performs the schema migrations available in this version of Authelia which are greater than the current
schema version of the database. By default this will migrate up to the latest available, but you can customize this.`

	cmdAutheliaStorageMigrateUpExample = `authelia storage migrate up
authelia storage migrate up --config config.yml
authelia storage migrate up --target 20 --config config.yml
authelia storage migrate up --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageMigrateDownShort = "Perform a migration down"

	cmdAutheliaStorageMigrateDownLong = `Perform a migration down.

This subcommand performs the schema migrations available in this version of Authelia which are less than the current
schema version of the database.`

	cmdAutheliaStorageMigrateDownExample = `authelia storage migrate down --target 20
authelia storage migrate down --target 20 --config config.yml
authelia storage migrate down --target 20 --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw`

	cmdAutheliaStorageLogsShort = "Manage various types of logs"

	cmdAutheliaStorageLogsLong = "Commands for managing different types of logs stored in the database"

	cmdAutheliaStorageLogsAuthShort = "Manage authentication logs"

	cmdAutheliaStorageLogsAuthLong = "Commands for managing authentication logs including pruning old entries and viewing statistics"

	cmdAutheliaStorageLogsAuthStatsShort = "Show authentication logs statistics"

	cmdAutheliaStorageLogsAuthStatsLong = "Display statistics about the authentication logs table including success/failure rates and record counts."

	cmdAutheliaStorageLogsAuthStatsExample = `authelia storage logs auth stats`

	cmdAutheliaStorageLogsAuthPruneShort = "Prune old authentication logs"

	cmdAutheliaStorageLogsAuthPruneLong = `Prune authentication logs based on age criteria.

This command helps manage the authentication_logs table which grows indefinitely.
Use this to remove authentication records older than a specified period to
prevent excessive database growth and improve performance.`

	cmdAutheliaStorageLogsAuthPruneExample = `authelia storage logs auth prune --older-than 90d
authelia storage logs auth prune --older-than 6m --batch-size 50000
authelia storage logs auth prune --older-than 1y --dry-run`

	cmdAutheliaConfigShort = "Perform config related actions"

	cmdAutheliaConfigLong = `Perform config related actions.

This subcommand contains other subcommands related to the configuration.`

	cmdAutheliaConfigExample = `authelia config --help`

	cmdAutheliaConfigTemplateShort = "Template a configuration file or files with enabled filters"

	cmdAutheliaConfigTemplateLong = `Template a configuration file or files with enabled filters.

This subcommand allows debugging the filtered YAML files with any of the available filters. It should be noted this
command needs to be executed with the same environment variables and working path as when normally running Authelia to
be useful.`

	cmdAutheliaConfigTemplateExample = `authelia config template --config.experimental.filters=template --config=config.yml`

	cmdAutheliaConfigValidateShort = "Check a configuration against the internal configuration validation mechanisms"

	cmdAutheliaConfigValidateLong = `Check a configuration against the internal configuration validation mechanisms.

This subcommand allows validation of the YAML and Environment configurations so that a configuration can be checked
prior to deploying it.`

	cmdAutheliaConfigValidateExample = `authelia config validate
authelia config validate --config config.yml`

	cmdAutheliaConfigValidateLegacyExample = `authelia validate-config
authelia validate-config --config config.yml`

	cmdAutheliaCryptoShort = "Perform cryptographic operations"

	cmdAutheliaCryptoLong = `Perform cryptographic operations.

This subcommand allows performing cryptographic certificate, key pair, etc tasks.`

	cmdAutheliaCryptoExample = `authelia crypto --help`

	cmdAutheliaCryptoRandShort = "Generate a cryptographically secure random string"

	cmdAutheliaCryptoRandLong = `Generate a cryptographically secure random string.

This subcommand allows generating cryptographically secure random strings for use for encryption keys, HMAC keys, etc.`

	cmdAutheliaCryptoRandExample = `authelia crypto rand --help
authelia crypto rand --length 80
authelia crypto rand -n 80
authelia crypto rand --charset alphanumeric
authelia crypto rand --charset alphabetic
authelia crypto rand --charset ascii
authelia crypto rand --charset numeric
authelia crypto rand --charset numeric-hex
authelia crypto rand --characters 0123456789ABCDEF
authelia crypto rand directory/file1 directory/file2
authelia crypto rand --file directory/file3,directory/file4`

	cmdAutheliaCryptoHashShort = "Perform cryptographic hash operations"

	cmdAutheliaCryptoHashLong = `Perform cryptographic hash operations.

This subcommand allows performing hashing cryptographic tasks.`

	cmdAutheliaCryptoHashExample = `authelia crypto hash --help`

	cmdAutheliaCryptoHashValidateShort = "Perform cryptographic hash validations"

	cmdAutheliaCryptoHashValidateLong = `Perform cryptographic hash validations.

This subcommand allows performing cryptographic hash validations. i.e. checking hash digests against a password.`

	cmdAutheliaCryptoHashValidateExample = `authelia crypto hash validate --help
authelia crypto hash validate --password 'p@ssw0rd' -- '$5$rounds=500000$WFjMpdCQxIkbNl0k$M0qZaZoK8Gwdh8Cw5diHgGfe5pE0iJvxcVG3.CVnQe.'`

	cmdAutheliaCryptoHashGenerateShort = "Generate cryptographic hash digests"

	cmdAutheliaCryptoHashGenerateLong = `Generate cryptographic hash digests.

This subcommand allows generating cryptographic hash digests.

See the help for the subcommands if you want to override the configuration or defaults.`

	cmdAutheliaCryptoHashGenerateExample = `authelia crypto hash generate --help`

	fmtCmdAutheliaCryptoHashGenerateSubShort = "Generate cryptographic %s hash digests"

	fmtCmdAutheliaCryptoHashGenerateSubLong = `Generate cryptographic %s hash digests.

This subcommand allows generating cryptographic %s hash digests.`

	fmtCmdAutheliaCryptoHashGenerateSubExample = `authelia crypto hash generate %s --help`

	cmdAutheliaCryptoCertificateShort = "Perform certificate cryptographic operations"

	cmdAutheliaCryptoCertificateLong = `Perform certificate cryptographic operations.

This subcommand allows performing certificate cryptographic tasks.`

	cmdAutheliaCryptoCertificateExample = `authelia crypto certificate --help`

	fmtCmdAutheliaCryptoCertificateSubShort = "Perform %s certificate cryptographic operations"

	fmtCmdAutheliaCryptoCertificateSubLong = `Perform %s certificate cryptographic operations.

This subcommand allows performing %s certificate cryptographic tasks.`

	fmtCmdAutheliaCryptoCertificateSubExample = `authelia crypto certificate %s --help`

	fmtCmdAutheliaCryptoCertificateGenerateRequestShort = "Generate an %s private key and %s"

	fmtCmdAutheliaCryptoCertificateGenerateRequestLong = `Generate an %s private key and %s.

This subcommand allows generating an %s private key and %s.`

	cmdAutheliaCryptoCertificateRSAGenerateExample = `authelia crypto certificate rsa generate --help`

	cmdAutheliaCryptoCertificateECDSAGenerateExample = `authelia crypto certificate ecdsa generate --help`

	cmdAutheliaCryptoCertificateEd25519GenerateExample = `authelia crypto certificate ed25519 request --help`

	cmdAutheliaCryptoCertificateRSARequestExample = `authelia crypto certificate rsa request --help`

	cmdAutheliaCryptoCertificateECDSARequestExample = `authelia crypto certificate ecdsa request --help`

	cmdAutheliaCryptoCertificateEd25519RequestExample = `authelia crypto certificate ed25519 request --help`

	cmdAutheliaCryptoPairShort = "Perform key pair cryptographic operations"

	cmdAutheliaCryptoPairLong = `Perform key pair cryptographic operations.

This subcommand allows performing key pair cryptographic tasks.`

	cmdAutheliaCryptoPairExample = `authelia crypto pair --help`

	cmdAutheliaCryptoPairSubShort = "Perform %s key pair cryptographic operations"

	cmdAutheliaCryptoPairSubLong = `Perform %s key pair cryptographic operations.

This subcommand allows performing %s key pair cryptographic tasks.`

	cmdAutheliaCryptoPairRSAExample = `authelia crypto pair rsa --help`

	cmdAutheliaCryptoPairECDSAExample = `authelia crypto pair ecdsa --help`

	cmdAutheliaCryptoPairEd25519Example = `authelia crypto pair ed25519 --help`

	fmtCmdAutheliaCryptoPairGenerateShort = "Generate a cryptographic %s key pair"

	fmtCmdAutheliaCryptoPairGenerateLong = `Generate a cryptographic %s key pair.

This subcommand allows generating an %s key pair.`

	cmdAutheliaCryptoPairRSAGenerateExample = `authelia crypto pair rsa generate --help`

	cmdAutheliaCryptoPairECDSAGenerateExample = `authelia crypto pair ecdsa generate --help`

	cmdAutheliaCryptoPairEd25519GenerateExample = `authelia crypto pair ed25519 generate --help`

	cmdAutheliaDebugShort = "Perform debug functions"

	cmdAutheliaDebugLong = `Perform debug related functions.

This subcommand contains other subcommands related to debugging.`

	cmdAutheliaDebugExample = `authelia debug --help`

	cmdAutheliaDebugTLSShort = "Perform a TLS debug operation"

	cmdAutheliaDebugTLSLong = `Perform a TLS debug operation.

This subcommand allows checking a remote server's TLS configuration and the ability to validate the certificate.`

	cmdAutheliaDebugTLSExample = `authelia debug tls tcp://smtp.example.com:465`

	cmdAutheliaDebugExpressionShort = "Perform a user attribute expression debug operation"

	cmdAutheliaDebugExpressionLong = `Perform a user attribute expression debug operation.

This subcommand allows checking a user attribute expression against a specific user.`

	cmdAutheliaDebugExpressionExample = `authelia debug expression username "'abc' in groups"`

	cmdAutheliaDebugOIDCShort = "Perform a OpenID Connect 1.0 debug operation"

	cmdAutheliaDebugOIDCLong = `Perform a OpenID Connect 1.0 debug operation.

This subcommand allows checking certain OpenID Connect 1.0 scenarios.`

	cmdAutheliaDebugOIDCExample = `authelia debug oidc --help`

	cmdAutheliaDebugOIDCClaimsShort = "Perform a OpenID Connect 1.0 claims hydration debug operation"

	cmdAutheliaDebugOIDCClaimsLong = `Perform a OpenID Connect 1.0 claims hydration debug operation.

This subcommand allows checking an OpenID Connect 1.0 claims hydration scenario by providing certain information about a request.`

	cmdAutheliaDebugOIDCClaimsExample = `authelia debug oidc claims --help`
)

const (
	storageMigrateDirectionUp   = "up"
	storageMigrateDirectionDown = "down"
)

const (
	storageLogs          = "logs"
	storageLogsAuth      = "auth"
	storageLogsAuthPrune = "prune"
	storageLogsAuthStats = "stats"

	cmdFlagLogsOlderThan = "older-than"
	cmdFlagLogsBatchSize = "batch-size"
	cmdFlagLogsDryRun    = "dry-run"
)

const (
	cmdFlagNameDirectory       = "directory"
	cmdFlagNameModeDirectories = "mode-dirs"

	cmdFlagNamePathCA  = "path.ca"
	cmdFlagNameBundles = "bundles"
	cmdFlagNameLegacy  = "legacy"

	cmdFlagNameFileExtensionLegacy    = "file.extension.legacy"
	cmdFlagNameFilePrivateKey         = "file.private-key"
	cmdFlagNameFilePublicKey          = "file.public-key"
	cmdFlagNameFileCertificate        = "file.certificate"
	cmdFlagNameFileBundleChain        = "file.bundle.chain"
	cmdFlagNameFileBundlePrivKeyChain = "file.bundle.priv-chain"
	cmdFlagNameFileCAPrivateKey       = "file.ca-private-key"
	cmdFlagNameFileCACertificate      = "file.ca-certificate"
	cmdFlagNameFileCSR                = "file.csr"

	cmdFlagNameExtendedUsage = "extended-usage"
	cmdFlagNameSignature     = "signature"
	cmdFlagNameCA            = "ca"
	cmdFlagNameSANs          = "sans"

	cmdFlagNameCommonName         = "common-name"
	cmdFlagNameOrganization       = "organization"
	cmdFlagNameOrganizationalUnit = "organizational-unit"
	cmdFlagNameCountry            = "country"
	cmdFlagNameProvince           = "province"
	cmdFlagNameLocality           = "locality"
	cmdFlagNameStreetAddress      = "street-address"
	cmdFlagNamePostcode           = "postcode"

	cmdFlagNameNotBefore = "not-before"
	cmdFlagNameNotAfter  = "not-after"
	cmdFlagNameDuration  = "duration"

	cmdFlagNameBits  = "bits"
	cmdFlagNameCurve = "curve"

	cmdFlagNamePassword         = "password"
	cmdFlagNameRandom           = "random"
	cmdFlagNameRandomLength     = "random.length"
	cmdFlagNameRandomCharSet    = "random.charset"
	cmdFlagNameRandomCharacters = "random.characters"
	cmdFlagNameNoConfirm        = "no-confirm"
	cmdFlagNameVariant          = "variant"
	cmdFlagNameCost             = "cost"
	cmdFlagNameIterations       = "iterations"
	cmdFlagNameParallelism      = "parallelism"
	cmdFlagNameBlockSize        = "block-size"
	cmdFlagNameMemory           = "memory"
	cmdFlagNameKeySize          = "key-size"
	cmdFlagNameSaltSize         = "salt-size"
	cmdFlagNameProfile          = "profile"

	cmdConfigDefaultContainer = "/config/configuration.yml"
	cmdConfigDefaultDaemon    = "/etc/authelia/configuration.yml"

	cmdFlagNameConfig    = "config"
	cmdFlagEnvNameConfig = "X_AUTHELIA_CONFIG"

	cmdFlagNameConfigExpFilters = "config.experimental.filters"
	cmdFlagEnvNameConfigFilters = "X_AUTHELIA_CONFIG_FILTERS"

	cmdFlagNameCharSet     = "charset"
	cmdFlagValueCharSet    = "alphanumeric"
	cmdFlagUsageCharset    = "sets the charset for the random password, options are 'ascii', 'alphanumeric', 'alphabetic', 'numeric', 'numeric-hex', and 'rfc3986'"
	cmdFlagNameCharacters  = "characters"
	cmdFlagUsageCharacters = "sets the explicit characters for the random string"
	cmdFlagNameLength      = "length"
	cmdFlagUsageLength     = "sets the character length for the random string"

	cmdFlagNameNewEncryptionKey = "new-encryption-key"

	cmdFlagNameFile        = "file"
	cmdFlagNameModeFiles   = "mode-files"
	cmdFlagNameUsers       = "users"
	cmdFlagNameServices    = "services"
	cmdFlagNameSectors     = "sectors"
	cmdFlagNameIdentifier  = "identifier"
	cmdFlagNameService     = "service"
	cmdFlagNameSector      = "sector"
	cmdFlagNameDescription = "description"
	cmdFlagNameAll         = "all"
	cmdFlagNameKeyID       = "kid"
	cmdFlagNameVerbose     = "verbose"
	cmdFlagNameSecret      = "secret"
	cmdFlagNameSecretSize  = "secret-size"
	cmdFlagNamePeriod      = "period"
	cmdFlagNameDigits      = "digits"
	cmdFlagNameAlgorithm   = "algorithm"
	cmdFlagNameIssuer      = "issuer"
	cmdFlagNameForce       = "force"
	cmdFlagNamePath        = "path"
	cmdFlagNameTarget      = "target"
	cmdFlagNameDestroyData = "destroy-data"

	cmdFlagNameEncryptionKey      = "encryption-key"
	cmdFlagNameSQLite3Path        = "sqlite.path"
	cmdFlagNameMySQLAddress       = "mysql.address"
	cmdFlagNameMySQLDatabase      = "mysql.database"
	cmdFlagNameMySQLUsername      = "mysql.username"
	cmdFlagNameMySQLPassword      = "mysql.password"
	cmdFlagNamePostgreSQLAddress  = "postgres.address"
	cmdFlagNamePostgreSQLDatabase = "postgres.database"
	cmdFlagNamePostgreSQLSchema   = "postgres.schema"
	cmdFlagNamePostgreSQLUsername = "postgres.username"
	cmdFlagNamePostgreSQLPassword = "postgres.password"
)

const (
	cmdUseHash          = "hash"
	cmdUseHashArgon2    = "argon2"
	cmdUseHashSHA2Crypt = "sha2crypt"
	cmdUseHashPBKDF2    = "pbkdf2"
	cmdUseHashBcrypt    = "bcrypt"
	cmdUseHashScrypt    = "scrypt"

	cmdUseExport         = "export"
	cmdUseImportFileName = "import <filename>"

	cmdUseCrypto      = "crypto"
	cmdUseRand        = "rand"
	cmdUseCertificate = "certificate"
	cmdUseGenerate    = "generate"
	cmdUseValidate    = "validate"
	cmdUseFmtValidate = "%s [flags] -- <digest>"
	cmdUseRequest     = "request"
	cmdUsePair        = "pair"
	cmdUseRSA         = "rsa"
	cmdUseECDSA       = "ecdsa"
	cmdUseEd25519     = "ed25519"
	cmdUseUser        = "user"
	cmdUseIP          = "ip"
)

const (
	cryptoCertPubCertOut = "certificate"
	cryptoCertCSROut     = "certificate signing request"

	prefixFilePassword = "authentication_backend.file.password"
)

var (
	errStorageSchemaOutdated     = errors.New("storage schema outdated")
	errStorageSchemaIncompatible = errors.New("storage schema incompatible")
)

const (
	identifierServiceOpenIDConnect = "openid"
	invalid                        = "invalid"
)

var (
	validIdentifierServices = []string{identifierServiceOpenIDConnect}
)

const (
	helpTopicConfigFilters = `Configuration Filters are a system for templating configuration files.

Using the --config.experimental.filters flag users can define multiple filters to apply to all configuration files that
are loaded by Authelia. These filters are applied after loading the file data from the filesystem, but before they are
parsed by the relevant file format parser.

The filters are processed in the order specified, and the content of each configuration file is logged as a base64 raw
string when the log level is set to trace.

The following filters are available:

	template:

		This filter uses the go template system to filter the file. In addition to the standard functions, several
		custom functions exist to facilitate this process.

		For a full list of functions see: https://www.authelia.com/reference/guides/templating/#functions

	expand-env:

		DEPRECATED: This filter expands environment variables in place where specified in the configuration. For example
        the string ${DOMAIN_NAME} will be replaced with the value from the DOMAIN_NAME environment variable or an empty
		string.`

	helpTopicConfig = `Configuration can be specified in multiple layers where each layer is a different source from
the last. The layers are loaded in the order below where each layer potentially overrides the individual settings from
previous layers with the individual settings it provides (i.e. if the same setting is specified twice).

Layers:
  - File/Directory Paths
  - Environment Variables
  - Secrets

File/Directory Paths:

	File/Directory Paths can be specified either via the '--config' CLI argument or the 'X_AUTHELIA_CONFIG' environment
	variable. If both the environment variable AND the CLI argument are specified the environment variable is completely
	ignored. These values both take lists separated by commas.

	Directories that are loaded via this method load all files with relevant extensions from the directory, this is not
    recursive. This means all files with these extensions must be Authelia configuration files with valid syntax.

	The paths specified are loaded in order, where individual settings specified by later files potentially overrides
	individual settings by later files (i.e. if the same setting is specified twice). Files specified Files in
	directories are loaded in lexicographic order.

	The files loaded via this method can be interpolated or templated via the configuration filters. Read more about
	this topic by running: authelia -h authelia filters

Environment Variables:

	Most configuration options in Authelia can be specified via an environment variable. The available options and the
	specific environment variable mapping can be found here: https://www.authelia.com/configuration/methods/environment/

Secrets:

	Some configuration options in Authelia can be specified via an environment variable which refers to the location of
	a file; also known as a secret. Every configuration key that ends with the following strings can be loaded in this
	way: 'key', 'secret', 'password', 'token'.

	The available options and the specific secret mapping can be found here: https://www.authelia.com/configuration/methods/secrets/`

	helpTopicTimeLayouts = `Several commands take date time inputs which are parsed. These inputs are parsed with
specific layouts in mind and these layouts are handled in order.

Format:

	The layouts use a format where specific sequence of characters are representative of a portion of each timestamp.

	See the go documentation for more information on how these layouts work, however the layouts are fairly self
	explanatory and you can just use standard unix timestamps if desired.

Layouts:

	Unix (µs): 1675899060000000
	Unix (ms): 1675899060000
	Unix (s): 1675899060
	Simple: Jan 2 15:04:05 2006
	Date Time: 2006-01-02 15:04:05
	RFC3339: 2006-01-02T15:04:05Z07:00
	RFC1123 with numeric timezone: Mon, 02 Jan 2006 15:04:05 -0700
	Ruby Date: Mon Jan 02 15:04:05 -0700 2006
	ANSIC: Mon Jan _2 15:04:05 2006
	Date: 2006-01-02`

	//nolint:gosec // Not a credential, it's the text of a help topic.
	helpTopicHashPassword = `The 'authelia hash-password' command has been replaced with the
'authelia crypto hash generate' command. Run 'authelia crypto hash generate --help'
for more information.

It was replaced for a few reasons. Specifically it was confusing to users
due to arguments which only had an effect on one algorithm and not the other,
and the new command makes the available options a lot clearer. In addition
the old command was not compatible with all of the available algorithms the
new one is compatible for and retrofitting it would be incredibly difficult.`
)

const (
	fmtYAMLConfigTemplateHeader = `
---
##
## Authelia rendered configuration file (file filters).
##
## Filters: %s
##
`

	fmtYAMLConfigTemplateFileHeader = `
---
##
## File Source Path: %s
##

`
)

const (
	wordYes = "Yes"
	wordNo  = "No"
)

const (
	suffixAlgorithm           = ".algorithm"
	suffixSHA2CryptVariant    = ".sha2crypt.variant"
	suffixSHA2CryptIterations = ".sha2crypt.iterations"
	suffixSHA2CryptSaltLength = ".sha2crypt.salt_length"
	suffixPBKDF2Variant       = ".pbkdf2.variant"
	suffixPBKDF2Iterations    = ".pbkdf2.iterations"
	suffixPBKDF2KeyLength     = ".pbkdf2.key_length"
	suffixPBKDF2SaltLength    = ".pbkdf2.salt_length"
	suffixBcryptVariant       = ".bcrypt.variant"
	suffixBcryptCost          = ".bcrypt.cost"
	suffixScryptVariant       = ".scrypt.variant"
	suffixScryptIterations    = ".scrypt.iterations"
	suffixScryptBlockSize     = ".scrypt.block_size"
	suffixScryptParallelism   = ".scrypt.parallelism"
	suffixScryptKeyLength     = ".scrypt.key_length"
	suffixScryptSaltLength    = ".scrypt.salt_length"
	suffixArgon2Variant       = ".argon2.variant"
	suffixArgon2Iterations    = ".argon2.iterations"
	suffixArgon2Memory        = ".argon2.memory"
	suffixArgon2Parallelism   = ".argon2.parallelism"
	suffixArgon2KeyLength     = ".argon2.key_length"
	suffixArgon2SaltLength    = ".argon2.salt_length"
)

var (
	reYAMLComment = regexp.MustCompile(`^---\n([.\n]*)`)
)
