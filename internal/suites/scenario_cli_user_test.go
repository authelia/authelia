package suites

func NewCLIScenario(suiteName string, dockerEnvironment *DockerEnvironment) *CLIAuthScenario {
	return &CLIAuthScenario{
		CommandSuite: &CommandSuite{
			BaseSuite: &BaseSuite{
				Name: suiteName,
			},
			DockerEnvironment: dockerEnvironment,
		},
	}
}

type CLIAuthScenario struct {
	*CommandSuite
}

func (s *CLIAuthScenario) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *CLIAuthScenario) TestUserCRUD() {
	var (
		username      = "test"
		password      = "password"
		cmdCreateUser = []string{"authelia",
			"user",
			"add",
			username,
			`--password "` + password + `"`,
			`--email="` + username + `@example.com"`,
			`--display-name="Test User"`,
			`--group=group1`,
			`--group=group2`,
		}
		cmdListUsers         = []string{"authelia", "user", "list"}
		cmdShowUser          = []string{"authelia", "user", "show", username}
		cmdChangePassword    = []string{"authelia", "user", "password", username, "new_password"}
		cmdDisableUser       = []string{"authelia", "user", "disable", username}
		cmdEnableUser        = []string{"authelia", "user", "enable", username}
		cmdDeleteUser        = []string{"authelia", "user", "del", username}
		cmdChangeEmail       = []string{"authelia", "user", "email", username, username + "@authelia.com"}
		cmdChangeDisplayName = []string{"authelia", "user", "display-name", username, `"Another Name"`}
	)

	defer s.Exec("authelia-backend", cmdDeleteUser) //nolint: errcheck

	testCases := []struct {
		name string
		test func()
	}{
		{
			"ShouldListusers",
			func() {
				output, err := s.Exec("authelia-backend", cmdListUsers)
				s.NoError(err)
				s.Regexp(`.*Username\s+Display Name\s+Email\s+Groups\s+Disabled.*`, output)
				s.GreaterOrEqual(len(output), 382)
			},
		},
		{
			"ShouldCreateUser",
			func() {
				output, err := s.Exec("authelia-backend", cmdCreateUser)

				s.NoError(err)
				s.Contains(output, "user added.")
			},
		},
		{
			"ShouldFailIfUserExists",
			func() {
				output, err := s.Exec("authelia-backend", cmdCreateUser)
				s.Error(err)
				s.Contains(output, "user already exists")
			},
		},
		{
			"ShouldDisplayUserInfo",
			func() {
				output, err := s.Exec("authelia-backend", cmdShowUser)
				s.NoError(err)
				s.Regexp(`.*Display Name:\s+Test User.*`, output)
			},
		},
		{
			"ShouldChangePassword",
			func() {
				output, err := s.Exec("authelia-backend", cmdChangePassword)
				s.NoError(err)
				s.Contains(output, "password changed")
			},
		},
		{
			"ShouldChangeEmail",
			func() {
				output, err := s.Exec("authelia-backend", cmdChangeEmail)
				s.NoError(err)
				s.Contains(output, "email changed")

				output, err = s.Exec("authelia-backend", cmdShowUser)
				s.NoError(err)
				s.Regexp(`.*Email:\s+test@authelia.com.*`, output)
			},
		},
		{
			"ShouldChangeDisplayName",
			func() {
				output, err := s.Exec("authelia-backend", cmdChangeDisplayName)
				s.NoError(err)
				s.Contains(output, "display name changed")

				output, err = s.Exec("authelia-backend", cmdShowUser)
				s.NoError(err)
				s.Regexp(`.*Display Name:\s+Another Name.*`, output)
			},
		},
		{
			"ShouldDisableUser",
			func() {
				output, err := s.Exec("authelia-backend", cmdDisableUser)
				s.NoError(err)
				s.Contains(output, "user disabled.")

				output, err = s.Exec("authelia-backend", cmdShowUser)
				s.NoError(err)
				s.Regexp(`.*Disabled:\s+true.*`, output)
			},
		},
		{
			"ShouldFailChangePasswordIfDisabled",
			func() {
				output, err := s.Exec("authelia-backend", cmdDisableUser)
				s.NoError(err)
				s.Contains(output, "user disabled.")

				output, err = s.Exec("authelia-backend", cmdChangePassword)
				s.Error(err)
				s.Contains(output, "user not found")
			},
		},
		{
			"ShouldEnableUser",
			func() {
				output, err := s.Exec("authelia-backend", cmdEnableUser)
				s.NoError(err)
				s.Contains(output, "user enabled.")

				output, err = s.Exec("authelia-backend", cmdShowUser)
				s.NoError(err)
				s.Regexp(`.*Disabled:\s+false.*`, output)
			},
		},
		{
			"ShouldDeleteDisabledUser",
			func() {
				output, err := s.Exec("authelia-backend", cmdDisableUser)
				s.NoError(err)
				s.Contains(output, "user disabled")

				output, err = s.Exec("authelia-backend", cmdDeleteUser)
				s.NoError(err)
				s.Contains(output, "user deleted")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.test()
		})
	}
}
