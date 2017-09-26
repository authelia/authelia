Feature: User can access certain subdomains with basic auth

  @need-registered-user-john
  Scenario: User is redirected to service after first factor if allowed
    When I visit "https://auth.test.local:8080/?redirect=https%3A%2F%2Fbasicauth.test.local%3A8080%2Fsecret.html&only_basic_auth=true"
    And I login with user "john" and password "password"
    Then I'm redirected to "https://basicauth.test.local:8080/secret.html"

  @need-registered-user-john
  Scenario: Redirection after first factor fails if basic_auth not allowed. It redirects user to first factor.
    When I visit "https://auth.test.local:8080/?redirect=https%3A%2F%2Fadmin.test.local%3A8080%2Fsecret.html&only_basic_auth=true"
    And I login with user "john" and password "password"
    Then I'm redirected to "https://auth.test.local:8080/?redirect=https%3A%2F%2Fadmin.test.local%3A8080%2Fsecret.html"

  @need-registered-user-john
  Scenario: User is redirected to second factor after first factor
    When I visit "https://auth.test.local:8080/?redirect=https%3A%2F%2Fadmin.test.local%3A8080%2Fsecret.html"
    And I login with user "john" and password "password"
    Then I'm redirected to "https://auth.test.local:8080/secondfactor?redirect=https%3A%2F%2Fadmin.test.local%3A8080%2Fsecret.html"
