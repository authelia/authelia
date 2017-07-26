Feature: Authelia keeps user sessions despite the application restart

  Scenario: Session is still valid after Authelia restarts
    When I register TOTP and login with user "john" and password "password"
    And the application restarts
    Then I have access to:
      | url                                          |
      | https://public.test.local:8080/secret.html   |
      | https://secret.test.local:8080/secret.html   |
      | https://secret1.test.local:8080/secret.html  |
      | https://secret2.test.local:8080/secret.html  |
      | https://mx1.mail.test.local:8080/secret.html |
      | https://mx2.mail.test.local:8080/secret.html |