Feature: Authelia keeps user sessions despite the application restart

  @need-authenticated-user-john
  Scenario: Session is still valid after Authelia restarts
    When the application restarts
    Then I have access to "https://admin.example.com:8080/secret.html"

  @need-registered-user-john
  Scenario: Secrets are stored even when Authelia restarts
    When the application restarts
    And I visit "https://admin.example.com:8080/secret.html" and get redirected "https://login.example.com:8080/?rd=https%3A%2F%2Fadmin.example.com%3A8080%2Fsecret.html"
    And I login with user "john" and password "password" 
    And I use "REGISTERED" as TOTP token handle
    And I click on "Sign in"
    Then I'm redirected to "https://admin.example.com:8080/secret.html"