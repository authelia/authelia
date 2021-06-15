package authelia

// PluginType is a string type used only in the PluginInformation struct.
type PluginType string

const (
	// AuthenticationPlugin is a PluginType for authentication.
	AuthenticationPlugin PluginType = "Authentication"

	// NotificationPlugin is a PluginType for notifications.
	NotificationPlugin PluginType = "Notification"
)
