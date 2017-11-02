Feature: User can access certain subdomains with single factor

  Scenario: User is redirected to service after first factor if allowed
    When I visit "https://login.example.com:8080/?redirect=https%3A%2F%2Fsingle_factor.example.com%3A8080%2Fsecret.html"
    And I login with user "john" and password "password"
    Then I'm redirected to "https://single_factor.example.com:8080/secret.html"

  Scenario: Redirection after first factor fails if single_factor not allowed. It redirects user to first factor.
    When I visit "https://login.example.com:8080/?redirect=https%3A%2F%2Fadmin.example.com%3A8080%2Fsecret.html"
    And I login with user "john" and password "password"
    Then I'm redirected to "https://login.example.com:8080/?redirect=https%3A%2F%2Fadmin.example.com%3A8080%2Fsecret.html"

  Scenario: User can login using basic authentication
    When I request "https://single_factor.example.com:8080/secret.html" with username "john" and password "password" using basic authentication
    Then I receive the secret page