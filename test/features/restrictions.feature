Feature: Non authenticated users have no access to certain pages

  Scenario: Anonymous user has no access to protected pages
    Then I get the following status code when requesting:
      | url                                                            | code | method |
      | https://login.example.com:8080/secondfactor                      | 401  | GET    |
      | https://login.example.com:8080/secondfactor/u2f/identity/start   | 401  | GET    |
      | https://login.example.com:8080/secondfactor/u2f/identity/finish  | 401  | GET    |
      | https://login.example.com:8080/secondfactor/totp/identity/start  | 401  | GET    |
      | https://login.example.com:8080/secondfactor/totp/identity/finish | 401  | GET    |
      | https://login.example.com:8080/loggedin                          | 401  | GET    |
      | https://login.example.com:8080/api/totp                          | 401  | POST   |
      | https://login.example.com:8080/api/u2f/sign_request              | 401  | GET    |
      | https://login.example.com:8080/api/u2f/sign                      | 401  | POST   |
      | https://login.example.com:8080/api/u2f/register_request          | 401  | GET    |
      | https://login.example.com:8080/api/u2f/register                  | 401  | POST   |
