Feature: Non authenticated users have no access to certain pages

  Scenario: Anonymous user has no access to protected pages
    Then I get the following status code when requesting:
      | url                                                            | code | method |
      | https://auth.test.local:8080/secondfactor                      | 401  | GET    |
      | https://auth.test.local:8080/secondfactor/u2f/identity/start   | 401  | GET    |
      | https://auth.test.local:8080/secondfactor/u2f/identity/finish  | 401  | GET    |
      | https://auth.test.local:8080/secondfactor/totp/identity/start  | 401  | GET    |
      | https://auth.test.local:8080/secondfactor/totp/identity/finish | 401  | GET    |
      | https://auth.test.local:8080/loggedin                          | 401  | GET    |
      | https://auth.test.local:8080/api/totp                          | 401  | POST   |
      | https://auth.test.local:8080/api/u2f/sign_request              | 401  | GET    |
      | https://auth.test.local:8080/api/u2f/sign                      | 401  | POST   |
      | https://auth.test.local:8080/api/u2f/register_request          | 401  | GET    |
      | https://auth.test.local:8080/api/u2f/register                  | 401  | POST   |


  @needs-single_factor-config
  @need-registered-user-john
  Scenario: User does not have acces to second factor related endpoints when in single factor mode
    Given I post "https://auth.test.local:8080/api/firstfactor" with body:
      | key         | value     |
      | username    | john      |
      | password    | password  |
    Then I get the following status code when requesting:
      | url                                                            | code | method |
      | https://auth.test.local:8080/secondfactor                      | 401  | GET    |
      | https://auth.test.local:8080/secondfactor/u2f/identity/start   | 401  | GET    |
      | https://auth.test.local:8080/secondfactor/u2f/identity/finish  | 401  | GET    |
      | https://auth.test.local:8080/secondfactor/totp/identity/start  | 401  | GET    |
      | https://auth.test.local:8080/secondfactor/totp/identity/finish | 401  | GET    |
      | https://auth.test.local:8080/api/totp                          | 401  | POST   |
      | https://auth.test.local:8080/api/u2f/sign_request              | 401  | GET    |
      | https://auth.test.local:8080/api/u2f/sign                      | 401  | POST   |
      | https://auth.test.local:8080/api/u2f/register_request          | 401  | GET    |
      | https://auth.test.local:8080/api/u2f/register                  | 401  | POST   |