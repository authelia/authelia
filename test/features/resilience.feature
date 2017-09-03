Feature: Authelia keeps user sessions despite the application restart

  Scenario: Session is still valid after Authelia restarts
    When I register TOTP and login with user "john" and password "password"
    And the application restarts
    Then I have access to:
      | url                                          |
      | https://secret.test.local:8080/secret.html   |

  Scenario: Secrets are stored even when Authelia restarts
    Given I visit "https://auth.test.local:8080/"
    And I login with user "john" and password "password"
    And I register a TOTP secret called "Sec0"
    When the application restarts
    And I visit "https://secret.test.local:8080/secret.html" and get redirected "https://auth.test.local:8080/"
    And I login with user "john" and password "password" 
    And I use "Sec0" as TOTP token handle
    And I click on "TOTP"
    Then I have access to:
      | url                                          |
      | https://secret.test.local:8080/secret.html   |