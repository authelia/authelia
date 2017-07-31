Feature: Non authenticated users have no access to certain pages

  Scenario Outline: User has no access to protected pages
    When I visit "<url>"
    Then I get an error <error code>

    Examples:
      | url                                                            | error code |
      | https://auth.test.local:8080/secondfactor                      | 401        |
      | https://auth.test.local:8080/verify                            | 401        |
      | https://auth.test.local:8080/secondfactor/u2f/identity/start   | 401        |
      | https://auth.test.local:8080/secondfactor/u2f/identity/finish  | 403        |
      | https://auth.test.local:8080/secondfactor/totp/identity/start  | 401        |
      | https://auth.test.local:8080/secondfactor/totp/identity/finish | 403        |
      | https://auth.test.local:8080/password-reset/identity/start     | 403        |
      | https://auth.test.local:8080/password-reset/identity/finish    | 403        |
  