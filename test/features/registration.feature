Feature: Register secret for second factor

  Scenario: Register a TOTP secret with correct label and issuer
    Given I visit "https://login.example.com:8080/"
    And I login with user "john" and password "password"
    When I register a TOTP secret called "Sec0"
    Then the otpauth url has label "john" and issuer "authelia.com"

  @needs-totp_issuer-config
  Scenario: Register a TOTP secret with correct label and custom issuer
    Given I visit "https://login.example.com:8080/"
    And I login with user "john" and password "password"
    When I register a TOTP secret called "Sec0"
    Then the otpauth url has label "john" and issuer "custom.com"