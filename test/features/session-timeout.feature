@needs-inactivity-config
Feature: Session is closed after a certain amount of time

  @need-authenticated-user-john
  Scenario: An authenticated user is disconnected after a certain inactivity period
    Given I have access to "https://public.example.com:8080/secret.html"
    When I sleep for 6 seconds
    And I visit "https://public.example.com:8080/secret.html"
    Then I'm redirected to "https://login.example.com:8080/?rd=https://public.example.com:8080/secret.html"

  @need-authenticated-user-john
  Scenario: An authenticated user is disconnected after session expiration period
    Given I have access to "https://public.example.com:8080/secret.html"
    When I sleep for 4 seconds
    And I visit "https://public.example.com:8080/secret.html"
    And I sleep for 4 seconds
    And I visit "https://public.example.com:8080/secret.html"
    And I sleep for 4 seconds
    And I visit "https://public.example.com:8080/secret.html"
    Then I'm redirected to "https://login.example.com:8080/?rd=https://public.example.com:8080/secret.html"
