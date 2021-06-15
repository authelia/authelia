package main

import (
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/notification"
	"github.com/authelia/authelia/internal/oidc"
	"github.com/authelia/authelia/internal/regulation"
	"github.com/authelia/authelia/internal/session"
	"github.com/authelia/authelia/internal/storage"
	"github.com/authelia/authelia/internal/utils"
	"github.com/authelia/authelia/v4"
)

func configureProviders(config *schema.Configuration) (providers middlewares.Providers, nonFatalErrs, errs []error) {
	certPool, errs, nonFatalErrs := utils.NewX509CertPool(config.CertificatesDirectory)

	if len(errs) != 0 {
		return providers, nonFatalErrs, errs
	}

	storageProvider, err := configureStorageProvider(&config.Storage)
	if err != nil {
		return providers, nonFatalErrs, append(errs, err)
	}

	userProvider, err := configureUserProvider(config, certPool)
	if err != nil {
		return providers, nonFatalErrs, append(errs, err)
	}

	notificationProvider, err := configureNotificationProvider(config, certPool)
	if err != nil {
		return providers, nonFatalErrs, append(errs, err)
	}

	clock := utils.RealClock{}
	authorizer := authorization.NewAuthorizer(config.AccessControl)
	sessionProvider := session.NewProvider(config.Session, certPool)
	regulator := regulation.NewRegulator(config.Regulation, storageProvider, clock)
	oidcProvider, err := oidc.NewOpenIDConnectProvider(config.IdentityProviders.OIDC)

	if err != nil {
		return providers, nonFatalErrs, append(errs, fmt.Errorf("Error initializing OpenID Connect Provider: %+v", err))
	}

	return middlewares.Providers{
		Authorizer:           authorizer,
		UserProvider:         userProvider,
		Regulator:            regulator,
		OpenIDConnect:        oidcProvider,
		StorageProvider:      storageProvider,
		NotificationProvider: notificationProvider,
		SessionProvider:      sessionProvider,
	}, nonFatalErrs, errs
}

func configureStorageProvider(config *schema.StorageConfiguration) (provider storage.Provider, err error) {
	switch {
	case config.PostgreSQL != nil:
		provider = storage.NewPostgreSQLProvider(*config.PostgreSQL)
	case config.MySQL != nil:
		provider = storage.NewMySQLProvider(*config.MySQL)
	case config.Local != nil:
		provider = storage.NewSQLiteProvider(config.Local.Path)
	default:
		return provider, errors.New("Unrecognized storage provider")
	}

	return provider, nil
}

func configureUserProvider(config *schema.Configuration, certPool *x509.CertPool) (provider authelia.UserProvider, err error) {
	switch {
	case config.AuthenticationBackend.File != nil:
		provider = authentication.NewFileUserProvider(config.AuthenticationBackend.File)
	case config.AuthenticationBackend.LDAP != nil:
		provider = authentication.NewLDAPUserProvider(*config.AuthenticationBackend.LDAP, certPool)
	case config.AuthenticationBackend.Plugin != nil:
		provider, err = loadUserProviderPlugin(config.AuthenticationBackend.Plugin.Name, config.PluginsDirectory)
		if err != nil {
			return provider, err
		}
	default:
		return provider, errors.New("Unrecognized authentication provider")
	}

	return provider, nil
}

func configureNotificationProvider(config *schema.Configuration, certPool *x509.CertPool) (provider authelia.NotificationProvider, err error) {
	switch {
	case config.Notifier.SMTP != nil:
		provider = notification.NewSMTPNotifier(*config.Notifier.SMTP, certPool)
	case config.Notifier.FileSystem != nil:
		provider = notification.NewFileNotifier(*config.Notifier.FileSystem)
	case config.Notifier.Plugin != nil:
		provider, err = loadNotificationProviderPlugin(config.Notifier.Plugin.Name, config.PluginsDirectory)
		if err != nil {
			return provider, err
		}
	default:
		return provider, errors.New("Unrecognized notifier provider")
	}

	if !config.Notifier.DisableStartupCheck {
		_, err := provider.StartupCheck()
		if err != nil {
			return provider, fmt.Errorf("Error during notifier startup check: %s", err)
		}
	}

	return provider, nil
}
