@needs-single_factor-config
Feature: Server is configured as a single factor only server

  @need-registered-user-john
  Scenario: User is redirected to service after first factor if allowed
    When I visit "https://auth.test.local:8080/?redirect=https%3A%2F%2Fpublic.test.local%3A8080%2Fsecret.html"
    And I login with user "john" and password "password"
    Then I'm redirected to "https://public.test.local:8080/secret.html"

  @need-registered-user-john
  Scenario: User is correctly redirected according to default redirection URL
    When I visit "https://auth.test.local:8080"
    And I login with user "john" and password "password"
    Then I'm redirected to "https://auth.test.local:8080/loggedin"
    And I sleep for 5 seconds
    Then I'm redirected to "https://home.test.local:8080/"
