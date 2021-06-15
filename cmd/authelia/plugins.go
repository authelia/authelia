package main

import (
	"errors"
	"fmt"
	"plugin"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4"
)

func loadNotificationProviderPlugin(name, directory string) (provider authelia.NotificationProvider, err error) {
	pluginPath := fmt.Sprintf("%s/%s.notifier.authelia.com.so", directory, name)

	notifierPlugin, err := plugin.Open(pluginPath)
	if err != nil {
		return provider, fmt.Errorf("Error opening notification provider plugin %s: %+v", pluginPath, err)
	}

	uinfo, err := notifierPlugin.Lookup("AutheliaPluginInformation")
	if err != nil {
		return provider, fmt.Errorf("Error during notification plugin lookup: could not lookup plugin information (maybe it is out of date): %+v", err)
	}

	info, ok := uinfo.(authelia.PluginInformation)
	if !ok {
		return provider, errors.New("Error during notification plugin discovery: plugin information is malformed or missing (maybe it is out of date)")
	} else {
		if info.Type != authelia.NotificationPlugin {
			return provider, fmt.Errorf("Error during notification plugin check: plugin should be type 'Notification' but it's type %s", info.Type)
		}
	}

	up, err := notifierPlugin.Lookup("NotificationProvider")
	if err != nil {
		return provider, fmt.Errorf("Error during notification plugin lookup: could not lookup notification provider (maybe it is out of date): %+v", err)
	}

	if p, ok := up.(authelia.NotificationProvider); !ok {
		return provider, errors.New("Error during notification plugin discovery: the plugin doesn't implement the interface (maybe it is out of date)")
	} else {
		provider = p

		logrus.Infof("Notifier provider plugin loaded: %s v%s by %s", info.Name, info.Version, info.Author)
	}

	return provider, nil
}

func loadUserProviderPlugin(name, directory string) (provider authelia.UserProvider, err error) {
	pluginPath := fmt.Sprintf("%s/%s.user.authelia.com.so", directory, name)

	notifierPlugin, err := plugin.Open(pluginPath)
	if err != nil {
		return provider, fmt.Errorf("Error opening user provider plugin %s: %+v", pluginPath, err)
	}

	uinfo, err := notifierPlugin.Lookup("AutheliaPluginInformation")
	if err != nil {
		return provider, fmt.Errorf("Error during user plugin lookup: could not lookup plugin information (maybe it is out of date): %+v", err)
	}

	info, ok := uinfo.(authelia.PluginInformation)
	if !ok {
		return provider, errors.New("Error during user plugin discovery: plugin information is malformed or missing (maybe it is out of date)")
	} else {
		if info.Type != authelia.AuthenticationPlugin {
			return provider, fmt.Errorf("Error during user plugin check: plugin should be type 'Authentication' but it's type %s", info.Type)
		}
	}

	up, err := notifierPlugin.Lookup("UserProvider")
	if err != nil {
		return provider, fmt.Errorf("Error during user plugin lookup: could not lookup user provider (maybe it is out of date): %+v", err)
	}

	if p, ok := up.(authelia.UserProvider); !ok {
		return provider, errors.New("Error during user plugin discovery: the plugin doesn't implement the interface (maybe it is out of date)")
	} else {
		provider = p

		logrus.Infof("Authentication provider plugin loaded: %s v%s by %s", info.Name, info.Version, info.Author)
	}

	return provider, nil
}
