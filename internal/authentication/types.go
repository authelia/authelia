package authentication

// UserDetails represent the details retrieved for a given user.
type UserDetails struct {
	Username string
	Emails   []string
	Groups   []string
}
