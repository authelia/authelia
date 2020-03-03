package authentication

// UserDetails represent the details retrieved for a given user.
type UserDetails struct {
	Emails []string
	Groups []string
}
