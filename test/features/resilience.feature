Feature: Authelia keeps user sessions despite the application restart

  @need-authenticated-user-john
  Scenario: Session is still valid after Authelia restarts
    When the application restarts
    Then I have access to:
      | url                                          |
      | https://admin.test.local:8080/secret.html   |

  @need-registered-user-john
  Scenario: Secrets are stored even when Authelia restarts
    When the application restarts
    And I visit "https://admin.test.local:8080/secret.html" and get redirected "https://auth.test.local:8080/?redirect=https%3A%2F%2Fadmin.test.local%3A8080%2Fsecret.html"
    And I login with user "john" and password "password" 
    And I use "REGISTERED" as TOTP token handle
    And I click on "TOTP"
    Then I'm redirected to "https://admin.test.local:8080/secret.html"