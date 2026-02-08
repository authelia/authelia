package notification

//go:generate mockgen -package notification -destination smtp_client_mock_test.go -mock_names SMTPClient=MockSMTPClient github.com/authelia/authelia/v4/internal/notification SMTPClient
//go:generate mockgen -package notification -destination smtp_client_factory_mock_test.go -mock_names SMTPClientFactory=MockSMTPClientFactory github.com/authelia/authelia/v4/internal/notification SMTPClientFactory
