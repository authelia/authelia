package suites

import "fmt"

// BaseDomain the base domain
var BaseDomain = "example.com:8080"

// LoginBaseURL the base URL of the login portal
var LoginBaseURL = fmt.Sprintf("https://login.%s/#/", BaseDomain)

// SingleFactorBaseURL the base URL of the singlefactor domain
var SingleFactorBaseURL = fmt.Sprintf("https://singlefactor.%s", BaseDomain)

// AdminBaseURL the base URL of the admin domain
var AdminBaseURL = fmt.Sprintf("https://admin.%s", BaseDomain)

// MailBaseURL the base URL of the mail domain
var MailBaseURL = fmt.Sprintf("https://mail.%s", BaseDomain)

// HomeBaseURL the base URL of the home domain
var HomeBaseURL = fmt.Sprintf("https://home.%s/", BaseDomain)
