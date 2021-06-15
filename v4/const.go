package authelia

// PluginType is a string type used only in the PluginInformation struct.
type PluginType string

const (
	// UserPlugin is a PluginType for user authentication and details.
	UserPlugin PluginType = "User"

	// NotificationPlugin is a PluginType for notifications.
	NotificationPlugin PluginType = "Notification"
)
